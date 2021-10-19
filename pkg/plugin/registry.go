package plugin

import (
	"fmt"
)

//RegistrableVar key we use to look up our plugins
var RegistrableVar = "Registrable"

// MetadataPlugin defines plugin interface
type MetadataPlugin interface {
	AddMeta(data map[string]interface{}) error
}

//Registry struct holds our registry
type Registry struct {
	Plugins []MetadataPlugin
}

//NewRegistry creates a new Registry instance
func NewRegistry() *Registry {
	return &Registry{
		Plugins: []MetadataPlugin{},
	}
}

//Register registers a new plugin
func (r *Registry) Register(p Plugin) error {
	m, err := p.Lookup(RegistrableVar)
	if err != nil {
		fmt.Println("unable to find the registration symbol", err.Error())
		return err
	}

	if metaPlugin, ok := m.(MetadataPlugin); ok {
		r.Plugins = append(r.Plugins, metaPlugin)
	}

	return nil
} 