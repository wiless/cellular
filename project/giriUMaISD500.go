package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	cell "github.com/wiless/cellular"

	"github.com/grd/statistics"
	"github.com/wiless/cellular/antenna"
	"github.com/wiless/cellular/deployment"
	"github.com/wiless/cellular/pathloss"
	"github.com/wiless/vlib"
)

var matlab *vlib.Matlab

var templateAAS *antenna.SettingAAS

type Point struct {
	X, Y float64
}

type LinkInfo struct {
	RxID              int
	NodeTypes         []string
	LinkGain          vlib.VectorF
	LinkGainNode      vlib.VectorI
	InterferenceLinks vlib.VectorF
}

var angles vlib.VectorF = vlib.VectorF{45, -45, -135, -45}
var singlecell deployment.DropSystem

type Winner struct {
	wsettings pathloss.ModelSetting
}

func (w *Winner) Set(pathloss.ModelSetting) {

}
func (w Winner) Get() pathloss.ModelSetting {
	return pathloss.ModelSetting{}
}
func (w Winner) LossInDbNodes(txnode, rxnode deployment.Node) (plDb float64, valid bool) {
	return 0, true
}
func (w Winner) LossInDb3D(txnode, rxnode vlib.Location3D) (plDb float64, valid bool) {
	return 0, true
}

func main() {
	matlab = vlib.NewMatlab("deployment")
	matlab.Silent = true
	matlab.Json = true

	seedvalue := time.Now().Unix()
	/// Setting to fixed seed
	seedvalue = 0
	rand.Seed(seedvalue)

	templateAAS = antenna.NewAAS()
	templateAAS.SetDefault()

	// modelsett:=pathloss.NewModelSettingi()

	var mymodel Winner

	// mymodel.ModelSetting.SetDefault()
	// mymodel.ModelSetting.Param[0] = 2
	DeployLayer1(&singlecell)

	singlecell.SetAllNodeProperty("BS", "AntennaType", 0)
	singlecell.SetAllNodeProperty("UE", "AntennaType", 1) /// Set All Pico to use antenna Type 1

	singlecell.SetAllNodeProperty("BS", "FreqGHz", vlib.VectorF{0.4, 0.85, 1.8}) /// Set All Pico to use antenna Type 0
	singlecell.SetAllNodeProperty("UE", "FreqGHz", vlib.VectorF{0.4, 0.85, 1.8}) /// Set All Pico to use antenna Type 0

	// lininfo := CalculatePathLoss(&singlecell, &model)

	rxids := singlecell.GetNodeIDs("UE")
	type MFNMetric []cell.LinkMetric
	MetricPerRx := make(map[int]MFNMetric)
	var AllMetrics MFNMetric
	wsystem := cell.NewWSystem()
	wsystem.BandwidthMHz = 10
	MaxCarriers := 1
	for _, rxid := range rxids {
		metrics := wsystem.EvaluteMetric(&singlecell, &mymodel, rxid, myfunc)
		if len(metrics) > 1 {
			log.Printf("%s[%d] Supports %d Carriers", "UE", rxid, len(metrics))
			MaxCarriers = int(math.Max(float64(MaxCarriers), float64(len(metrics))))
			// log.Printf("%s[%d] Links %#v ", "UE", rxid, metrics)
		}
		AllMetrics = append(AllMetrics, metrics...)
		MetricPerRx[rxid] = metrics
	}
	// vlib.SaveMapStructure2(MetricPerRx, "linkmetric.json", "UE", "LinkMetric", true)
	vlib.SaveStructure(AllMetrics, "linkmetric2.json", true)
	//Generate SINR values for CDF
	SINR := make(map[float64]vlib.VectorF)

	for _, metric := range MetricPerRx {
		for f := 0; f < len(metric); f++ {

			temp := SINR[metric[f].FreqInGHz]
			temp.AppendAtEnd(metric[f].BestSINR)
			SINR[metric[f].FreqInGHz] = temp
		}
	}
	cnt := 0
	for f, sinr := range SINR {
		log.Printf("\n F%d=%f \nSINR_%d= %v", cnt, f, cnt, sinr)
		cnt++
	}
	fmt.Println("\n")
	matlab.Close()
	fmt.Println("\n")
}

/// Calculate Pathloss

