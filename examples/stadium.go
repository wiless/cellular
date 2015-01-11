package main

import (
	"fmt"
	cell "github.com/wiless/cellular"
	"github.com/wiless/cellular/antenna"
	"github.com/wiless/cellular/deployment"
	"github.com/wiless/cellular/pathloss"
	"github.com/wiless/vlib"
	"log"
	"math"
	"math/rand"

	"time"
)

var matlab *vlib.Matlab

var defaultAntenna *antenna.SettingAAS
var systemAntennas map[int]antenna.SettingAAS

// Dimension
// Outer Diameter : 283.01887m = 141.50944
// Inner Diameter : 174.5283m = 87.26415

var angles vlib.VectorF = vlib.VectorF{45, -45, -135, -45}

func init() {
	CreateDefaultAntenna()
	matlab = vlib.NewMatlab("deployment")
	matlab.Silent = true
	matlab.Json = true
	rand.Seed(time.Now().Unix())

}

func main() {
	// fmt.Printf("The sample mean is %g", mean)

	var singlecell deployment.DropSystem

	// modelsett:=pathloss.NewModelSettingi()
	var model pathloss.PathLossModel
	model.ModelSetting.SetDefault()
	model.ModelSetting.Param[0] = 2

	// SingleCellDeploy(&singlecell)

	/// Save deployment
	// vlib.SaveStructure(&singlecell, "stadiumOut.json", true)
	// fb, err := singlecell.MarshalJSON()
	// fmt.Println(err, "\n", string(fb))
	// fmt.Printf("\nOLDSYSTEM = %v \n\n", system)
	// var newsystem deployment.0DropSystem
	vlib.LoadStructure("stadium.json", &singlecell)

	/// Generate Different Antenna Types for every Transmit Node
	systemAntennas = make(map[int]antenna.SettingAAS)

	vlib.LoadMapStructure("antennas.json", systemAntennas)

	knownAntennaTypes := vlib.GetIntKeys(systemAntennas)
	// fmt.Println("Keys = ", knownAntennaTypes)
	/// validate Antennatypes requested in nodes
	for key, val := range singlecell.Nodes {
		found, _ := vlib.Contains(knownAntennaTypes, val.AntennaType)
		if !found {
			log.Printf("Unknown AntennaType %d for %s[%d] \n Setting to Default 0", val.AntennaType, val.Type, val.ID)
			val.AntennaType = 0
			singlecell.Nodes[key] = val

		}

	}

	// vlib.SaveMapStructure(systemAntennas, "antennas.json", true)
	rxids := singlecell.GetNodeIDs("UE")
	type MFNMetric []cell.LinkMetric
	MetricPerRx := make(map[int]MFNMetric)
	var AllMetrics MFNMetric
	for _, rxid := range rxids {
		metrics := EvaluteMetric(&singlecell, &model, rxid)
		if len(metrics) > 1 {
			log.Printf("%s[%d] Supports %d Carriers", "UE", rxid, len(metrics))
			// log.Printf("%s[%d] Links %#v ", "UE", rxid, metrics)
		}
		AllMetrics = append(AllMetrics, metrics...)
		MetricPerRx[rxid] = metrics
	}
	vlib.SaveMapStructure2(MetricPerRx, "cell.LinkMetric.json", "UE", "cell.LinkMetric", true)
	vlib.SaveStructure(AllMetrics, "cell.LinkMetric2.json", true)

	// CreateChannelLinks()

	// ueLinkInfo := CalculatePathLoss(&singlecell, &model)

	// rssi := vlib.NewVectorF(len(ueLinkInfo))
	// for indx, val := range ueLinkInfo {

	// 	temp := vlib.InvDbF(val.LinkGain)
	// 	MaxSignal := vlib.Max(temp)

	// 	TotalInterference := (vlib.Sum(temp) - MaxSignal) + vlib.Sum(vlib.InvDbF(val.InterferenceLinks))
	// 	SIR := MaxSignal / TotalInterference
	// 	rssi[indx] = vlib.Db(SIR)
	// }
	// // matlab.Export("rssi", rssi)
	// matlab.Export("SIR", rssi)
	// matlab.ExportStruct("LinkInfo", ueLinkInfo)

	matlab.Close()
	fmt.Println("\n")
}

/// Calculate Pathloss

