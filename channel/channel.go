// Simple SISO Channel interface that creates links and emulates multipath channel between transmitters and receivers
// Will soon be moved to github.com/wiless/gocomm package
package channel

import (
	cell "github.com/wiless/cellular"
	"github.com/wiless/gocomm"
	"github.com/wiless/gocomm/core"
	"github.com/wiless/vlib"
	"log"
	"sync"
	"time"
)

func init() {
	log.Println("Initiated cellular.channel")
}

type BufferState int

var BufferStates = [...]string{
	"DataReady",
	"DataSent",
	"DataACKed",
}

func (b BufferState) String() string {
	if int(b) >= len(BufferStates) {
		log.Printf("Found this %d Bufferstate ", int(b))
		return "Unknown!!"
	}
	return BufferStates[b]
}

const (
	DataReady BufferState = iota
	DataSent
	DataACKed
)

type Channel struct {
	sflinks     []SFN
	freqs       vlib.VectorF
	txnodes     map[int]cell.Transmitter
	rxnodes     map[int]cell.Receiver
	freqindxMap map[float64]vlib.VectorI
	initialized bool
}

type SFN struct {
	links          []cell.LinkMetric
	chparams       [][]core.ChannelParam
	freqGHz        float64
	txPortIDs      vlib.VectorI
	rxPortIDs      vlib.VectorI
	txMgr          TransmitterBufferManager
	rxMgr          ReceiverBufferManager
	associatedRx   map[int]vlib.VectorI // Lookup of all rxids affected by a transmission of a transmitter
	associatedTx   map[int]vlib.VectorI // Lookup of all txids affecting a receiver
	Port2LinkIndex map[int]int          // Lookup of the row index for the given receiver-id
	wg             sync.WaitGroup
}

type TransmitterBuffer struct {
	id int

	source gocomm.Complex128AChannel
	sync.Mutex
	counter int
	rawdata vlib.VectorC
	data    gocomm.SComplex128AObj
	state   BufferState
}

type RecieverBuffer struct {
	source gocomm.Complex128AChannel

	sync.Mutex
	counter int
	data    vlib.VectorC
	TotalTx int
}

type TransmitterBufferManager struct {
	txInputBuffer map[int]*TransmitterBuffer
	feedbackTx2Rx chan int
	feedbackRx2Tx chan int
}

type Status struct {
	Total   int
	counter int
}

type ReceiverBufferManager struct {
	rxOutputBuffer map[int]*RecieverBuffer
	feedbackTx2Rx  chan int
	feedbackRx2Tx  chan int
	sync.Mutex
	TxReadyStatus map[int]*Status
}

func (r *ReceiverBufferManager) ShouldACK(txid int) bool {
	result := r.TxReadyStatus[txid].counter == r.TxReadyStatus[txid].Total
	if result {
		r.Lock()
		r.TxReadyStatus[txid].counter = 0
		r.Unlock()
	}
	return result
}

func (r *ReceiverBufferManager) UpdateCounter(txid int) {
	r.Lock()
	r.TxReadyStatus[txid].counter++
	//log.Printf("RxMgr : %d of %d Receivers of Tx-%d have Processed Data", r.TxReadyStatus[txid].counter, r.TxReadyStatus[txid].Total, txid)
	r.Unlock()
}

func (r *ReceiverBufferManager) Create(rxid int, totalTxIDs int) {
	newbuf := new(RecieverBuffer)
	newbuf.source = gocomm.NewComplex128AChannel()
	newbuf.TotalTx = totalTxIDs
	newbuf.counter = totalTxIDs
	r.rxOutputBuffer[rxid] = newbuf

}

func (r *RecieverBuffer) Write(obj gocomm.SComplex128AObj) {
	r.counter = r.TotalTx
	r.source <- obj

}
func (r *RecieverBuffer) Accumulate(samples vlib.VectorC) {
	r.Lock()

	if r.data.Size() > samples.Size() {
		samples.Resize(r.data.Size())
	}
	if r.data.Size() < samples.Size() {
		r.data.Resize(samples.Size())
	}

	r.data.PlusEqual(samples)
	r.counter--
	r.Unlock()
}

func (r *ReceiverBufferManager) GetCh(rid int) gocomm.Complex128AChannel {
	obj, ok := r.rxOutputBuffer[rid]
	if !ok {
		log.Panicln("ReceiverBufferManager::Get() - No such Rx Buffer for rxid=", rid)
	}

	return obj.source
}

