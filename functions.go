package cellular

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/wiless/cellular/antenna"
	"github.com/wiless/cellular/deployment"
	"github.com/wiless/cellular/pathloss"
	"github.com/wiless/channelmodel"
	"github.com/wiless/vlib"
)

type WSystem struct {
	FrequencyGHz float64
	BandwidthMHz float64
	NoisePSDdBm  float64
	ActiveCells  vlib.VectorI
	OtherLossFn  func(txnode, rxnode deployment.Node) float64
}

var DEFAULTERR_PL float64 = 999999

func NewWSystem() WSystem {
	var result WSystem
	result.BandwidthMHz = 10.0
	result.NoisePSDdBm = -173.9
	return result
}

type AntennaOfTxNode func(txnodeID int) antenna.SettingAAS

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

				ant := afn(txnodeID)
				ant.FreqHz = f * 1.0e9
				ant.Centre = txnode.Location
				ant.HTiltAngle = txnode.Direction

				ant.CreateElements(txnode.Location)
				// log.Println(txnode.Orientation)
				// antenna.HTiltAngle, antenna.VTiltAngle = txnode.Orientation[0], txnode.Orientation[1]
				// antenna.CreateElements(txnode.Location)
				//	log.Println("Checking Locations of Tx and Rx : ", txnode.Location, rxnode.Location)
				// lossDb := model.LossInDb(distance)
				//txnode.Location.Z = txnode.Height
				// model.LossInDb3D(txnode.Location, rxnode.Location)
				log.Println("what is this ===#=====#=========*=== ", ant)
				time.Sleep(1 * time.Second)
				lossDb, _ := model.LossInDb3D(txnode.Location, rxnode.Location, f)
				aasgain, _, _ := ant.AASGain(rxnode.Location)

				var otherLossDb float64 = 0
				if w.OtherLossFn != nil {
					otherLossDb = w.OtherLossFn(txnode, rxnode)
				}
				// log.Print(vlib.Db(aasgain), txnode.TxPowerDBm, lossDb, otherLossDb)
				rxRSRP := vlib.Db(aasgain) + txnode.TxPowerDBm - lossDb - otherLossDb
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
	if w.ActiveCells.Size() == 0 {

		for i := 0; i < len(txnodeTypes); i++ {
			alltxNodeIds.AppendAtEnd(singlecell.GetNodeIDs(txnodeTypes[i])...)
		}
	} else {
		alltxNodeIds = w.ActiveCells
	}

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

				ant := afn(txnodeID)
				ant.Centre = txnode.Location
				ant.FreqHz = systemFrequencyGHz * 1.0e9
				// log.Println("Antenna for ", txnodeID, ant.VTiltAngle)
				// log.Println(txnode.Orientation)
				// antenna.HTiltAngle, antenna.VTiltAngle = txnode.Orientation[0], txnode.Orientation[1]
				// antenna.CreateElements(txnode.Location)
				//	log.Println("Checking Locations of Tx and Rx : ", txnode.Location, rxnode.Location)
				// lossDb := model.LossInDb(distance)
				//txnode.Location.Z = txnqodeth.Height
				// model.LossInDb3D(txnode.Location, rxnode.Location)
				// if rxid == 401 {
				// fmt.Println(txnode)
				// fmt.Printf("\nRXNODE %d , Received Antenna Gain BS-%d distance is %f ", rxid, txnodeID)
				// fmt.Printf("\n TxNode Location :  %v \n Antenna location %v \n Rx Location ", txnode.Location, ant.Centre, rxnode.Location)

				// dist, thetaH, thetaV := vlib.RelativeGeo(txnode.Location, rxnode.Location)
				// dist, thetaH, thetaV := ant.Centre(ant.Location, rxnode.Location)
				// ant.CreateElements(txnode.Location)
				// fmt.Println("\nAntenna External w.r.t TX ", dist, thetaH, thetaV)
				// dist, thetaH, thetaV = vlib.RelativeGeo(ant.Centre, rxnode.Location)
				// fmt.Println("\nAntenna External w.r.t Antenna Centre", dist, thetaH, thetaV)
				// elementLocations := ant.GetElements()
				// for i, v := range elementLocations {
				// 	dist, thetaH, thetaV = vlib.RelativeGeo(v, rxnode.Location)
				// 	fmt.Printf("\nAntenna Elements %d @ %v \n Metrics  w.r.t Antenna Centre : %f %f %f", i, v, dist, thetaH, thetaV)
				// }
				// // }

				lossDb, plerr := model.LossInDb3D(txnode.Location, rxnode.Location, systemFrequencyGHz)
				if !plerr {
					log.Fatal("Cannot work")
				}

				aasgain, _, _ := ant.AASGain(rxnode.Location)

				// fmt.Printf("\nOther values are aas=%v,txpower=%v,Ploss=%v , PLERROR =%v", vlib.Db(aasgain), txnode.TxPowerDBm, lossDb, plerr)
				rxRSRP := vlib.Db(aasgain) + txnode.TxPowerDBm - lossDb

				if rxRSRP > -59 {
					fmt.Printf("\n asdfasdfs EVAL1 %d RSSI =%v, AAS =%f ,PL = %f, otherLoss=%f ", rxid, rxRSRP, vlib.Db(aasgain), lossDb)
				}
				// fmt.Printf("\n Distance is %f", dist)
				// fmt.Printf("\n Angle is H,V %f,%f", thetaH, thetaV)

				// fmt.Printf("\n Storing RSRP for %d is %v", txnodeID, rxRSRP)

				if math.IsInf(rxRSRP, 0) {
					log.Panicln("============= %%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%")
				}

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

