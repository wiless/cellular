package main

import (
	"github.com/wiless/gocomm/sink"
	"math/rand"

	"github.com/wiless/cellular"
	"github.com/wiless/cellular/TxRx"
	"github.com/wiless/vlib"

	"github.com/wiless/cellular/channel"
	// "github.com/wiless/gocomm/sink"

	"log"

	// "os"

	"time"
)

var matlab *vlib.Matlab

func init() {
	matlab = vlib.NewMatlab("channel")
	matlab.Silent = true
	matlab.Json = true
	// rand.Seed(time.Now().Unix())
	rand.Seed(0)
	log.Println("Seed is : ", rand.Uint32())
	// rand.Seed(time.Now().Unix())

}

func main() {
	starttime := time.Now()

	Ntx := 2
	tx := make([]TxRx.SimpleTransmitter, Ntx)
	for i := 0; i < Ntx; i++ {
		tx[i].Init()
		tx[i].SetID(i + 100)
		tx[i].Nblocks = 200
		tx[i].BlockLen = 32

	}

	rx := make([]TxRx.SimpleReceiver, Ntx)
	for i := 0; i < Ntx; i++ {
		rx[i].Init()
		rx[i].SetID(200 + i)
	}
	//var channelEmulator channel.Channel
	// channelEmulator.CreateFromFile("linkmetric2.json")

	links := make([]cellular.LinkMetric, Ntx)
	SNRdB := 10.0
	for i := 0; i < Ntx; i++ {
		if i == 0 {
			SNRdB = 105
		} else {
			SNRdB = 105
		}
		links[i] = cellular.CreateSimpleLink(rx[i].GetID(), tx[i].GetID(), SNRdB)
		// log.Printf("Links between %d->%d : %#v", tx[i].GetID(), rx[i].GetID(), links[i])
	}
	// time.Sleep(100 * time.Second)

	channelEmulator := channel.NewWirelessChannel(links)

	// sfid := channelEmulator.SFNids()[0]
	{
		for indx := 0; indx < Ntx; indx++ {
			var systx cellular.Transmitter
			systx = &tx[indx]
			channelEmulator.AddTransmiter(systx)
		}

		for indx := 0; indx < Ntx; indx++ {
			var sysrx cellular.Receiver
			sysrx = &rx[indx]
			channelEmulator.AddReceiver(sysrx)
		}

	}

	// txprobe0 := tx[0].GetProbe(0)
	// go sink.CROcomplexAScatter(txprobe0)
	// txprobe1 := tx[1].GetProbe(0)
	// go sink.CROcomplexAScatter(txprobe1)

	rxprobe0 := rx[0].GetProbe(0)
	go sink.CROcomplexAScatter(rxprobe0)
	// rxprobe1 := rx[1].GetProbe(0)
	// go sink.CROcomplexAScatter(rxprobe1)

	channelEmulator.Init()
	channelEmulator.Start()
	log.Println("Done..")
	matlab.Close()
	log.Println("Time Elapsed ", time.Since(starttime))
}
