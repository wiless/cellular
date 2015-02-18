package main

import (
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

	Ntx := 10
	tx := make([]TxRx.SimpleTransmitter, Ntx)

	for i := 0; i < Ntx; i++ {
		tx[i].Init()
		tx[i].SetID(i + 200)
		tx[i].Nblocks = 30

	}
	rx := make([]TxRx.SimpleReceiver, Ntx)

	for i := 0; i < Ntx; i++ {
		rx[i].Init()
		rx[i].SetID(100 + i)
	}
	//var sisochannel channel.Channel
	// sisochannel.CreateFromFile("linkmetric2.json")

	links := make([]cellular.LinkMetric, Ntx)
	for i := 0; i < Ntx; i++ {

		links[i] = cellular.CreateSimpleLink(rx[i].GetID(), tx[i].GetID(), 10)
		// log.Printf("Links between %d->%d : %#v", tx[i].GetID(), rx[i].GetID(), links[i])
	}
	sisochannel := channel.NewWirelessChannel(links)
	sfid := sisochannel.SFNids()[0]
	{
		txnodesIds := sisochannel.GetTxNodeIDs(sfid)
		for indx, txid := range txnodesIds {
			tx[indx].SetID(txid)
			var systx cellular.Transmitter
			systx = &tx[indx]
			sisochannel.AddTransmiter(systx)
		}
		rxnodesIds := sisochannel.GetRxNodeIDs(sfid)
		for indx, rxid := range rxnodesIds {
			rx[indx].SetID(rxid)
			var sysrx cellular.Receiver
			sysrx = &rx[indx]
			sisochannel.AddReceiver(sysrx)
		}

	}
	sisochannel.Init()
	sisochannel.Start()
	log.Println("Done..")
	matlab.Close()
	log.Println("Time Elapsed ", time.Since(starttime))
}
