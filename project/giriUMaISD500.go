package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/wiless/cellular/pathloss"

	cell "github.com/wiless/cellular"

	"github.com/wiless/cellular/antenna"
	"github.com/wiless/cellular/deployment"
	"github.com/wiless/vlib"
)

var matlab *vlib.Matlab
var defaultAAS antenna.SettingAAS
var templateAAS []antenna.SettingAAS

var singlecell deployment.DropSystem
var secangles = vlib.VectorF{0.0, 120.0, -120.0}
var nSectors = 3
var CellRadius = 250.0
var nUEPerCell = 200
var nCells = 7
var CarriersGHz = vlib.VectorF{1.8}

func init() {

	defaultAAS.SetDefault()
	defaultAAS.N = 1
	defaultAAS.FreqHz = CarriersGHz[0]
	defaultAAS.BeamTilt = 0
	defaultAAS.DisableBeamTit = false
	defaultAAS.VTiltAngle = 30
	defaultAAS.ESpacingVFactor = .5
	defaultAAS.HTiltAngle = 0
	defaultAAS.MfileName = "output.m"
	defaultAAS.Omni = false

	defaultAAS.HoldOn = false
	defaultAAS.AASArrayType = antenna.LinearPhaseArray
	defaultAAS.CurveWidthInDegree = 30.0
	defaultAAS.CurveRadius = 1.00
}

func main() {
	matlab = vlib.NewMatlab("deployment")
	matlab.Silent = true
	matlab.Json = true

	seedvalue := time.Now().Unix()
	/// comment the below line to have different seed everytime
	seedvalue = 0
	rand.Seed(seedvalue)

	var hatamodel pathloss.OkumuraHata
	DeployLayer1(&singlecell)

	singlecell.SetAllNodeProperty("BS", "AntennaType", 0)
	singlecell.SetAllNodeProperty("UE", "AntennaType", 1) /// Set All Pico to use antenna Type 1

	singlecell.SetAllNodeProperty("BS", "FreqGHz", CarriersGHz) /// Set All Pico to use antenna Type 0
	singlecell.SetAllNodeProperty("UE", "FreqGHz", CarriersGHz) /// Set All Pico to use antenna Type 0

	rxids := singlecell.GetNodeIDs("UE")
	type MFNMetric []cell.LinkMetric
	MetricPerRx := make(map[int]MFNMetric)
	var AllMetrics MFNMetric
	wsystem := cell.NewWSystem()
	wsystem.BandwidthMHz = 10
	MaxCarriers := 1
	for _, rxid := range rxids {
		metrics := wsystem.EvaluteMetric(&singlecell, &hatamodel, rxid, myfunc)
		if len(metrics) > 1 {
			// log.Printf("%s[%d] Supports %d Carriers", "UE", rxid, len(metrics))
			MaxCarriers = int(math.Max(float64(MaxCarriers), float64(len(metrics))))
			// log.Printf("%s[%d] Links %#v ", "UE", rxid, metrics)
		}
		AllMetrics = append(AllMetrics, metrics...)
		MetricPerRx[rxid] = metrics
	}
	// vlib.SaveMapStructure2(MetricPerRx, "linkmetric.json", "UE", "LinkMetric", true)
	// vlib.SaveStructure(AllMetrics, "linkmetric2.json", true)

	//Generate SINR values for CDF
	SINR := make(map[float64]vlib.VectorF)
	log.Println("Total Freqs", SINR)
	counter := 0
	for _, metric := range MetricPerRx {

		for f := 0; f < len(metric); f++ {

			temp := SINR[metric[f].FreqInGHz]
			// temp.AppendAtEnd(metric[f].BestRSRP - (metric[f].N0))
			temp.AppendAtEnd(metric[f].BestSINR)

			SINR[metric[f].FreqInGHz] = temp
			// if counter < 10 {
			// 	fmt.Printf("\nrxid=%d indx %d Freq %f Value %v, %f", rxid, f, metric[f].FreqInGHz, metric[f].BestSINR, SINR[metric[f].FreqInGHz])
			// }
		}
		counter++
	}
	matlab.Close()
	cnt := 0
	matlab = vlib.NewMatlab("sinrVal.m")
	legendstring := ""
	for f, sinr := range SINR {
		log.Printf("\n F%d=%f \nSINR%d= %v", cnt, f, cnt, len(sinr))
		str := fmt.Sprintf("sinr%d", int(f*1000))
		// str = strings.Replace(str, ".", "", -1)
		matlab.Export(str, sinr)
		matlab.Command("cdfplot(" + str + ");hold all;")
		legendstring += str + " "
		cnt++
	}

	matlab.Command(fmt.Sprintf("legend %v", legendstring))
	matlab.Close()
	fmt.Println("\n")
}

