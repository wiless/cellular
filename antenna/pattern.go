package antenna

import (
	"math"
	cmplx "math/cmplx"

	"github.com/wiless/vlib"
)

// func BSPatternDb(theta, phi, gain float64) float64 {

// 	// fmt.Println("Origina ", degree)
// 	if degree > 180 {
// 		rem := math.Mod(degree, 180.0)
// 		degree = -180 + rem

// 	} else if degree < -180 {
// 		rem := math.Mod(degree, 180.0)
// 		//	fmt.Println("Remainder for ", degree, rem)
// 		degree = 180 + rem
// 	}
// 	theta := -(degree)
// 	theta3Db := (s.HBeamWidth)
// 	SLAV := s.SLAV
// 	tilt := -(s.HTiltAngle)
// 	//  Reference TS25.996 - Section 4.5 - BS Antenna Pattern
// 	val := math.Pow(10, -math.Min(12.0*math.Pow((theta-tilt)/theta3Db, 2), SLAV)/10.0)
// 	return val

// }

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

//CombPatternDb calculates returns all the beam gain for each TxRU
func CombPatternDb(theta, phi float64, ant SettingAAS) (aag map[int]vlib.MatrixF, bestBeamID int, Az, El float64) {

	theta = Wrap180To180(theta)
	phi = Wrap0To180(phi)
	var ag float64
	Az, El, ag = BSPatternDb(theta, phi)
	// fmt.Println("Antenna Element Gain:", ag)
	hspace := ant.ESpacingHFactor
	vspace := ant.ESpacingVFactor
	dtilt := ant.ElectronicTilt // degree
	descan := ant.Dscan         //degree
	// var maxgain float64 // Maximum BeamGain out of the Beam Set
	var sum = complex(0.0, 0.0)
	nv := ant.BSAntennaConfig[0] / ant.BSAntennaConfig[5]
	nh := ant.BSAntennaConfig[1] / ant.BSAntennaConfig[6]
	var maxgain float64
	bestBeamID = 0
	// var aag map[int]vlib.MatrixF
	nbeams := len(ant.Dscan) * len(ant.ElectronicTilt)
	aag = make(map[int]vlib.MatrixF, nbeams)
	/////
	c := 1.0 / float64(nv*nh)
	for i := 0; i < len(dtilt); i++ { //  dtilt is a vector of Zenith Angles of the Beam Set
		for j := 0; j < len(descan); j++ { // descan is a vector of Azimuth Angles of the Beam Set
			beamid := j + len(descan)*i
			for m := 1; m <= nv; m++ {
				for n := 1; n <= nh; n++ {
					phiP := -math.Cos(dtilt[i]*math.Pi/180) + math.Cos(phi*math.Pi/180)
					phiR := -math.Sin(dtilt[i]*math.Pi/180)*math.Sin(descan[j]*math.Pi/180) + math.Sin(phi*math.Pi/180)*math.Sin(theta*math.Pi/180)
					// fmt.Println("///////////////////", phiP, phiR)
					w := cmplx.Exp(complex(0, 2*math.Pi*(float64(n-1)*vspace*phiP)))
					v := cmplx.Exp(complex(0, 2*math.Pi*(float64(m-1)*hspace*phiR)))
					// fmt.Println("W:", w, "V", v)
					sum = sum + w*v
				}
			}
			/////////////
			txRUGains := vlib.NewMatrixF(ant.BSAntennaConfig[5], ant.BSAntennaConfig[6])
			for k := 0; k < ant.BSAntennaConfig[5]; k++ {
				for l := 0; l < ant.BSAntennaConfig[6]; l++ {
					txRUGains[k][l] = 10*math.Log10(c*math.Pow(cmplx.Abs(sum), 2)) + ag // Composite Beam Gain + Antenna Element Gain
					temp := txRUGains[k][l]
					if maxgain < temp {
						maxgain = temp
						bestBeamID = beamid
					}
				}
			}
			aag[beamid] = txRUGains
			//////////
		}
	}
	return aag, bestBeamID, Az, El
}

// func DFTCombPatternDb(theta, phi float64, ant SettingAAS) (aag map[int]vlib.MatrixF, bestBeamID int, Az, El float64) {

