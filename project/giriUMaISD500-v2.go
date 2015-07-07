package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
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
var systemAntennas map[int]*antenna.SettingAAS

var singlecell deployment.DropSystem
var secangles = vlib.VectorF{0.0, 120.0, -120.0}

// var nSectors = 1
var CellRadius = 500.0
var nUEPerCell = 400
var nCells = 19
var CarriersGHz = vlib.VectorF{1.8}

func init() {

	defaultAAS.SetDefault()

	vlib.LoadStructure("omni.json", &defaultAAS)
	// vlib.LoadStructure("sector.json", &defaultAAS)

	// vlib.LoadStructure("omni.json", defaultAAS)

	// defaultAAS.N = 1
	// defaultAAS.FreqHz = CarriersGHz[0]
	// defaultAAS.BeamTilt = 0
	// defaultAAS.DisableBeamTit = false
	// defaultAAS.VTiltAngle = 30
	// defaultAAS.ESpacingVFactor = .5
	// defaultAAS.HTiltAngle = 0
	// defaultAAS.MfileName = "output.m"
	// defaultAAS.Omni = true
	// defaultAAS.GainDb = 10
	// defaultAAS.HoldOn = false
	// defaultAAS.AASArrayType = antenna.LinearPhaseArray
	// defaultAAS.CurveWidthInDegree = 30.0
	// defaultAAS.CurveRadius = 1.00

	// vlib.SaveStructure(defaultAAS, "defaultAAS.json", true)
}

func main() {
	matlab = vlib.NewMatlab("deployment")
	matlab.Silent = true
	matlab.Json = true

	seedvalue := time.Now().Unix()
	/// comment the below line to have different seed everytime
	// seedvalue = 0
	rand.Seed(seedvalue)

	var plmodel pathloss.OkumuraHata
	// var plmodel walfisch.WalfischIke
	// var plmodel pathloss.SimplePLModel

	DeployLayer1(&singlecell)

	// singlecell.SetAllNodeProperty("BS", "AntennaType", 0)
	// singlecell.SetAllNodeProperty("UE", "AntennaType", 1)
	singlecell.SetAllNodeProperty("UE", "FreqGHz", CarriersGHz)
	layerBS := []string{"BS0", "BS1", "BS2"}
	layer2BS := []string{"OBS0", "OBS1", "OBS2"}

	var bsids vlib.VectorI

	for indx, bs := range layerBS {
		singlecell.SetAllNodeProperty(bs, "FreqGHz", CarriersGHz)
		singlecell.SetAllNodeProperty(bs, "TxPowerDBm", 22.0)
		singlecell.SetAllNodeProperty(bs, "Direction", secangles[indx])
		newids := singlecell.GetNodeIDs(bs)
		bsids.AppendAtEnd(newids...)
		fmt.Printf("\n %s : %v", bs, newids)

	}

	for indx, bs := range layer2BS {
		singlecell.SetAllNodeProperty(bs, "FreqGHz", CarriersGHz)
		singlecell.SetAllNodeProperty(bs, "TxPowerDBm", 22.0)
		singlecell.SetAllNodeProperty(bs, "Direction", secangles[indx])
		newids := singlecell.GetNodeIDs(bs)
		bsids.AppendAtEnd(newids...)
		fmt.Printf("\n %s : %v", bs, newids)

	}

	CreateAntennas(singlecell, bsids)
	vlib.SaveStructure(systemAntennas, "antennaArray.json", true)
	vlib.SaveStructure(singlecell.GetSetting(), "dep.json", true)
	vlib.SaveStructure(singlecell.Nodes, "nodelist.json", true)

	/// System 1 @ 400MHz
	// Dump UE locations
	{

		fid, _ := os.Create("uelocations.dat")
		ueids := singlecell.GetNodeIDs("UE")
		fmt.Fprintf(fid, "%% ID\tX\tY\tZ")
		for _, id := range ueids {
			node := singlecell.Nodes[id]
			fmt.Fprintf(fid, "\n %d \t %f \t %f \t %f \t %f \t %f ", id, node.Location.X, node.Location.Y, node.Location.Z)

		}
		fid.Close()

	}

	// Dump bs nodelocations
	{
		fid, _ := os.Create("bslocations.dat")

		fmt.Fprintf(fid, "%% ID\tX\tY\tZ\tPower\tdirection")
		for _, id := range bsids {
			node := singlecell.Nodes[id]
			fmt.Fprintf(fid, "\n %d \t %f \t %f \t %f \t %f \t %f ", id, node.Location.X, node.Location.Y, node.Location.Z, node.TxPowerDBm, node.Direction)

		}
		fid.Close()

	}

	wsystem := cell.NewWSystem()
	wsystem.BandwidthMHz = 10
	wsystem.FrequencyGHz = CarriersGHz[0]

	rxids := singlecell.GetNodeIDs("UE")
	RxMetrics400 := make(map[int]cell.LinkMetric)

	baseCells := vlib.VectorI{0, 1, 2}
	baseCells = baseCells.Scale(nCells)

	// wsystem.ActiveCells.AppendAtEnd(baseCells.Add(4)...)
	//wsystem.ActiveCells.AppendAtEnd(baseCells.Add(1)...)
	//wsystem.ActiveCells.AppendAtEnd(baseCells.Add(4)...)

	// cell := 2
	// startid := 0 + nUEPerCell*(cell)
	// endid := nUEPerCell * (cell + 1)
	// cell0UE := rxids[startid:endid]
	// log.Printf("\n ************** UEs of Cell %d := %v", cell, cell0UE)
	for _, rxid := range rxids {
		metric := wsystem.EvaluteLinkMetric(&singlecell, &plmodel, rxid, myfunc)
		RxMetrics400[rxid] = metric

	}

	// Dump antenna nodelocations
	{
		fid, _ := os.Create("antennalocations.dat")

		fmt.Fprintf(fid, "%% ID\tX\tY\tZ\tHDirection\tHWidth")
		for _, id := range bsids {
			ant := myfunc(id)
			// if id%7 == 0 {
			// 	node.TxPowerDBm = 0
			// } else {
			// 	node.TxPowerDBm = 44
			// }
			fmt.Fprintf(fid, "\n %d \t %f \t %f \t %f \t %f \t %f ", id, ant.Centre.X, ant.Centre.Y, ant.Centre.Z, ant.HTiltAngle, ant.HBeamWidth)

		}
		fid.Close()

	}
	vlib.DumpMap2CSV("table400.dat", RxMetrics400)
	vlib.SaveStructure(RxMetrics400, "metric400MHz.json", true)

	// /// System 2 @ 1800MHz
	// RxMetrics1800 := make(map[int]cell.LinkMetric)
	// wsystem.BandwidthMHz = 10
	// wsystem.FrequencyGHz = 0.4
	// singlecell.SetAllNodeProperty("BS", "TxPowerDBm", 22.0)
	// for _, rxid := range rxids {
	// 	metric := wsystem.EvaluteLinkMetric(&singlecell, &hatamodel, rxid, myfunc)
	// 	RxMetrics1800[rxid] = metric
	// }
	// vlib.SaveStructure(RxMetrics1800, "metric1800MHz.json", true)
	// DumpMap2CSV("table1800.dat", RxMetrics1800)

	matlab.Close()

	// matlab = vlib.NewMatlab("sinrVal.m")
	// legendstring := ""
	// var SINR vlib.VectorF
	// for _, sinr := range SINR {

	// 	str := fmt.Sprintf("sinr%d", int(wsystem.FrequencyGHz*1000))
	// 	// str = strings.Replace(str, ".", "", -1)
	// 	matlab.Export(str, sinr)
	// 	matlab.Command("cdfplot(" + str + ");hold all;")
	// 	legendstring += str + " "

	// }
	// matlab.Export("TxPower", singlecell.GetNodeType("BS").TxPowerDBm)
	// matlab.Export("AntennaGainDb", defaultAAS.GainDb)
	// matlab.Command(fmt.Sprintf("legend %v", legendstring))
	// matlab.Close()
	fmt.Println("\n")
}

