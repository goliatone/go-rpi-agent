package main 

import (
	"net"
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
	
	ifs, _ := net.Interfaces()
	out := ""
 	for _, v := range ifs {
		if v.Name != iface {
			continue
		}

    	h := v.HardwareAddr.String()
    	if len(h) == 0 {
        	continue
     	}
		out = h
 	}
	return out, nil
}
