/* Code contribution by istdev
 */
// Implements a simple phase delay from different n-Antenna elements
package pathloss

import (
	"log"
	"math"

	"github.com/wiless/cellular/deployment"
	"github.com/wiless/vlib"
)

type OkumuraHata struct {
	wsettings ModelSetting
}

func (w *OkumuraHata) Set(ModelSetting) {

}
func (w OkumuraHata) Get() ModelSetting {
	return ModelSetting{}
}
func (w OkumuraHata) LossInDbNodes(txnode, rxnode deployment.Node, freqGHz float64) (plDb float64, valid bool) {

	return 0, true
}
func (w OkumuraHata) LossInDb3D(src, dest vlib.Location3D, freqGHz float64) (plDb float64, valid bool) {
	FreqMHz := freqGHz * 1.0e3
	distance := src.DistanceFrom(dest) / 1.0e3

	var result float64
	result = -1
	if FreqMHz >= 150 && FreqMHz < 1500 && distance > 0.05 {
		var Ch float64
		// Ch = 0.8 + (1.1*math.Log10(FreqMHz)-0.7)*dest.Z - 1.56*math.Log10(FreqMHz)
		if FreqMHz >= 150.0 && FreqMHz <= 200.0 {
			Ch = 8.29*math.Pow(math.Log10(1.54*dest.Z), 2) - 1.1
		} else if FreqMHz > 200.0 && FreqMHz <= 1500.0 {
			Ch = 3.2*math.Pow(math.Log10(11.75*dest.Z), 2) - 4.97
		}
		result = 69.55 + 26.16*math.Log10(FreqMHz) - 13.82*math.Log10(src.Z) - Ch + (44.9-6.55*math.Log10(src.Z))*math.Log10(distance)

	} else if FreqMHz >= 1500 && FreqMHz < 2000 && distance > 0.05 {
		a := (1.1*math.Log10(FreqMHz)-0.7)*dest.Z - (1.56*math.Log10(FreqMHz) - 0.8)
		result = 46.3 + 33.9*math.Log10(FreqMHz) - 13.82*math.Log10(src.Z) - a + (44.9-6.55*math.Log10(src.Z))*math.Log10(distance) + 3

	} else if FreqMHz >= 150 && FreqMHz < 2000 && distance <= 0.05 {
		result = 20*math.Log10(distance) + 20*math.Log10(FreqMHz) + 32.45

	} else {
		log.Panic("Path loss model does not valid for given frequency")
		return 0, false
	}

	return result, true
}
