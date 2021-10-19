package main 

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
	return "1567738079529", nil
}