func (t *TransmitterBufferManager) Set(tid int, ch gocomm.Complex128AChannel) {

	tb, ok := t.txInputBuffer[tid]
	if !ok {
		log.Panicln("TxBufMgr.Set(): Unknown/not-created Txid ", tid)
	}

	tb.source = ch

}
func (t *TransmitterBufferManager) Get(tid int) gocomm.Complex128AChannel {
	return t.txInputBuffer[tid].source
}

func (t *TransmitterBuffer) SetState(bf BufferState) {
	t.Lock()
	t.state = bf
	t.Unlock()
}

func (t *TransmitterBuffer) Update() {
	tmp := <-t.source
	log.Println("Txbuffer Update() : ", tmp.Message, " time = ", tmp.TimeStamp)
	t.WriteObj(tmp)
}

func (t *TransmitterBuffer) WriteObj(obj gocomm.SComplex128AObj) {
	t.Lock()
	t.data = obj
	t.rawdata = obj.Ch
	t.state = DataReady
	t.counter++
	t.Unlock()
}

func (t *TransmitterBuffer) WriteSamples(v vlib.VectorC) {
	t.Lock()
	t.rawdata = v
	t.data.Ch = v
	t.Unlock()
}

func (t *TransmitterBuffer) ReadObj() gocomm.SComplex128AObj {
	return t.data
}

func (t *TransmitterBuffer) ReadSamples() vlib.VectorC {
	return t.rawdata
}

func (s *SFN) startRxManager() {

	cnt := 0
	var mgrwg sync.WaitGroup
	for {
		// log.Println("RxMgr: Listening .. ")
		// for i := 0; i < 5; i++ {
		// 	time.Sleep(1 * time.Millisecond)
		// 	fmt.Printf("*.*.")
		// }
		// fmt.Printf("\n")
		txid, ok := <-s.rxMgr.feedbackTx2Rx
		log.Printf("RxMgr: Counter %d : Found Broadcast of Transmitter %d", cnt, txid)

		affectedRxids, ok := s.associatedRx[txid]
		if !ok {
			log.Println("RxMgr : Unknown **** TxID sent  by Txmanager *****")
			return
		}

		txobj := s.txMgr.txInputBuffer[txid].ReadObj()
		log.Printf("RxMgr: Packets from Tx : %d = Timestamp : %f ", txid, txobj.TimeStamp)

		for _, rxid := range affectedRxids {
			// log.Printf("RxMgr : Rx-%d Processing Packet %d  from %d", rxid, s.txMgr.txInputBuffer[txid].counter, txid)
			mgrwg.Add(1)
			go func(rid int) {
				rxbufr := s.rxMgr.rxOutputBuffer[rid]
				chindx, ok := s.Port2LinkIndex[rxid]
				if !ok {
					log.Panicln("RxMgr: No link information for rxid ", chindx)
				}

				txsamples := txobj.Ch
				/// Ideally Do convolution
				// conv and ...
				// for j, t := range s.links[chindx].TxNodeIDs {
				// 	if t == txid {
				// 		s.chparams[chindx][j].Coeff
				// 	}

				// }

				//
				///
				rxsamples := txsamples

				/// Accumulate into rxbuffer

				rxbufr.Accumulate(rxsamples)
				// log.Printf("RxMgr Rx-%d is Accumulating samples from %d ", rid, txid)

				if rxbufr.counter == 0 {
					/// Data ready to be processed by the receiver ,
					// log.Printf("RxMgr Rx-%d SENT ", rid)
					var rxobj gocomm.SComplex128AObj
					/// Change extra params if needed
					rxobj = txobj

					/// Add Noise as given in link
					NoisePower := vlib.InvDb(s.links[chindx].N0)
					noise := vlib.RandNCVec(rxsamples.Size(), NoisePower)
					rxsamples = rxsamples.Scale(2)
					if rxid == 200 {
						scale := vlib.GetEJtheta(22)
						rxsamples = rxsamples.ScaleC(scale)
						// re := vlib.NewOnesF(len(rxsamples))
						// im := vlib.NewVectorF(len(rxsamples)).Add(rxobj.TimeStamp / 200.0)
						// faded = vlib.ToVectorC2(re, im)
						//
						// log.Println("**************************************  did rotations with ", rxsamples[0:10])
						// log.Println("**************************************  did rotations with ", faded[0:10])
						rxobj.Ch = rxsamples // .AddVector(noise)
					}
					rxobj.Ch = rxsamples.AddVector(noise)

					cnt++

					/// Write to Reciever
					rxbufr.Write(rxobj)
				} else {
					// log.Printf("RxMgr Rx-%d Cant transmit data YET for Tx related to %d", rid, txid)

				}
				s.rxMgr.UpdateCounter(txid) /// counter++

				mgrwg.Done()
			}(rxid)

		}
		mgrwg.Wait()

		if s.rxMgr.ShouldACK(txid) {
			log.Printf("RxMgr Sending ACK for TxID %d (Total Transmissions %d)", txid, cnt)
			s.rxMgr.feedbackRx2Tx <- txid
			time.Sleep(250 * time.Millisecond)
		}

	}

	s.wg.Done()
}

