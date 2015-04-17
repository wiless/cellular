// Implements a simple phase delay from different n-Antenna elements
package pathloss

import (
	"encoding/json"
	"log"

	// "flag"

	"math"
	"math/cmplx"

	"github.com/wiless/vlib"
)

type PathLossType int

type SimplePLModel struct {
	ModelSetting
}

type PathLossModel SimplePLModel

// interface {}

const (
	Exponential PathLossType = iota
	FreeSpace
)

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

func (p *SimplePLModel) LossInDb(distance float64) float64 {
	switch p.Type {
	case Exponential:
		{
			if distance <= p.CutOffDistance {
				return 0
			}
			// L = 10\ n\ \log_{10}(d)+C
			n, C := p.Param[0], p.Param[1]

			/// Not the exact step, just a simple dependency n is small for <1GHz
			n = n * p.FreqHz / 1e9
			result := 10.0*n*math.Log10(distance) + C
			return result
		}
	case FreeSpace:
		{
			if distance <= p.CutOffDistance {
				return 0
			}
			// L = 20\ \log_{10}\left(\frac{4\pi d}{\lambda}\right)
			factor := p.Param[0]
			result := 20 * math.Log10(factor*distance)
			return result
		}
	default:
		return -100
	}
}

func (p *SimplePLModel) LossInDbBetween(src, dest complex128) float64 {
	distance := cmplx.Abs(dest - src)
	return p.LossInDb(distance)
}

func (p *SimplePLModel) LossInDbBetween3D(src, dest vlib.Location3D) float64 {
	distance := Distance3D(src, dest)
	log.Println("Inside Loss in dB between 3d", src, dest)
	return p.LossInDb(distance)
}

func (p *SimplePLModel) AllLossInDbBetween3D(src vlib.Location3D, dest []vlib.Location3D) vlib.VectorF {
	result := vlib.NewVectorF(len(dest))
	for i := 0; i < len(dest); i++ {
		distance := src.DistanceFrom(dest[i])
		result[i] = p.LossInDb(distance)
	}
	return result
}

func (p *SimplePLModel) AllLossInDbBetween(src complex128, dest vlib.VectorC) vlib.VectorF {

	result := vlib.NewVectorF(dest.Size())
	for i := 0; i < dest.Size(); i++ {
		distance := cmplx.Abs(src - dest[i])
		result[i] = p.LossInDb(distance)
	}
	return result

}

func Distance3D(src, dest vlib.Location3D) float64 {
	distance := src.DistanceFrom(dest)
	return distance
}

func Distance(src, dest complex128) float64 {
	return cmplx.Abs(src - dest)
}
