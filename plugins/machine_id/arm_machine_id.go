package main 

import (
	"io/ioutil"
	"strings"
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

	machineID, err := getMachineID()
	
	if err != nil {
		return err
	}

	MachineID := Identifier{"machine-id", machineID, "dbus machine-id"}
	data["Interfaces"] = append(data["Interfaces"].([]Identifier), MachineID)	

	return nil
}


//machine-id is found in either /etc/machine-id or
// /var/lib/dbus/machine-id 
func getMachineID() (string, error) {

	readMachineID := func (filepath string) (string, error) {
		contents, err := ioutil.ReadFile(filepath)	
		if err != nil {
			return "", err
		}
		s := string(contents)
		s = strings.TrimSpace(s)
		return s, nil
	}

	contents, err := readMachineID("/var/lib/dbus/machine-id")
	
	if err != nil {
		return "", err
	}

	if contents != "" {
		return contents, nil
	}

	return readMachineID("/etc/machine-id")
}