func (t *TransmitterBufferManager) ReadyForNextSlot() bool {
	result := true
	for _, tbr := range t.txInputBuffer {
		// log.Printf("Tx : %d : %s ", txid, tbr.state)
		result = result && (tbr.state == DataACKed || tbr.state == DataReady)

	}
	return result

}

func (t *TransmitterBufferManager) Start() {
	var trigTxID int
	var ok bool
	/// send once in sequential and then keep waiting..
	var wg sync.WaitGroup

	for tid, tx := range t.txInputBuffer {
		wg.Add(1)

		go func(tidx int, txb *TransmitterBuffer) {
			txb.Update()
			txb.SetState(DataSent)
			log.Printf("TxMgr - ONETIME : Fetching txid %d (%d), State:%v , Packet : %f  Message=%s ", tidx, txb.id, txb.state, txb.data.TimeStamp, txb.data.Message)
			t.feedbackTx2Rx <- tidx
			wg.Done()
		}(tid, tx)
	}
	wg.Wait()
	// log.Println("Atleast one packet has been sent")
	go func() {
		for {
			if t.ReadyForNextSlot() {
				for txid, txbfr := range t.txInputBuffer {
					wg.Add(1)
					go func(tid int, bfr *TransmitterBuffer) {
						bfr.SetState(DataSent)
						t.feedbackTx2Rx <- tid
						wg.Done()
					}(txid, txbfr)

				}
				wg.Wait()

			}

		}
	}()

	for {
		// log.Println("TxBuf Mgr: Waiting for some feedback !!")

		trigTxID, ok = <-t.feedbackRx2Tx
		t.txInputBuffer[trigTxID].SetState(DataACKed)
		log.Printf("TxMgr: Received ACK from %d Packet ID=%d, NewState:%v", trigTxID, t.txInputBuffer[trigTxID].counter-1, t.txInputBuffer[trigTxID].state)
		// Fetch New Data
		if !ok {
			log.Println("Unknown TxID requested by Rxmanager !!")
			return
		}
		wg.Add(1)
		go func(tid int) {
			// log.Printf("TxMgr: Refilling Data for Txid %d, Status %v,Packet ID %d", tid, t.txInputBuffer[tid].state, t.txInputBuffer[tid].counter-1)
			t.txInputBuffer[tid].Update()
			// log.Printf("TxMgr: After Refilling Data for Txid %d, Status %v,Packet ID %d", tid, t.txInputBuffer[tid].state, t.txInputBuffer[tid].counter-1)
			wg.Done()
		}(trigTxID)
		wg.Wait()
	}
	log.Println("All job done at TxMgr")
}

func (s *SFN) StartBufferManager() {
	s.wg.Add(1)
	go s.startRxManager()

	s.txMgr.Start()
	s.wg.Wait()
}

func (s *SFN) createDefaultPDP() {
	s.chparams = make([][]core.ChannelParam, len(s.links))
	s.Port2LinkIndex = make(map[int]int)
	s.associatedRx = make(map[int]vlib.VectorI)
	s.associatedTx = make(map[int]vlib.VectorI)
	s.txPortIDs.Resize(0)

	tmptx := make(map[int]bool)
	for i := 0; i < len(s.chparams); i++ {
		rxid := s.links[i].RxNodeID

		// For every link, concurent RX and connect
		NtxNodes := len(s.links[i].TxNodeIDs)
		s.chparams[i] = make([]core.ChannelParam, NtxNodes)
		s.Port2LinkIndex[rxid] = i
		if _, ok := s.associatedTx[rxid]; ok {
			log.Println("Duplicate Link found for %d !! ", rxid)
		} else {
			s.associatedTx[rxid] = s.links[i].TxNodeIDs
			s.rxPortIDs.AppendAtEnd(rxid)
		}

		for j, tid := range s.links[i].TxNodeIDs {
			s.chparams[i][j] = core.DefaultChannel()
			s.chparams[i][j].PowerInDBm = s.links[i].TxNodesRSRP[j]
			rvec := s.associatedRx[tid]
			rvec.AppendAtEnd(rxid)
			s.associatedRx[tid] = rvec
			tmptx[tid] = true
		}

		//log.Printf("\n%d @ %f :  %#v", s.links[i].RxNodeID, s.links[i].FreqInGHz, s.chparams[i])
	}
	/// ensure only once the TxIds are entered
	for key, _ := range tmptx {
		s.txPortIDs.AppendAtEnd(key)
	}
	log.Println("Default PDP created for : ", len(s.chparams))
}

