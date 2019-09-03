// +build darwin
package main 

import (
	"os"
	"fmt"
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

	fmt.Println("Hello plugin world!")
	host, err := os.Hostname()

	if err != nil {
		return err
	}

	Host := Identifier{"host", host + ".local", "Hostname"}
	data["Interfaces"] = append(data["Interfaces"].([]Identifier), Host)

	return nil
}
