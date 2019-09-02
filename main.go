package main

import (
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

//Metadata holds data we will send over to the registry
type Metadata map[string]interface{}

func init() {
	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", "rpi-agent")
		flag.PrintDefaults()
		fmt.Printf("Version: %s, Build: %s\n", VERSION, BUILD_DATE)
	}
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
		s,  // service type and protocl
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
	log.Println("Sutting down service...")
}

// HTTP service to introspect RPi instance
func startService(deviceUUID string) {
	host, _ := os.Hostname()
	metadata, err := getAddress()
	handleError(err, "Error GetAddress: ")

	serial, err := getSerial()
	handleError(err, "Error GetSerial: ")

	metadata["serial"] = serial

	machineID, err := getMachineID()
	handleError(err, "Error GetMachineID: ")

	metadata["machine-id"] = machineID

	startTime := time.Now()
	// metadata["agent_start"] = startTime.Unix()
	metadata["agent_start"] = startTime.Format(time.RFC3339)

	metadata["agent_uptime"] = time.Since(startTime)

	metadata["hostname"] = host + ".local"

	mac, err := getMac("wlan0")
	metadata["mac_wlan0"] = mac

	mac, err = getMac("eth0")
	metadata["mac_eth0"] = mac

	output := make(Metadata)

	output["metadata"] = metadata
	output["Uuid"] = deviceUUID
	output["name"] = getNameFromHostname(host, "")
	output["status"] = "online"
	output["alias"] = serial

	if registryURL != nil {
		registerAgent(*registryURL, output)
	}

	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		//Update our uptime
		metadata["agent_uptime"] = time.Now().Sub(startTime)

		rw.Header().Set("Content-Type", "application/json")
		json.NewEncoder(rw).Encode(output)
	})

	p := fmt.Sprintf(":%d", *port)
	log.Println("Starting HTTP instrospection service...")
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

	tmpl, err := t.ParseFiles(*tplPath)
	
	if err != nil {
		log.Fatal("Parse: ", err)
		return nil,err
	}

	jsonMeta, err := json.MarshalIndent(meta, "", "    ")
	if err == nil {
		log.Printf("metadata: %s", string(jsonMeta))
	}
	

	var output bytes.Buffer 
	if err = tmpl.Execute(&output, meta); err != nil {
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

//machine-id is found in either /etc/machine-id or
// /var/lib/dbus/machine-id 
func getMachineID() (string, error) {

	readMachineID := func (filepath string) (string, error) {
		contents, err := ioutil.ReadFile(filepath)	
		if err != nil {
			return "", err
		}
		return string(contents), nil
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

func getAddress() (Metadata, error) {
	output := make(Metadata)
	inter, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, ifa := range inter {
		addrs, err := ifa.Addrs()

		if err != nil {
			return nil, err
		}

		for _, a := range addrs {
			if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					output["ip_"+string(ifa.Name)] = ipnet.IP.String()
				}
			}
		}
	}

	return output, nil
}

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
