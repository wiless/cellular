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

//CombPatternDb calculates combine antenna pattern gain
func CombPatternDb(theta, phi, Ag, dtilt float64, Nv, Nh int) (az, el, Aa, result, old float64) {
	phi = Wrap0To180(phi)
	theta = Wrap180To180(theta)
	hspace := 0.5
	vspace := 0.8
	dtilt = 100   // degree
	descan := 0.0 //degree
	var sum = complex(0.0, 0.0)

	for m := 0; m < Nh; m++ {
		for n := 0; n < Nv; n++ {
			w := complex(1/math.Pow(float64(Nh*Nv), 1/2), 0) * cmplx.Exp(complex(0, 2*math.Pi*(float64(n)*vspace*math.Sin((math.Pi/180)*dtilt)-float64(m)*hspace*math.Cos((math.Pi/180)*dtilt)*math.Sin((math.Pi/180)*descan))))
			v := cmplx.Exp(complex(0, 2*math.Pi*(float64(n)*vspace*math.Cos((math.Pi/180)*phi)+float64(m)*hspace*math.Sin((math.Pi/180)*phi)*math.Sin((math.Pi/180)*theta))))
			sum = sum + w*v
		}
	}

	result = 10 * math.Log10(math.Pow(cmplx.Abs(sum), 2))
	az = theta
	el = phi
	Aa = Ag + result
	return az, el, Aa, result, Ag
}

//Analogy Beamforming antenna array gain
func AnalogBeamDb(theta, phi float64, nv, nh int) (az, el, Aa float64, result vlib.VectorF, old float64) {
	pi := math.Pi
	azimuth := vlib.VectorF{-5 * pi / 16, -3 * pi / 16, -pi / 16, pi / 16, 3 * pi / 16, 5 * pi / 16}
	azimuth = azimuth.Scale(180 / pi)
	zenith := vlib.VectorF{5 * pi / 8, 7 * pi / 8}
	zenith = zenith.Scale(180 / pi)
	phi = Wrap0To180(phi)
	// theta = Wrap180To180(theta)
	reaz := azimuth.Sub(-theta)
	aasgainDB := vlib.NewVectorF(0)
	// fmt.Println("Start")
	// _, _, aasgain := BSPatternDb(theta, phi)
	for j := 0; j < azimuth.Len(); j++ {
		for i := 0; i < zenith.Len(); i++ {
			reel := phi - (zenith[i])

			_, _, aasgain := BSPatternDb(reaz[j], reel)
			// aasgain = 0.0
			// _, _, aas, _, _ := UMaCombPatternDb(reaz[j], azimuth[j], zenith[i], reel, aasgain, 4, 8)
			_, _, aas, _, _ := UMaCombPatternDb(azimuth[j], azimuth[j], zenith[i], zenith[i], aasgain, 4, 8)
			// fmt.Println("azimuth: ", reaz[j], "zenith: ", reel, "aas: ", aas, "aasgainDb: ", aasgain, "composite aas: ", aas-aasgain)
			aasgainDB.AppendAtEnd(aas)
		}
	}
	// fmt.Println("End")
	aas := vlib.Max(aasgainDB)
	return theta, phi, aas, aasgainDB, 0
}

//CombPatternDb calculates combine antenna pattern gain
func UMaCombPatternDb(theta, descan, zenith, phi, Ag float64, Nv, Nh int) (az, el, Aa, result, old float64) {
	hspace := 0.5
	vspace := 0.8
	phi = Wrap0To180(phi)
	theta = Wrap180To180(theta)
	// dtilt := zenith // degree
	var sum = complex(0.0, 0.0)

	for m := 0; m < Nh; m++ {
		for n := 0; n < Nv; n++ {
			w := complex(1/math.Pow(float64(Nv), 1/2), 0) * cmplx.Exp(complex(0, -1*2*math.Pi*(float64(n)*vspace*math.Cos((math.Pi/180)*phi))))
			v := complex(1/math.Pow(float64(Nh), 1/2), 0) * cmplx.Exp(complex(0, -1*2*math.Pi*(float64(m)*hspace*math.Sin((math.Pi/180)*phi)*math.Sin((math.Pi/180)*theta))))
			sum = sum + w*v
		}
	}

	result = 10 * math.Log10(math.Pow(cmplx.Abs(sum), 2))
	az = theta
	el = phi
	Aa = Ag + result
	return az, el, Aa, result, Ag
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