// 	theta = Wrap180To180(theta)
// 	phi = Wrap0To180(phi)
// 	var ag float64
// 	Az, El, ag = BSPatternDb(theta, phi)
// 	// fmt.Println("Antenna Element Gain:", ag)
// 	hspace := ant.ESpacingHFactor
// 	vspace := ant.ESpacingVFactor
// 	dtilt := ant.ElectronicTilt // degree
// 	descan := ant.Dscan         //degree
// 	// var maxgain float64 // Maximum BeamGain out of the Beam Set
// 	var sum = complex(0.0, 0.0)
// 	nv := ant.BSAntennaConfig[0] / ant.BSAntennaConfig[5]
// 	nh := ant.BSAntennaConfig[1] / ant.BSAntennaConfig[6]
// 	var maxgain float64
// 	bestBeamID = 0
// 	// var aag map[int]vlib.MatrixF
// 	nbeams := len(ant.Dscan) * len(ant.ElectronicTilt)
// 	aag = make(map[int]vlib.MatrixF, nbeams)
// 	/////
// 	c := 1.0 / float64(nv*nh)
// 	for i := 0; i < len(dtilt); i++ { //  dtilt is a vector of Zenith Angles of the Beam Set
// 		for j := 0; j < len(descan); j++ { // descan is a vector of Azimuth Angles of the Beam Set
// 			beamid := j + len(descan)*i
// 			for m := 1; m <= nv; m++ {
// 				for n := 1; n <= nh; n++ {
// 					w := complex(1/math.Pow(float64(nv), 1/2), 0) * cmplx.Exp(complex(0, -1*2*math.Pi*(float64(n-1)*vspace*math.Cos((math.Pi/180)*(dtilt[i]-phi)))))
// 					v := complex(1/math.Pow(float64(nh), 1/2), 0) * cmplx.Exp(complex(0, -1*2*math.Pi*(float64(m-1)*hspace*math.Sin((math.Pi/180)*(dtilt[i]-phi))*math.Sin((math.Pi/180)*(descan[j]-theta)))))
// 					sum = sum + w*v
// 				}
// 			}
// 			/////////////
// 			txRUGains := vlib.NewMatrixF(ant.BSAntennaConfig[5], ant.BSAntennaConfig[6])
// 			for k := 0; k < ant.BSAntennaConfig[5]; k++ {
// 				for l := 0; l < ant.BSAntennaConfig[6]; l++ {
// 					txRUGains[k][l] = 10*math.Log10(c*math.Pow(cmplx.Abs(sum), 2)) + ag // Composite Beam Gain + Antenna Element Gain
// 					temp := txRUGains[k][l]
// 					if maxgain < temp {
// 						maxgain = temp
// 						bestBeamID = beamid
// 					}
// 				}
// 			}
// 			aag[beamid] = txRUGains
// 			//////////
// 		}
// 	}
// 	return aag, bestBeamID, Az, El
// }

// func txBeamGain(ant SettingAAS, dtilt, descan float64, nv, nh int) (sum float64) {
// 	c := 1.0 / float64(nv*nh)
// 	for i := 0; i < len(dtilt); i++ { //  dtilt is a vector of Zenith Angles of the Beam Set
// 		for j := 0; j < len(descan); j++ { // descan is a vector of Azimuth Angles of the Beam Set
// 			for m := 1; m < nv; m++ {
// 				for n := 1; n < nh; n++ {
// 					phiP := -math.Cos(dtilt[i]*math.Pi/180) + math.Cos(phi*math.Pi/180)
// 					phiR := -math.Sin(dtilt[i]*math.Pi/180)*math.Sin(descan[j]*math.Pi/180) + math.Sin(phi*math.Pi/180)*math.Sin(theta*math.Pi/180)
// 					w := cmplx.Exp(complex(0, 2*math.Pi*(float64(n-1)*vspace*phiP)))
// 					v := cmplx.Exp(complex(0, 2*math.Pi*(float64(m-1)*hspace*phiR)))
// 					sum = sum + w*v
// 				}
// 			}
// 		}
// 	}
// 	return sum
// }

// //CombPatternDb calculates combine antenna pattern gain
// func CombPatternDb(theta, phi, Ag, dtilt float64, Nv, Nh int) (az, el, Aa, result, old float64) {
// 	phi = Wrap0To180(phi)
// 	theta = Wrap180To180(theta)
// 	hspace := 0.5
// 	vspace := 0.8
// 	dtilt = 100   // degree
// 	descan := 0.0 //degree
// 	var sum = complex(0.0, 0.0)

