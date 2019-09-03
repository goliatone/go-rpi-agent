package main 

import (
	"os"
)

//Metadata interface
type Metadata map[string]interface{}

//Identifier are different key value pairs that help identify this device
type Identifier struct {
	Name string
	Value string
	Description string
}

//MetadataPlugin is a simple interface that defines
//the AddMeta and AddStatus methods
type MetadataPlugin struct {}

//AddMeta ...
func (m MetadataPlugin) AddMeta(data Metadata) error {
	if _, ok := data["Interfaces"].([]Identifier); !ok {
		i := []Identifier{}
	  data["Interfaces"] = i
	}
	
	host, err := os.Hostname()

	if err != nil {
		return err
	}

	Host := Identifier{"host", host + ".local", "Hostname"}
	data["Interfaces"] = append(data["Interfaces"].([]Identifier), Host)

	return nil
}