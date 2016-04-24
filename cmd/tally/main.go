package main

import (
	"e2/tally"
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/qmsk/e2/client"
	"github.com/qmsk/e2/discovery"
	"github.com/wujiang/embd"
	_ "github.com/wujiang/embd/host/rpi" // This loads the RPi drive
	"log"
)

var options = struct {
	DiscoveryOptions discovery.Options `group:"E2 Discovery"`
	ClientOptions    client.Options    `group:"E2 XML"`
	TallyOptions     tally.Options     `group:"Tally"`
}{}

var parser = flags.NewParser(&options, flags.Default)
var relays [8]embd.DigitalPin

func main() {
	// Initialize GPIO
	fmt.Printf("Initializing GPIO\n")
	pins := [8]int{21, 20, 16, 12, 26, 19, 13, 6}
	for i, pin := range pins {
		fmt.Printf("Relay %d pin %d: ", i+1, pin)
		relay, err := embd.NewDigitalPin(pins[i])
		if err != nil {
			panic(err)
		}
		fmt.Printf(" open")
		defer relay.Close()

		if err := relay.SetDirection(embd.Out); err != nil {
			panic(err)
		}
		fmt.Printf(" output")
		if err := relay.Write(embd.Low); err != nil {
			panic(err)
		}
		fmt.Printf(" low\n")
		relays[i] = relay
	}

	if _, err := parser.Parse(); err != nil {
		log.Fatalf("%v\n", err)
	}

	tally, err := options.TallyOptions.Tally(options.ClientOptions, options.DiscoveryOptions)
	if err != nil {
		log.Fatalf("Tally: %v\n", err)
	}

	if err := tally.Run(); err != nil {
		log.Fatalf("Tally.Run: %v\n", err)
	} else {
		log.Printf("Exit")
	}
}