// 	for m := 0; m < Nh; m++ {
// 		for n := 0; n < Nv; n++ {
// 			w := complex(1/math.Pow(float64(Nh*Nv), 1/2), 0) * cmplx.Exp(complex(0, 2*math.Pi*(float64(n)*vspace*math.Sin((math.Pi/180)*dtilt)-float64(m)*hspace*math.Cos((math.Pi/180)*dtilt)*math.Sin((math.Pi/180)*descan))))
// 			v := cmplx.Exp(complex(0, 2*math.Pi*(float64(n)*vspace*math.Cos((math.Pi/180)*phi)+float64(m)*hspace*math.Sin((math.Pi/180)*phi)*math.Sin((math.Pi/180)*theta))))
// 			sum = sum + w*v
// 		}
// 	}

// 	result = 10 * math.Log10(math.Pow(cmplx.Abs(sum), 2))
// 	az = theta
// 	el = phi
// 	Aa = Ag + result
// 	return az, el, Aa, result, Ag
// }

// //Analogy Beamforming antenna array gain
// func AnalogBeamDb(theta, phi float64, nv, nh int) (az, el, Aa float64, result vlib.VectorF, old float64) {
// 	pi := math.Pi
// 	azimuth := vlib.VectorF{-5 * pi / 16, -3 * pi / 16, -pi / 16, pi / 16, 3 * pi / 16, 5 * pi / 16}
// 	azimuth = azimuth.Scale(180 / pi)
// 	zenith := vlib.VectorF{5 * pi / 8, 7 * pi / 8}
// 	zenith = zenith.Scale(180 / pi)
// 	phi = Wrap0To180(phi)
// 	// theta = Wrap180To180(theta)
// 	reaz := azimuth.Sub(-theta)
// 	aasgainDB := vlib.NewVectorF(0)
// 	// fmt.Println("Start")
// 	// _, _, aasgain := BSPatternDb(theta, phi)
// 	for j := 0; j < azimuth.Len(); j++ {
// 		for i := 0; i < zenith.Len(); i++ {
// 			reel := phi - (zenith[i])

// 			_, _, aasgain := BSPatternDb(reaz[j], reel)
// 			// aasgain = 0.0
// 			// _, _, aas, _, _ := UMaCombPatternDb(reaz[j], azimuth[j], zenith[i], reel, aasgain, 4, 8)
// 			_, _, aas, _, _ := UMaCombPatternDb(azimuth[j], azimuth[j], zenith[i], zenith[i], aasgain, 4, 8)
// 			// fmt.Println("azimuth: ", reaz[j], "zenith: ", reel, "aas: ", aas, "aasgainDb: ", aasgain, "composite aas: ", aas-aasgain)
// 			aasgainDB.AppendAtEnd(aas)
// 		}
// 	}
// 	// fmt.Println("End")
// 	aas := vlib.Max(aasgainDB)
// 	return theta, phi, aas, aasgainDB, 0
// }

// //CombPatternDb calculates combine antenna pattern gain
// func UMaCombPatternDb(theta, descan, zenith, phi, Ag float64, Nv, Nh int) (az, el, Aa, result, old float64) {
// 	hspace := 0.5
// 	vspace := 0.8
// 	phi = Wrap0To180(phi)
// 	theta = Wrap180To180(theta)
// 	// dtilt := zenith // degree
// 	var sum = complex(0.0, 0.0)

// 	for m := 0; m < Nh; m++ {
// 		for n := 0; n < Nv; n++ {
// 			w := complex(1/math.Pow(float64(Nv), 1/2), 0) * cmplx.Exp(complex(0, -1*2*math.Pi*(float64(n)*vspace*math.Cos((math.Pi/180)*phi))))
// 			v := complex(1/math.Pow(float64(Nh), 1/2), 0) * cmplx.Exp(complex(0, -1*2*math.Pi*(float64(m)*hspace*math.Sin((math.Pi/180)*phi)*math.Sin((math.Pi/180)*theta))))
// 			sum = sum + w*v
// 		}
// 	}

// 	result = 10 * math.Log10(math.Pow(cmplx.Abs(sum), 2))
// 	az = theta
// 	el = phi
// 	Aa = Ag + result
// 	return az, el, Aa, result, Ag
// }

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
