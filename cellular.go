package cellular

import (
	// "sync"

	// "github.com/wiless/gocomm"
	"github.com/wiless/vlib"
)

type GenericStruct map[string]interface{}

// type LinkInfo struct {
// 	RxNodeID          int
// 	NodeTypes         []string
// 	LinkGain          vlib.VectorF
// 	LinkGainNode      vlib.VectorI
// 	InterferenceLinks vlib.VectorF
// }

type LinkMetric struct {
	RxNodeID     int
	FreqInGHz    float64
	BandwidthMHz float64
	N0           float64
	TxNodeIDs    vlib.VectorI
	TxNodesRSRP  vlib.VectorF
	RSSI         float64
	BestRSRP     float64
	BestRSRPNode int
	BestSINR     float64
	RoIDbm       float64
	//AgainDb      float64
	BestCouplingLoss  float64
	MaxTxAg           float64 // Tx AAS Gain
	MaxRxAg           float64 // Rx AAS Gain
	AssoTxAg          float64 // Tx AAS Gain for Associated Link
	AssoRxAg          float64 // Rx AAS Gain for Associated Link
	MaxTransmitBeamID int
}

func (l *LinkMetric) SetParams(fGHz, bwMHz float64) {
	// BandwidthMHz := 20.0
	NoisePSDdBmPerHz := -173.9
	l.N0 = NoisePSDdBmPerHz + vlib.Db(bwMHz*1e6)
	l.FreqInGHz = fGHz

}

//CreateLink creates a single tx-rx link with a given SNR with bandwidth=10MHz, Signal power assumed as 0dBm
//and N0 calculated based on 10MHz bandwidth
func CreateLink(rxid, txid int, snrDb float64) LinkMetric {
	var result LinkMetric

	result.SetParams(2.1, 10)
	result.RxNodeID = rxid
	result.TxNodeIDs.AppendAtEnd(txid)
	rssi := snrDb + result.N0
	result.TxNodesRSRP.AppendAtEnd(rssi)
	result.RoIDbm = -1000
	return result
}

func CreateSimpleLink(rxid, txid int, snrDb float64) LinkMetric {
	var result LinkMetric
	result = CreateLink(rxid, txid, snrDb)
	result.N0 = -snrDb
	rssi := snrDb + result.N0
	result.TxNodesRSRP[0] = rssi
	return result
}

// type Transmitter interface {
// 	SetWaitGroup(wg *sync.WaitGroup)
// 	GetChannel() gocomm.Complex128AChannel
// 	GetID() int
// 	StartTransmit()
// 	GetSeed() int64
// 	IsActive() bool
// }

// type Receiver interface {
// 	GetID() int
// 	SetWaitGroup(wg *sync.WaitGroup)
// 	StartReceive(rxch gocomm.Complex128AChannel)
// 	IsActive() bool
// }
