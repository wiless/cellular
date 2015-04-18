// Implements a simple phase delay from different n-Antenna elements
package pathloss

import (
	"encoding/json"
	"log"
	"math"

	"github.com/wiless/cellular/deployment"

	"github.com/wiless/vlib"
)

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
	CutOffDistance float64
	Param          []float64 /// Factors relatedto
	isInitialized  bool
	AddShadowLoss  bool
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
