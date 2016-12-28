/* Code contribution by istdev
 */
// Implements a simple phase delay from different n-Antenna elements
package pathloss

import (
	"github.com/wiless/cellular/deployment"
	"github.com/wiless/vlib"
)

type RMa struct {
	wsettings ModelSetting
}

func (w *RMa) Set(ModelSetting) {

}
func (w RMa) Get() ModelSetting {
	return ModelSetting{}
}

// type Model interface {
// 	Set(ModelSetting)
// 	Get() ModelSetting
// 	LossInDbNodes(txnode, rxnode deployment.Node, freqGHz float64) (plDb float64, valid bool)
// 	LossInDb3D(txnode, rxnode vlib.Location3D, freqGHz float64) (plDb float64, valid bool)
// }

func (r *RMa) LossInDb3D(src, dest vlib.Location3D, freqGHz float64) (plDb float64, valid bool) {
	// FreqMHz := freqGHz * 1.0e3                 // Frequency is in MHz
	// distance := src.DistanceFrom(dest) / 1.0e3 // Convert to km (most equations have d in km)

	// var result float64
	// result = -1

	// result = 46.3 + 33.9*math.Log10(FreqMHz) - 13.82*math.Log10(src.Z) - a + (44.9-6.55*math.Log10(src.Z))*math.Log10(distance) + 3

	return 0, true
}

func (r *RMa) LossInDbNodes(txnode, rxnode deployment.Node, freqGHz float64) (plDb float64, valid bool) {

	return 0, true
}
