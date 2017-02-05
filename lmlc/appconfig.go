package main

import (
	"log"

	"github.com/spf13/viper"
)

//AppConfig  Struct for the app parameteres
type AppConfig struct {
	CellRadius float64
	TxPowerDbm float64
}

// ReadAppConfig reads all the configuration for the app
func ReadAppConfig() {
	log.Println(viper.AllSettings())

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

	// Load from the external configuration files
	CellRadius = viper.GetFloat64("CellRadius")
	TxPowerDbm = viper.GetFloat64("TxpowerDBm")

	log.Println(CellRadius)
	log.Println(TxPowerDbm)

}
