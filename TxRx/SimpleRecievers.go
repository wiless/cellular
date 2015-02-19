package TxRx

import (
	"fmt"
	"github.com/wiless/vlib"
	"os"
	"strconv"

	// "github.com/wiless/cellular"
	// "github.com/wiless/cellular/channel"
	"github.com/wiless/gocomm"
	// "github.com/wiless/gocomm/chipset"

	// "github.com/wiless/gocomm/modem"
	// "github.com/wiless/gocomm/sink"

	"log"
	"sync"

	// "os"
)

type SimpleReceiver struct {
	nid      int
	key      string
	wg       *sync.WaitGroup
	filename string
	probes   []gocomm.Complex128AChannel
}

func (s *SimpleReceiver) NProbes() int {
	// Dummy init function
	return len(s.probes)

}

func (s *SimpleReceiver) Init() {
	// Dummy init function
	s.key = string(vlib.RandString(8))
	s.probes = make([]gocomm.Complex128AChannel, 1)
	s.probes[0] = gocomm.NewComplex128AChannel()
}

func (s *SimpleReceiver) GetProbe(prbId int) gocomm.Complex128AChannel {
	if prbId >= s.NProbes() {
		log.Panicln("Rx:GetProbe Index out of bound")
	}
	return s.probes[prbId]
}

// Simple rx node
func (c *SimpleReceiver) StartReceive(rxch gocomm.Complex128AChannel) {
	c.filename = strconv.Itoa(c.GetID()) + "_" + c.key + ".dat"
	w, _ := os.Create(c.filename)

	for i := 0; ; i++ {
		// log.Printf("CSFReceiver: Rx-%d Waiting to read data at Input ", c.GetID())
		rdata := <-rxch
		// c.probes[0] <- rdata
		select {
		case c.probes[0] <- rdata:
			log.Println("==========Message sent to probe==== R R RR ")
		default:
			log.Println("no message sent")
		}

		fmt.Fprintf(w, "%d : %v\n", i, rdata)
		log.Printf("SimpleReceiver (%d): Received Packet ID  %f ", c.GetID(), rdata.TimeStamp)
		if i == rdata.GetMaxExpected()-1 {
			break
		}

	}
	if c.wg != nil {
		log.Println("Done receiever job of ", c.GetID())
		c.wg.Done()
	}

}

func (c SimpleReceiver) GetID() int {
	return c.nid
}
func (c *SimpleReceiver) SetID(id int) {
	c.nid = id
}
func (c SimpleReceiver) IsActive() bool {
	return true
}

func (c *SimpleReceiver) SetWaitGroup(wg *sync.WaitGroup) {
	c.wg = wg
}
