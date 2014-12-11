package deployment

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	sim "github.com/wiless/cellular"
	"github.com/wiless/vlib"
	"log"
	"math"
	"math/rand"
)

// func (c *Complex) MarshalJSON() ([]byte, error) {
// 	return []byte(fmt.Sprintf("A,%v", complex128(*c))), nil
// }

// func (c *Complex) UnmarshalJSON([]byte) error {
// 	fmt.Print("Something")
// 	return nil
// }

// func (v *VectorF) MarshalJSON() ([]byte, error) {
// 	str := fmt.Sprintf("x=%f", v)
// 	return []byte(str), nil
// 	// return json.Marshal([]float64(v))
// }
// func (c vlib.Complex) String() string {
// 	return fmt.Sprintf("S,%v", complex128(c))
// }

type Node struct {
	Type     string
	id       int
	Location vlib.Location3D
	Height   float64
	Meta     string
	Indoor   bool
}

type DropParameter struct {
	Centre     complex128
	Type       DropType
	Randomnoss bool // if true, uniformly distributed else equallyspaced in region

	//Radius in meters
	Radius      float64 `json:"radius"`
	InnerRadius float64

	/// Angles are in degree
	RotationDegree      float64
	InnerRotationDegree float64

	// Number of Drops
	NCount int
}

func (d DropParameter) MarshalJSON() ([]byte, error) {
	// var mydata map[string]interface{}
	// mydata = map[string]interface{}(d)

	mydata, err := vlib.ToMap(d)
	// fmt.Printf("\n Drop Parameter %#v", mydata)
	if err != nil {
		fmt.Println(err)
		return []byte{}, err
	}
	fx, rerr := json.Marshal(mydata)
	return fx, rerr
}

type NodeType struct {
	Name    string
	Hmin    float64
	Hmax    float64
	Count   int
	startID int
	nodeIDs vlib.VectorI
	Params  DropParameter
}

type DropType int

var DropTypes = [...]string{
	"Circular",
	"Hexagonal",
	"Rectangular",
	"Annular",
}

func (c DropType) String() string {
	return DropTypes[c]
}

type DropSystem struct {
	*dDropSetting
	Nodes  map[int]Node
	lastID int
}

func (d *DropSystem) UnmarshalJSON(jsondata []byte) error {
	bfr := bytes.NewBuffer(jsondata)
	dec := json.NewDecoder(bfr)
	var customobject map[string]interface{}
	customobject = make(map[string]interface{})
	dec.Decode(&customobject)
	d.lastID = int(customobject["LastID"].(float64))
	temp := customobject["DropSetting"].(dDropSetting)

	fmt.Printf("\n DropSetting : %v", temp)
	bfr.WriteString(`{`)
	bfr.WriteString(`"DropSetting":`)
	// enc.Encode(d.dDropSetting)
	// // bfr.Bytes()[bfr.Len()-2] = ' '
	// bfr.WriteString(`,"Nodes":[`)
	// cnt := 0
	// maxcount := len(d.Nodes)
	// for key, val := range d.Nodes {
	// 	obj := struct {
	// 		ID      int
	// 		NodeObj Node
	// 	}{key, *val}
	// 	enc.Encode(obj)

	// 	cnt++
	// 	if cnt == maxcount {
	// 		break
	// 	} else {
	// 		bfr.WriteByte(',')
	// 	}

	// }

	// bfr.WriteByte(']')
	// bfr.WriteString(`,"LastID":`)
	// enc.Encode(d.lastID)
	// bfr.WriteString("}\n")
	// return bfr.Bytes(), nil
	return nil
}

