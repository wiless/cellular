package cellular

import (
	"log"

	"github.com/wiless/cellular/antenna"

	"github.com/wiless/cellular/deployment"
	"github.com/wiless/cellular/pathloss"
	"github.com/wiless/vlib"
)

type AntennaOfTxNode func(txnodeID int) antenna.SettingAAS

type WSystem struct {
	BandwidthMHz float64
	NoisePSDdBm  float64
}

func NewWSystem() WSystem {
	var result WSystem
	result.BandwidthMHz = 10.0
	result.NoisePSDdBm = -173.9
	return result
}

func (w WSystem) EvaluteMetric(singlecell *deployment.DropSystem, model *pathloss.PathLossModel, rxid int, afn AntennaOfTxNode) []LinkMetric {
	BandwidthMHz := w.BandwidthMHz
	NoisePSDdBm := w.NoisePSDdBm
	N0 := NoisePSDdBm + vlib.Db(BandwidthMHz*1e6)
	var PerFreqLink map[float64]LinkMetric
	PerFreqLink = make(map[float64]LinkMetric)
	rxnode := singlecell.Nodes[rxid]
	// nfrequencies := len(rxnode.Frequency)
	// log.SetOutput(os.Stderr)
	log.Printf("%s[%d] Supports %3.2f GHz", rxnode.Type, rxnode.ID, rxnode.FreqGHz)
	txnodeTypes := singlecell.GetTxNodeNames()

	var alltxNodeIds vlib.VectorI
	for i := 0; i < len(txnodeTypes); i++ {
		alltxNodeIds.AppendAtEnd(singlecell.GetNodeIDs(txnodeTypes[i])...)
	}

	for _, f := range rxnode.FreqGHz {
		var link LinkMetric

		link.FreqInGHz = f
		link.RxNodeID = rxid
		link.BestRSRP = -1000
		link.RoIDbm = -1000
		link.N0 = N0
		link.BandwidthMHz = BandwidthMHz
		model.FreqHz = f * 1e9

		nlinks := 0
		for _, val := range alltxNodeIds {
			txnodeID := val
			txnode := singlecell.Nodes[val]

			if found, _ := vlib.Contains(txnode.FreqGHz, f); found {
				nlinks++
				link.TxNodeIDs.AppendAtEnd(txnodeID)
				antenna := afn(txnodeID)
				antenna.Freq = f * 1.0e9

				antenna.HTiltAngle, antenna.VTiltAngle = txnode.Orientation[0], txnode.Orientation[1]

				antenna.CreateElements(txnode.Location)
				distance, _, _ := vlib.RelativeGeo(txnode.Location, rxnode.Location)

				lossDb := model.LossInDb(distance)
				aasgain, _, _ := antenna.AASGain(rxnode.Location) /// linear scale
				totalGainDb := vlib.Db(aasgain) - lossDb
				link.TxNodesRSRP.AppendAtEnd(totalGainDb)

				log.Printf("%s[%d] : TxNode %d : Link @ %3.2fGHz  : %-4.3fdB", rxnode.Type, rxid, val, f, totalGainDb)

			} else {
				log.Printf("%s[%d] : TxNode %d : No Link on %3.2fGHz", rxnode.Type, rxid, val, f)

			}
		}

		/// Do the statistics here
		if nlinks > 0 {
			link.N0 = N0
			link.BandwidthMHz = BandwidthMHz

			rsrpLinr := vlib.InvDbF(link.TxNodesRSRP)
			totalrssi := vlib.Sum(rsrpLinr) + vlib.InvDb(link.N0)
			maxrsrp := vlib.Max(rsrpLinr)

			// if nlinks == 1 {
			// 	link.BestSINR = vlib.Db(maxrsrp) - N0
			// 	// +1000 /// s/i = MAX value
			// } else {
			link.BestSINR = vlib.Db(maxrsrp / (totalrssi - maxrsrp))
			// }
			val, sindx := vlib.Sorted(link.TxNodesRSRP)

			// fmt.Println("Sorted TxNodes & Values : ", link.TxNodeIDs, link.TxNodesRSRP)
			link.TxNodesRSRP = val
			link.TxNodeIDs = link.TxNodeIDs.At(sindx)
			// fmt.Println("Sorted TxNodes & Values : ", link.TxNodeIDs, link.TxNodesRSRP)

			link.RSSI = vlib.Db(totalrssi)
			link.BestRSRP = vlib.Db(maxrsrp)
			link.BestRSRPNode = link.TxNodeIDs[0]
			PerFreqLink[f] = link
		}

	}
	result := make([]LinkMetric, len(PerFreqLink))
	var cnt int = 0
	for _, val := range PerFreqLink {

		result[cnt] = val
		cnt++
	}

	// if len(PerFreqLink) != 0 {

	// 	log.Printf("%#v", PerFreqLink)
	// }

	return result
}
