package antenna

import (
	"math"
	cmplx "math/cmplx"

	"github.com/wiless/vlib"
)

 
func (ant SettingAAS) panelorientation(theta float64, indx int) float64 {
	if len(ant.PanelAz) < 2 {
		return theta
	}
	if theta >= 0 {
		theta = -ant.PanelAz[indx] + theta
	} else {
		theta = ant.PanelAz[indx] + theta
	}
	return theta
}

func (ant SettingAAS) CombPatternDb(theta, phi float64) (aag map[int]map[int]vlib.MatrixF, bestPanelID, bestBeamID int, Az, El float64) {
	numpanel := ant.AntennaConfig[3] * ant.AntennaConfig[4]
	aag = make(map[int]map[int]vlib.MatrixF)
	var panelgain map[int]vlib.MatrixF
	panelgain = make(map[int]vlib.MatrixF)
	maxpanelgain := -1000.0
	bestID := 0
	for i := 0; i < numpanel; i++ {
		theta = ant.panelorientation(theta, i)
		panelgain, bestID, Az, El = ant.combPatternDb(theta, phi)
		for id, val := range panelgain {
			if aag[i] == nil {
				aag[i] = make(map[int]vlib.MatrixF)
			}
			aag[i][id] = val
		}
		if panelgain[bestID][0][0] > maxpanelgain {
			bestPanelID = i
			bestBeamID = bestID
			maxpanelgain = panelgain[bestID][0][0]
		}
	}
	return aag, bestPanelID, bestBeamID, Az, El
}

func (ant SettingAAS) combPatternDb(theta, phi float64) (aag map[int]vlib.MatrixF, bestBeamID int, Az, El float64) {

	theta = Wrap180To180(theta)
	phi = Wrap0To180(phi)
	var ag float64
	Az, El, ag = ElementGainDb(theta, phi, ant)
	hspace := ant.ESpacingHFactor
	vspace := ant.ESpacingVFactor
	var sum = complex(0.0, 0.0)

	dtilt := ant.ElectronicTilt // degree
	descan := ant.Dscan         //degree

	nv := ant.AntennaConfig[0] / ant.AntennaConfig[5]
	nh := ant.AntennaConfig[1] / ant.AntennaConfig[6]

 	maxgain := -1000.0
	bestBeamID = 0
	nbeams := len(ant.Dscan) * len(ant.ElectronicTilt)
	aag = make(map[int]vlib.MatrixF, nbeams)

 	var c = complex(math.Sqrt(1/float64(nv*nh)), 0)
	for i := 0; i < len(dtilt); i++ { //  dtilt is a vector of Zenith Angles of the Beam Set
		for j := 0; j < len(descan); j++ { // descan is a vector of Azimuth Angles of the Beam Set
			beamid := j + len(descan)*i
			sum = 0.0
			for m := 1; m <= nv; m++ {
				for n := 1; n <= nh; n++ {
					phiP := -math.Cos(dtilt[i]*math.Pi/180) + math.Cos(phi*math.Pi/180)
					phiR := -math.Sin(dtilt[i]*math.Pi/180)*math.Sin(descan[j]*math.Pi/180) + math.Sin(phi*math.Pi/180)*math.Sin(theta*math.Pi/180)
					w := cmplx.Exp(complex(0, 2*math.Pi*(float64(m-1)*vspace*phiP)))
					v := cmplx.Exp(complex(0, 2*math.Pi*(float64(n-1)*hspace*phiR)))
 					sum = sum + c*cmplx.Conj(w*v)
 				}
			}
			txRUGains := vlib.NewMatrixF(ant.AntennaConfig[5], ant.AntennaConfig[6])
			for k := 0; k < ant.AntennaConfig[5]; k++ {
				for l := 0; l < ant.AntennaConfig[6]; l++ {
 					txRUGains[k][l] = ag + (10 * math.Log10(math.Pow(cmplx.Abs(sum), 2))) // Composite Beam Gain + Antenna Element Gain
					temp := txRUGains[k][l]
					if maxgain < temp {
						maxgain = temp
						bestBeamID = beamid
					}
				}
			}
			aag[beamid] = txRUGains
		}
	}
	return aag, bestBeamID, Az, El

}

// Antenna Gain per panel
// func (ant SettingAAS) CombPatternDb(theta, phi float64) (aag map[int]Panel, bestBeamID int, Az, El float64) {

// 	total_panels := ant.AntennaConfig[3] * ant.AntennaConfig[4]
// 	aag = make(map[int][int]vlib.MatrixF, total_panels)
// 	for panel := 0; panel < total_panels; panel++ {
// 		aag_temp, bestBeamID, Az, El := ant.combPatternDb(theta, phi)
// 		aag[panel] = aag_temp
// 	}

// 	return aag, bestBeamID, Az, El
// }

// type Panel struct {
// 	BeamGain   map[int]vlib.MatrixF
// 	BestBeamID int
// }

