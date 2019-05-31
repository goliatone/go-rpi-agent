
package main

import (
	"context"
	"flag"
	"log"
	"time"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"github.com/grandcat/zeroconf"
)

var (
	service  = flag.String("service", "_workstation._tcp", "Set the service category to look for devices.")
	domain   = flag.String("domain", "local", "Set the search domain. For local networks, default is fine.")
	waitTime = flag.Int("wait", 10, "Duration in [s] to run discovery.")
)

func main() {
	flag.Parse()

	// Discover all services on the network (e.g. _workstation._tcp)
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Fatalln("Failed to initialize resolver:", err.Error())
	}

	entries := make(chan *zeroconf.ServiceEntry)
	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			o, _ := json.Marshal(entry)
			log.Println(string(o))
		}
		log.Println("No more entries.")
	}(entries)
	
	ctx := context.Background()
	err = resolver.Browse(ctx, *service, *domain, entries)
	if err != nil {
		log.Fatalln("Failed to browse:", err.Error())
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	var tc <-chan time.Time
	if *waitTime > 0 {
		tc = time.After(time.Second * time.Duration(*waitTime))
	}

	select{
	case <- sig:
		//Exit by user
	case <- tc:
		//Timeout
	}

	log.Println("Shutting down browser...")
	//Give some time for any debug stuff...
	time.Sleep(1 * time.Second)
}