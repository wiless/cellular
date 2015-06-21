package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/wiless/cellular/pathloss"

	cell "github.com/wiless/cellular"

	"github.com/wiless/cellular/antenna"
	"github.com/wiless/cellular/deployment"
	"github.com/wiless/vlib"
)

var matlab *vlib.Matlab
var defaultAAS antenna.SettingAAS
var templateAAS map[int]*antenna.SettingAAS

var singlecell deployment.DropSystem
var secangles = vlib.VectorF{0.0, 120.0, -120.0}
var nSectors = 3
var CellRadius = 250.0
var nUEPerCell = 550
var nCells = 1
var CarriersGHz = vlib.VectorF{1.8}

func init() {

	defaultAAS.SetDefault()
	defaultAAS.N = 1
	defaultAAS.FreqHz = 3.4 // CarriersGHz[0]
	defaultAAS.BeamTilt = 0
	defaultAAS.DisableBeamTit = false
	defaultAAS.VTiltAngle = 30
	defaultAAS.ESpacingVFactor = .5
	defaultAAS.HTiltAngle = 0
	defaultAAS.MfileName = "output.m"
	defaultAAS.Omni = true
	defaultAAS.GainDb = 10
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
	// singlecell.SetAllNodeProperty("BS", "TxPower", vlib.InvDb(23))

	singlecell.SetAllNodeProperty("BS", "FreqGHz", CarriersGHz) /// Set All Pico to use antenna Type 0
	singlecell.SetAllNodeProperty("UE", "FreqGHz", CarriersGHz) /// Set All Pico to use antenna Type 0

	rxids := singlecell.GetNodeIDs("UE")

	AllRxMetrics := make(map[int]cell.LinkMetric)

	wsystem := cell.NewWSystem()
	wsystem.BandwidthMHz = 10
	wsystem.FrequencyGHz = 1.8

	for _, rxid := range rxids {
		metric := wsystem.EvaluteLinkMetric(&singlecell, &hatamodel, rxid, myfunc)

		// 	// log.Printf("%s[%d] Supports %d Carriers", "UE", rxid, len(metrics))
		// 	MaxCarriers = int(math.Max(float64(MaxCarriers), float64(len(metrics))))
		// 	// log.Printf("%s[%d] Links %#v ", "UE", rxid, metrics)
		// }
		// AllMetrics = append(AllMetrics, metrics...)
		AllRxMetrics[rxid] = metric
	}
	// vlib.SaveMapStructure2(MetricPerRx, "linkmetric.json", "UE", "LinkMetric", true)
	vlib.SaveStructure(AllRxMetrics, "linkmetric2.json", true)

	//Generate SINR values for CDF
	var SINR vlib.VectorF
	log.Println("Total Freqs", SINR)

	w, fer := os.Create("nodeinfo.dat")
	if fer != nil {
		log.Print("Error Creating CSV file ", fer)
	}
	cwr := csv.NewWriter(w)
	// var record [4]string
	cwr.Comma = '\t'
	w.WriteString("%NodeID\tFreqHz\tX\tY\tSINR\n")
	for _, metric := range AllRxMetrics {
		// temp.AppendAtEnd(metric[f].BestRSRP - (metric[f].N0))
		SINR.AppendAtEnd(metric.BestSINR)
		loc := singlecell.Nodes[metric.RxNodeID].Location
		record := strings.Split(fmt.Sprintf("%d\t%f\t%f\t%f\t%f", metric.RxNodeID, metric.FreqInGHz, loc.X, loc.Y, metric.BestSINR), "\t")

		cwr.Write(record)
		// if counter < 10 {
		// 	fmt.Printf("\nrxid=%d indx %d Freq %f Value %v, %f", rxid, f, metric[f].FreqInGHz, metric[f].BestSINR, SINR[metric[f].FreqInGHz])
		// }
	}

	cwr.Flush()
	w.Close()
	matlab.Close()

	matlab = vlib.NewMatlab("sinrVal.m")
	legendstring := ""
	for _, sinr := range SINR {

		str := fmt.Sprintf("sinr%d", int(wsystem.FrequencyGHz*1000))
		// str = strings.Replace(str, ".", "", -1)
		matlab.Export(str, sinr)
		matlab.Command("cdfplot(" + str + ");hold all;")
		legendstring += str + " "

	}
	matlab.Export("TxPower", singlecell.GetNodeType("BS").TxPowerDBm)
	matlab.Export("AntennaGainDb", defaultAAS.GainDb)
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

	setting.SetCoverage(deployment.CircularCoverage(AreaRadius))
	setting.AddNodeType(deployment.NodeType{Name: "BS", TxPowerDBm: 10, Hmin: 30.0, Hmax: 30.0, Count: nCells * nSectors})
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
	templateAAS = make(map[int]*antenna.SettingAAS)
	// sectorBW := 360.0 / float64(nSectors)
	for _, i := range bsids {
		temp := antenna.NewAAS()
		*temp = defaultAAS

		templateAAS[i] = temp
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
	fmt.Println("Allantnena", templateAAS)

	mtype := reflect.TypeOf(templateAAS)
	// fmt.Println(mtype.Key(), mtype.Elem(), mtype.Elem().Kind(), strings.TrimLeft(mtype.Elem().String(), "*"))

	vlib.SaveMapStructure2(templateAAS, "antennaArray.json", "nodeid", strings.TrimLeft(mtype.Elem().String(), "*"))
	vlib.SaveStructure(system, "dep.json", true)

}

func myfunc(nodeID int) antenna.SettingAAS {
	// atype := singlecell.Nodes[txnodeID]
	/// all nodeid same antenna
	obj, ok := templateAAS[nodeID]
	if !ok {
		log.Printf("No antenna created !! for %d ", nodeID)
		return defaultAAS
	} else {

		// fmt.Printf("\nNode %d , sector %v", nodeID, vlib.ModInt(nodeID, 3))
		return *obj
	}
}