func (d *DropSystem) MarshalJSON() ([]byte, error) {

	bfr := bytes.NewBuffer(nil)
	enc := json.NewEncoder(bfr)
	bfr.WriteString(`{`)
	bfr.WriteString(`"DropSetting":`)
	enc.Encode(d.dDropSetting)
	// bfr.Bytes()[bfr.Len()-2] = ' '
	bfr.WriteString(`,"Nodes":[`)
	cnt := 0
	maxcount := len(d.Nodes)
	for key, val := range d.Nodes {
		obj := struct {
			ID      int
			NodeObj Node
		}{key, val}
		enc.Encode(obj)

		cnt++
		if cnt == maxcount {
			break
		} else {
			bfr.WriteByte(',')
		}

	}

	bfr.WriteByte(']')
	bfr.WriteString(`,"LastID":`)
	enc.Encode(d.lastID)
	bfr.WriteString("}\n")
	return bfr.Bytes(), nil
}

// func (c *Complex) UnmarshalJSON([]byte) error {
// 	fmt.Print("Something")
// 	return nil
// }

type Area struct {
	Celltype   DropType
	Dimensions vlib.VectorF
}

type dDropSetting struct {
	NodeTypes []NodeType
	// minDistance    map[NodePair]float64 `json:"-"`
	CoverageRegion Area // For circular its just radius, Rectangular, its length, width
	isInitialized  bool
	TxNodeNames    []string
	RxNodeNames    []string
}

func (d *dDropSetting) NodeCount(ntype string) int {
	for _, val := range d.NodeTypes {
		if val.Name == ntype {

			return val.Count
		}
	}
	return -1
}

func (d *dDropSetting) GetNodeIndex(ntype string) int {
	for indx, val := range d.NodeTypes {
		if val.Name == ntype {

			return indx
		}
	}
	return -1
}

func (d *dDropSetting) SetNodeCount(ntype string, count int) {

	// fmt.Println(d.NodeTypeLookup)
	// fmt.Println(d.NodeTypes)
	indx := d.GetNodeIndex(ntype)
	if indx != -1 {
		d.NodeTypes[indx].Count = count
	}

	// }

}

func (d *dDropSetting) SetCoverage(area Area) {
	d.CoverageRegion = area
}

func (d *DropSystem) NewNode(ntype string) *Node {
	indx := d.GetNodeIndex(ntype)
	notype := &d.NodeTypes[indx]
	node := new(Node)
	node.Type = notype.Name
	node.Indoor = false
	if notype.Hmin == notype.Hmax {
		node.Height = notype.Hmin
	} else {
		node.Height = rand.Float64()*(notype.Hmax-notype.Hmin) + notype.Hmin
	}
	// node.id = notype.startID
	node.id = d.lastID
	// fmt.Printf("\n Node Type is %#v", notype)
	// fmt.Printf("\n Creating a Node of type %s , with ID %d for Coverage Type %s", ntype, node.id, d.CoverageRegion.CellType)

	d.lastID++

	return node
}

func (d *dDropSetting) SetTxNodeNames(names ...string) {
	d.TxNodeNames = names
}

func (d *dDropSetting) SetRxNodeNames(names ...string) {
	d.RxNodeNames = names
}

func (d *dDropSetting) GetTxNodeNames() []string {
	return d.TxNodeNames
}

func (d *dDropSetting) GetRxNodeNames() []string {
	return d.RxNodeNames
}
func from2D(loc complex128, height float64) [3]float64 {
	var result [3]float64
	result[0] = real(loc)
	result[1] = imag(loc)
	result[2] = height
	return result
}

func NewNodeType(name string, heights ...float64) *NodeType {
	result := new(NodeType)
	result.Name = name
	switch len(heights) {
	case 0:
		result.Hmax, result.Hmax = 0, 0
	case 1:
		result.Hmin, result.Hmax = heights[0], heights[0]
	default: /// Any arguments >=2
		result.Hmin, result.Hmax = heights[0], heights[1]
		// case 3:
		// 	result.Hmin, result.Hmax = heights[0],heights[1]

	}

	return result
}

func (d *DropSystem) SetSetting(setting *dDropSetting) {
	d.dDropSetting = setting
}
func (d *DropSystem) GetSetting() *dDropSetting {
	return d.dDropSetting
}

