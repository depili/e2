package client

import (
	"fmt"
	"github.com/qmsk/e2/discovery"
	"log"
)

// Returns client Options for the given discovery packet
func (options Options) DiscoverOptions(discoveryPacket discovery.Packet) (Options, error) {
	options.Address = discoveryPacket.IP.String()
	options.XMLPort = fmt.Sprintf("%d", discoveryPacket.XMLPort)

	return options, nil
}

// If there is no URL given, use Discovery to find any E2 systems.
// Returns a new Options for the first E2 found.
func (options Options) DiscoverClient(discoveryOptions discovery.Options) (Options, error) {
	if options.Address != "" {
		return options, nil
	} else if discovery, err := discoveryOptions.Discovery(); err != nil {
		return options, err
	} else {
		defer discovery.Stop()

		log.Printf("Discovering systems on %v...\n", discovery)

		for packet := range discovery.Run() {
			if options, err := options.DiscoverOptions(packet); err != nil {
				log.Printf("Discovery invalid: %v\n", err)
			} else {
				log.Printf("Discovered system: %v\n", options.Address)

				return options, nil
			}
		}

		return options, fmt.Errorf("Discovery failed")
	}
}
