package antenna

import (
	"math"
	cmplx "math/cmplx"
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
func BSPatternDb(theta, phi float64) (az, el, Ag float64) {
	phi = Wrap0To180(phi)
	theta = Wrap180To180(theta)
	MaxGaindBi := 8.0 //
	theta3dB := 65.0  // degree
	SLAmax := 30.0
	Am := SLAmax
	Ah := -math.Min(12.0*math.Pow(theta/theta3dB, 2.0), Am)

	MechTiltGCS := 90.0 // Pointing to Horizon..axis..
	Av := -math.Min(12.0*math.Pow((phi-MechTiltGCS)/theta3dB, 2.0), SLAmax)
	result := -math.Min(-(Av + Ah), Am)
	//result = Ah
	az = theta
	el = phi
	Ag = result + MaxGaindBi
	return az, el, Ag
}

//CombPatternDb calculates combine antenna pattern gain
func CombPatternDb(theta, phi, Ag float64, Nv, Nh int) (az, el, Aa, result, old float64) {
	phi = Wrap0To180(phi)
	theta = Wrap180To180(theta)
	hspace := 0.5
	vspace := 0.9
	dtilt := 9.0   // degree
	descan := 25.0 //degree
	var sum = complex(0.0, 0.0)
	//var sum = 0.0

	for m := 1; m < Nh; m++ {
		for n := 1; n < Nv; n++ { //Nv rows Nh columns
			//w := (1 / math.Pow(float64(Nh*Nv), 0.5)) * math.Exp((math.Pow(1, 0.5))*2*math.Pi*(float64(n-1)*vspace*math.Sin(dtilt*math.Pi/180)-float64(m-1)*hspace*math.Cos(dtilt*math.Pi/180)*math.Sin(descan*math.Pi/180)))
			//v := math.Exp((math.Pow(1, 0.5)) * 2 * math.Pi * (float64(n-1)*vspace*math.Cos(phi*math.Pi/180) + float64(m-1)*hspace*math.Sin(phi*math.Pi/180)*math.Sin(theta*math.Pi/180)))

			w := complex(1/math.Pow(float64(Nh*Nv), 1/2), 0) * cmplx.Exp(complex(0, 2*math.Pi*(float64(n-1)*vspace*math.Sin(dtilt*math.Pi/180)-float64(m-1)*hspace*math.Cos(dtilt*math.Pi/180)*math.Sin(descan*math.Pi/180))))
			v := cmplx.Exp(complex(0, 2*math.Pi*(float64(n-1)*vspace*math.Cos(phi*math.Pi/180)+float64(m-1)*hspace*math.Sin(phi*math.Pi/180)*math.Sin(theta*math.Pi/180))))

			sum = sum + w*v
			//sum = sum + w*v

		}

	}

	result = 10 * math.Log10(math.Pow(cmplx.Abs(sum), 2))
	//result = 10 * math.Log10(math.Pow(math.Abs(sum), 2))

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
	MaxGaindBi := 5.0 //
	theta3dB := 90.0  // degree
	SLAmax := 25.0
	Am := 25.0
	MechTiltGCS := 90.0 /// Need to be set to 180.. pointing to Ground, when Antenna mounted on ceiling..

	Ah := -math.Min(12.0*math.Pow(theta/theta3dB, 2.0), Am)
	Av := -math.Min(12.0*math.Pow((phi-MechTiltGCS)/theta3dB, 2.0), SLAmax)
	result := -math.Min(-(Av+Ah), Am) + MaxGaindBi

	az = theta
	el = phi
	Ag = result
	return az, el, Ag

}

// UEPatternDb generates the antenna gain for given theta,phi
// based OMNI Directional gain..
func UEPatternOmniDb(theta, phi, gain float64) float64 {
	return gain
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
