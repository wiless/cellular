package main

import (
	// "encoding/json"
	// "fmt"
	"github.com/wiless/cellular/antenna"
)

func main() {
	var setting antenna.SettingAAS

	setting.SetDefault()
	setting.N = 5
	setting.Freq = 2e9
	setting.BeamTilt = 0
	setting.DisableBeamTit = false
	setting.VTiltAngle = 0
	setting.MfileName = "output.m"
	setting.HoldOn = true
	antenna.RunAAS(setting)

}
