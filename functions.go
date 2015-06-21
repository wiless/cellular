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
	FrequencyGHz float64
	BandwidthMHz float64
	NoisePSDdBm  float64
}

func NewWSystem() WSystem {
	var result WSystem
	result.BandwidthMHz = 10.0
	result.NoisePSDdBm = -173.9
	return result
}

/// EvaluteMetric iteratively calls the path-loss m
func (w WSystem) EvaluteMetric(singlecell *deployment.DropSystem, model pathloss.Model, rxid int, afn AntennaOfTxNode) []LinkMetric {
	BandwidthMHz := w.BandwidthMHz
	NoisePSDdBm := w.NoisePSDdBm

	N0 := NoisePSDdBm + vlib.Db(BandwidthMHz*1e6)
	var PerFreqLink map[float64]LinkMetric
	PerFreqLink = make(map[float64]LinkMetric)
	rxnode := singlecell.Nodes[rxid]
	// nfrequencies := len(rxnode.Frequency)
	// log.SetOutput(os.Stderr)
	// log.Printf("%s[%d] Supports %3.2f GHz", rxnode.Type, rxnode.ID, rxnode.FreqGHz)
	txnodeTypes := singlecell.GetTxNodeNames()

	var alltxNodeIds vlib.VectorI
	for i := 0; i < len(txnodeTypes); i++ {
		alltxNodeIds.AppendAtEnd(singlecell.GetNodeIDs(txnodeTypes[i])...)
	}
	// fmt.Println("All txnodes are  : ", alltxNodeIds)
	for _, f := range rxnode.FreqGHz {
		var link LinkMetric

		link.FreqInGHz = f
		link.RxNodeID = rxid
		link.BestRSRP = -1000
		link.RoIDbm = -1000
		link.N0 = N0
		link.BandwidthMHz = BandwidthMHz
		// model.SetFreqHz = f * 1e9
		link.TxNodeIDs.Resize(0)
		nlinks := 0
		for _, val := range alltxNodeIds {
			txnodeID := val
			txnode := singlecell.Nodes[val]

			if found, _ := vlib.Contains(txnode.FreqGHz, f); found {

				nlinks++
				link.TxNodeIDs.AppendAtEnd(txnodeID)

				antenna := afn(txnodeID)
				antenna.FreqHz = f * 1.0e9
				// log.Println(txnode.Orientation)
				// antenna.HTiltAngle, antenna.VTiltAngle = txnode.Orientation[0], txnode.Orientation[1]
				// antenna.CreateElements(txnode.Location)
				//	log.Println("Checking Locations of Tx and Rx : ", txnode.Location, rxnode.Location)
				// lossDb := model.LossInDb(distance)
				//txnode.Location.Z = txnode.Height
				// model.LossInDb3D(txnode.Location, rxnode.Location)
				lossDb, _ := model.LossInDb3D(txnode.Location, rxnode.Location, f)
				aasgain, _, _ := antenna.AASGain(rxnode.Location)
				rxRSRP := vlib.Db(aasgain) + txnode.TxPowerDBm - lossDb
				link.TxNodesRSRP.AppendAtEnd(rxRSRP)

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
			//fmt.Println("\n  vlib.Sum(rsrpLinr), vlib.InvDb(link.N0) ", vlib.Db(vlib.Sum(rsrpLinr)), (link.N0))
			// if nlinks == 1 {
			// 	link.BestSINR = vlib.Db(maxrsrp) - N0
			// 	// +1000 /// s/i = MAX value
			// } else {
			if totalrssi == maxrsrp {
				link.BestSINR = vlib.Db(maxrsrp)
				if link.BestSINR > 200 {
					link.BestSINR = 1000
				}

			} else {
				link.BestSINR = vlib.Db(maxrsrp) - vlib.Db(totalrssi-maxrsrp)
			}

			link.RSSI = vlib.Db(totalrssi)

			sortedRxrp, indx := link.TxNodesRSRP.Sorted2()
			link.TxNodeIDs = link.TxNodeIDs.At(indx.Flip()...) // Sort it
			link.TxNodesRSRP = sortedRxrp.Flip()
			link.BestRSRP = link.TxNodesRSRP[0]
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

/// EvaluteMetric iteratively calls the path-loss m
func (w WSystem) EvaluteLinkMetric(singlecell *deployment.DropSystem, model pathloss.Model, rxid int, afn AntennaOfTxNode) LinkMetric {
	BandwidthMHz := w.BandwidthMHz
	NoisePSDdBm := w.NoisePSDdBm
	systemFrequencyGHz := w.FrequencyGHz

	N0 := NoisePSDdBm + vlib.Db(BandwidthMHz*1e6)
	var link LinkMetric
	rxnode := singlecell.Nodes[rxid]

	txnodeTypes := singlecell.GetTxNodeNames()

	var alltxNodeIds vlib.VectorI
	for i := 0; i < len(txnodeTypes); i++ {
		alltxNodeIds.AppendAtEnd(singlecell.GetNodeIDs(txnodeTypes[i])...)
	}
	// fmt.Println("All txnodes are  : ", alltxNodeIds)

	if rxnode.FreqGHz.Contains(systemFrequencyGHz) {

		link.FreqInGHz = systemFrequencyGHz
		link.RxNodeID = rxid
		link.BestRSRP = -1000
		link.RoIDbm = -1000
		link.N0 = N0
		link.BandwidthMHz = BandwidthMHz

		// model.SetFreqHz = f * 1e9
		link.TxNodeIDs.Resize(0)
		nlinks := 0
		for _, val := range alltxNodeIds {
			txnodeID := val
			txnode := singlecell.Nodes[val]

			if found := txnode.FreqGHz.Contains(systemFrequencyGHz); found {

				nlinks++
				link.TxNodeIDs.AppendAtEnd(txnodeID)

				antenna := afn(txnodeID)
				antenna.FreqHz = systemFrequencyGHz * 1.0e9
				// log.Println(txnode.Orientation)
				// antenna.HTiltAngle, antenna.VTiltAngle = txnode.Orientation[0], txnode.Orientation[1]
				// antenna.CreateElements(txnode.Location)
				//	log.Println("Checking Locations of Tx and Rx : ", txnode.Location, rxnode.Location)
				// lossDb := model.LossInDb(distance)
				//txnode.Location.Z = txnode.Height
				// model.LossInDb3D(txnode.Location, rxnode.Location)
				lossDb, _ := model.LossInDb3D(txnode.Location, rxnode.Location, systemFrequencyGHz)
				aasgain, _, _ := antenna.AASGain(rxnode.Location)
				rxRSRP := vlib.Db(aasgain) + txnode.TxPowerDBm - lossDb
				link.TxNodesRSRP.AppendAtEnd(rxRSRP)

			} else {
				log.Printf("%s[%d] : TxNode %d : No Link on %3.2fGHz", rxnode.Type, rxid, val, systemFrequencyGHz)

			}
		}

		/// Do the statistics here
		if nlinks > 0 {
			link.N0 = N0
			link.BandwidthMHz = BandwidthMHz
			rsrpLinr := vlib.InvDbF(link.TxNodesRSRP)
			totalrssi := vlib.Sum(rsrpLinr) + vlib.InvDb(link.N0)
			maxrsrp := vlib.Max(rsrpLinr)
			//fmt.Println("\n  vlib.Sum(rsrpLinr), vlib.InvDb(link.N0) ", vlib.Db(vlib.Sum(rsrpLinr)), (link.N0))
			// if nlinks == 1 {
			// 	link.BestSINR = vlib.Db(maxrsrp) - N0
			// 	// +1000 /// s/i = MAX value
			// } else {
			if totalrssi == maxrsrp {
				link.BestSINR = vlib.Db(maxrsrp)
				if link.BestSINR > 200 {
					link.BestSINR = 1000
				}

			} else {
				link.BestSINR = vlib.Db(maxrsrp) - vlib.Db(totalrssi-maxrsrp)
			}

			link.RSSI = vlib.Db(totalrssi)

			sortedRxrp, indx := link.TxNodesRSRP.Sorted2()
			link.TxNodeIDs = link.TxNodeIDs.At(indx.Flip()...) // Sort it
			link.TxNodesRSRP = sortedRxrp.Flip()
			link.BestRSRP = link.TxNodesRSRP[0]
			link.BestRSRPNode = link.TxNodeIDs[0]

		}

	}

	// if len(PerFreqLink) != 0 {

	// 	log.Printf("%#v", PerFreqLink)
	// }

	return link
}
