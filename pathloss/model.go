// Implements a simple phase delay from different n-Antenna elements
package pathloss

import (
	"encoding/json"
	"log"
	"math"
	"strings"

	"github.com/wiless/cellular/deployment"

	"github.com/wiless/vlib"
)

type PLModel interface {
	Set(ModelSetting)
	Get() ModelSetting
	LossInDbNodes(txnode, rxnode deployment.Node, freqGHz float64) (plDb float64, valid bool)
	LossInDb3D(txnode, rxnode vlib.Location3D, freqGHz float64) (plDb float64, valid bool)
}
type Model interface {
	Set(ModelSetting)
	Get() ModelSetting
	LossInDbNodes(txnode, rxnode deployment.Node, freqGHz float64) (plDb float64, valid bool)
	LossInDb3D(txnode, rxnode vlib.Location3D, freqGHz float64) (plDb float64, valid bool)
}

type PathLossType int

var PathLossTypes = [...]string{
	"Exponential",
	"FreeSpace",
}

func (p PathLossType) String() string {
	return PathLossTypes[p]
}

type ModelSetting struct {
	Type           PathLossType
	FreqHz         float64
	freqGHz        float64
	CutOffDistance float64
	Param          []float64 /// Factors relatedto
	pNames         []string
	Name           string
	isInitialized  bool
	AddShadowLoss  bool
	param          map[string]float64 /// always use capital letters for parameter name
}

func (m *ModelSetting) SetFGHz(fGHz float64) *ModelSetting {
	m.FreqHz = fGHz * 1e9
	m.freqGHz = fGHz
	return m
}

// Value returns the vaue of the paramter set for the model
func (m *ModelSetting) Value(pname string) float64 {
	if m.param == nil {
		return 0
	}
	pname = strings.ToUpper(pname)
	value := m.param[pname]
	return value
}

func (m *ModelSetting) Parameters() []string {
	return m.pNames
}

func (m *ModelSetting) AddParam(name string, value float64) *ModelSetting {
	if m.param == nil {
		m.param = make(map[string]float64)

	}
	name = strings.ToUpper(name)
	m.param[name] = value
	m.pNames = append(m.pNames, name)
	m.Param = append(m.Param, value)
	return m
}

func (m *ModelSetting) FGHz() (fGHz float64) {
	if m.freqGHz == 0 {
		m.freqGHz = m.FreqHz / 1.0e9
	}
	return m.freqGHz
}

func (m *ModelSetting) SetDefault() {
	m.Type = Exponential
	m.FreqHz = 2.0e9
	m.CutOffDistance = 0
	m.AddShadowLoss = false
	m.Init()
}

func (m *ModelSetting) Init() {
	m.isInitialized = true
	switch m.Type {
	case Exponential:
		m.Param = []float64{2, 0}
		return
	case FreeSpace:
		// L = 20\ \log_{10}\left(\frac{4\pi d}{\lambda}\right)

		c := 3.0e8
		Lamda := c / m.FreqHz
		m.Param = []float64{4 * math.Pi / Lamda}
		return
	default:
	}

}

func NewModelSetting() *ModelSetting {
	result := new(ModelSetting)
	result.param = make(map[string]float64)
	result.SetDefault()
	return result
}

func (s *ModelSetting) Set(str string) {
	err := json.Unmarshal([]byte(str), s)
	if err != nil {
		log.Print("Error ", err)
	}
}

const (
	Exponential PathLossType = iota
	FreeSpace
)
