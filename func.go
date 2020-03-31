package cellular

import (
	"fmt"
	"math"
	"math/rand"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/wiless/cellular/antenna"
	"github.com/wiless/cellular/deployment"
	CM "github.com/wiless/channelmodel"
	"github.com/wiless/vlib"
)

type WSystem struct {
	FrequencyGHz float64
	BandwidthMHz float64
	NoisePSDdBm  float64
	ActiveCells  vlib.VectorI
	OtherLossFn  func(plmodel CM.PLModel, txnode, rxnode deployment.Node, isLOS bool) float64
}

var DEFAULTERR_PL float64 = 999999

func NewWSystem() WSystem {
	var result WSystem
	result.BandwidthMHz = 10.0
	result.NoisePSDdBm = -174.0
	return result
}

type AntennaOfTxNode func(txnodeID int) antenna.SettingAAS

// EvaluteLinkMetricV2 evaluates the link metric with New PL model interface
func (w *WSystem) EvaluateLinkMetricV2(singlecell *deployment.DropSystem, model CM.PLModel, rxid int, afn AntennaOfTxNode) LinkMetric {

	BandwidthMHz := w.BandwidthMHz
	NoisePSDdBm := w.NoisePSDdBm
	systemFrequencyGHz := w.FrequencyGHz

	N0 := NoisePSDdBm - 30 + vlib.Db(BandwidthMHz*1e6)
	// fmt.Println("Noise Power is ", NoisePSDdBm, "After Bandwidth ",BandwidthMHz, N0)
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
		var rxdebugnode = false
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

				var lossDb float64
				var dist float64
				var d2In float64 = 0
				var otherLossDb float64 = 0
				var islos bool
				extraloss := 0.0
				inloss := 0.0
				lossDb = DEFAULTERR_PL
				rxRSRP := -DEFAULTERR_PL
				var plerr error

				if model.IsSupported(systemFrequencyGHz) && txnode.Active {
					dist = txnode.Location.Distance2DFrom(rxnode.Location)

					if rxnode.Indoor && model.Env() == "RMa" {
						d2In = rand.Float64() * 10.0

					} else if rxnode.Indoor && model.Env() == "UMa" {

						d2In = rand.Float64() * 25.0
					}

					lossDb, islos, plerr = model.PLbetweenIndoor(txnode.Location, rxnode.Location, d2In)
					if rxnode.Indoor {
						inloss = model.O2ILossDb(systemFrequencyGHz, d2In)
						otherLossDb += inloss
					}

					if w.OtherLossFn != nil {
						extraloss = w.OtherLossFn(model, txnode, rxnode, islos)
						otherLossDb += extraloss
					}

					if plerr != nil {
						log.Infof("EvaluateMetricV2 : (%d,%d) %v > %v", txnode.ID, rxnode.ID, lossDb, plerr)
						lossDb = DEFAULTERR_PL
					}
				} else {
					if !model.IsSupported(systemFrequencyGHz) {
						log.Fatalf("The Current Path loss Model %#v Doest not support Frequency %vGHz", model, systemFrequencyGHz)
					}
				}

				if txnode.Active {
					d3d, az, el := vlib.RelativeGeo(txnode.Location, rxnode.Location)
					antennaHBeamMax := 0.0
					el = -el + 90.0
					GCSaz := az + (txnode.Direction - antennaHBeamMax)
					GCSel := el - txnode.VTilt

					var Az, El, aasgainDB float64
					if model.Env() == "InH" {
						Az, El, aasgainDB = antenna.BSPatternIndoorHS_Db(GCSaz, GCSel)
					} else {
						Az, El, aasgainDB = antenna.BSPatternDb(GCSaz, GCSel)
					}
					_ = Az
					_ = El
					//	_, _, Aagain, result, Ag := antenna.CombPatternDb(Az, El, aasgainDB, 10, 4)

					// HGAINmaxDBi := 8.0 //
					_ = d3d
					//fmt.Printf("\n%d:%d (az,el)=[%v %v] distance=%v, SectorOrientation: %v, true AZ=(%v) EL(%v)%vdB ", txnodeID, rxid, az, el, d3d, txnode.Direction, GCSaz, GCSel, aasgainDB-8.0)
					// fmt.Printf("\n[Tx (%d),Rx(%d)]Antenna Gain aas=%v,txpower=%v,H,V (%v,%v)", txnode.ID, rxnode.ID, aasgainDB, txnode.TxPowerDBm, az, el)
					// az, el, aasgain2 := antenna.BSPatternDb(az, el)
					// fmt.Printf("\nNEW [Tx (%d),Rx(%d)]Antenna Gain aas=%v,txpower=%v,H,V (%v,%v)", txnode.ID, rxnode.ID, aasgain2, txnode.TxPowerDBm, az, el)
					// if aasgain2 != aasgainDB {
					// 	fmt.Println("\n  MIS MATCH ", aasgain2, aasgainDB)
					// }
					// // aasgainDB = aasgain2
					//Again = aasgainDB

					rxRSRP = aasgainDB + txnode.TxPowerDBm - 30 - lossDb - otherLossDb

					// if rxid == len(alltxNodeIds) {
					// 	fid, _ := os.Create("Rxlocation.dat")
					// 	fmt.Fprintf(fid, "%%ID\t\t\tRxid\t\t\tD3d\t\t\tRx\t\t\tRy\t\t\tRz\t\tPathloss\t\tIsLOS\t\tOtherlosses")
					// 	fmt.Fprintf(fid, "\n %d \t\t %d \t\t %f \t\t %f \t\t %f \t\t %f\t\t %f \t\t %t \t\t %f ", txnodeID, rxid, d3d, rxnode.Location.X, rxnode.Location.Y, rxnode.Location.Z, lossDb, islos, otherLossDb)
					// 	fid.Close()

					// 	fid1, _ := os.Create("Gain.dat")
					// 	fmt.Fprintf(fid1, "%%ID\t\tRxid\t\td3d\t\tRx\t\tRy\t\tRz\t\tAasgainDB\t\tAz\t\tEl")
					// 	fmt.Fprintf(fid1, "\n %d \t %d \t %f \t %f \t %f \t %f \t %f \t %f \t %f", txnodeID, rxid, d3d, rxnode.Location.X, rxnode.Location.Y, rxnode.Location.Z, aasgainDB, Az, math.Floor(El*1000)/1000)
					// 	fid1.Close()

					// } else {
					// 	fid, _ := os.OpenFile("Rxlocation.dat", os.O_APPEND|os.O_WRONLY, 0600)
					// 	fmt.Fprintf(fid, "\n %d \t\t %d \t\t %f \t\t %f \t\t %f \t\t %f\t\t %f \t\t %t \t\t %f ", txnodeID, rxid, d3d, rxnode.Location.X, rxnode.Location.Y, rxnode.Location.Z, lossDb, islos, otherLossDb)
					// 	fid.Close()

					// 	fid1, _ := os.OpenFile("Gain.dat", os.O_APPEND|os.O_WRONLY, 0600)
					// 	fmt.Fprintf(fid1, "\n %d \t %d \t %f \t %f \t %f \t %f \t %f \t %f \t %f", txnodeID, rxid, d3d, rxnode.Location.X, rxnode.Location.Y, rxnode.Location.Z, aasgainDB, Az, math.Floor(El*1000)/1000)
					// 	fid1.Close()
					// }

					rxdebugnode = true
					if rxdebugnode && rxRSRP > -90 {
						_ = dist
						// fmt.Printf("\r EVAL2 Rx-Tx (LOS:%v) %d-%d rxRSRP =%v,Power=%f,AAS =%f ,PL = %f, otherLoss=%f , dist =%v, d2In: =%v", islos, rxid, txnodeID, rxRSRP, txnode.TxPowerDBm, aasgainDB, lossDb, otherLossDb, dist, d2In)
						if rxnode.Indoor || rxnode.InCar {
							//	fmt.Println("\n Found in Indoor ", d2In, inloss, extraloss)
						}

					}
				}

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
			//link.AgainDb = Again

		}

	}

	return link
}