// EvaluteLinkMetricV2 evalutes the link metric with New PL model interface
func (w *WSystem) EvaluteLinkMetricV2(singlecell *deployment.DropSystem, model CM.PLModel, rxid int, afn AntennaOfTxNode) LinkMetric {
	BandwidthMHz := w.BandwidthMHz
	NoisePSDdBm := w.NoisePSDdBm
	systemFrequencyGHz := w.FrequencyGHz

	N0 := NoisePSDdBm + vlib.Db(BandwidthMHz*1e6)
	var link LinkMetric
	rxnode := singlecell.Nodes[rxid]

	txnodeTypes := singlecell.GetTxNodeNames()

	var alltxNodeIds vlib.VectorI
	if w.ActiveCells.Size() == 0 {

		for i := 0; i < len(txnodeTypes); i++ {
			alltxNodeIds.AppendAtEnd(singlecell.GetNodeIDs(txnodeTypes[i])...)
		}
	} else {
		alltxNodeIds = w.ActiveCells
	}

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

				ant := afn(txnodeID)
				ant.Centre = txnode.Location
				ant.FreqHz = systemFrequencyGHz * 1.0e9
				// log.Println("Antenna for ", txnodeID, ant.VTiltAngle)
				// log.Println(txnode.Orientation)
				// antenna.HTiltAngle, antenna.VTiltAngle = txnode.Orientation[0], txnode.Orientation[1]
				// antenna.CreateElements(txnode.Location)
				//	log.Println("Checking Locations of Tx and Rx : ", txnode.Location, rxnode.Location)
				// lossDb := model.LossInDb(distance)
				//txnode.Location.Z = txnqodeth.Height
				// model.LossInDb3D(txnode.Location, rxnode.Location)
				// if rxid == 401 {
				// fmt.Println(txnode)
				// fmt.Printf("\nRXNODE %d , Received Antenna Gain BS-%d distance is %f ", rxid, txnodeID)
				// fmt.Printf("\n TxNode Location :  %v \n Antenna location %v \n Rx Location ", txnode.Location, ant.Centre, rxnode.Location)

				// dist, thetaH, thetaV := vlib.RelativeGeo(txnode.Location, rxnode.Location)
				// dist, thetaH, thetaV := ant.Centre(ant.Location, rxnode.Location)
				// ant.CreateElements(txnode.Location)
				// fmt.Println("\nAntenna External w.r.t TX ", dist, thetaH, thetaV)
				// dist, thetaH, thetaV = vlib.RelativeGeo(ant.Centre, rxnode.Location)
				// fmt.Println("\nAntenna External w.r.t Antenna Centre", dist, thetaH, thetaV)
				// elementLocations := ant.GetElements()
				// for i, v := range elementLocations {
				// 	dist, thetaH, thetaV = vlib.RelativeGeo(v, rxnode.Location)
				// 	fmt.Printf("\nAntenna Elements %d @ %v \n Metrics  w.r.t Antenna Centre : %f %f %f", i, v, dist, thetaH, thetaV)
				// }
				// // }
				var lossDb float64
				var dist float64
				if model.IsSupported(systemFrequencyGHz) {

					pldb, _, plerr := model.PLbetween(txnode.Location, rxnode.Location)
					dist = txnode.Location.Distance2DFrom(rxnode.Location)
					if plerr != nil {
						log.Printf("PathLoss Error : %v, setting to FIXED %v", plerr, DEFAULTERR_PL)
						pldb = DEFAULTERR_PL
					}
					lossDb = pldb
				} else {
					log.Fatalf("The Current Path loss Model %#v Doest not support Frequency %vGHz", model, systemFrequencyGHz)
				}

				aasgain, _, _ := ant.AASGain(rxnode.Location)

				// fmt.Printf("\nOther values are aas=%v,txpower=%v,Ploss=%v , PLERROR =%v", vlib.Db(aasgain), txnode.TxPowerDBm, lossDb, plerr)

				var otherLossDb float64 = 0
				if w.OtherLossFn != nil {
					otherLossDb = w.OtherLossFn(txnode, rxnode)
				}
				rxRSRP := vlib.Db(aasgain) + txnode.TxPowerDBm - lossDb - otherLossDb
				AASGain := vlib.Db(aasgain)
				if AASGain > ant.GainDb {
					fmt.Printf("\n EVAL2 %d RSSI =%v, AAS =%f [%f] ,PL = %f , , otherLoss=%f , dist =%v", rxid, rxRSRP, AASGain, aasgain, lossDb, otherLossDb, dist)
				}
				// log.Print(rxRSRP, "=", vlib.Db(aasgain), txnode.TxPowerDBm, lossDb, otherLossDb)
				// rxRSRP := vlib.Db(aasgain) + txnode.TxPowerDBm - lossDb
				// fmt.Printf("\n Distance is %f", dist)
				// fmt.Printf("\n Angle is H,V %f,%f", thetaH, thetaV)

				// fmt.Printf("\n Storing RSRP for %d is %v", txnodeID, rxRSRP)

				if math.IsInf(rxRSRP, 0) {
					log.Panicln("============= %%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%% %")
				}

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