/// Calculate Pathloss

func DeployLayer1(system *deployment.DropSystem) {
	setting := system.GetSetting()
	if setting == nil {
		setting = deployment.NewDropSetting()
	}

	AreaRadius := CellRadius

	// templateAAS = antenna.NewAAS()
	// templateAAS.SetDefault()
	// templateAAS.Omni = true

	setting.SetCoverage(deployment.CircularCoverage(AreaRadius))
	setting.AddNodeType(deployment.NodeType{Name: "BS", Hmin: 30.0, Hmax: 30.0, Count: nCells * nSectors})
	setting.AddNodeType(deployment.NodeType{Name: "UE", Hmin: 1.1, Hmax: 1.1, Count: nUEPerCell * nCells})

	// setting.AddNodeType(waptype)
	/// You can save the settings of this deployment by uncommenting this line
	system.SetSetting(setting)
	system.Init()

	setting.SetTxNodeNames("BS")
	setting.SetRxNodeNames("UE")
	/// Drop BS Nodes
	bslocations := make([]vlib.Location3D, system.NodeCount("BS"))
	{

		clocations := deployment.HexGrid(nCells, vlib.FromCmplx(deployment.ORIGIN), CellRadius, 30)
		/// three nodes with single cell centere

		for i := 0; i < nCells; i++ {

			for k := 0; k < nSectors; k++ {
				bslocations[i*3+k] = clocations[i]
			}
		}

		system.SetAllNodeLocation("BS", vlib.Location3DtoVecC(bslocations)) /// UPDATE just the XY positions

		// system.DropNodeType("BS")
		// find UE locations
		var uelocations vlib.VectorC
		for indx, bsloc := range clocations {
			log.Printf("Deployed for cell %d ", indx)
			ulocation := deployment.HexRandU(bsloc.Cmplx(), CellRadius, nUEPerCell, 30)
			uelocations = append(uelocations, ulocation...)
		}
		system.SetAllNodeLocation("UE", uelocations)
	}

	/// Create Antennas for each BS-NODE

	bsids := system.GetNodeIDs("BS")
	templateAAS = make([]antenna.SettingAAS, system.NodeCount("BS"))
	// sectorBW := 360.0 / float64(nSectors)

	for i := 0; i < len(templateAAS); i++ {
		templateAAS[i] = *antenna.NewAAS()
		templateAAS[i] = defaultAAS
		templateAAS[i].FreqHz = CarriersGHz[0] * 1.e9
		// templateAAS[i].HBeamWidth = 65
		templateAAS[i].HTiltAngle = secangles[vlib.ModInt(i, 3)]
		if nSectors == 1 {
			templateAAS[i].Omni = true
		} else {
			templateAAS[i].Omni = false
		}
		templateAAS[i].CreateElements(system.Nodes[bsids[i]].Location)

		hgain := vlib.NewVectorF(360)
		cnt := 0
		cmd := `delta=pi/180;
phaseangle=0:delta:2*pi-delta;`
		matlab.Command(cmd)
		for d := 0; d < 360; d++ {
			hgain[cnt] = templateAAS[i].ElementDirectionHGain(float64(d))
			cnt++
		}

		matlab.Export("gain"+strconv.Itoa(i), hgain)

		cmd = fmt.Sprintf("polar(phaseangle,gain%d);hold all", i)
		matlab.Command(cmd)
	}
	vlib.SaveStructure(templateAAS, "antennaArray.json")
	vlib.SaveStructure(system, "dep.json", true)

}

func myfunc(nodeID int) antenna.SettingAAS {
	// atype := singlecell.Nodes[txnodeID]
	/// all nodeid same antenna

	// fmt.Printf("\nNode %d , sector %v", nodeID, vlib.ModInt(nodeID, 3))
	return templateAAS[nodeID]
}