func ElementGainDb(theta, phi float64, ant SettingAAS) (az, el, Ag float64) {
	phi = Wrap0To180(phi)
	theta = Wrap180To180(theta)
	MaxGaindBi := ant.GainDb   //    0 for ue and 8 for bs
	theta3dB := ant.HBeamWidth // degree
	phi3dB := ant.VBeamWidth
	SLAmax := ant.SLAV
	Am := SLAmax
	Ah := -math.Min(12.0*math.Pow(theta/theta3dB, 2.0), Am)
 	// fmt.Println("Horizontal Gain: ", Ah)
	MechTiltGCS := ant.BeamTilt // Pointing to Horizon..axis..
	Av := -math.Min(12.0*math.Pow((phi-MechTiltGCS)/phi3dB, 2.0), SLAmax)
	result := -math.Min(-(Av + Ah), Am)
	//result = Ah
	az = Ah
	el = Av

	Ag = result + MaxGaindBi


	return az, el, Ag
}

// Wrap0To180 wraps the input angle to 0 to 180
func Wrap0To180(degree float64) float64 {
	if degree >= 0 && degree <= 180 {
		return degree
	}
	if degree < 0 {
		degree = -degree
	}
	if degree >= 360 {
		degree = math.Mod(degree, 360)
	}
	if degree > 180 {

		degree = 360 - degree
	}
	return degree
}

// Wrap180To180 wraps the input angle to -180 to 180
func Wrap180To180(degree float64) float64 {
	if degree >= -180 && degree <= 180 {
		return degree
	}
	if degree > 180 {
		rem := math.Mod(degree, 180.0)
		degree = -180 + rem

	} else if degree < -180 {
		rem := math.Mod(degree, 180.0)
		//	fmt.Println("Remainder for ", degree, rem)
		degree = 180 + rem
	}
	return degree
}

// BSPatternDb generates the antenna gain for given theta,phi in degree
// based on Table 8-6 in Report ITU-R M.2412
// returns effective Antenna Gain Ag, Horizontal gain az, Elevation Gain el

func BSPatternDb(theta, phi float64) (az, el, Ag float64) {
	phi = Wrap0To180(phi)
	theta = Wrap180To180(theta)
	MaxGaindBi := 8.0 //    0 for ue and 8 for bs
	theta3dB := 65.0  // degree
	SLAmax := 30.0
	Am := SLAmax
	Ah := -math.Min(12.0*math.Pow(theta/theta3dB, 2.0), Am)

	MechTiltGCS := 90.0 // Pointing to Horizon..axis..
	Av := -math.Min(12.0*math.Pow((phi-MechTiltGCS)/theta3dB, 2.0), SLAmax)
	result := -math.Min(-math.Floor(Av+Ah), Am)
	//result = Ah
	az = Ah
	el = Av
	Ag = result + MaxGaindBi
	return az, el, Ag
}

// BSPatternIndoorHS_Db generates the antenna gain for given theta,phi
// based on Table 8-7 in Report ITU-R M.2412
func BSPatternIndoorHS_Db(theta, phi float64) (az, el, Ag float64) {
	phi = Wrap0To180(phi)
	theta = Wrap180To180(theta)
	MaxGaindBi := 5.0
	theta3dB := 90.0
	SLAmax := 25.0
	Am := 25.0
	MechTiltGCS := 110.0 /// Need to be set to 180.. pointing to Ground, when Antenna mounted on ceiling..

	Ah := -math.Min(12.0*math.Pow(theta/theta3dB, 2.0), Am)
	Av := -math.Min(12.0*math.Pow((phi-MechTiltGCS)/theta3dB, 2.0), SLAmax)
	result := -math.Min(-math.Floor(Av+Ah), Am) + MaxGaindBi

	az = Ah
	el = Av
	Ag = result
	return az, el, Ag
}

// UEPatternDb generates the antenna gain for given theta,phi
// based OMNI Directional gain..
func UEPatternOmniDb(theta, phi float64) float64 {
	phi = Wrap0To180(phi)
	theta = Wrap180To180(theta)
	MaxGaindBi := 0.0 //    0 for ue and 8 for bs
	theta3dB := 65.0  // degree
	SLAmax := 30.0
	Am := SLAmax
	Ah := 0.0
	//Ah := -math.Min(12.0*math.Pow(theta/theta3dB, 2.0), Am)

	MechTiltGCS := 90.0 // Pointing to Horizon..axis..
	Av := -math.Min(12.0*math.Pow((180-phi-MechTiltGCS)/theta3dB, 2.0), SLAmax)

	result := -math.Min(-math.Floor(Av+Ah), Am)
	//result = Ah

	Ag := result + MaxGaindBi

	return Ag
}

// UEPatternDb generates the antenna gain for given theta,phi
// based Table 8-8 in Report ITU-R M.2412 for fGHZ  > 4GHz i.e 30GHz & 70GHz
func UEPatternDb(theta, phi float64) (az, el, Ag float64) {
	phi = Wrap0To180(phi)
	theta = Wrap180To180(theta)
	MaxGaindBi := 5.0 //

	theta3dB := 90.0 // degree
	SLAmax := 25.0
	Am := 25.0

	Ah := -math.Min(12.0*math.Pow(theta/theta3dB, 2.0), Am)
	Av := -math.Min(12.0*math.Pow((phi-90.0)/theta3dB, 2.0), SLAmax)
	result := -math.Min(-(Av+Ah), Am) + MaxGaindBi

	az = theta
	el = phi
	Ag = result
	return az, el, Ag
}
