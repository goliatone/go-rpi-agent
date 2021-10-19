package main

import (
	"plugin"
	"path/filepath"
	"path"
	"bytes"
	"flag"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"regexp"
	"strings"
	"time"
	"text/template"

	"github.com/grandcat/zeroconf"
	"github.com/twinj/uuid"
	// "github.com/goliatone/plugin"
)

//TODO: make configurable
const (
	uuidFile = "/usr/local/src/rpi-agent/metadata/.device_uuid"
)

//VERSION Generated via ld flags
var VERSION string

//BUILD_DATE Generated via ld flags
var BUILD_DATE string

var (
	name   	 	= flag.String("name", "rpi-agent", "Set the agents name, default is rpi-agent.")
	domain  	= flag.String("domain", "local", "Set the search domain. For local networks, default is fine.")
	service		= flag.String("service", "_rpi._tcp", "Set the service category to look for devices.")
	port    	= flag.Int("port", 8080, "Service port.")
	registryURL = flag.String("registry", "", "Registry service URL.")
	tplPath		= flag.String("registry-tpl", "/opt/rpi-agent/templates/default.tpl.json", "Path to registry payload template.")
	showVersion = flag.Bool("version", false, "Show version")
)

//TODO: Have these in separate file
//Metadata holds data we will send over to the registry
type Metadata map[string]interface{}

//Identifier are different key value pairs that help identify this device
type Identifier struct {
	Name string
	Value string
	Description string
}

const REGISTRABLE_VAR = "Registrable"

// MetadataPlugin defines plugin interface
type MetadataPlugin interface {
	AddMeta(data map[string]interface{}) error
}

var plugins []MetadataPlugin

// var PluginRegistry Registry

func init() {
	// PluginRegistry = NewRegistry()
	// Load()
	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", "rpi-agent")
		flag.PrintDefaults()
		fmt.Printf("Version: %s, Build: %s\n", VERSION, BUILD_DATE)
	}


	// data := make(Metadata)

	f, err := filepath.Glob("./plugins/*.so")
	if err != nil {
		fmt.Println("panic here")
		panic(err)
	}
	for _, filename := range(f) {
		fmt.Printf("loading plugin %s\n", filename)
		p, err := plugin.Open(filename)
		if err != nil {
			panic(err)
		}
		pMeta, err := p.Lookup(REGISTRABLE_VAR)
		if err != nil {
			fmt.Printf("not registrable %s\n", filename)
			panic(err)
		}

		if metaPlugin, ok := pMeta.(MetadataPlugin); ok {
			// err = metaPlugin.RegisterDecoder(r.Decoder.Register)
			fmt.Printf("adding plugin to plugin list %s\n", filename)
			plugins = append(plugins, metaPlugin)	
			// metaPlugin.AddMeta(data)
		}
	}

	// fmt.Printf("%v", data)
	// os.Exit(0)
}

func main() {
	flag.Parse()

	if *showVersion == true {
		log.Printf("version %s build %s", VERSION, BUILD_DATE)
		os.Exit(0)
	}

	//TODO: Take parameters for name, service...
	deviceUUID := getDefaultUUID()
	fmt.Println("device uuid: " + deviceUUID)

	// Start out http service
	go startService(deviceUUID)

	// Extra information about our service
	txtRecord := []string{
		"version=" + VERSION,
		"build_date=" + BUILD_DATE,
		"uuid=" + deviceUUID,
	}

	n := *name
	s := *service
	d := *domain
	service, err := zeroconf.Register(
		n, // service instance name
		s,  // service type and protocol
		fmt.Sprintf("%s.", d),     // service domain
		*port,         // service port
		txtRecord,    // service metadata
		nil,          // register on all network interfaces
	)

	//Seconds, default value is 3200
	service.TTL(3200)

	defer service.Shutdown()

	if err != nil {
		log.Fatal(err)
	}

	//Clean exit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	select {
	case <- sig:
	}
	log.Println("Shutting down service...")
}

