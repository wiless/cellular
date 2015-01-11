package main

import (
	"fmt"
	"github.com/wiless/cellular/channel"
	"github.com/wiless/vlib"
	"math/rand"
	// "os"

	"time"
)

var matlab *vlib.Matlab

func init() {
	matlab = vlib.NewMatlab("channel")
	matlab.Silent = true
	matlab.Json = true
	rand.Seed(time.Now().Unix())

}

func main() {
	// var result []cell.LinkMetric
	// vlib.LoadStructure("linkmetric2.json", &result)
	var sisochannel channel.Channel
	sisochannel.CreateFromFile("linkmetric2.json")
	sisochannel.Init()
	// CreateChannelLinks()

	// w, _ := os.Create("dump.txt")
	// for idx, val := range result {
	// 	fmt.Fprintf(w, "\n %d :  %#v\n", idx, val)
	// }

	matlab.Close()
	fmt.Println("\n")
}
