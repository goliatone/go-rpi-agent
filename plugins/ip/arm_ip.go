package main 

import (
	"net"
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