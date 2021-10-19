package main 

import (
	"net"
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

	inter, err := net.Interfaces()
	if err != nil {
		return  err
	}

	for _, ifa := range inter {
		addrs, err := ifa.Addrs()

		if err != nil {
			return err
		}

		for _, a := range addrs {
			if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					IPAddress := Identifier{"ip", ipnet.IP.String(), "ip_"+string(ifa.Name)}	
					data["Interfaces"] = append(data["Interfaces"].([]Identifier), IPAddress)
				}
			}
		}
	}

	return nil
}
