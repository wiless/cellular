package antenna

import (
	"math"

	log "github.com/Sirupsen/logrus"
)

// BSPatternDb generates the antenna gain for given theta,phi in degree
// based on Table 8-6 in Report ITU-R M.2412
func BSPatternDb(theta, phi, gain float64) float64 {
	if !(phi >= 0 && phi <= 180) {
		log.Warnf("Elevation beyond range")
	}
	if !(theta >= -180 && theta <= 180) {
		log.Warnf("Azimuth beyond range")
	}

	MaxGaindBi := 8.0 //
	_ = MaxGaindBi
	theta3dB := 65.0 // degree
	SLAv := 30.0
	Am := 30.0
	Av := -math.Min(12.0*math.Pow((phi-90)/theta3dB, 2.0), SLAv)
	Ah := math.Min(-12.0*math.Pow(theta/theta3dB, 2.0), Am)
	result := -math.Min(-math.Floor(Av+Ah), Am) + MaxGaindBi
	result = Ah
	return result
}

// BSPatternIndoorHS_Db generates the antenna gain for given theta,phi
// based on Table 8-7 in Report ITU-R M.2412
func BSPatternIndoorHS_Db(theta, phi, gain float64) float64 {
	MaxGaindBi := 5.0 //
	_ = MaxGaindBi
	theta3dB := 90.0 // degree
	SLAv := 25.0
	Am := 25.0
	Avr := -math.Min(12.0*math.Pow((phi-90)/theta3dB, 2.0), SLAv)
	Avh := -math.Min(12.0*math.Pow(theta/theta3dB, 2.0), Am)
	result := -math.Min(-math.Floor(Avr+Avh), Am) + MaxGaindBi
	return result
}

// UEPatternDb generates the antenna gain for given theta,phi
// based OMNI Directional gain..
func UEPatternOmniDb(theta, phi, gain float64) float64 {
	return gain
}

// UEPatternDb generates the antenna gain for given theta,phi
// based Table 8-8 in Report ITU-R M.2412 for fGHZ  > 4GHz i.e 30GHz & 70GHz
func UEPatternDb(theta, phi, gain float64) float64 {
	MaxGaindBi := 5.0 //
	_ = MaxGaindBi
	theta3dB := 90.0 // degree
	SLAv := 25.0
	Am := 25.0
	Avr := -math.Min(12.0*math.Pow((phi-90)/theta3dB, 2.0), SLAv)
	Avh := -math.Min(12.0*math.Pow(theta/theta3dB, 2.0), Am)
	result := -math.Min(-math.Floor(Avr+Avh), Am) + MaxGaindBi
	return result
}
