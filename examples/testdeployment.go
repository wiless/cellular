package main

import (
	"fmt"
	"github.com/grd/statistics"
	"github.com/wiless/cellular/antenna"
	"github.com/wiless/cellular/deployment"
	"github.com/wiless/cellular/pathloss"
	"github.com/wiless/vlib"
	"log"
	"math/rand"
	"time"
)

var matlab *vlib.Matlab

var templateAAS *antenna.SettingAAS

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

	rand.Seed(time.Now().Unix())
	// fmt.Printf("The sample mean is %g", mean)

	var singlecell deployment.DropSystem

	// modelsett:=pathloss.NewModelSettingi()
	var model pathloss.PathLossModel
	model.ModelSetting.SetDefault()
	model.ModelSetting.Param[0] = 2
	SingleCellDeploy(&singlecell)

	templateAAS = antenna.NewAAS()
	templateAAS.SetDefault()
	templateAAS.N = 1
	templateAAS.BeamTilt = 0
	templateAAS.HTiltAngle = 45
	templateAAS.VTiltAngle = 0
	templateAAS.DisableBeamTit = true
	templateAAS.Omni = false
	ueLinkInfo := CalculatePathLoss(&singlecell, &model)
	rssi := vlib.NewVectorF(len(ueLinkInfo))
	for indx, val := range ueLinkInfo {

		temp := vlib.InvDbF(val.LinkGain)
		MaxSignal := vlib.Max(temp)

		TotalInterference := (vlib.Sum(temp) - MaxSignal) + vlib.Sum(vlib.InvDbF(val.InterferenceLinks))
		SIR := MaxSignal / TotalInterference
		rssi[indx] = vlib.Db(SIR)
	}
	// matlab.Export("rssi", rssi)
	matlab.Export("SIR", rssi)
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

func SingleCellDeploy(system *deployment.DropSystem) {

	setting := deployment.NewDropSetting()
	temp := deployment.NewDropSetting()
	temp.SetDefaults()

	CellRadius := 1500.0
	AreaRadius := CellRadius
	setting.SetCoverage(deployment.CircularCoverage(AreaRadius))

	WAPNodes := 10
	NCluster := 1
	ClusterSize := 1

	setting.AddNodeType(deployment.NodeType{Name: "BS", Hmin: 25.0, Hmax: 25.0, Count: 4})
	setting.AddNodeType(deployment.NodeType{Name: "UE", Hmin: 0.0, Hmax: 1.5, Count: 5500})
	setting.AddNodeType(deployment.NodeType{Name: "PICO", Hmin: 0.0, Hmax: 10.0, Count: NCluster * ClusterSize})
	waptype := deployment.NodeType{Name: "WAP", Hmin: 5, Hmax: 50, Count: WAPNodes}
	setting.AddNodeType(waptype)
	/// You can save the settings of this deployment by uncommenting this line
	system.SetSetting(setting)

	system.Init()
	setting.SetTxNodeNames("BS", "WAP", "PICO")
	setting.SetRxNodeNames("UE")
	vlib.SaveStructure(setting, "deployment.json", false)

	// newsetting := deployment.NewDropSetting()
	/// Drop UE Nodes
	{

		locations := deployment.AnnularRingEqPoints(deployment.ORIGIN, 700, system.NodeCount("BS"))
		system.SetAllNodeLocation("BS", locations)
		//system.DropNodeType("BS")
	}

	/// Drops the UE in the default coverage mode i.e. Circular Coverage as set above
	system.DropNodeType("UE")
	system.DropNodeType("WAP")
	// fmt.Printf("\nUE = %f", system.Locations("UE"))
	// fmt.Printf("\nBS = %f", system.Locations("BS"))

	/// Custom dropping of nodes
	var wlocation vlib.VectorC
	wappos := deployment.RandPointR(AreaRadius)

	/// Drop WAP nodes in a rectangular region centered at wappos - some random point in
	wlocation = deployment.RectangularEqPoints(wappos, 50, rand.Float64()*360, WAPNodes)

	/// Drop WAP nodes in an Annular region between radius 100 to 200m, from Origin
	wlocation = deployment.AnnularRingPoints(deployment.ORIGIN, 700, 1000, WAPNodes)

	/// Drop WAP nodes in Equally spaced in an Annular radius of  200m, from Origin
	// wlocation = deployment.AnnularRingEqPoints(deployment.ORIGIN, 1500, WAPNodes)

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
	// system.SetNodeLocation("BS", 0, complex(0, 0))
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
	S=floor((SIR+110)/deltasize);
cindx=floor(SIR/deltaRssi);
scatter3(real(ue),imag(ue),SIR,64,cindx,'filled');
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