func (s *SFN) GetTxNodeIDs() vlib.VectorI {
	return s.txPortIDs
}

func (s *SFN) GetRxNodeIDs() vlib.VectorI {
	return s.rxPortIDs
}

func NewWirelessChannelFromFile(file string) *Channel {
	result := new(Channel)
	result.CreateFromFile(file)
	return result
}

func NewWirelessChannel(links []cell.LinkMetric) *Channel {
	result := new(Channel)
	result.classifySFN(links)
	return result
}

func (c *Channel) GetTxNodeIDs(sfid int) vlib.VectorI {
	return c.sflinks[sfid].GetTxNodeIDs()
}

func (c *Channel) GetRxNodeIDs(sfid int) vlib.VectorI {
	return c.sflinks[sfid].GetRxNodeIDs()

}

// CheckTransmitters checks if a Transmitter is set for all the txnodeids set through linkmetrics
func (c *Channel) CheckTransmitters() bool {

	for i := 0; i < len(c.sflinks); i++ {
		vec := c.sflinks[i].GetTxNodeIDs()
		for _, val := range vec {
			_, ok := c.txnodes[val]
			if !ok {
				log.Println("No Transmitter set for id ", val)

				return false
			}
		}
	}
	return true
}

// CheckTransmitters checks if a Transmitter is set for all the txnodeids set through linkmetrics
func (c *Channel) CheckReceivers() bool {

	for i := 0; i < len(c.sflinks); i++ {
		vec := c.sflinks[i].GetRxNodeIDs()
		for _, val := range vec {
			_, ok := c.rxnodes[val]
			if !ok {
				log.Println("No Receiver set for id ", val)
				return false
			}
		}
	}
	return true
}

// Start triggers all the transmitters and receivers in all the SFN to transmit and receive data
func (c *Channel) Start(sfids ...int) {

	if len(sfids) == 0 {
		sfids = vlib.NewSegmentI(0, len(c.sflinks))
		log.Println("Start all the SFN in the system : ", sfids)
	}

	if !c.initialized {
		log.Panicln("Channel Object Not initialized..Forgot to call .Init() ?? ")
	}

	var wg sync.WaitGroup
	for _, sfid := range sfids {
		// c.sflinks[sfid].wg = wg
		// log.Println("Start for SFID = ", sfid)

		{
			txnodeIDs := c.sflinks[sfid].GetTxNodeIDs()
			rxnodeIDs := c.sflinks[sfid].GetRxNodeIDs()

			// /// Should start all for all the SFN
			log.Println("TxNodes  : ", txnodeIDs)
			log.Println("RxNodes  : ", rxnodeIDs)

			for _, tid := range txnodeIDs {
				readCH := c.txnodes[tid].GetChannel()
				c.sflinks[sfid].txMgr.Set(tid, readCH)

			}

			/// Start Concurrent Transmission process
			for indx, tx := range c.txnodes {
				tx.SetWaitGroup(&wg)
				wg.Add(1)
				// log.Println("Setting WG = ", &wg)
				log.Printf("%d Tx Started... %#v", indx, tx.GetID())
				go tx.StartTransmit()
				// log.Printf("%d Tx Started...done.. %#v", indx, tx.GetID())
			}

			/// Start Concurrent Receiver process
			for _, rid := range rxnodeIDs {
				rx, ok := c.rxnodes[rid]
				if !ok {
					log.Panicln("Surprising !! No Receiver attached for ", rid)
				}

				rx.SetWaitGroup(&wg)
				wg.Add(1)
				log.Printf("Rx Started... %d", rid)
				writeCH := c.sflinks[sfid].rxMgr.GetCh(rid)
				go rx.StartReceive(writeCH)
			}

			go c.sflinks[sfid].StartBufferManager()

		}
		wg.Wait()

	}
	// time.Sleep(2 * time.Second)
	log.Println("Finished channel emulation")
}

