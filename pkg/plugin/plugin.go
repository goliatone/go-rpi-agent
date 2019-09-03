package plugin

import (
	"plugin"
	"strings"
	"fmt"
	"path"
	"path/filepath"
)

//Plugin interface for plugin symbols lookup
type Plugin interface {
	Lookup(name string) (plugin.Symbol, error)
}

//Options options for loader
type Options struct {
	Directory string 
	Pattern string
}

//Load will load all plugins in a given location
func Load(opts Options, reg *Registry) (int, error) {
	pattern := path.Join(opts.Directory, opts.Pattern)
	
	plugins, err := Glob(pattern)
	if err != nil {
		return 0, err
	}
	return load(plugins, reg)
}

//Glob will look for the contents of a directory and load plugin
//fils
func Glob(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}

func load(plugins []string, reg *Registry) (int, error) {
	errors := []error{}

	loaded := 0
	for k, pluginName := range plugins {
		if err := open(pluginName, reg); err != nil {
			errors = append(errors, makeError(k, pluginName, err))
			continue
		}
		loaded++
	}

	if len(errors) > 0 {
		return loaded, loaderError{errors}
	}
	return loaded, nil
}

func open(name string, reg *Registry) (err error) {

	defer func() {
		if r := recover(); r != nil {
			var ok bool 
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
		}
	}()

	var p Plugin 
	p, err = openPlugin(name)
	if err != nil {
		return 
	}

	err = reg.Register(p)

	return
}

var openPlugin = defaultOpenPlugin

func defaultOpenPlugin(name string) (Plugin, error) {
	return plugin.Open(name)
}



func makeError(i int, name string, err error) error {
	return fmt.Errorf("opening plugin %d (%s): %s", i, name, err.Error())
}

type loaderError struct {
	errors []error
}

func (l loaderError) Error() string {
	m := make([]string, len(l.errors))
	for i,err := range l.errors {
		m[i] = err.Error()
	}
	return fmt.Sprintf("Error loading plugins: \n%s",  strings.Join(m, "\n"))
}
