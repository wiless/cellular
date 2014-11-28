package main

import (
	"fmt"
	"github.com/grd/statistics"
	"log"
	"math/rand"
	"time"
	"wiless/cellular/antenna"
	"wiless/cellular/deployment"
	"wiless/cellular/pathloss"
	"wiless/vlib"
)

var matlab *vlib.Matlab

var templateAAS *antenna.SettingAAS

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

	CellRadius := 1500.0
	AreaRadius := CellRadius * 3.0
	setting.SetCoverage(deployment.RectangularCoverage(AreaRadius))

	WAPNodes := 10
	NCluster := 5
	ClusterSize := 10

	setting.AddNodeType(deployment.NodeType{Name: "BS", Hmin: 20.0, Hmax: 20.0, Count: 1})
	setting.AddNodeType(deployment.NodeType{Name: "UE", Hmin: 10.0, Hmax: 10.0, Count: 2500})
	setting.AddNodeType(deployment.NodeType{Name: "WAP", Hmin: 0.0, Hmax: 0.0, Count: WAPNodes})
	setting.AddNodeType(deployment.NodeType{Name: "PICO", Hmin: 0.0, Hmax: 0.0, Count: NCluster * ClusterSize})
	setting.AddNodeType(deployment.NodeType{Name: "NOKIA", Hmin: 5, Hmax: 5, Count: 30})
	/// You can save the settings of this deployment by uncommenting this line
	system.SetSetting(setting)
	system.Init()
	setting.SetTxNodeNames("BS", "WAP", "PICO")
	setting.SetRxNodeNames("UE")
	vlib.SaveStructure(setting, "nodetype.txt", true)

	/// Drop UE Nodes
	system.DropNodeType("BS")
	/// Drops the UE in the default coverage mode i.e. Circular Coverage as set above
	system.DropNodeType("UE")
	// fmt.Printf("\nUE = %f", system.Locations("UE"))
	// fmt.Printf("\nBS = %f", system.Locations("BS"))

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
	plot(real(bs),imag(bs),'ro','MarkerFaceColor','red','MarkerSize',10);
	hold all;
	plot(real(ue),imag(ue),'.');
	plot(real(wap),imag(wap),'m*');
	plot(real(pico),imag(pico),'k*');
	grid on;`

	matlab.Command(plotcmd)
	looptxt := `for k=1:length(bs)
text(real(bs(k)),imag(bs(k)),'BS')    
end`
	matlab.Q(looptxt)

	/// Plot scatter
	scattercmd := `figure;C=colormap;
	deltaRssi=80/64;
	deltasize=80/14;
	S=floor((rssi+110)/deltasize);
cindx=floor(rssi/deltaRssi);
scatter3(real(ue),imag(ue),rssi,64,cindx,'filled');
colorbar;
view(2)
`
	matlab.Q(scattercmd)
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
	model.ModelSetting.Param[0] = 4
	SingleCellDeploy(&singlecell)

	templateAAS = antenna.NewAAS()
	templateAAS.SetDefault()
	templateAAS.N = 8
	templateAAS.BeamTilt = 0
	templateAAS.HTiltAngle = -35
	templateAAS.VTiltAngle = 0
	templateAAS.DisableBeamTit = false

	ueLinkInfo := CalculatePathLoss(&singlecell, &model)
	rssi := vlib.NewVectorF(len(ueLinkInfo))
	for indx, val := range ueLinkInfo {
		rssi[indx] = val.MinPathLos[0]
	}
	matlab.Export("rssi", rssi)
	matlab.ExportStruct("LinkInfo", ueLinkInfo)

	// matlab.ExportStruct("nodeTypes", singlecell.GetTxNodeNames())
	// log.Println("Exporting Node :")
	// matlab.ExportStruct("nodeinfo", singlecell.Nodes)
	// matlab.ExportStruct("nodeinfo1", *singlecell.Nodes[1])
	// for i := 0; i < len(singlecell.Nodes); i++ {
	// 	matlab.ExportStruct("Nodes"+strconv.Itoa(i), *singlecell.Nodes[i])
	// }

	matlab.Close()
	fmt.Println("\n")
}

/// Calculate Pathloss

func CalculatePathLoss(singlecell *deployment.DropSystem, model *pathloss.PathLossModel) []LinkInfo {

	txNodeNames := singlecell.GetTxNodeNames()
	txNodeNames = []string{"BS"}

	rxNodeNames := singlecell.GetRxNodeNames()
	log.Println(txNodeNames, rxNodeNames)

	// rxlocs := singlecell.Locations("UE")
	rxlocs3D := singlecell.Locations3D("UE")
	RxLinkInfo := make([]LinkInfo, len(rxlocs3D))

	var pathLossPerRxNode map[int]vlib.VectorF
	pathLossPerRxNode = make(map[int]vlib.VectorF)
	log.Println(pathLossPerRxNode)
	for i := 0; i < len(rxlocs3D); i++ {
		rxlocation := rxlocs3D[i]
		var info LinkInfo
		info.RxID = i

		func(rxlocation vlib.Location3D, txNodeNames []string) {
			info.NodeTypes = make([]string, len(txNodeNames))
			info.MinPathLos = vlib.NewVectorF(len(txNodeNames))
			info.MinPathLosNode = vlib.NewVectorI(len(txNodeNames))
			for indx, name := range txNodeNames {
				txlocs := singlecell.Locations(name)
				txLocs3D := singlecell.Locations3D(name)

				allpathlossPerTxType := vlib.NewVectorF((txlocs.Size()))

				info.NodeTypes[indx] = name

				for k := 0; k < txlocs.Size(); k++ {

					templateAAS.CreateElements(txLocs3D[k])

					// srcLocation := txlocs[k]
					distance, _, _ := vlib.RelativeGeo(txLocs3D[k], rxlocation)
					// distance := cmplx.Abs(rxlocation - srcLocation)
					lossDb := model.LossInDb(distance)
					aasgain, _, _ := templateAAS.AASGain(rxlocation)

					// f("%d  %v  %v distance = %v", i, src, rxlocs[k], rxlocs[k]-src)
					// matstr := fmt.Sprintf("Distance(%d,%d)", rxnodeId+1, k+1)
					totalGain := vlib.Db(aasgain) - lossDb
					// fmt.Printf("\n(%s,%d)PL, Gain = %f %f %f", name, k, lossDb, aasgain, totalGain)
					allpathlossPerTxType[k] = totalGain

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
	return RxLinkInfo
}
