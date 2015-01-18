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

type SFN struct {
	links    []cell.LinkMetric
	chparams [][]core.ChannelParam
	freqGHz  float64
}

func (s *SFN) CreateDefaultPDP() {
	s.chparams = make([][]core.ChannelParam, len(s.links))
	for i := 0; i < len(s.chparams); i++ {
		s.chparams[i] = make([]core.ChannelParam, len(s.links[i].TxNodeIDs))
		for j, _ := range s.links[i].TxNodeIDs {
			s.chparams[i][j] = core.DefaultChannel()
			s.chparams[i][j].PowerInDBm = s.links[i].TxNodesRSRP[j]
		}
		log.Printf("\n%d @ %f :  %#v", s.links[i].RxNodeID, s.links[i].FreqInGHz, s.chparams[i])
	}
	log.Println("Default PDP created for : ", len(s.chparams))
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

func (c *Channel) Start() {

	var wg sync.WaitGroup
	var rxch gocomm.Complex128AChannel

	for indx, tx := range c.txnodes {
		tx.SetWaitGroup(&wg)
		rxch = tx.GetChannel()
		wg.Add(1)
		log.Printf("%d Tx Started... %#v", indx, tx.GetID())
		go tx.StartTransmit()
	}

	/// Actually DATA to be weighted and combined here before writing into rx channel
	for indx, rx := range c.rxnodes {
		rx.SetWaitGroup(&wg)
		wg.Add(1)
		log.Printf("%d Rx Started... %#v", indx, rx.GetID())
		go rx.StartReceive(rxch)
	}
	wg.Wait()
	log.Println("Done")
}
func (c *Channel) SetTransmiter(tx cell.Transmitter) {
	if val, ok := c.txnodes[tx.GetID()]; ok {
		log.Println("Overwriting Node ", tx.GetID(), val)
	} else {
		c.txnodes[tx.GetID()] = tx
		log.Println("Added Node ", tx.GetID())
	}

}
func (c *Channel) SetReceiver(rx cell.Receiver) {
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
		c.sflinks[i].CreateDefaultPDP()
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
