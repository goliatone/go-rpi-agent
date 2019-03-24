package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"

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

func main() {

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

	service, err := zeroconf.Register(
		"hid-reader", // service instance name
		"_rpi._tcp",  // service type and protocl
		"local.",     // service domain
		8080,         // service port
		txtRecord,    // service metadata
		nil,          // register on all network interfaces
	)

	if err != nil {
		log.Fatal(err)
	}

	defer service.Shutdown()

	// Sleep forever
	select {}
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

// HTTP service to introspect RPi instance
func startService(deviceUUID string) {
	host, _ := os.Hostname()
	metadata, err := getAddress()
	handleError(err, "Error GetAddress: ")

	serial, err := getSerial()
	handleError(err, "Error GetAddress: ")

	metadata["serial"] = serial

	metadata["hostname"] = host + ".local"

	mac, err := getMac("wlan0")
	metadata["mac_wlan0"] = mac

	mac, err = getMac("eth0")
	metadata["mac_eth0"] = mac

	output := make(map[string]interface{})

	output["metadata"] = metadata
	output["uuid"] = deviceUUID
	output["name"] = getNameFromHostname(host, "")
	output["status"] = "online"
	output["alias"] = serial

	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		json.NewEncoder(rw).Encode(output)
	})

	log.Println("starting HTTP instrospection service...")
	log.Printf("service available at: %s:8080", host)

	if err := http.ListenAndServe(":8080", nil); err != nil {
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

func post(url string, jsonData []byte) (*http.Response, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return resp, nil
}

func getAddress() (map[string]string, error) {
	output := make(map[string]string)
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
