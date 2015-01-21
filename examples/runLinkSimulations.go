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

	var swg SinWaveGenerator

	var csvr CSVReceiver
	swg.nid, csvr.nid = 0, 1

	// var sisochannel channel.Channel
	// sisochannel.CreateFromFile("linkmetric2.json")

	links := make([]cellular.LinkMetric, 1)

	links[0] = cellular.CreateSimpleLink(csvr.GetID(), swg.GetID(), 10)

	sisochannel := channel.NewWirelessChannel(links)

	swg.Init()
	{

		var tx cellular.Transmitter
		var rx cellular.Receiver
		tx = &swg
		rx = &csvr

		sisochannel.AddTransmiter(tx)
		sisochannel.AddReceiver(rx)

	}

	sisochannel.Init()
	sisochannel.Start()

	log.Println("Done..")

	matlab.Close()
	fmt.Println("\n")
}

func (s SinWaveGenerator) GetChannel() gocomm.Complex128AChannel {
	return s.sch
}

func (s *SinWaveGenerator) Init() {
	s.sch = gocomm.NewComplex128AChannel()
	s.Nblocks = 10
}
func (s *SinWaveGenerator) StartTransmit() {
	fmt.Println("Current WG = ", s.wg)
	if s.sch == nil {
		log.Panicln("SinWaveGenerator Not Intialized !! No channel yet")
	}
	log.Println("Ready to send ??")
	var chdata gocomm.SComplex128AObj
	chdata.MaxExpected = s.Nblocks
	chdata.Message = "BS"
	chdata.Ts = 1
	N := 32                   // 32bits=16SYMBOLS per TTI
	qpsk := modem.NewModem(2) // QPSK Modem
	log.Println("Ready to send ??")
	for i := 0; i < s.Nblocks; i++ {
		chdata.Next(qpsk.ModulateBits(vlib.RandB(N)))
		log.Println("Sending Tx..", i, " with   ", len(chdata.Ch), " symbols ")
		s.sch <- chdata
	}
	if s.wg != nil {
		s.wg.Done()
	}

}

func (s SinWaveGenerator) GetID() int {
	return s.nid
}

func (s SinWaveGenerator) GetSeed() int64 {
	return 0
}
func (s SinWaveGenerator) IsActive() bool {
	return true
}
func (s *SinWaveGenerator) SetWaitGroup(wg *sync.WaitGroup) {
	log.Println("Before WG = ", s.wg)
	s.wg = wg
	log.Println("After WG = ", wg)
}

// Simple rx node
func (c *CSVReceiver) StartReceive(rxch gocomm.Complex128AChannel) {
	w, _ := os.Create("output.dat")

	for i := 0; ; i++ {
		log.Println("Waiting to read data from my input pin ", i)
		rdata := <-rxch
		fmt.Fprintf(w, "\n%d : %#v", i, rdata)
		log.Println("CSVReceive ..", i)
		if i == rdata.GetMaxExpected()-1 {
			break
		}

	}
	if c.wg != nil {
		c.wg.Done()
	}

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
