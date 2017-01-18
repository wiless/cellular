package main

import (
	"fmt"
	"log"
	"math"
	"math/cmplx"
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

type MatInfo struct {
	BaseID             int `json:"baseID"`
	SecID              int `json:"secID"`
	UserID             int `json:"userID"`
	RSSI               float64
	IfStation          vlib.VectorI `json:"ifStation"`
	IfRSSI             vlib.VectorF `json:"ifRSSI"`
	ThermalNoise       float64      `json:"thermalNoise"`
	SINR               float64
	RestOfInterference float64 `json:"restOfInterference"`
}

var matlab *vlib.Matlab
var defaultAAS antenna.SettingAAS
var systemAntennas map[int]*antenna.SettingAAS

var singlecell deployment.DropSystem
var secangles = vlib.VectorF{0.0, 120.0, -120.0}

// var nUEPerCell = 1000
var nCells = 19
var trueCells = 1

var CellRadius float64 = 1000

var CarriersGHz = vlib.VectorF{0.7}
var RXTYPES = []string{"MUE"}
var VTILT = 0.0

var NVillages = 3
var VillageRadius = 400.0
var VillageDistance = 2500.0

var GPradius = 550.0
var GPusers = 0        //525
var NUEsPerVillage = 0 //125
var NMobileUEs = 1000  // 100

func init() {

	defaultAAS.SetDefault()

	//	vlib.LoadStructure("omni.json", &defaultAAS)
	vlib.LoadStructure("sector.json", &defaultAAS)

	// vlib.LoadStructure("omni.json", defaultAAS)

	// defaultAAS.N = 1
	defaultAAS.FreqHz = CarriersGHz[0]
	// defaultAAS.BeamTilt = 0
	// defaultAAS.DisableBeamTit = false
	defaultAAS.VTiltAngle = VTILT
	// defaultAAS.ESpacingVFactor = .5
	// defaultAAS.HTiltAngle = 0
	// defaultAAS.MfileName = "output.m"
	// defaultAAS.Omni = true
	// defaultAAS.GainDb = 10
	// defaultAAS.HoldOn = false
	// defaultAAS.AASArrayType = antenna.LinearPhaseArray
	// defaultAAS.CurveWidthInDegree = 30.0
	// defaultAAS.CurveRadius = 1.00

	vlib.SaveStructure(defaultAAS, "defaultAAS.json", true)
}

func main() {
	matlab = vlib.NewMatlab("deployment")
	matlab.Silent = true
	matlab.Json = false

	seedvalue := time.Now().Unix()
	/// comment the below line to have different seed everytime
	// seedvalue = 0
	rand.Seed(seedvalue)

	var plmodel pathloss.OkumuraHata
	// var plmodel walfisch.WalfischIke
	// var plmodel pathloss.SimplePLModel
	// var plmodel pathloss.RMa

	DeployLayer1(&singlecell)

	// singlecell.SetAllNodeProperty("BS", "AntennaType", 0)
	// singlecell.SetAllNodeProperty("UE", "AntennaType", 1)

	rxnodeTypes := singlecell.GetNodeTypesOfMode(deployment.ReceiveOnly)
	log.Println("Found these RX nodes ", rxnodeTypes)

	for _, uetype := range rxnodeTypes {
		singlecell.SetAllNodeProperty(uetype, "FreqGHz", CarriersGHz)
	}
	// CASE A1 & A2
	// CASE B1 & B2

	layerBS := []string{"BS0", "BS1", "BS2"}
	// layer2BS := []string{"OBS0", "OBS1", "OBS2"}

	var bsids vlib.VectorI

	for indx, bs := range layerBS {
		singlecell.SetAllNodeProperty(bs, "FreqGHz", CarriersGHz)
		singlecell.SetAllNodeProperty(bs, "TxPowerDBm", 43.0)
		singlecell.SetAllNodeProperty(bs, "Direction", secangles[indx])
		newids := singlecell.GetNodeIDs(bs)
		bsids.AppendAtEnd(newids...)
		fmt.Printf("\n %s : %v", bs, newids)

	}

	// for indx, bs := range layer2BS {
	// 	singlecell.SetAllNodeProperty(bs, "FreqGHz", CarriersGHz)
	// 	singlecell.SetAllNodeProperty(bs, "TxPowerDBm", 22.0)
	// 	singlecell.SetAllNodeProperty(bs, "Direction", secangles[indx])
	// 	newids := singlecell.GetNodeIDs(bs)
	// 	bsids.AppendAtEnd(newids...)
	// 	fmt.Printf("\n %s : %v", bs, newids)

	// }

	CreateAntennas(singlecell, bsids)
	vlib.SaveStructure(systemAntennas, "antennaArray.json", true)
	vlib.SaveStructure(singlecell.GetSetting(), "dep.json", true)
	vlib.SaveStructure(singlecell.Nodes, "nodelist.json", true)

	rxtypes := RXTYPES

	/// System 1 @ 400MHz
	// Dump UE locations
	{

		fid, _ := os.Create("uelocations.dat")
		ueids := singlecell.GetNodeIDs(rxtypes...)
		log.Println("RXid range : ", ueids[0], ueids[len(ueids)-1], len(ueids))
		fmt.Fprintf(fid, "%% ID\tX\tY\tZ")
		for _, id := range ueids {
			node := singlecell.Nodes[id]
			fmt.Fprintf(fid, "\n %d \t %f \t %f \t %f ", id, node.Location.X, node.Location.Y, node.Location.Z)

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
	wsystem.BandwidthMHz = 20
	wsystem.FrequencyGHz = CarriersGHz[0]

	rxids := singlecell.GetNodeIDs(rxtypes...)
	log.Println("RXid range : ", rxids[0], rxids[len(rxids)-1], len(rxids))
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

		fmt.Fprintf(fid, "%% ID\tX\tY\tZ\tHDirection\tHWidth\tVTilt")
		for _, id := range bsids {
			ant := myfunc(id)
			// if id%7 == 0 {
			// 	node.TxPowerDBm = 0
			// } else {
			// 	node.TxPowerDBm = 44
			// }
			fmt.Fprintf(fid, "\n %d \t %f \t %f \t %f \t %f \t %f \t %f", id, ant.Centre.X, ant.Centre.Y, ant.Centre.Z, ant.HTiltAngle, ant.HBeamWidth, ant.VTiltAngle)

		}
		fid.Close()

	}

	vlib.DumpMap2CSV("table700MHz.dat", RxMetrics400)
	vlib.SaveStructure(RxMetrics400, "metric700MHz.json", true)

	/// Code Dump for Throughput Calculation

	// log.Printf("\n ************** UEs of Cell %d := %v", cell, cell0UE)
	log.Println("RXID range ", rxids[0], rxids[len(rxids)-1])

	// BaseID             int `json:"baseID"`
	// SecID              int `json:"secID"`
	// UserID             int `json:"userID"`
	// RSSI               float64
	// IfStation          vlib.VectorI `json:"ifStation"`
	// IfRSSI             vlib.VectorF `json:"ifRSSI"`
	// ThermalNoise       float64      `json:"thermalNoise"`
	// SINR               float64
	// RestOfInterference float64 `json:"restOfInterference"`

	MatlabResult := make([]MatInfo, len(rxids))
	for indx, rxid := range rxids {
		metric := RxMetrics400[rxid]
		var minfo MatInfo
		minfo.UserID = metric.RxNodeID
		minfo.SecID = 0 // metric.BestRSRPNode
		minfo.BaseID = metric.BestRSRPNode
		minfo.RSSI = 0 // normalized

		MAXINTER := 8
		minfo.IfStation = metric.TxNodeIDs[1:MAXINTER] // the first entry is best
		var ifrssi vlib.VectorF
		ifrssi = metric.TxNodesRSRP[1:]
		ifrssi = ifrssi.Sub(metric.TxNodesRSRP[0])

		residual := ifrssi[MAXINTER:]
		residual = vlib.InvDbF(residual)
		ifrssi = ifrssi[0:MAXINTER]

		minfo.IfRSSI = ifrssi // the first entry is best
		minfo.ThermalNoise = metric.N0
		minfo.SINR = metric.BestSINR
		minfo.RestOfInterference = vlib.Db(vlib.Sum(residual))

		MatlabResult[indx] = minfo
	}
	vlib.SaveStructure(MatlabResult, "matlabdump.json", true)

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
	// _, ids := deployment.HexWrapGrid(nCells, vlib.Origin3D, CellRadius, 30, trueCells)
	// matlab.Export("VirtualCellID", ids)

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
	fmt.Println("\n ============================")

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

			BSHEIGHT := 35.0
			/// NodeType should come from API calls
			newnodetype := deployment.NodeType{Name: "BS0", Hmin: BSHEIGHT, Hmax: BSHEIGHT, Count: nCells}
			newnodetype.Mode = deployment.TransmitOnly
			setting.AddNodeType(newnodetype)

			newnodetype = deployment.NodeType{Name: "BS1", Hmin: BSHEIGHT, Hmax: BSHEIGHT, Count: nCells}
			newnodetype.Mode = deployment.TransmitOnly
			setting.AddNodeType(newnodetype)

			newnodetype = deployment.NodeType{Name: "BS2", Hmin: BSHEIGHT, Hmax: BSHEIGHT, Count: nCells}
			newnodetype.Mode = deployment.TransmitOnly
			setting.AddNodeType(newnodetype)

			// newnodetype = deployment.NodeType{Name: "OBS0", Hmin: 30.0, Hmax: 30.0, Count: nCells}
			// newnodetype.Mode = deployment.TransmitOnly
			// setting.AddNodeType(newnodetype)

			// newnodetype = deployment.NodeType{Name: "OBS1", Hmin: 30.0, Hmax: 30.0, Count: nCells}
			// newnodetype.Mode = deployment.TransmitOnly
			// setting.AddNodeType(newnodetype)

			// newnodetype = deployment.NodeType{Name: "OBS2", Hmin: 30.0, Hmax: 30.0, Count: nCells}
			// newnodetype.Mode = deployment.TransmitOnly
			// setting.AddNodeType(newnodetype)

			/// NodeType should come from API calls

			/// ALL NODES considered
			// newnodetype = deployment.NodeType{Name: "UE", Hmin: 1.1, Hmax: 1.1, Count: nUEPerCell * nCells}
			// newnodetype.Mode = deployment.ReceiveOnly
			// setting.AddNodeType(newnodetype)

			/// CASE A1 & A2
			// newnodetype = deployment.NodeType{Name: "UE", Hmin: 1.1, Hmax: 1.1, Count: nUEPerCell * nCells}
			// newnodetype.Mode = deployment.ReceiveOnly
			// setting.AddNodeType(newnodetype)

			/// CASE B1 & B2
			newnodetype = deployment.NodeType{Name: "UE", Hmin: 1.1, Hmax: 1.1, Count: GPusers * trueCells}
			newnodetype.Mode = deployment.ReceiveOnly
			setting.AddNodeType(newnodetype)

			newnodetype = deployment.NodeType{Name: "VUE", Hmin: 1.1, Hmax: 1.1, Count: NUEsPerVillage * NVillages * trueCells}
			newnodetype.Mode = deployment.ReceiveOnly
			setting.AddNodeType(newnodetype)

			newnodetype = deployment.NodeType{Name: "MUE", Hmin: 1.1, Hmax: 1.1, Count: NMobileUEs * trueCells}
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

	// clocations := deployment.HexGrid(nCells, vlib.Origin3D, CellRadius, 30)
	clocations, _ := deployment.HexWrapGrid(nCells, vlib.Origin3D, CellRadius, 30, 19)
	system.SetAllNodeLocation("BS0", vlib.Location3DtoVecC(clocations))
	system.SetAllNodeLocation("BS1", vlib.Location3DtoVecC(clocations))
	system.SetAllNodeLocation("BS2", vlib.Location3DtoVecC(clocations))

	// CASE A1 & A2
	// uelocations := LoadUELocations(system)
	// system.SetAllNodeLocation("UE", uelocations)

	// CASE B1 & B2
	// Workaround else should come from API calls or Databases
	uelocations := LoadUELocationsGP(system)
	vuelocations := LoadUELocationsV(system)
	muelocations := LoadUELocations(system)

	system.SetAllNodeLocation("UE", uelocations)
	system.SetAllNodeLocation("VUE", vuelocations)
	system.SetAllNodeLocation("MUE", muelocations)

}

func LoadUELocationsV(system *deployment.DropSystem) vlib.VectorC {

	var uelocations vlib.VectorC
	hexCenters := deployment.HexGrid(trueCells, vlib.FromCmplx(deployment.ORIGIN), CellRadius, 30)
	for indx, bsloc := range hexCenters {
		// log.Printf("Deployed for cell %d at %v", indx, bsloc.Cmplx())
		_ = indx
		// 3-Villages in the HEXAGONAL CELL
		//villageCentre := deployment.HexRandU(bsloc, CellRadius, NVillages, 30)

		// Practical
		//	villageCentres := deployment.AnnularRingPoints(bsloc.Cmplx(), 1500, 3000, NVillages)
		villageCentres := deployment.AnnularRingEqPoints(bsloc.Cmplx(), VillageDistance, NVillages) /// On
		offset := vlib.RandUFVec(NVillages).ShiftAndScale(0, 500.0)                                 // add U(0,1500)  scale by 1 to 2.0
		rotate := vlib.RandUFVec(NVillages).ScaleAndShift(math.Pi/10, -math.Pi/20)                  // +- 10 degrees
		_ = rotate
		_ = offset
		for v, vc := range villageCentres {
			// Add Random offset U(0,1500) Radially
			c := vc + cmplx.Rect(offset[v], cmplx.Phase(vc)) // +rotate[v]

			// log.Printf("Adding Village %d of GP %d , VC  %v , Radial Offset %v , %v, RESULT %v", v, indx, vc, offset[v], (cmplx.Phase(vc)), cmplx.Abs(c-vc))
			log.Printf("Adding Village %d of GP %d  : %d users", v, indx, NUEsPerVillage)
			villageUElocations := deployment.CircularPoints(c, VillageRadius, NUEsPerVillage)

			uelocations = append(uelocations, villageUElocations...)
		}

	}

	return uelocations
}

func LoadUELocationsGP(system *deployment.DropSystem) vlib.VectorC {

	var uelocations vlib.VectorC
	hexCenters := deployment.HexGrid(trueCells, vlib.FromCmplx(deployment.ORIGIN), CellRadius, 30)
	for indx, bsloc := range hexCenters {
		log.Printf("Dropping GP %d UEs for cell %d", GPusers, indx)

		// AT GP
		uelocation := deployment.CircularPoints(bsloc.Cmplx(), GPradius, GPusers)
		uelocations = append(uelocations, uelocation...)

	}

	return uelocations

}

func LoadUELocations(system *deployment.DropSystem) vlib.VectorC {

	var uelocations vlib.VectorC
	hexCenters := deployment.HexGrid(trueCells, vlib.FromCmplx(deployment.ORIGIN), CellRadius, 30)
	for indx, bsloc := range hexCenters {
		log.Printf("Dropping Uniform %d UEs for cell %d", NMobileUEs, indx)

		ulocation := deployment.HexRandU(bsloc.Cmplx(), CellRadius, NMobileUEs, 30)
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
		// fmt.Printf("\nNode %d , Omni= %v, Dirction=(H%v,V%v) and center is %v", nodeID, obj.Omni, obj.HTiltAngle, obj.VTiltAngle, obj.Centre)
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

		// copy(systemAntennas[i], defaultAAS)
		vlib.LoadStructure("sector.json", systemAntennas[i])
		*systemAntennas[i] = defaultAAS

		// systemAntennas[i].FreqHz = CarriersGHz[0] * 1.e9
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
		// vgain := vlib.NewVectorF(360)

		cnt := 0
		cmd := `delta=pi/180;
		phaseangle=0:delta:2*pi-delta;`
		matlab.Command(cmd)
		for d := 0; d < 360; d++ {
			hgain[cnt] = systemAntennas[i].ElementDirectionHGain(float64(d))
			//		hgain[cnt] = systemAntennas[i].ElementEffectiveGain(thetaH, thetaV)
			cnt++
		}

		matlab.Export("gain"+strconv.Itoa(i), hgain)
		// fmt.Printf("\nBS %d, Antenna : %#v", i, systemAntennas[i])

		cmd = fmt.Sprintf("polar(phaseangle,gain%d);hold all", i)
		matlab.Command(cmd)
	}
}