func (d *dDropSetting) AddNodeType(ntype NodeType) {

	d.NodeTypes = append(d.NodeTypes, ntype)
}

func (d *dDropSetting) SetDefaults() {
	d.SetCoverage(CircularCoverage(100))

	bs := *NewNodeType("BS", 20)
	ue := *NewNodeType("UE", 0)
	d.AddNodeType(bs)
	d.AddNodeType(ue)

}

func NewDropSetting() *dDropSetting {
	result := new(dDropSetting)
	result.isInitialized = false
	return result
}

func (d *dDropSetting) Init() {
	// d.NodeCount = make(map[string]int)
	// d.NodeMap = make(map[string]NodeType)
	d.isInitialized = true

	for indx, _ := range d.NodeTypes {
		d.NodeTypes[indx].nodeIDs = vlib.NewVectorI(d.NodeTypes[indx].Count)
		//	fmt.Println("The node types are  : indx, nodetype ", indx, notype)
	}
}

func (d *DropSystem) Init() {
	d.dDropSetting.Init()
	d.Nodes = make(map[int]Node)
	count := 0

	for indx, _ := range d.NodeTypes {
		d.NodeTypes[indx].startID = count
		d.NodeTypes[indx].nodeIDs.Resize(d.NodeTypes[indx].Count)
		//fmt.Println("\nWill Create %s Nodes %d : %v", d.NodeTypes[indx].Name, d.NodeTypes[indx].Count, d.NodeTypes[indx].nodeIDs)

		for i := 0; i < d.NodeTypes[indx].Count; i++ {
			node := d.NewNode(d.NodeTypes[indx].Name)
			d.Nodes[node.id] = *node
			d.NodeTypes[indx].nodeIDs[i] = node.id
		}
		count += d.NodeTypes[indx].Count

		//fmt.Printf("\n Nodes %s created %d ", d.NodeTypes[indx].Name, d.NodeTypes[indx].nodeIDs.Size())
	}
	//fmt.Println("SYSTEM IS = ", d)

}

func (d *DropSystem) GetNodeType(ntype string) *NodeType {

	indx := d.GetNodeIndex(ntype)

	if indx != -1 {
		return &d.NodeTypes[indx]
	} else {
		log.Panicln("DropSystem::GetNodeType() : No Such Type ", ntype)

		return nil
	}
}

func (d *DropSystem) DropNodeType(nodetype string) {

	// notype := d.GetNodeType(nodetype)
	// ntype.nodeIDs = vlib.NewVectorI(ntype.count)
	N := d.NodeCount(nodetype)
	// fmt.Printf("\nDrop Nodes of Type %v : %d", d.NodeTypes[0], N)
	// d.NodeTypes[0].startID = 20
	// fmt.Printf("\nModified  Nodes of Type %v : %d", d.NodeTypes[0], N)

	// ntype.startID = d.lastID

	// if circular
	switch d.CoverageRegion.Celltype {
	case Circular:

		radius := d.CoverageRegion.Dimensions[0]
		locations := CircularPoints(complex(0, 0), radius, N)
		d.SetAllNodeLocation(nodetype, locations)

	case Rectangular:
		length := d.CoverageRegion.Dimensions[0]
		locations := RectangularNPoints(complex(0, 0), length, length, 0, N)
		d.SetAllNodeLocation(nodetype, locations)
	default:
		radius := d.CoverageRegion.Dimensions[0]
		locations := CircularPoints(complex(0, 0), radius, N)
		d.SetAllNodeLocation(nodetype, locations)

	}
	d.PopulateHeight(nodetype)
}

func (d *DropSystem) PopulateHeight(ntype string) {

	notype := d.GetNodeType(ntype)
	var random bool = false
	var height float64
	height = notype.Hmin
	if notype.Hmax != notype.Hmin {
		random = true
	}

	// result := vlib.NewVectorC(notype.Count)
	Hrange := notype.Hmax - notype.Hmin
	for i := 0; i < notype.nodeIDs.Size(); i++ {
		if random {
			height = Hrange*rand.Float64() + notype.Hmin
		}
		d.Nodes[notype.nodeIDs[i]].Location.SetHeight(height)
	}

}

