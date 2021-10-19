package main 

import (
	"io/ioutil"
	"strings"
)

// Registrable is the name to lookup after loading the plugin for the module registering
var Registrable registrable

type registrable int

//Identifier are different key value pairs that help identify this device
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

	mac, err := getMac("wlan0")
	if err != nil {
		return err
	}
	
	MacWlan0 := Identifier{"mac48", mac, "MAC 48 Ethernet interface wlan0"}
	data["Interfaces"] = append(data["Interfaces"].([]Identifier), MacWlan0)

	mac, err = getMac("eth0")

	if err != nil {
		return err
	}

	MacEth0 := Identifier{"mac48", mac, "MAC 48 Ethernet interface eth0"}
	
	data["Interfaces"] = append(data["Interfaces"].([]Identifier), MacEth0)

	return nil
}


func getMac(iface string) (string, error) {
	filepath := strings.Replace("/sys/class/net/#iface#/address", "#iface#", iface, -1)
	contents, err := ioutil.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	mac := strings.Replace(string(contents), "\n", "", -1)
	return mac, nil
}
