package config

import (
	"github.com/magiconair/properties"
	"log"
	"os"
)

func LoadProperties() *properties.Properties {
	myDir, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	log.Println(myDir)

	propertiesInstance := properties.MustLoadFile(myDir+"/resources/service.properties", properties.UTF8)
	return propertiesInstance
}
