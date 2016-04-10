package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/qmsk/e2/client"
	"github.com/qmsk/e2/discovery"
	"github.com/wujiang/embd"
	_ "github.com/wujiang/embd/host/rpi" // This loads the RPi drive
	"log"
	"regexp"
	"strconv"
)

var options = struct {
	DiscoveryOptions discovery.Options `group:"E2 Discovery"`
	ClientOptions    client.Options    `group:"E2 JSON-RPC"`
}{}

var parser = flags.NewParser(&options, flags.Default)

func main() {
	r, _ := regexp.Compile("tally=(\\d+)")

	if err := embd.InitGPIO(); err != nil {
		panic(err)
	}
	defer embd.CloseGPIO()

	pins := [8]int{21, 20, 16, 12, 26, 19, 13, 6}
	var relays [8]embd.DigitalPin
	for i := 0; i < 8; i++ {
		relay, err := embd.NewDigitalPin(pins[i])
		if err != nil {
			panic(err)
		}
		defer relay.Close()

		if err := relay.SetDirection(embd.Out); err != nil {
			panic(err)
		}
		if err := relay.Write(embd.Low); err != nil {
			panic(err)
		}
		relays[i] = relay
	}

	if _, err := parser.Parse(); err != nil {
		log.Fatalf("%v\n", err)
	}

	if clientOptions, err := options.ClientOptions.DiscoverClient(options.DiscoveryOptions); err != nil {
		log.Fatalf("Client %#v: Discover %#v: %v\n", options.ClientOptions, options.DiscoveryOptions, err)
	} else if xmlClient, err := clientOptions.XMLClient(); err != nil {
		log.Fatalf("Client %#v: XMLClient: %v", clientOptions, err)
	} else if listenChan, err := xmlClient.Listen(); err != nil {
		log.Fatalf("XMLClient %v: Listen: %v", xmlClient, err)
	} else {
		for system := range listenChan {
			tallies := [8]bool{false, false, false, false, false, false, false, false}

			fmt.Printf("\033[H\033[2J")

			for sourceID, source := range system.SrcMgr.SourceCol {
				fmt.Printf("Source %d: %v\n", sourceID, source.Name)
				tally := -1
				contact := system.SrcMgr.InputCfgCol[source.InputCfgIndex].ConfigContact
				match := r.FindStringSubmatch(contact)
				if len(match) == 2 {
					tally, err = strconv.Atoi(match[1])
					if err != nil || tally > 8 || tally < 1 {
						tally = -1
					} else {
						fmt.Printf("\tTally: %d\n", tally)
					}
				}
				for screenID, screen := range system.DestMgr.ScreenDestCol {
					for _, layer := range screen.LayerCollection {
						if layer.LastSrcIdx != sourceID {
							continue
						}

						if layer.PvwMode > 0 {
							fmt.Printf("\tPreview %d: %v\n", screenID, screen.Name)
						}
						if layer.PgmMode > 0 {
							fmt.Printf("\tProgram %d: %v\n", screenID, screen.Name)
							if tally != -1 {
								tallies[tally-1] = true
							}
						}
					}
				}
			}
			fmt.Printf("\nTallies:\n")
			for i := 0; i < 8; i++ {
				fmt.Printf("\t%d", i+1)
				if tallies[i] {
					if err := embd.DigitalWrite(pins[i], embd.High); err != nil {
						panic(err)
					}
					fmt.Printf(" ON")
				} else {
					if err := embd.DigitalWrite(pins[i], embd.Low); err != nil {
						panic(err)
					}
					fmt.Printf(" OFF")
				}
			}
		}
	}
}