/// Calculate Pathloss

func DeployLayer1(system *deployment.DropSystem) {
	setting := system.GetSetting()

	if setting == nil {
		setting = deployment.NewDropSetting()

		GENERATE := true
		if GENERATE {
			// AreaRadius := CellRadius
			/// Should come from API
			// setting.SetCoverage(deployment.CircularCoverage(AreaRadius))

			/// NodeType should come from API calls
			newnodetype := deployment.NodeType{Name: "BS0", Hmin: 30.0, Hmax: 30.0, Count: nCells}
			newnodetype.Mode = deployment.TransmitOnly
			setting.AddNodeType(newnodetype)

			newnodetype = deployment.NodeType{Name: "BS1", Hmin: 30.0, Hmax: 30.0, Count: nCells}
			newnodetype.Mode = deployment.TransmitOnly
			setting.AddNodeType(newnodetype)

			newnodetype = deployment.NodeType{Name: "BS2", Hmin: 30.0, Hmax: 30.0, Count: nCells}
			newnodetype.Mode = deployment.TransmitOnly
			setting.AddNodeType(newnodetype)

			newnodetype = deployment.NodeType{Name: "OBS0", Hmin: 30.0, Hmax: 30.0, Count: nCells}
			newnodetype.Mode = deployment.TransmitOnly
			setting.AddNodeType(newnodetype)

			newnodetype = deployment.NodeType{Name: "OBS1", Hmin: 30.0, Hmax: 30.0, Count: nCells}
			newnodetype.Mode = deployment.TransmitOnly
			setting.AddNodeType(newnodetype)

			newnodetype = deployment.NodeType{Name: "OBS2", Hmin: 30.0, Hmax: 30.0, Count: nCells}
			newnodetype.Mode = deployment.TransmitOnly
			setting.AddNodeType(newnodetype)
			/// NodeType should come from API calls
			newnodetype = deployment.NodeType{Name: "UE", Hmin: 1.1, Hmax: 1.1, Count: nUEPerCell * nCells}
			newnodetype.Mode = deployment.ReceiveOnly
			setting.AddNodeType(newnodetype)

			// vlib.SaveStructure(setting, "depSettings.json", true)

		} else {
			vlib.LoadStructure("depSettings.json", setting)
			fmt.Printf("\n %#v", setting.NodeTypes)
		}
		system.SetSetting(setting)

	}

	system.Init()

	// Workaround else should come from API calls or Databases
	// bslocations := LoadBSLocations(system)
	// system.SetAllNodeLocation("BS", vlib.Location3DtoVecC(bslocations))

	// area := deployment.RectangularCoverage(600)
	// deployment.DropSetting.SetCoverage(area)

	clocations := deployment.HexGrid(nCells, vlib.Origin3D, CellRadius, 30)
	system.SetAllNodeLocation("BS0", vlib.Location3DtoVecC(clocations))
	system.SetAllNodeLocation("BS1", vlib.Location3DtoVecC(clocations))
	system.SetAllNodeLocation("BS2", vlib.Location3DtoVecC(clocations))

	var loc vlib.Location3D
	loc.X = 1500
	loc.Y = 1500
	loc.Z = 0
	clocations = deployment.HexGrid(nCells, loc, CellRadius, 30)
	system.SetAllNodeLocation("OBS0", vlib.Location3DtoVecC(clocations))
	system.SetAllNodeLocation("OBS1", vlib.Location3DtoVecC(clocations))
	system.SetAllNodeLocation("OBS2", vlib.Location3DtoVecC(clocations))

	// Workaround else should come from API calls or Databases
	uelocations := LoadUELocations(system)
	system.SetAllNodeLocation("UE", uelocations)

}

