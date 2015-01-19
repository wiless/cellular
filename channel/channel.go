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
)

func init() {
	log.Println("Initiated cellular.channel")
}

type TransmitterBuffer struct {
	sync.Mutex
	data vlib.VectorC
}

func (t *TransmitterBuffer) Write(v vlib.VectorC) {
	t.Lock()
	t.data = v
	t.Unlock()
}
func (t *TransmitterBuffer) Read() vlib.VectorC {

}

type SFN struct {
	links     []cell.LinkMetric
	chparams  [][]core.ChannelParam
	freqGHz   float64
	txIDs     vlib.VectorI
	rxIDs     vlib.VectorI
	rxSamples map[int]vlib.VectorC
}

func (s *SFN) createDefaultPDP() {
	s.chparams = make([][]core.ChannelParam, len(s.links))
	tmprx := make(map[int]bool)
	for i := 0; i < len(s.chparams); i++ {
		s.chparams[i] = make([]core.ChannelParam, len(s.links[i].TxNodeIDs))

		if val, ok := tmprx[s.links[i].RxNodeID]; ok {
			log.Println("Duplicate Link found for %d !! ", val)
		} else {
			tmprx[s.links[i].RxNodeID] = true
		}

		tmptx := make(map[int]bool)
		for j, tid := range s.links[i].TxNodeIDs {
			s.chparams[i][j] = core.DefaultChannel()
			s.chparams[i][j].PowerInDBm = s.links[i].TxNodesRSRP[j]
			tmptx[tid] = true
		}
		for key, _ := range tmptx {
			s.txIDs.AppendAtEnd(key)
		}
		log.Printf("\n%d @ %f :  %#v", s.links[i].RxNodeID, s.links[i].FreqInGHz, s.chparams[i])
	}

	for key, _ := range tmprx {
		s.rxIDs.AppendAtEnd(key)
	}

	log.Println("Default PDP created for : ", len(s.chparams))

}

func (s *SFN) GetTxNodeIDs() vlib.VectorI {
	return s.txIDs
}

func (s *SFN) GetRxNodeIDs() vlib.VectorI {
	return s.rxIDs
}

type Channel struct {
	sflinks     []SFN
	freqs       vlib.VectorF
	txnodes     map[int]cell.Transmitter
	rxnodes     map[int]cell.Receiver
	freqindxMap map[float64]vlib.VectorI
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

	/// Check if all transmitters are set for each nodes
	if !c.CheckTransmitters() {
		log.Panicln("Some transmitters not associated !!")
	}
	if !c.CheckReceivers() {
		log.Panicln("Some receivers not associated !!")
	}

	var wg sync.WaitGroup
	for _, sfid := range sfids {

		go func() {
			/// Should start all for all the SFN
			log.Println("TxNodes  : ", c.sflinks[sfid].GetTxNodeIDs())
			log.Println("RxNodes  : ", c.sflinks[sfid].GetRxNodeIDs())

			var rxch gocomm.Complex128AChannel

			txnodeIDs := c.sflinks[sfid].GetTxNodeIDs()
			rxnodeIDs := c.sflinks[sfid].GetRxNodeIDs()
			for indx, tid := range txnodeIDs {

				tx, ok := c.txnodes[tid]
				if !ok {
					log.Panicln("Surprising !! No Transmitter attached for ", tid)
				}

				tx.SetWaitGroup(&wg)
				rxch = tx.GetChannel()

				wg.Add(1)
				log.Printf("%d Tx Started... %#v", indx, tx.GetID())
				go tx.StartTransmit()

			}

			for indx, rid := range rxnodeIDs {
				rx, ok := c.rxnodes[rid]
				if !ok {
					log.Panicln("Surprising !! No Receiver attached for ", rid)
				}
				// for indx, rx := range c.rxnodes {
				rx.SetWaitGroup(&wg)
				wg.Add(1)
				log.Printf("%d Rx Started... %#v", indx, rx.GetID())
				go rx.StartReceive(rxch)
			}

		}()
		wg.Wait()

	}

	log.Println("Done")
}

// AddTransmitter adds the transmitter tx and assoicates with the txnodeid from tx.GetID()
func (c *Channel) AddTransmiter(tx cell.Transmitter) {
	if val, ok := c.txnodes[tx.GetID()]; ok {
		log.Println("Overwriting Node ", tx.GetID(), val)
	} else {
		c.txnodes[tx.GetID()] = tx
		log.Println("Added Node ", tx.GetID())
	}

}

// AddReceiver adds the receiver rx and assoicates with the rxnodeid from rx.GetID()
func (c *Channel) AddReceiver(rx cell.Receiver) {
	if val, ok := c.rxnodes[rx.GetID()]; ok {
		log.Println("Overwriting Node ", rx.GetID(), val)
	} else {
		c.rxnodes[rx.GetID()] = rx
		log.Println("Added Node ", rx.GetID())
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

}

func (c *Channel) SFN() int {
	return c.freqs.Size()
}

/// After loading all links this must be last func to be called before running the channel
func (c *Channel) Init() {
	c.txnodes = make(map[int]cell.Transmitter)
	c.rxnodes = make(map[int]cell.Receiver)
	for i := 0; i < len(c.sflinks); i++ {
		// c.sflinks[i].links[]
	}
}
