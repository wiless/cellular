package main

import (
	"fmt"
	"github.com/wiless/vlib"
	"math/rand"
	"os"

	"time"
)

var matlab *vlib.Matlab

// Dimension
// Outer Diameter : 283.01887m = 141.50944
// Inner Diameter : 174.5283m = 87.26415
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

func init() {
	matlab = vlib.NewMatlab("channel")
	matlab.Silent = true
	matlab.Json = true
	rand.Seed(time.Now().Unix())

}

type DataType struct {
	UE         int `json:"ObjectID"`
	LinkMetric []struct {
		RxNodeID     int
		FreqInGHz    float32
		BandwidthMHz int
		N0           float32
		TxNodeIDs    []int
		TxNodesRSRP  []float32
		RSSI         float32
		BestRSRP     float32
		BestRSRPNode int
		BestSINR     float32
		RoIDbm       int
	} `json:"Object"`
}

func main() {
	// type MFNMetric []LinkMetric

	// MetricPerRx := make(map[int]MFNMetric)
	var result []DataType
	// result = make([]DataType)
	vlib.LoadStructure("linkmetric.json", &result)
	// CreateChannelLinks()
	w, _ := os.Create("dump.txt")
	for idx, val := range result {
		fmt.Fprintf(w, "\n %d :  %#v\n", idx, val)
	}

	matlab.Close()
	fmt.Println("\n")
}
