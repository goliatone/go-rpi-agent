package main 

import (
	"os"
)

// Registrable is the name to lookup after loading the plugin for the module registering
var Registrable registrable

type registrable int

//Identifier should be imported from main package
type Identifier struct {
	Name string
	Value string
	Description string
}

//AddMeta ...
func (p *registrable) AddMeta(data map[string]interface{}) error {
	if _, ok := data["Interfaces"].([]Identifier); !ok {
		i := []Identifier{}
	  data["Interfaces"] = i
	}

	host, err := os.Hostname()

	if err != nil {
		return err
	}

	Host := Identifier{"host", host, "Hostname"}
	data["Interfaces"] = append(data["Interfaces"].([]Identifier), Host)

	return nil
}