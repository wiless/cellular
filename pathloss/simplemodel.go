package pathloss

import (
	"fmt"
	"math"
	"math/cmplx"

	"github.com/wiless/cellular/deployment"
	"github.com/wiless/vlib"
)

type SimplePLModel struct {
	ModelSetting
}

func (w *SimplePLModel) Set(ModelSetting) {

}
func (w SimplePLModel) Get() ModelSetting {
	return ModelSetting{}
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

func (p *SimplePLModel) LossInDbBetweenNodes(src, dest deployment.Node) (plDb float64, valid bool) {

	return p.LossInDbBetween3D(src.Location, dest.Location)
}

func (p *SimplePLModel) LossInDbBetween3D(src, dest vlib.Location3D) (plDb float64, valid bool) {
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
		return math.NaN(), false
	}

	// rm := 26.16 * math.Log10(FreqMHz)
	// log.Printf("\n Frequency %v  %v", FreqMHz, result)
	// return p.LossInDb(distance)
	// log.Printf("\n Height %v", src)
	return result, true
}

func (w SimplePLModel) LossInDbNodes(txnode, rxnode deployment.Node, freqGHz float64) (plDb float64, valid bool) {

	return 0, true
}

func (p *SimplePLModel) LossInDb3D(src, dest vlib.Location3D, freqGHz float64) (plDb float64, valid bool) {
	FreqMHz := freqGHz * 1000.0
	p.FreqHz = freqGHz * 1e9
	distance := src.DistanceFrom(dest) / 1.0e3
	var result float64
	result = 9999
	// dest.Z = 1.1
	fmt.Println("\n PATH LOSS HEIGHT src,dest", src.Z, dest.Z)
	if p.FreqHz >= 1.5e8 && p.FreqHz < 1.5e9 {
		var Ch float64
		// Ch = 0.8 + (1.1*math.Log10(FreqMHz)-0.7)*dest.Z - 1.56*math.Log10(FreqMHz)
		if FreqMHz >= 150.0 && FreqMHz <= 200.0 {
			Ch = 8.29*math.Pow(math.Log10(1.54*dest.Z), 2) - 1.1
		} else if FreqMHz > 200.0 && FreqMHz <= 1500.0 {
			Ch = 3.2*math.Pow(math.Log10(11.75*dest.Z), 2) - 4.97
		}
		result = 69.55 + 26.16*math.Log10(FreqMHz) - 13.82*math.Log10(src.Z) - Ch + (44.9-6.55*math.Log10(src.Z))*math.Log10(distance)
		if math.IsInf(result, 0) {
			fmt.Println("Something ************ IS WRONG *****************   < 1.5 Ghz Ch = ", Ch, "inline val", math.Log10(11.75*dest.Z))
		}
	} else if p.FreqHz >= 1.5e9 && p.FreqHz < 2.0e9 {
		a := (1.1*math.Log10(FreqMHz)-0.7)*dest.Z - (1.56*math.Log10(FreqMHz) - 0.8)
		result = 46.3 + 33.9*math.Log10(FreqMHz) - 13.82*math.Log10(src.Z) - a + (44.9-6.55*math.Log10(src.Z))*math.Log10(distance) + 3
		if math.IsInf(result, 0) {
			fmt.Println("Something ************ IS WRONG ***************** > 1.5GHz  ")
		}
	} else {
		fmt.Printf("Something ************ UINKNOWN CASE  ***************** > 1.5GHz  ")
		return math.NaN(), false
	}

	// rm := 26.16 * math.Log10(FreqMHz)
	// log.Printf("\n Frequency %v  %v", FreqMHz, result)
	// return p.LossInDb(distance)
	// log.Printf("\n Height %v", src)
	return result, true
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
