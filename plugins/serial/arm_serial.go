package main 

import (
	"os"
	"io/ioutil"
	"regexp"
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

  	serial , err := getSerial()
  	if err != nil {
		return err
  	}

  	Serial := Identifier{"serial", serial, "Serial Number"}
  	data["Interfaces"] = append(data["Interfaces"].([]Identifier), Serial)	
  
  	return nil
}

func getSerial() (string, error) {
	contents, err := ioutil.ReadFile("/proc/cpuinfo")
	if err != nil {
		return "", err
	}
	includeRegex, err := regexp.Compile(`Serial\s+:\s(\w+)`)
	if err != nil {
		return "", err
	}

	if includeRegex.Match(contents) {
		includeFile := includeRegex.FindStringSubmatch(string(contents))
		if len(includeFile) == 2 {
			return includeFile[1], nil
		}
	}

	return "", nil
}