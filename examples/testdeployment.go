package main

import (
	"fmt"
	"github.com/grd/statistics"
	"log"
	"math/cmplx"
	"math/rand"
	"strconv"
	"time"
	"wiless/cellular/deployment"
	"wiless/cellular/pathloss"
	"wiless/vlib"
)

var matlab *vlib.Matlab

type LinkInfo struct {
	RxID           int
	NodeTypes      []string
	MinPathLos     vlib.VectorF
	MinPathLosNode vlib.VectorI
}

func SingleCellDeploy(system *deployment.DropSystem) {

	setting := deployment.NewDropSetting()
	temp := deployment.NewDropSetting()
	temp.SetDefaults()

	CellRadius := 100.0
	AreaRadius := CellRadius * 3.0
	setting.SetCoverage(deployment.CircularCoverage(AreaRadius))

	WAPNodes := 10
	NCluster := 5
	ClusterSize := 10

	setting.AddNodeType(deployment.NodeType{Name: "BS", Hmin: 20.0, Hmax: 20.0, Count: 1})
	setting.AddNodeType(deployment.NodeType{Name: "UE", Hmin: 0.0, Hmax: 10.0, Count: 150})
	setting.AddNodeType(deployment.NodeType{Name: "WAP", Hmin: 0.0, Hmax: 0.0, Count: WAPNodes})
	setting.AddNodeType(deployment.NodeType{Name: "PICO", Hmin: 0.0, Hmax: 0.0, Count: NCluster * ClusterSize})

	/// You can save the settings of this deployment by uncommenting this line
	// vlib.SaveStructure(setting, "nodetype.txt", true)

	system.SetSetting(setting)

	setting.SetTxNodeNames("BS", "WAP", "PICO")
	setting.SetRxNodeNames("UE")

	system.Init()

	/// Drop UE Nodes
	/// Drops the UE in the default coverage mode i.e. Circular Coverage as set above
	system.DropNodeType("UE")

	/// Custom dropping of nodes
	var wlocation vlib.VectorC
	wappos := deployment.RandPointR(AreaRadius)

	/// Drop WAP nodes in a rectangular region centered at wappos - some random point in
	wlocation = deployment.RectangularEqPoints(wappos, 50, rand.Float64()*360, WAPNodes)

	/// Drop WAP nodes in an Annular region between radius 100 to 200m, from Origin
	wlocation = deployment.AnnularRingPoints(deployment.ORIGIN, 100, 200, WAPNodes)

	/// Drop WAP nodes in Equally spaced in an Annular radius of  200m, from Origin
	wlocation = deployment.AnnularRingEqPoints(deployment.ORIGIN, 200, WAPNodes)

	/// Save the wappos variable in matlab
	matlab.Export("wappos", wappos)

	/// Add a Text marker on the figure after all operatus just before matlab.Close()
	for i := 0; i < len(wlocation); i++ {
		str := fmt.Sprintf("W%d", i)
		matlab.Q(matlab.AddText(wlocation[i], str))
	}

	// nids := system.GetNodeIDs("PICO")
	// fmt.Printf("PICO nodes ids ", nids)
	/// Droping ClusterSize picos in each Cluster , total cluster = NCluster

	for i := 0; i < NCluster; i++ {
		/// Centre point of each cluster is random
		clusterCentre := deployment.RandPointR(AreaRadius)

		/// Random points in Circular region of radius 20 , centred at clusterCentre
		plocation := deployment.CircularPoints(clusterCentre, 20, ClusterSize)

		/// Index of the PicoNodes 0...ClusterSize-1, and there on
		nodeIndexes := vlib.NewSegmentI(i*ClusterSize, ClusterSize)

		/// Set locations of nodes of types "PICO" with
		system.SetNodeLocationOf("PICO", nodeIndexes, plocation)

		/// Mark the Cluster name as C0,C1, etc
		cname := fmt.Sprintf("C%d", i)
		matlab.Q(matlab.AddText(clusterCentre, cname))
	}

	// matlab.Q(matlab.AddText(deployment.ORIGIN, "BS"))
	///
	system.SetNodeLocation("BS", 0, complex(0, 0))
	system.SetAllNodeLocation("WAP", wlocation)

	matlab.Export("bs", system.Locations("BS"))
	matlab.Export("ue", system.Locations("UE"))
	matlab.Export("wap", system.Locations("WAP"))
	matlab.Export("pico", system.Locations("PICO"))

	plotcmd := `hold off;
	plot(real(bs),imag(bs),'ro');
	hold all;
	plot(real(ue),imag(ue),'.');
	plot(real(wap),imag(wap),'m*');
	plot(real(pico),imag(pico),'ro');
	grid on;`

	matlab.Command(plotcmd)
	matlab.AddText(wappos, "WAPcentre")

	/// MOVING BS on HEX co-ords
	// {
	// 	origincords := deployment.HexagonalPoints(complex(0, 0), CellRadius)
	// 	// system.SetAllNodeLocation("BS", origincords)

	// 	var otherCords vlib.VectorC

	// 	for i := 0; i < len(origincords); i++ {
	// 		var centre complex128
	// 		if i == 5 {
	// 			centre = (origincords[i] + origincords[0])
	// 		} else {
	// 			centre = (origincords[i] + origincords[i+1])
	// 		}

	// 		otherCords = append(otherCords, deployment.HexagonalPoints(centre, CellRadius)...)
	// 	}
	// 	// matlab.Export("pos", otherCords)

	// }

}