// AddTransmitter adds the transmitter tx and assoicates with the txnodeid from tx.GetID()
func (c *Channel) AddTransmiter(tx cell.Transmitter) {
	if val, ok := c.txnodes[tx.GetID()]; ok {
		log.Println("Tx Overwriting Node ", tx.GetID(), val)
	} else {
		c.txnodes[tx.GetID()] = tx

	}

}

// AddReceiver adds the receiver rx and assoicates with the rxnodeid from rx.GetID()
func (c *Channel) AddReceiver(rx cell.Receiver) {
	if val, ok := c.rxnodes[rx.GetID()]; ok {
		log.Println("Rx Overwriting Node ", rx.GetID(), val)
	} else {
		c.rxnodes[rx.GetID()] = rx
		// log.Println("Receiver Added Node ", rx.GetID())
	}
}

func (c *Channel) CreateFromFile(file string) {
	var tmplinks []cell.LinkMetric

	vlib.LoadStructure(file, &tmplinks)
	c.classifySFN(tmplinks)

}

/// to be called only when freqindxMap is created
func (c *Channel) classifySFN(links []cell.LinkMetric) {
	c.freqindxMap = make(map[float64]vlib.VectorI)

	for i, v := range links {
		index := c.freqindxMap[v.FreqInGHz]
		index.AppendAtEnd(i)
		c.freqindxMap[v.FreqInGHz] = index
	}

	c.sflinks = make([]SFN, len(c.freqindxMap))
	c.freqs = vlib.NewVectorF(len(c.freqindxMap))
	var i int = 0
	for f, ivec := range c.freqindxMap {
		c.sflinks[i].links = make([]cell.LinkMetric, len(ivec))
		c.sflinks[i].freqGHz = f
		c.freqs[i] = f
		for j, v := range ivec {
			c.sflinks[i].links[j] = links[v]
		}
		c.sflinks[i].createDefaultPDP()
		// log.Println("=================== ", f)
		// log.Println(c.sflinks[i])
		i++
	}
	c.txnodes = make(map[int]cell.Transmitter)
	c.rxnodes = make(map[int]cell.Receiver)
}

func (c *Channel) SFNids() vlib.VectorI {
	result := vlib.NewVectorI(len(c.sflinks))
	for key, _ := range c.sflinks {
		result.AppendAtEnd(key)
	}
	return result
}

func (c *Channel) SFN() int {
	return c.freqs.Size()
}

/// After loading all links this must be last func to be called before running the channel
func (c *Channel) Init() {

	for i := 0; i < len(c.sflinks); i++ {
		if !c.CheckTransmitters() {
			log.Panicln("All txports not associcated with Transmitters")
		}
		if !c.CheckReceivers() {
			log.Panicln("All rxports not associcated with Receivers")
		}

		downlinkfb := make(chan int, 100)
		uplinkfb := make(chan int, 100)

		c.sflinks[i].txMgr.feedbackTx2Rx = downlinkfb
		c.sflinks[i].txMgr.feedbackRx2Tx = uplinkfb

		c.sflinks[i].rxMgr.feedbackTx2Rx = downlinkfb
		c.sflinks[i].rxMgr.feedbackRx2Tx = uplinkfb

		c.sflinks[i].rxMgr.rxOutputBuffer = make(map[int]*RecieverBuffer)
		c.sflinks[i].txMgr.txInputBuffer = make(map[int]*TransmitterBuffer)

		/// Set Status monitor at RxBufMgr
		c.sflinks[i].rxMgr.TxReadyStatus = make(map[int]*Status)

		/// Create  RxOutputBuffer
		for _, val := range c.sflinks[i].rxPortIDs {
			totalTxIDs := c.sflinks[i].associatedTx[val].Size()
			c.sflinks[i].rxMgr.Create(val, totalTxIDs)
		}

		/// Create TxInputBuffer and
		for _, val := range c.sflinks[i].txPortIDs {
			c.sflinks[i].txMgr.txInputBuffer[val] = new(TransmitterBuffer)
			c.sflinks[i].txMgr.txInputBuffer[val].id = val
			txstat := new(Status)

			txstat.Total = c.sflinks[i].associatedRx[val].Size()
			txstat.counter = 0
			c.sflinks[i].rxMgr.TxReadyStatus[val] = txstat

		}
		// log.Printf("Buffer Info %#v  %#v", c.sflinks[i].txMgr, c.sflinks[i].rxMgr)

	}

	c.initialized = true
}