func CalculatePathLoss(singlecell *deployment.DropSystem, model *pathloss.SimplePLModel) []LinkInfo {

	txNodeNames := singlecell.GetTxNodeNames()
	txNodeNames = []string{"BS"} /// do only for BS

	rxNodeNames := singlecell.GetRxNodeNames()
	log.Println(txNodeNames, rxNodeNames)

	// rxlocs := singlecell.Locations("UE")
	rxlocs3D := singlecell.Locations3D("UE")
	RxLinkInfo := make([]LinkInfo, len(rxlocs3D))

	/// Generate Shadow Grid

	fmt.Printf("SETTING %s", singlecell.CoverageRegion.Celltype)

	// rows:=20
	// cols:=20
	// shwGrid := vlib.NewMatrixF(rows, cols)
	// for i := 0; i < len(rxlocs3D); i++ {
	// 	rxlocation := rxlocs3D[i]
	// 	var info LinkInfo
	// 	info.RxID = i
	// }

	var pathLossPerRxNode map[int]vlib.VectorF
	pathLossPerRxNode = make(map[int]vlib.VectorF)
	log.Println(pathLossPerRxNode)
	for i := 0; i < len(rxlocs3D); i++ {
		rxlocation := rxlocs3D[i]
		var info LinkInfo
		info.RxID = i

		func(rxlocation vlib.Location3D, txNodeNames []string) {
			info.NodeTypes = make([]string, len(txNodeNames))
			info.LinkGain = vlib.NewVectorF(len(txNodeNames))
			info.LinkGainNode = vlib.NewVectorI(len(txNodeNames))
			info.InterferenceLinks = vlib.NewVectorF(len(txNodeNames))
			for indx, name := range txNodeNames {
				txlocs := singlecell.Locations(name)
				txLocs3D := singlecell.Locations3D(name)

				allpathlossPerTxType := vlib.NewVectorF((txlocs.Size()))

				info.NodeTypes[indx] = name
				N := txlocs.Size()

				for k := 0; k < N; k++ {
					// angle := float64((k) * 360 / N)
					if name == "BS" {
						// templateAAS.HTiltAngle = 0 //angles[k]
						templateAAS.Omni = false
					} else {
						// templateAAS.HTiltAngle = 0
						templateAAS.Omni = true
					}

					templateAAS.CreateElements(txLocs3D[k])
					distance, _, _ := vlib.RelativeGeo(txLocs3D[k], rxlocation)
					lossDb := model.LossInDb(distance)
					aasgain, _, _ := templateAAS.AASGain(rxlocation) /// linear scale
					totalGainDb := vlib.Db(aasgain) - lossDb
					allpathlossPerTxType[k] = totalGainDb

					// fmt.Printf("\n Distance %f : loss %f dB", distance, lossDb)
					// matlab.Export(matstr, data)
				}
				data := statistics.Float64(allpathlossPerTxType)
				info.LinkGain[indx], info.LinkGainNode[indx] = statistics.Max(&data) // dB
				info.InterferenceLinks[indx] = vlib.Db(vlib.Sum(vlib.InvDbF(allpathlossPerTxType)) - vlib.InvDb(info.LinkGain[indx]))

			}

		}(rxlocation, txNodeNames)
		RxLinkInfo[i] = info
		// fmt.Printf("\n Info[%d] : %#v", i, info)
	}
	return RxLinkInfo
}

func DeployLayer1(system *deployment.DropSystem) {
	setting := system.GetSetting()
	if setting == nil {
		setting = deployment.NewDropSetting()
	}

	CellRadius := 200.0
	AreaRadius := CellRadius
	setting.SetCoverage(deployment.CircularCoverage(AreaRadius))
	setting.AddNodeType(deployment.NodeType{Name: "BS", Hmin: 40.0, Hmax: 40.0, Count: 7})
	setting.AddNodeType(deployment.NodeType{Name: "UE", Hmin: 1.1, Hmax: 10.0, Count: 30 * 7})

	// setting.AddNodeType(waptype)
	/// You can save the settings of this deployment by uncommenting this line
	system.SetSetting(setting)
	system.Init()

	setting.SetTxNodeNames("BS")
	setting.SetRxNodeNames("UE")
	/// Drop BS Nodes
	{
		locations := deployment.HexGrid(system.NodeCount("BS"), vlib.FromCmplx(deployment.ORIGIN), 100, 30)
		system.SetAllNodeLocation("BS", vlib.Location3DtoVecC(locations))
		// system.DropNodeType("BS")
		// find UE locations
		var uelocations vlib.VectorC
		for indx, bsloc := range locations {
			log.Printf("Deployed for cell %d ", indx)
			ulocation := deployment.HexRandU(bsloc.Cmplx(), 100, 30, 30)
			uelocations = append(uelocations, ulocation...)
		}
		system.SetAllNodeLocation("UE", uelocations)
	}

	vlib.SaveStructure(&system, "giridep.json", true)

}

func myfunc(nodeID int) antenna.SettingAAS {
	// atype := singlecell.Nodes[txnodeID]
	/// all nodeid same antenna
	return *templateAAS
}