// EvaluateLinkMetricV3 evaluates the link metric with PL model interface and also dumps details of the coupling loss into linklogfname (csv)
// if linklogfname="" it does not dump
// EvaluateLinkMetricV3 evaluates the link metric with PL model interface and also dumps details of the coupling loss into linklogfname (csv)
// if linklogfname="" it does not dump
func (w *WSystem) EvaluateLinkMetricV3(singlecell *deployment.DropSystem, model CM.PLModel, rxid int, afn AntennaOfTxNode, fid *os.File) LinkMetric {

	var LOG = true
	if fid == nil {
		LOG = false
	}
	// var fid *os.File
	BandwidthMHz := w.BandwidthMHz
	NoisePSDdBm := w.NoisePSDdBm
	systemFrequencyGHz := w.FrequencyGHz
	rxnode := singlecell.Nodes[rxid]
	N0 := NoisePSDdBm + vlib.Db(BandwidthMHz*1e6) + rxnode.RxNoiseFigureDbm
	var link LinkMetric

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
		link.MaxTxAg = -1000.0
		link.AssoTxAg = -1000.0
		link.MaxRxAg = -1000.0
		link.AssoRxAg = -1000.0
		link.FreqInGHz = systemFrequencyGHz
		link.RxNodeID = rxid
		link.BestRSRP = -1000
		link.RoIDbm = -1000
		link.N0 = N0
		link.BandwidthMHz = BandwidthMHz
		var rxdebugnode = false
		link.TxNodeIDs.Resize(0)
		nlinks := 0

		var beamrsrp = make(map[int]*vlib.VectorF)
		for _, val := range alltxNodeIds {
			txnodeID := val
			txnode := singlecell.Nodes[val]
			if found := txnode.FreqGHz.Contains(systemFrequencyGHz); found {
				nlinks++
				link.TxNodeIDs.AppendAtEnd(txnodeID)
				//ant := afn(txnodeID)
				var rxant antenna.SettingAAS = afn(rxid)
				var txant antenna.SettingAAS = afn(txnodeID)
				var lossDb float64
				var dist float64
				var d2In float64 = 0
				var otherLossDb float64 = 0
				var islos bool
				extraloss := 0.0
				inloss := 0.0
				carloss := 0.0
				lossDb = DEFAULTERR_PL
				rxRSRP := -DEFAULTERR_PL
				var plerr error

				if model.IsSupported(systemFrequencyGHz) && txnode.Active {
					dist = txnode.Location.Distance2DFrom(rxnode.Location)
					if rxnode.Indoor && model.Env() == "RMa" {
						d2In = rand.Float64() * 10.0

					} else if rxnode.Indoor && model.Env() == "UMa" {

						d2In = rand.Float64() * 25.0
					}

					lossDb, islos, plerr = model.PLbetweenIndoor(txnode.Location, rxnode.Location, d2In)
					if rxnode.Indoor {
						inloss = model.O2ILossDb(systemFrequencyGHz, d2In)
						otherLossDb += inloss
					}
					if rxnode.InCar {
						carloss += model.O2ICarLossDb()
						otherLossDb += carloss
					}

					if w.OtherLossFn != nil {
						extraloss = w.OtherLossFn(model, txnode, rxnode, islos)
						otherLossDb += extraloss
					}

					if plerr != nil {
						log.Infof("EvaluateMetricV3 : (%d,%d) %v > %v", txnode.ID, rxnode.ID, lossDb, plerr)
						lossDb = DEFAULTERR_PL
					}
				} else {
					if !model.IsSupported(systemFrequencyGHz) {
						log.Fatalf("The Current Path loss Model %#v Doest not support Frequency %vGHz", model, systemFrequencyGHz)
					}
				}

				if txnode.Active {
					d3d, az, el := vlib.RelativeGeo(txnode.Location, rxnode.Location)
					el = -el + 90.0
					GCSaz := az - txnode.Direction
					if GCSaz > 180 {
						GCSaz = -360 + GCSaz
					}
					if GCSaz < -180 {
						GCSaz = 360 + GCSaz
					}
					if math.Abs(GCSaz) > 180 {
						fmt.Println("Error in Orientation", GCSaz)
					}
					GCSel := el
					var BSbestBeamID int
					var BSaasgainDB float64
					var UEaasgainDB float64

					if beamrsrp[txnodeID] == nil {
						beamrsrp[txnodeID] = new(vlib.VectorF)
					}

 					_, Rxaz, Rxel := vlib.RelativeGeo(rxnode.Location, txnode.Location)
					Rxel = -Rxel + 90.0
					RxGCSaz := Rxaz - rxnode.Direction
					if RxGCSaz > 180 {
						RxGCSaz = -360 + RxGCSaz
					}
					if RxGCSaz < -180 {
						RxGCSaz = 360 + RxGCSaz
					}
					if math.Abs(RxGCSaz) > 180 {
						fmt.Println("Error in Orientation", RxGCSaz)
					}
					RxGCSel := Rxel

					BSaasBeamgainDB, BSbestPanelID, BSbestBeamID, _, _ := txant.CombPatternDb(GCSaz, GCSel)

					if rxant.Omni == false {
						UEaasBeamgainDB, UEbestPanelID, UEbestBeamID, _, _ := rxant.CombPatternDb(RxGCSaz, RxGCSel)
						UEaasgainDB = UEaasBeamgainDB[UEbestPanelID][UEbestBeamID][0][0]
						// config.PrintStructsPretty(UEaasBeamgainDB)
						// fmt.Println("MaxPanelID: ", UEbestPanelID, "MaxBeamID: ", UEbestBeamID, "Max AAS Gain: ", UEaasgainDB)
						if UEaasgainDB > link.MaxRxAg {
							link.MaxRxAg = UEaasgainDB
						}
					} else {
						UEaasgainDB = 0.0
					}

					for _, valpanel := range BSaasBeamgainDB {
						for beamID, valbeam := range valpanel {
							tempRSRP := UEaasgainDB + valbeam[0][0] - lossDb - otherLossDb + txnode.TxPowerDBm
							if valbeam[0][0] > link.MaxTxAg {
								link.MaxTxAg = valbeam[0][0]
								link.MaxTransmitBeamID = beamID
							}
							beamrsrp[txnodeID].AppendAtEnd(tempRSRP)
						}
					}
					BSaasgainDB = BSaasBeamgainDB[BSbestPanelID][BSbestBeamID][0][0] // Picking gain from TxRU o,o assuming all TxRUs have same gain/ all beams
					// config.PrintStructsPretty(BSaasBeamgainDB)
					// fmt.Println("===================  Rxid:", rxnode.ID, "Txid: ", txnodeID, BSaasBeamgainDB[0][0].Get(0, 0))
					// fmt.Println("MaxPanelID: ", BSbestPanelID, "MaxBeamID: ", BSbestBeamID, "Max AAS Gain: ", BSaasgainDB)

					rxRSRP = UEaasgainDB + BSaasgainDB - lossDb - otherLossDb + txnode.TxPowerDBm

					if rxRSRP > vlib.Max(link.TxNodesRSRP) || link.AssoTxAg == -1000.0 {
						link.AssoTxAg = BSaasgainDB
 						link.AssoRxAg = UEaasgainDB
					}
					link.TxNodesRSRP.AppendAtEnd(rxRSRP)

					if LOG {
						fmt.Fprintf(fid, "\n%d,%d,%5.2f,%f,%t,%f,%f,%f,%f,%f,%f,%f,%f,%f,%f,%f,%f", rxid, txnodeID, d3d, rxnode.Location.Z, islos, rxRSRP-txnode.TxPowerDBm, lossDb, inloss, carloss, extraloss, txnode.TxPowerDBm, BSaasgainDB, UEaasgainDB, GCSaz, GCSel, RxGCSaz, RxGCSel)
					}

					rxdebugnode = false
					if rxdebugnode {
						_ = dist
						fmt.Printf("\r EVAL2 Rx-Tx (LOS:%v) %d-%d rxRSRP =%v,Power=%f,AAS =%f ,PL = %f, otherLoss=%f , dist =%v, d2In: =%v", islos, rxid, txnodeID, rxRSRP, txnode.TxPowerDBm, BSaasgainDB, lossDb, otherLossDb, dist, d2In)
						if rxnode.Indoor || rxnode.InCar {
							fmt.Println("\n Found in Indoor ", d2In, inloss, extraloss)
						}

					}
				}

				if math.IsInf(rxRSRP, 0) {
					log.Panicln("============= %%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%% %")
				}

				// link.TxNodesRSRP.AppendAtEnd(rxRSRP)

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
			sortedRxrp, indx := link.TxNodesRSRP.Sorted2()
			link.TxNodeIDs = link.TxNodeIDs.At(indx.Flip()...) // Sort it
			link.TxNodesRSRP = sortedRxrp.Flip()
			link.BestRSRP = link.TxNodesRSRP[0]
			// link.BestRSRPNode = singlecell.Nodes[link.TxNodeIDs[0]].Alias() + int(float64(len(alltxNodeIds)/3)*math.Floor(float64(link.TxNodeIDs[0]/(len(alltxNodeIds)/3))))
			link.BestRSRPNode = link.TxNodeIDs[0]
			link.BestCouplingLoss = link.BestRSRP - singlecell.Nodes[link.BestRSRPNode].TxPowerDBm
			if totalrssi == maxrsrp {
				link.BestSINR = vlib.Db(maxrsrp)
				if link.BestSINR > 200 {
					link.BestSINR = 1000
				}
			} else {
				rssi := 0.0
				interferenceBeam := 0
				for i := 1; i <= (nlinks - 1); i++ {
					beamrsrpLinr := vlib.InvDbF(*beamrsrp[link.TxNodeIDs[i]])
					interferenceBeam = rand.Intn(beamrsrpLinr.Len())
					rssi = rssi + beamrsrpLinr[interferenceBeam]
				}
				rssi = rssi + vlib.InvDb(link.N0)
				link.RSSI = rssi + maxrsrp
				link.BestSINR = vlib.Db(maxrsrp) - vlib.Db(rssi) //vlib.Db(totalrssi-maxrsrp)
			}
		}
	}

	return link
}
