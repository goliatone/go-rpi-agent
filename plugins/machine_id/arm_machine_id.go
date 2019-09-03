package main 

import (
	"os"
	"io/ioutil"
	"strings"
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
