package simulation

import (
	"github.com/wiless/vlib"
	"math"
	"math/cmplx"
)

type GenericStruct map[string]interface{}

func GetEJtheta(degree float64) complex128 {
	return cmplx.Exp(complex(0.0, -degree*math.Pi/180.0))
}

func Radian(degree float64) float64 {
	return degree * math.Pi / 180.0
}

type LinkInfo struct {
	RxNodeID          int
	NodeTypes         []string
	LinkGain          vlib.VectorF
	LinkGainNode      vlib.VectorI
	InterferenceLinks vlib.VectorF
}

type LinkMetric struct {
	RxNodeID     int
	FreqInGHz    float64
	BandwidthMHz float64
	N0           float64
	TxNodeIDs    vlib.VectorI
	TxNodesRSRP  vlib.VectorF
	RSSI         float64
	BestRSRP     float64
	BestRSRPNode int
	BestSINR     float64
	RoIDbm       float64
}