func LoadUELocations(system *deployment.DropSystem) vlib.VectorC {

	var uelocations vlib.VectorC
	hexCenters := deployment.HexGrid(nCells, vlib.FromCmplx(deployment.ORIGIN), CellRadius, 30)
	for indx, bsloc := range hexCenters {
		log.Printf("Deployed for cell %d ", indx)
		ulocation := deployment.HexRandU(bsloc.Cmplx(), CellRadius, nUEPerCell, 30)
		// for i, v := range ulocation {
		// 	ulocation[i] = v + bsloc.Cmplx()
		// }
		uelocations = append(uelocations, ulocation...)
	}
	return uelocations

}

func myfunc(nodeID int) antenna.SettingAAS {
	// atype := singlecell.Nodes[txnodeID]
	/// all nodeid same antenna
	obj, ok := systemAntennas[nodeID]
	if !ok {
		log.Printf("No antenna created !! for %d ", nodeID)
		return defaultAAS
	} else {

		// fmt.Printf("\nNode %d , Omni= %v, Dirction=%v and center is %v", nodeID, obj.Omni, obj.HTiltAngle, obj.Centre)
		return *obj
	}
}

func CreateAntennas(system deployment.DropSystem, bsids vlib.VectorI) {
	if systemAntennas == nil {
		systemAntennas = make(map[int]*antenna.SettingAAS)
	}

	// omni := antenna.NewAAS()
	// sector := antenna.NewAAS()

	// vlib.LoadStructure("omni.json", omni)
	// vlib.LoadStructure("sector.json", sector)

	for _, i := range bsids {
		systemAntennas[i] = antenna.NewAAS()
		vlib.LoadStructure("sector.json", systemAntennas[i])
		systemAntennas[i].FreqHz = CarriersGHz[0] * 1.e9
		// systemAntennas[i].HBeamWidth = 65

		systemAntennas[i].HTiltAngle = system.Nodes[i].Direction

		// if nSectors == 1 {
		// 	systemAntennas[i].Omni = true
		// } else {
		// 	systemAntennas[i].Omni = false
		// }
		systemAntennas[i].CreateElements(system.Nodes[i].Location)
		// fmt.Printf("\nType=%s , BSid=%d : System Antenna : %v", system.Nodes[i].Type, i, systemAntennas[i].Centre)

		hgain := vlib.NewVectorF(360)
		cnt := 0
		cmd := `delta=pi/180;
		phaseangle=0:delta:2*pi-delta;`
		matlab.Command(cmd)
		for d := 0; d < 360; d++ {
			hgain[cnt] = systemAntennas[i].ElementDirectionHGain(float64(d))
			// hgain[cnt] = systemAntennas[i].ElementEffectiveGain(thetaH, thetaV)
			cnt++
		}

		matlab.Export("gain"+strconv.Itoa(i), hgain)
		// fmt.Printf("\nBS %d, Antenna : %#v", i, systemAntennas[i])

		cmd = fmt.Sprintf("polar(phaseangle,gain%d);hold all", i)
		matlab.Command(cmd)
	}
}
