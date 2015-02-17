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
}

func (s *SimpleReceiver) Init() {
	// Dummy init function
	s.key = string(vlib.RandString(8))

}

// Simple rx node
func (c *SimpleReceiver) StartReceive(rxch gocomm.Complex128AChannel) {
	c.filename = strconv.Itoa(c.GetID()) + "_" + c.key + ".dat"
	w, _ := os.Create(c.filename)

	for i := 0; ; i++ {
		// log.Printf("CSFReceiver: Rx-%d Waiting to read data at Input ", c.GetID())
		rdata := <-rxch
		fmt.Fprintf(w, "\n%d : %#v", i, rdata)
		log.Printf("SimpleReceiver (%d): Received Packet ID  %f ", c.GetID(), rdata.TimeStamp)
		if i == rdata.GetMaxExpected()-1 {
			break
		}

	}
	if c.wg != nil {
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
