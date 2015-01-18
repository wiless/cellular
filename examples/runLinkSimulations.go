package main

import (
	"fmt"
	"github.com/wiless/cellular"
	"github.com/wiless/cellular/channel"
	"github.com/wiless/gocomm"
	"github.com/wiless/gocomm/modem"
	"github.com/wiless/vlib"

	"log"
	"math/rand"
	"os"
	"sync"
	// "os"

	"time"
)

var matlab *vlib.Matlab

func init() {
	matlab = vlib.NewMatlab("channel")
	matlab.Silent = true
	matlab.Json = true
	rand.Seed(time.Now().Unix())

}

type SinWaveGenerator struct {
	nid     int
	sch     gocomm.Complex128AChannel
	Nblocks int
	wg      *sync.WaitGroup
}

type CSVReceiver struct {
	nid int
	wg  *sync.WaitGroup
}

func main() {

	// var sisochannel channel.Channel
	// sisochannel.CreateFromFile("linkmetric2.json")
	links := make([]cellular.LinkMetric, 1)
	links[0] = cellular.CreateSimpleLink(0, 1, 10)
	sisochannel := channel.NewWirelessChannel(links)
	sisochannel.Init()

	var swg SinWaveGenerator
	var csvr CSVReceiver
	swg.Init()
	{

		var tx cellular.Transmitter
		var rx cellular.Receiver
		tx = &swg
		rx = &csvr
		sisochannel.SetTransmiter(tx)
		sisochannel.SetReceiver(rx)
		sisochannel.Start()
	}

	// sisochannel.SetReceiver(r)
	// CreateChannelLinks()
	log.Println("Done..")
	// for i := 0; ; i++ {
	// 	rdata := <-tx.GetChannel()

	// 	fmt.Println("Rx ", i, " = > ", rdata.Ch)
	// 	if i == rdata.GetMaxExpected()-1 {
	// 		log.Println("Finished ", i+1, " blocks")
	// 		break
	// 	}
	// }

	// w, _ := os.Create("dump.txt")
	// for idx, val := range result {
	// 	fmt.Fprintf(w, "\n %d :  %#v\n", idx, val)
	// }

	matlab.Close()
	fmt.Println("\n")
}

//  Simple transmit node
// Example link-level simuation

func (s SinWaveGenerator) GetChannel() gocomm.Complex128AChannel {
	return s.sch
}

func (s *SinWaveGenerator) Init() {
	s.sch = gocomm.NewComplex128AChannel()
	s.Nblocks = 10
}
func (s *SinWaveGenerator) StartTransmit() {
	if s.sch == nil {
		log.Panicln("SinWaveGenerator Not Intialized !! No channel yet")
	}
	var chdata gocomm.SComplex128AObj
	chdata.MaxExpected = s.Nblocks
	chdata.Message = "BS"
	chdata.Ts = 1
	N := 32                   // 32bits=16SYMBOLS per TTI
	qpsk := modem.NewModem(2) // QPSK Modem

	for i := 0; i < s.Nblocks; i++ {
		chdata.Next(qpsk.ModulateBits(vlib.RandB(N)))
		// log.Println("Tx..", i)
		s.sch <- chdata
	}
	s.wg.Done()
}

func (s SinWaveGenerator) GetID() int {
	return 0
}

func (s SinWaveGenerator) GetSeed() int64 {
	return 0
}
func (s SinWaveGenerator) IsActive() bool {
	return true
}
func (s *SinWaveGenerator) SetWaitGroup(wg *sync.WaitGroup) {
	s.wg = wg
}

// Simple rx node
func (c *CSVReceiver) StartReceive(rxch gocomm.Complex128AChannel) {
	w, _ := os.Create("output.dat")
	c.nid = 2
	for i := 0; ; i++ {
		rdata := <-rxch
		fmt.Fprintf(w, "\n%d : %#v", i, rdata)
		// log.Println("Rx..", i)
		if i == rdata.GetMaxExpected()-1 {
			break
		}

	}
	c.wg.Done()
}

func (c CSVReceiver) GetID() int {
	return c.nid
}

func (c CSVReceiver) IsActive() bool {
	return true
}

func (c *CSVReceiver) SetWaitGroup(wg *sync.WaitGroup) {
	c.wg = wg
}