func (d *DropSystem) Locations(ntype string) vlib.VectorC {

	notype := d.GetNodeType(ntype)

	result := vlib.NewVectorC(notype.Count)
	for i := 0; i < (notype.Count); i++ {

		result[i] = d.Nodes[notype.nodeIDs[i]].Location.Cmplx()
	}
	return result
}

func (d *DropSystem) Locations3D(ntype string) []vlib.Location3D {
	notype := d.GetNodeType(ntype)
	result := make([]vlib.Location3D, notype.Count)
	for i := 0; i < (notype.Count); i++ {
		result[i] = d.Nodes[notype.nodeIDs[i]].Location
	}
	return result
}

func (d *DropSystem) SetNodeLocation(ntype string, nid int, location complex128) {
	notype := d.GetNodeType(ntype)
	val := notype.nodeIDs[nid]
	d.Nodes[val].Location.FromCmplx(location)
	// d.Nodes[val].Location.SetHeight(d.Nodes[val].Height)

}
func (d *DropSystem) SetNodeLocationOf(ntype string, nodeIDs vlib.VectorI, locations vlib.VectorC) {
	for i := 0; i < nodeIDs.Size(); i++ {
		d.SetNodeLocation(ntype, nodeIDs[i], locations[i])
	}
}
func (d *DropSystem) SetAllNodeLocation(ntype string, locations vlib.VectorC) {
	notype := d.GetNodeType(ntype)

	// fmt.Println("No. of Total Nodes is ", len(d.Nodes))
	// fmt.Println("No. of Locations is ", locations.Size())
	// fmt.Println("No. of NodeIDs is ", notype)
	for indx, val := range notype.nodeIDs {
		d.Nodes[val].Location.FromCmplx(locations[indx])
	}
}

func RandPoint(centre complex128, radius float64) complex128 {
	var result complex128
	// r := math.Sqrt(rand.Float64())
	r := math.Sqrt(rand.Float64()) * radius
	theta := rand.Float64() * 360

	scale := complex(r, 0)
	point := scale * sim.GetEJtheta(theta)
	// x := r*math.Cos(theta);
	// y := r*math.Sin(theta);

	result = point + centre

	return result
}

func HexagonalPoints(centre complex128, length float64) vlib.VectorC {
	result := vlib.NewVectorC(6)
	// degree := 0.0
	for i := 0; i < 6; i++ {

		result[i] = sim.GetEJtheta(60.0*float64(i))*complex(length, 0) + centre
	}

	return result
}

// n = 10000;
// Rc2 = 20;
// Rc1 = 10;
// Xc = -30;
// Yc = -40;

// theta = rand(1,n)*(2*pi);
// r = sqrt((Rc2^2-Rc1^2)*rand(1,n)+Rc1^2);
// x = Xc + r.*cos(theta);
// y = Yc + r.*sin(theta);

// plot(x,y,'.'); axis square
func AnnularPoint(centre complex128, innerRadius, outerRadius float64) complex128 {
	var result complex128
	// r := math.Sqrt(rand.Float64())
	if outerRadius < innerRadius {
		innerRadius, outerRadius = outerRadius, innerRadius
	}
	radius := math.Pow(outerRadius, 2) - math.Pow(innerRadius, 2)
	r := math.Sqrt(rand.Float64()*radius + math.Pow(innerRadius, 2))
	theta := rand.Float64() * 360

	scale := complex(r, 0)
	point := scale * sim.GetEJtheta(theta)
	// x := r*math.Cos(theta);
	// y := r*math.Sin(theta);

	result = point + centre

	return result
}
func AnnularRingPoints(centre complex128, innerRadius, outerRadius float64, N int) vlib.VectorC {
	result := vlib.NewVectorC(N)
	// degree := 0.0
	for i := 0; i < N; i++ {
		result[i] = AnnularPoint(centre, innerRadius, outerRadius)
	}

	return result
}