// HTTP service to introspect RPi instance
func startService(deviceUUID string) {

	meta := make(Metadata)
	meta["Uuid"] = deviceUUID
	// meta["Name"] = getNameFromHostname(host, "")

	err := addHost(meta)
	handleError(err, "Error AddHost: ")

	err = addIps(meta)
	handleError(err, "Error GetAddress: ")

	err = addSerial(meta)
	handleError(err, "Error GetSerial: ")

	err = addMachineID(meta)
	handleError(err, "Error GetMachineID: ")

	err = addMac(meta)
	handleError(err, "Error AddMac: ")

	status := make(map[string]interface{})

	startTime := time.Now()
	status["agent_start"] = startTime.Format(time.RFC3339)
	status["agent_uptime"] = time.Since(startTime)
	
	status["connectivity"] = "online"

	meta["Status"] = status

	if registryURL != nil {
		registerAgent(*registryURL, meta)
	}

	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		//Update our uptime
		status["agent_uptime"] = time.Now().Sub(startTime)

		n := path.Base(*tplPath)
		t := template.New(n)
		tpl, err := t.ParseFiles(*tplPath)
	
		if err != nil {
			log.Fatal("Parse: ", err)
			return 
		}

		var output bytes.Buffer 
		if err = tpl.Execute(&output, meta); err != nil {
			log.Fatal("FATAL registerAgent: ", err)
			return
		}

		rw.Header().Set("Content-Type", "application/json")
		output.WriteTo(rw)
		// json.NewEncoder(rw).Encode(output)
	})

	//Some meta info...
	host, err := os.Hostname()
	p := fmt.Sprintf(":%d", *port)
	log.Println("Starting HTTP introspection service...")
	log.Printf("Service available at: %s%s", host, p)

	if err := http.ListenAndServe(p, nil); err != nil {
		log.Fatal(err)
	}
}

func getNameFromHostname(hostname string, prefix string) string {
	hostname = strings.TrimSuffix(hostname, ".local")
	return prefix + hostname
}

func handleError(err error, msg string) {
	if err != nil {
		os.Stderr.WriteString(msg + err.Error() + "\n")
		os.Exit(1)
	}
}

func registerAgent(url string, meta Metadata) (*http.Response, error) {

	n := path.Base(*tplPath)
	t := template.New(n)
	
	log.Printf("loading template from \"%s\"\n", *tplPath)

	tpl, err := t.ParseFiles(*tplPath)
	
	if err != nil {
		log.Fatal("Parse: ", err)
		return nil,err
	}

	jsonMeta, err := json.MarshalIndent(meta, "", "    ")
	if err == nil {
		log.Printf("metadata: %s", string(jsonMeta))
	}
	

	var output bytes.Buffer 
	if err = tpl.Execute(&output, meta); err != nil {
		log.Fatal("FATAL registerAgent: ", err)
		return nil,err
	}

	req, err := http.NewRequest("POST", url, &output)

	if err != nil {
		return nil, err
	}

	log.Printf("POST: %s %s\n", url, output.String())

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return resp, nil
}


// func readMachineID(filepath string) (string, error) {
// 	contents, err := ioutil.ReadFile(filepath)	
// 	if err != nil {
// 		return "", err
// 	}
// 	return string(contents), nil
// }

func getDefaultUUID() string {
	if dat, _ := ioutil.ReadFile(uuidFile); dat != nil {
		uuid := string(dat)
		return strings.TrimSpace(uuid)
	}
	def := uuid.NewV4().String()

	return def
}

func saveUUID(uuid string) {
	ioutil.WriteFile(uuidFile, []byte(uuid), 0644)
}

///////////////////////////////////////////////////
// TODO: Move to plugins :)

func addIps(data Metadata) error {
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

func addHost(data Metadata) error {
	if _, ok := data["Interfaces"].([]Identifier); !ok {
		i := []Identifier{}
	  data["Interfaces"] = i
	}

	host, err := os.Hostname()

	if err != nil {
		return err
	}

	Host := Identifier{"host", host + ".local", "Hostname"}
	data["Interfaces"] = append(data["Interfaces"].([]Identifier), Host)

	return nil
}

func addSerial(data Metadata) error {
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

func addMachineID(data Metadata) error {
	if _, ok := data["Interfaces"].([]Identifier); !ok {
		i := []Identifier{}
		data["Interfaces"] = i
  	}

	machineID, err := getMachineID()
	
	if err != nil {
		return err
	}

	MachineID := Identifier{"machine-id", machineID, "Dbus machine-id"}
	data["Interfaces"] = append(data["Interfaces"].([]Identifier), MachineID)	

	return nil
}

func addMac(data Metadata) error {
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

func getMac(iface string) (string, error) {
	filepath := strings.Replace("/sys/class/net/#iface#/address", "#iface#", iface, -1)
	contents, err := ioutil.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	mac := strings.Replace(string(contents), "\n", "", -1)
	return mac, nil
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