func main() {
	matlab = vlib.NewMatlab("deployment")
	matlab.Silent = true
	matlab.Json = true

	rand.Seed(time.Now().Unix())
	// fmt.Printf("The sample mean is %g", mean)

	var singlecell deployment.DropSystem

	// modelsett:=pathloss.NewModelSettingi()
	var model pathloss.PathLossModel
	model.ModelSetting.SetDefault()

	SingleCellDeploy(&singlecell)

	CalculatePathLoss(&singlecell, &model)

	matlab.ExportStruct("nodeTypes", singlecell.GetTxNodeNames())
	// log.Println("Exporting Node :")
	// matlab.ExportStruct("nodeinfo", singlecell.Nodes)
	// matlab.ExportStruct("nodeinfo1", *singlecell.Nodes[1])
	for i := 0; i < len(singlecell.Nodes); i++ {
		matlab.ExportStruct("Nodes"+strconv.Itoa(i), *singlecell.Nodes[i])
	}

	matlab.Close()
	fmt.Println("\n")

}

/// Calculate Pathloss

func CalculatePathLoss(singlecell *deployment.DropSystem, model *pathloss.PathLossModel) {

	txNodeNames := singlecell.GetTxNodeNames()
	rxNodeNames := singlecell.GetRxNodeNames()
	log.Print(txNodeNames, rxNodeNames)
	rxlocs := singlecell.Locations("UE")

	RxLinkInfo := make([]LinkInfo, len(rxlocs))
	var pathLossPerRxNode map[int]vlib.VectorF
	pathLossPerRxNode = make(map[int]vlib.VectorF)
	log.Print(pathLossPerRxNode)
	for i := 0; i < rxlocs.Size(); i++ {
		rxlocation := rxlocs[i]
		var info LinkInfo
		info.RxID = i
		func(rxlocation complex128, txNodeNames []string) {

			info.NodeTypes = make([]string, len(txNodeNames))
			info.MinPathLos = vlib.NewVectorF(len(txNodeNames))
			info.MinPathLosNode = vlib.NewVectorI(len(txNodeNames))
			for indx, name := range txNodeNames {
				txlocs := singlecell.Locations(name)
				allpathlossPerTxType := vlib.NewVectorF((txlocs.Size()))

				info.NodeTypes[indx] = name

				for k := 0; k < txlocs.Size(); k++ {
					srcLocation := txlocs[k]
					distance := cmplx.Abs(rxlocation - srcLocation)
					lossDb := model.LossInDb(distance)
					// fmt.Printf("%d  %v  %v distance = %v", i, src, rxlocs[k], rxlocs[k]-src)
					// matstr := fmt.Sprintf("Distance(%d,%d)", rxnodeId+1, k+1)

					allpathlossPerTxType[k] = lossDb
					// fmt.Printf("\n Distance %f : loss %f dB", distance, lossDb)
					// matlab.Export(matstr, data)
				}
				data := statistics.Float64(allpathlossPerTxType)
				info.MinPathLos[indx], info.MinPathLosNode[indx] = statistics.Min(&data)

			}

		}(rxlocation, txNodeNames)
		RxLinkInfo[i] = info
		// fmt.Printf("\n Info[%d] : %#v", i, info)
	}
}