func AnnularRingEqPoints(centre complex128, outerRadius float64, N int) vlib.VectorC {
	result := vlib.NewVectorC(N)
	// degree := 0.0
	angleOffset := 360.0 / float64(N)
	angle := 0.0
	for i := 0; i < N; i++ {
		point := complex(outerRadius, 0) * sim.GetEJtheta(angle)

		result[i] = point + centre
		angle += angleOffset
	}

	return result
}

func CircularPoints(centre complex128, radius float64, N int) vlib.VectorC {
	result := vlib.NewVectorC(N)
	// degree := 0.0
	for i := 0; i < N; i++ {
		result[i] = RandPoint(centre, radius)
	}

	return result
}

func RectangularPoint(centre complex128, width, height, angleInDegree float64) complex128 {
	dx := rand.Float64()*width - width/2.0
	dy := rand.Float64()*height - height/2.0
	point := complex(dx, dy) /// centred at 0,0
	result := point - centre

	return result * sim.GetEJtheta(angleInDegree)

}
func RectangularNPoints(centre complex128, width, height, angleInDegree float64, N int) vlib.VectorC {
	result := vlib.NewVectorC(N)
	for i := 0; i < N; i++ {
		result[i] = RectangularPoint(centre, width, height, angleInDegree)
	}
	return result
}

func RectangularEqPoints(centre complex128, length, angle float64, N int) vlib.VectorC {
	result := vlib.NewVectorC(N)
	offset := length / float64(N)
	pos := 0.0
	for i := 0; i < N; i++ {
		result[i] = complex(pos, 0)
		pos += offset
	}
	result = result.ScaleC(sim.GetEJtheta(angle))

	mean := -vlib.MeanC(result) + centre
	result = result.AddC(mean)

	return result
}

/// Simplest point for origin centred, rectangular region of length size
func RandPointR(size float64) complex128 {
	return RectangularPoint(ORIGIN, size, size, 0)
}

func CircularCoverage(radius float64) Area {
	return Area{Circular, vlib.VectorF{radius}}
}

func RectangularCoverage(length float64) Area {
	return Area{Rectangular, vlib.VectorF{length, length}}
}

func (d *DropSystem) GetNodeIDs(ntype string) vlib.VectorI {
	indx := d.GetNodeIndex(ntype)
	if indx != -1 {
		return d.NodeTypes[indx].nodeIDs
	}
	return vlib.NewVectorI(0)
}

// wlocation = deployment.RectangularEqPoints(wappos, 50, rand.Float64()*360, WAPNodes)
// 	wlocation = deployment.AnnularRingPoints(deployment.ORIGIN, 100, 200, WAPNodes)
// 	wlocation = deployment.AnnularRingEqPoints(deployment.ORIGIN, 200, WAPNodes)

func (d *DropSystem) Drop(dp *DropParameter, result *vlib.VectorC) error {
	switch dp.Type {
	case Circular:
		locations := CircularPoints(dp.Centre, dp.Radius, dp.NCount)
		result = &locations
		return nil
	case Rectangular:
		locations := RectangularEqPoints(dp.Centre, dp.Radius, dp.RotationDegree, dp.NCount)
		result = &locations
		return nil
	case Hexagonal:
		locations := HexagonalPoints(dp.Centre, dp.Radius)
		result = &locations
		return nil
	case Annular:
		locations := AnnularRingPoints(dp.Centre, dp.InnerRadius, dp.Radius, dp.NCount)
		result = &locations
		return nil
	default:
		return errors.New("Unknown DropType")
	}
}

const (
	ORIGIN = complex(0, 0)
)
const (
	Circular DropType = iota
	Hexagonal
	Rectangular
	Annular
)
