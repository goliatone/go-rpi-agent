package main 

import (
	"io/ioutil"
	"regexp"
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