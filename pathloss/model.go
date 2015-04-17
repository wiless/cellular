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

type PathLossModel struct {
	ModelSetting
}

func (p *PathLossModel) LossInDb(distance float64) float64 {
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

func (p *PathLossModel) LossInDbBetween(src, dest complex128) float64 {
	distance := cmplx.Abs(dest - src)
	return p.LossInDb(distance)
}

func (p *PathLossModel) LossInDbBetween3D(src, dest vlib.Location3D) float64 {
	FreqMHz := p.FreqHz / 1.0e6
	distance := src.DistanceFrom(dest) / 1.0e3
	var result float64
	if p.FreqHz >= 1.5e8 && p.FreqHz < 1.5e9 {
		var Ch float64
		// Ch = 0.8 + (1.1*math.Log10(FreqMHz)-0.7)*dest.Z - 1.56*math.Log10(FreqMHz)
		if FreqMHz >= 150.0 && FreqMHz <= 200.0 {
			Ch = 8.29*math.Pow(math.Log10(1.54*dest.Z), 2) - 1.1
		} else if FreqMHz > 200.0 && FreqMHz <= 1500.0 {
			Ch = 3.2*math.Pow(math.Log10(11.75*dest.Z), 2) - 4.97
		}
		result = 69.55 + 26.16*math.Log10(FreqMHz) - 13.82*math.Log10(src.Z) - Ch + (44.9-6.55*math.Log10(src.Z))*math.Log10(distance)
	} else if p.FreqHz >= 1.5e9 && p.FreqHz < 2.0e9 {
		a := (1.1*math.Log10(FreqMHz)-0.7)*dest.Z - (1.56*math.Log10(FreqMHz) - 0.8)
		result = 46.3 + 33.9*math.Log10(FreqMHz) - 13.82*math.Log10(src.Z) - a + (44.9-6.55*math.Log10(src.Z))*math.Log10(distance) + 3
	} else {
		log.Panic("Path loss model does not valid for given frequency")
	}

	// rm := 26.16 * math.Log10(FreqMHz)
	// log.Printf("\n Frequency %v  %v", FreqMHz, result)
	// return p.LossInDb(distance)
	// log.Printf("\n Height %v", src)
	return result
}

func (p *PathLossModel) AllLossInDbBetween3D(src vlib.VectorF, dest []vlib.VectorF) vlib.VectorF {
	result := vlib.NewVectorF(len(dest))
	for i := 0; i < len(dest); i++ {
		distance := Distance3D(src, dest[i])
		result[i] = p.LossInDb(distance)
	}
	return result
}

func (p *PathLossModel) AllLossInDbBetween(src complex128, dest vlib.VectorC) vlib.VectorF {

	result := vlib.NewVectorF(dest.Size())
	for i := 0; i < dest.Size(); i++ {
		distance := cmplx.Abs(src - dest[i])
		result[i] = p.LossInDb(distance)
	}
	return result

}

func Distance3D(src, dest vlib.VectorF) float64 {
	if len(src) != 3 || len(dest) != 3 {
		return -1
	}
	result := vlib.Sub(src, dest)
	return vlib.Norm2(result)
}

func Distance(src, dest complex128) float64 {
	return cmplx.Abs(src - dest)
}
