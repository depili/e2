package tally

import (
	"fmt"
	"github.com/qmsk/e2/client"
	"github.com/qmsk/e2/discovery"
	"github.com/wujiang/embd"
	_ "github.com/wujiang/embd/host/rpi" // This loads the RPi drive
	"log"
)

var pins = [8]int{21, 20, 16, 12, 26, 19, 13, 6}

type Options struct {
	clientOptions    client.Options
	discoveryOptions discovery.Options
}

func (options Options) Tally(clientOptions client.Options, discoveryOptions discovery.Options) (*Tally, error) {
	options.clientOptions = clientOptions
	options.discoveryOptions = discoveryOptions

	var tally Tally

	return &tally, tally.start(options)
}

// Concurrent tally support for multiple sources and destinations
type Tally struct {
	options Options

	discovery     *discovery.Discovery
	discoveryChan chan discovery.Packet

	/* run() state */
	// active systems
	sources map[string]Source

	// updates to sources
	sourceChan chan Source

	// Relays
	relays [8]embd.DigitalPin
}

func (tally *Tally) start(options Options) error {
	tally.options = options
	tally.sources = make(map[string]Source)
	tally.sourceChan = make(chan Source)

	// Initialize GPIO
	log.Printf("Initializing GPIO\n")

	if err := embd.InitGPIO(); err != nil {
		panic(err)
	}

	for i, pin := range pins {
		fmt.Printf("Relay %d pin %d: ", i+1, pin)
		relay, err := embd.NewDigitalPin(pins[i])
		if err != nil {
			panic(err)
		}
		fmt.Printf(" open")

		if err := relay.SetDirection(embd.Out); err != nil {
			panic(err)
		}
		fmt.Printf(" output")
		if err := relay.Write(embd.Low); err != nil {
			panic(err)
		}
		fmt.Printf(" low\n")
		tally.relays[i] = relay
	}

	if discovery, err := options.discoveryOptions.Discovery(); err != nil {
		return fmt.Errorf("discovery:DiscoveryOptions.Discovery: %v", err)
	} else {
		tally.discovery = discovery
		tally.discoveryChan = discovery.Run()
	}

	return nil
}

// mainloop, owns Tally state
func (tally *Tally) Run() error {
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
				return fmt.Errorf("Tally.update: %v\n", err)
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
	log.Printf("Set relays\n")
	relay_states := [8]bool{false, false, false, false, false, false, false, false}
	for i, status := range state.Tally {
		if i < 9 && i > 0 {
			relay_states[i-1] = status.Program
		}
	}

	for i, enable := range relay_states {
		fmt.Printf("\t%d: ", i+1)
		if enable {
			if err := tally.relays[i].Write(embd.High); err != nil {
				panic(err)
			}
			fmt.Printf("\033[7m\033[1mON\033[21m\033[27m ")
		} else {
			if err := tally.relays[i].Write(embd.Low); err != nil {
				panic(err)
			}
			fmt.Printf("off ")
		}
	}
	fmt.Printf("\n")

	return nil
}
