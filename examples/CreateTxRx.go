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

	Ntx := 2
	tx := make([]TxRx.SimpleTransmitter, Ntx)
	// wg := new(sync.WaitGroup)

	for i := 0; i < Ntx; i++ {
		// wg.Add(1)
		tx[i].Init()
		tx[i].SetID(i + 200)
		tx[i].Nblocks = 5
		// tx[i].SetWaitGroup(wg)
		// 	go tx[i].StartTransmit()
	}
	rx := make([]TxRx.SimpleReceiver, Ntx)

	for i := 0; i < Ntx; i++ {
		// wg.Add(1)
		rx[i].Init()
		rx[i].SetID(100 + i)

		// rx[i].SetWaitGroup(wg)
		// 	go tx[i].StartTransmit()
	}
	//var sisochannel channel.Channel
	// sisochannel.CreateFromFile("linkmetric2.json")

	links := make([]cellular.LinkMetric, Ntx)
	for i := 0; i < Ntx; i++ {

		links[i] = cellular.CreateSimpleLink(rx[i].GetID(), tx[i].GetID(), 10)
		// log.Printf("Links between %d->%d : %#v", tx[i].GetID(), rx[i].GetID(), links[i])
	}
	sisochannel := channel.NewWirelessChannel(links)
	// log.Println("SFNS", sisochannel.SFNids())
	sfid := sisochannel.SFNids()[0]
	{
		txnodesIds := sisochannel.GetTxNodeIDs(sfid)
		log.Println(txnodesIds)

		for indx, txid := range txnodesIds {
			// log.Println("Did I change Txid indx : txid ", indx, txid)
			tx[indx].SetID(txid)
			var systx cellular.Transmitter
			systx = &tx[indx]

			sisochannel.AddTransmiter(systx)

			// log.Printf("%d Tx Added %d", indx, txid)
		}
		rxnodesIds := sisochannel.GetRxNodeIDs(sfid)
		log.Println(rxnodesIds)
		for indx, rxid := range rxnodesIds {
			rx[indx].SetID(rxid)
			// log.Println("Did I change rxid indx : rxid ", indx, rxid)
			var sysrx cellular.Receiver
			sysrx = &rx[indx]
			sisochannel.AddReceiver(sysrx)
			// log.Printf("%d Rx Added %d", indx, rxid)
		}

	}
	sisochannel.Init()
	sisochannel.Start()

	// for i := 0; i < Ntx; i++ {
	// 	wg.Add(1)
	// 	go func(txch gocomm.Complex128AChannel) {
	// 		for cnt := 0; ; cnt++ {
	// 			chdata := <-txch
	// 			fmt.Printf("\n ,Tx %s : %f", chdata.Message, chdata.Ch)
	// 			if chdata.GetMaxExpected()-1 == cnt {
	// 				break
	// 			}
	// 		}
	// 	}(tx[i].GetChannel())

	// }
	// wg.Wait()
	matlab.Close()
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