func EvaluteMetric(singlecell *deployment.DropSystem, model *pathloss.PathLossModel, rxid int) []cell.LinkMetric {
	BandwidthMHz := 20.0
	NoisePSDdBm := -173.9
	N0 := NoisePSDdBm + vlib.Db(BandwidthMHz*1e6)
	var PerFreqLink map[float64]cell.LinkMetric
	PerFreqLink = make(map[float64]cell.LinkMetric)
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
		var link cell.LinkMetric

		link.FreqInGHz = f
		link.RxNodeID = rxid
		link.BestRSRP = -1000
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
				antenna := systemAntennas[txnode.AntennaType]
				antenna.Freq = f * 1.0e9

				antenna.HTiltAngle, antenna.VTiltAngle = txnode.Orientation[0], txnode.Orientation[1]
				// fmt.Printf("\n For Rx(%d) %s [%d]. antenna = %v", info.RxID, name, txnids[k], antenna)
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
	result := make([]cell.LinkMetric, len(PerFreqLink))
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

// func CalculatePathLoss(singlecell *deployment.DropSystem, model *pathloss.PathLossModel) []LinkInfo {
// 	txNodeNames := singlecell.GetTxNodeNames()
// 	rxNodeNames := singlecell.GetRxNodeNames()

// 	// rxlocs := singlecell.Locations("UE")
// 	rxlocs3D := singlecell.Locations3D(rxNodeNames[0])
// 	RxLinkInfo := make([]LinkInfo, len(rxlocs3D))

// 	/// Generate Shadow Grid

// 	// fmt.Printf("SETTING %s", singlecell.CoverageRegion.Celltype)

// 	// shwGrid := vlib.NewMatrixF(rows, cols)
// 	// for i := 0; i < len(rxlocs3D); i++ {
// 	// 	rxlocation := rxlocs3D[i]
// 	// 	var info LinkInfo
// 	// 	info.RxID = i
// 	// }

// 	//	var pathLossPerRxNode map[int]vlib.VectorF
// 	//pathLossPerRxNode = make(map[int]vlib.VectorF)
// 	//log.Println(pathLossPerRxNode)
// 	for i := 0; i < len(rxlocs3D); i++ {
// 		rxlocation := rxlocs3D[i]
// 		var info LinkInfo

// 		func(rxlocation vlib.Location3D, txNodeNames []string) {
// 			info.NodeTypes = make([]string, len(txNodeNames))
// 			info.LinkGain = vlib.NewVectorF(len(txNodeNames))
// 			info.LinkGainNode = vlib.NewVectorI(len(txNodeNames))
// 			info.InterferenceLinks = vlib.NewVectorF(len(txNodeNames))

// 			for indx, name := range txNodeNames {
// 				txlocs := singlecell.Locations(name)
// 				txLocs3D := singlecell.Locations3D(name)

// 				allpathlossPerTxType := vlib.NewVectorF((txlocs.Size()))

// 				info.NodeTypes[indx] = name
// 				N := txlocs.Size()
// 				txnids := singlecell.GetNodeIDs(name)
// 				for k := 0; k < N; k++ {
// 					node := singlecell.Nodes[txnids[k]]
// 					aid := node.AntennaType
// 					// antenna:= systemAntennas[txn]
// 					// angle := float64((k) * 360 / N)

// 					antenna := systemAntennas[aid]
// 					antenna.HTiltAngle, antenna.VTiltAngle = node.Orientation[0], node.Orientation[1]
// 					// fmt.Printf("\n For Rx(%d) %s [%d]. antenna = %v", info.RxID, name, txnids[k], antenna)
// 					antenna.CreateElements(txLocs3D[k])
// 					distance, _, _ := vlib.RelativeGeo(txLocs3D[k], rxlocation)
// 					lossDb := model.LossInDb(distance)
// 					aasgain, _, _ := antenna.AASGain(rxlocation) /// linear scale
// 					totalGainDb := vlib.Db(aasgain) - lossDb
// 					allpathlossPerTxType[k] = totalGainDb

// 					// fmt.Printf("\n Distance %f : loss %f dB", distance, lossDb)
// 					// matlab.Export(matstr, data)
// 				}
// 				data := statistics.Float64(allpathlossPerTxType)
// 				info.LinkGain[indx], info.LinkGainNode[indx] = statistics.Max(&data) // dB
// 				info.InterferenceLinks[indx] = vlib.Db(vlib.Sum(vlib.InvDbF(allpathlossPerTxType)) - vlib.InvDb(info.LinkGain[indx]))

// 			}

// 		}(rxlocation, txNodeNames)
// 		RxLinkInfo[i] = info
// 		fmt.Printf("\n Info[%d] : %#v", i, info)
// 	}

// 	return RxLinkInfo
// }

func SingleCellDeploy(system *deployment.DropSystem) {

	setting := deployment.NewDropSetting()
	temp := deployment.NewDropSetting()
	temp.SetDefaults()

	CellRadius := 141.50944
	AreaRadius := CellRadius
	setting.SetCoverage(deployment.CircularCoverage(AreaRadius))

	StadiumInnerRadius := 87.26415
	StadiumOuterRadius := 141.50944

	/// Total PICO nodes required
	// deltaOffset := 20.0 // 20m
	OuterArea := math.Pi * StadiumOuterRadius * StadiumOuterRadius
	InnerArea := math.Pi * StadiumInnerRadius * StadiumInnerRadius
	MinDistance := 20.0 / 2
	PicoCount := int(math.Ceil((OuterArea - InnerArea) / (math.Pi * MinDistance * MinDistance)))
	PicoCount = 2
	log.Println("Total Nodes Per Ring", PicoCount)

	setting.AddNodeType(deployment.NodeType{Name: "UE", Hmin: 1.0, Hmax: 10.0, Count: 10})
	setting.AddNodeType(deployment.NodeType{Name: "PICO", Hmin: 20.0, Hmax: 25.0, Count: PicoCount})
	/// You can save the settings of this deployment by uncommenting this line

	setting.SetTxNodeNames("PICO")
	setting.SetRxNodeNames("UE")
	system.SetSetting(setting)
	system.Init()

	vlib.SaveStructure(setting, "nodetype.txt", true)

	// jerr, jbytes := system.MarshalJSON()
	// jbytes, jerr := json.Marshal(system)
	// fmt.Println("===============")
	// fmt.Print(jerr, jbytes)
	// fmt.Println("===============")
	// jbytes, jerr = json.Marshal(setting)
	// fmt.Println("===============")
	// fmt.Print(jerr, string(jbytes))
	// fmt.Println("===============")

	/// Drop UE Nodes
	{
		locations := deployment.AnnularRingPoints(deployment.ORIGIN, StadiumInnerRadius, StadiumOuterRadius, system.NodeCount("UE"))
		system.SetAllNodeLocation("UE", locations)
	}

	/// Drop PICO Nodes
	{
		var PICOlocations vlib.VectorC
		random := true
		if !random {
			// offset := 10
			radius := StadiumInnerRadius + 10.0
			for i := 0; i < PicoCount; {
				count := int(math.Floor(2.0 * math.Pi * radius / 20.0))

				locations := deployment.AnnularRingEqPoints(deployment.ORIGIN, radius, count)
				PICOlocations.AppendAtEnd(locations...)
				i += count
				// fmt.Printf("\n Total %d , Current %d : %v", i, count, PICOlocations)

				radius += 20.0
			}

		} else {
			PICOlocations = deployment.AnnularRingPoints(deployment.ORIGIN, StadiumInnerRadius, StadiumOuterRadius, PicoCount)
		}
		system.SetAllNodeLocation("PICO", PICOlocations)

	}

	matlab.Export("ue", system.Locations("UE"))
	matlab.Export("pico", system.Locations("PICO"))

	plotcmd := `hold off;
	plot(real(ue),imag(ue),'.');
	hold all;
	plot(real(pico),imag(pico),'k*');
	grid on;`

	matlab.Command(plotcmd)
	// 	looptxt := `for k=1:length(bs)
	// text(real(bs(k)),imag(bs(k)),'BS')
	// end`
	// 	matlab.Q(looptxt)

	/// Plot scatter
	scattercmd := `figure;C=colormap;
	deltaRssi=80/64;
	deltasize=80/14;
	S=floor((SIR+110)/deltasize);
cindx=floor(SIR/deltaRssi);
scatter3(real(ue),imag(ue),SIR,64,cindx,'filled');
colorbar;
view(2)
`
	matlab.Q(scattercmd)

}

func CreateDefaultAntenna() {
	defaultAntenna = antenna.NewAAS()
	defaultAntenna.SetDefault()
	defaultAntenna.N = 1
	defaultAntenna.BeamTilt = 0
	defaultAntenna.HTiltAngle = 0
	defaultAntenna.VTiltAngle = 0
	defaultAntenna.DisableBeamTit = true
	defaultAntenna.Omni = false
}
