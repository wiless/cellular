package main

import (
	"log"

	"github.com/spf13/viper"
	"github.com/wiless/vlib"
)

//AppConfig  Struct for the app parameteres
type AppConfig struct {
	CellRadius float64
	TxPowerDbm float64
}

var C AppConfig

// ReadAppConfig reads all the configuration for the app
func ReadAppConfig() {
	log.Print("Reading APP config ")
	viper.AddConfigPath(indir)
	viper.SetConfigName("config")

	err := viper.ReadInConfig()
	if err != nil {
		log.Print("ReadInConfig ", err)
	}
	// Set all the default values
	{
		viper.SetDefault("TxPowerDbm", TxPowerDbm)
		viper.SetDefault("CellRadius", CellRadius)
	}
	err = viper.Unmarshal(&C)
	if err == nil {
		log.Print("Error unmarshalling ", err)
	}
	log.Printf("%#v ", C)
	// Load from the external configuration files
	CellRadius = viper.GetFloat64("CellRadius")
	TxPowerDbm = viper.GetFloat64("TxpowerDBm")

	SwitchOutput()
	vlib.SaveStructure(C, "OutputSetting.json", true)
	SwitchBack()

}
