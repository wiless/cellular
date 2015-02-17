package main

import (
	"fmt"
	"github.com/wiless/cellular"
	"github.com/wiless/cellular/channel"
	"github.com/wiless/gocomm"
	"github.com/wiless/gocomm/chipset"
	"github.com/wiless/gocomm/modem"
	"github.com/wiless/gocomm/sink"
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
	nid      int
	sch      gocomm.Complex128AChannel
	proxyPin gocomm.Complex128AChannel
	Nblocks  int
	wg       *sync.WaitGroup
}

type CSVReceiver struct {
	nid int
	wg  *sync.WaitGroup
}

func main() {
	starttime := time.Now()
	var sisochannel channel.Channel
	sisochannel.CreateFromFile("linkmetric2.json")

	// links := make([]cellular.LinkMetric, 1)
	// links[0] = cellular.CreateSimpleLink(csvr.GetID(), swg.GetID(), 10)
	// sisochannel := channel.NewWirelessChannel(links)

	// swg.nid, csvr.nid = 0, 1
	// sink.CRO(scale, NextSize, InCH)
	cmplxCH := gocomm.NewComplex128AChannel()
	go sink.CROcomplexAScatter(cmplxCH)
	var data gocomm.SComplex128AObj
	data.MaxExpected = 200
	// fd, err := os.OpenFile("/home/ssk/projects/goGUI/MyPipe", os.O_WRONLY, os.ModeNamedPipe)

	// log.Print("File Error ", err)
	// fmt.Fprintf(fd, "Hello from golang")

	// fd.Close()
	qpsk := modem.NewModem(2)
	noise := rand.Intn(1) == 0
	for i := 0; i < data.GetMaxExpected(); i++ {
		bits := vlib.RandB(128)
		if noise {

			data.Ch = vlib.AddC(qpsk.ModulateBits(bits), vlib.RandNCVec(64, .01))
		} else {
			data.Ch = qpsk.ModulateBits(bits)

		}

		// data.Ch = vlib.RandNCVec(64, .000001)
		// fmt.Printf("\r %d : %f", i, data.Ch[0:10])
		// fmt.Fprintf(fd, "\n%d: %d samples", i, len(data.Ch))
		cmplxCH <- data

		time.Sleep(400 * time.Millisecond)
	}

	// fd.Close()
	// log.Print(cmplxCH)
	return
	var myprobe sink.TwoPinProbe
	var proxypin gocomm.Complex128AChannel

	sfid := sisochannel.SFNids()[0]
	{
		txnodesIds := sisochannel.GetTxNodeIDs(sfid)
		log.Println(txnodesIds)

		for indx, txid := range txnodesIds {
			var swg SinWaveGenerator
			swg.Init()
			swg.nid = txid

			///
			if indx == 1 {

				proxypin = myprobe.ProxyPin(swg.GetChannel())
				swg.SetProxyPin(proxypin)
				go myprobe.Probe()
			}

			var tx cellular.Transmitter
			tx = &swg

			sisochannel.AddTransmiter(tx)

			// log.Printf("%d Tx Added %d", indx, txid)
		}
		rxnodesIds := sisochannel.GetRxNodeIDs(sfid)
		log.Println(rxnodesIds)
		for _, rxid := range rxnodesIds {
			var csvr CSVReceiver
			csvr.nid = rxid
			var rx cellular.Receiver
			rx = &csvr
			sisochannel.AddReceiver(rx)
			// log.Printf("%d Rx Added %d", indx, rxid)
		}

	}

	sisochannel.Init()
	sisochannel.Start()

	// func() {

	// 	for i := 0; ; i++ {
	// 		// log.Printf("CSFReceiver: Rx-%d Waiting to read data at Input ", c.GetID())
	// 		rdata := <-proxypin
	// 		log.Println("CRO : Received Packet ID  ", i)
	// 		if i == rdata.GetMaxExpected()-1 {
	// 			break
	// 		}

	// 	}

	// }()

	log.Println("Done..")

	matlab.Close()
	log.Println("Time Elapsed ", time.Since(starttime))
}

func (s SinWaveGenerator) GetChannel() gocomm.Complex128AChannel {
	if s.proxyPin == nil {
		return s.sch
	}
	return s.proxyPin
}

///
func (s *SinWaveGenerator) SetProxyPin(proxypin gocomm.Complex128AChannel) {
	s.proxyPin = proxypin
}

func (s *SinWaveGenerator) Init() {
	s.sch = gocomm.NewComplex128AChannel()
	s.proxyPin = nil
	s.Nblocks = 10
}
func (s *SinWaveGenerator) StartTransmit() {

	if s.sch == nil {
		log.Panicln("SinWaveGenerator Not Initialized !! No channel yet")
	}

	// log.Println("Ready to send ??")
	var chdata gocomm.SComplex128AObj
	chdata.MaxExpected = s.Nblocks
	chdata.Message = "BS"
	chdata.Ts = 1
	N := 32                   // 32bits=16SYMBOLS per TTI
	qpsk := modem.NewModem(2) // QPSK Modem
	// log.Println("SineWaveGen: Ready to send ??")
	for i := 0; i < s.Nblocks; i++ {
		chdata.Next(qpsk.ModulateBits(vlib.RandB(N)))
		log.Printf("SineWaveGen: Block-%d Writing into Go-chan Tx-%d with %d symbols ", i, s.GetID(), len(chdata.Ch))
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
	s.wg = wg
}

// Simple rx node
func (c *CSVReceiver) StartReceive(rxch gocomm.Complex128AChannel) {
	w, _ := os.Create("output.dat")

	for i := 0; ; i++ {
		// log.Printf("CSFReceiver: Rx-%d Waiting to read data at Input ", c.GetID())
		rdata := <-rxch
		fmt.Fprintf(w, "\n%d : %#v", i, rdata)
		log.Println("CSFReceiver: Received Packet ID  ", i)
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
