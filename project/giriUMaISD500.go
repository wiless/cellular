package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

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

func main() {
	matlab = vlib.NewMatlab("deployment")
	matlab.Silent = true
	matlab.Json = true

	seedvalue := time.Now().Unix()
	/// Setting to fixed seed
	seedvalue = 0
	rand.Seed(seedvalue)
	// fmt.Printf("The sample mean is %g", mean)

	var singlecell deployment.DropSystem

	// modelsett:=pathloss.NewModelSettingi()
	var model pathloss.PathLossModel
	model.ModelSetting.SetDefault()
	model.ModelSetting.Param[0] = 2
	DeployLayer1(&singlecell)

	lininfo := CalculatePathLoss(&singlecell, &model)

	vlib.SaveStructure(lininfo, "linkinfo.json", true)
	fmt.Println("\n")
	matlab.Close()
	fmt.Println("\n")
}

/// Calculate Pathloss

func CalculatePathLoss(singlecell *deployment.DropSystem, model *pathloss.PathLossModel) []LinkInfo {

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
						templateAAS.HTiltAngle = angles[k]
						templateAAS.Omni = false
					} else {
						templateAAS.HTiltAngle = 0
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

	CellRadius := 1500.0
	AreaRadius := CellRadius
	setting.SetCoverage(deployment.CircularCoverage(AreaRadius))
	setting.AddNodeType(deployment.NodeType{Name: "BS", Hmin: 25.0, Hmax: 25.0, Count: 7})
	setting.AddNodeType(deployment.NodeType{Name: "UE", Hmin: 0.0, Hmax: 1.5, Count: 30 * 7})

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
