package tally

import (
	"fmt"
	"github.com/qmsk/e2/client"
	"github.com/qmsk/e2/discovery"
	"log"
)

type Options struct {
	clientOptions    client.Options
	discoveryOptions discovery.Options
}

func (options Options) Tally(clientOptions client.Options, discoveryOptions discovery.Options, channel chan [4]bool) (*Tally, error) {
	options.clientOptions = clientOptions
	options.discoveryOptions = discoveryOptions

	var tally Tally

	return &tally, tally.start(options, channel)
}

// Concurrent tally support for multiple sources and destinations
type Tally struct {
	options Options

	discovery     *discovery.Discovery
	discoveryChan chan discovery.Packet
	tallyChan     chan [4]bool

	/* run() state */
	// active systems
	sources map[string]Source

	// updates to sources
	sourceChan chan Source
}

func (tally *Tally) start(options Options, tallyChan chan [4]bool) error {
	tally.options = options
	tally.sources = make(map[string]Source)
	tally.sourceChan = make(chan Source)
	tally.tallyChan = tallyChan

	if discovery, err := options.discoveryOptions.Discovery(); err != nil {
		return fmt.Errorf("discovery:DiscoveryOptions.Discovery: %v", err)
	} else {
		tally.discovery = discovery
		tally.discoveryChan = discovery.Run()
	}

	return nil
}

// mainloop, owns Tally state
func (tally *Tally) Run() {
	for {
		select {
		case discoveryPacket := <-tally.discoveryChan:
			if clientOptions, err := tally.options.clientOptions.DiscoverOptions(discoveryPacket); err != nil {
				log.Printf("Tally: invalid discovery client options: %v\n", err)
			} else if _, exists := tally.sources[clientOptions.String()]; exists {
				// already known
			} else if source, err := newSource(tally, clientOptions); err != nil {
				log.Printf("Tally: unable to connect to discovered system: %v\n", err)
			} else {
				log.Printf("Tally: connected to new source: %v\n", source)

				tally.sources[clientOptions.String()] = source
			}

		case source := <-tally.sourceChan:
			if err := source.err; err != nil {
				log.Printf("Tally: Source %v Error: %v\n", source, err)

				delete(tally.sources, source.String())
			} else {
				log.Printf("Tally: Source %v: Update\n", source)

				tally.sources[source.String()] = source
			}

			if err := tally.update(); err != nil {
				log.Fatalf("Tally.update: %v\n", err)
			}
		}
	}
}

// Compute new output state from sources
func (tally *Tally) update() error {
	var state = State{
		Inputs: make(map[Input]ID),
		Tally:  make(map[ID]Status),
	}

	for _, source := range tally.sources {
		if err := state.updateSystem(source.system, source.String()); err != nil {
			return err
		}
	}

	if err := state.update(); err != nil {
		return err
	}

	log.Printf("Tally.update: state:\n")
	state.Print()

	tally_states := [4]bool{false, false, false, false}
	for i, status := range state.Tally {
		if i < 5 && i > 0 {
			tally_states[i-1] = status.Program
		}
	}
	tally.tallyChan <- tally_states

	return nil
}
