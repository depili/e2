package main

import (
	"github.com/depili/e2/tally"
	"github.com/jessevdk/go-flags"
	"github.com/qmsk/e2/client"
	"github.com/qmsk/e2/discovery"
	"github.com/qmsk/e2/hetec-dcp"
	"log"
	"os"
)

var options = struct {
	DiscoveryOptions discovery.Options `group:"E2 Discovery"`
	ClientOptions    client.Options    `group:"E2 XML"`
	TallyOptions     tally.Options     `group:"Tally"`
	KvmOptions       dcp.Options       `group:"Hetec DCP Serial client"`
}{}

var parser = flags.NewParser(&options, flags.Default)

func main() {
	if _, err := parser.Parse(); err != nil {
		log.Fatalf("%v\n", err)
	}

	e2_chan := make(chan [4]bool, 5)
	tally, err := options.TallyOptions.Tally(options.ClientOptions, options.DiscoveryOptions, e2_chan)
	if err != nil {
		log.Fatalf("Tally: %v\n", err)
	}
	go tally.Run()

	kvm_chan := make(chan int, 5)
	go kvm_listen(kvm_chan)

	for {
		select {
		case kvm := <-kvm_chan:
			log.Printf("KVM console: %d", kvm)
		case tallies := <-e2_chan:
			log.Printf("E2 tallies: 1: %t 2: %t 3: %t 4: %t", tallies[0], tallies[1], tallies[2], tallies[3])
		}
	}

}

func kvm_listen(channel chan int) error {
	if client, err := options.KvmOptions.Client(); err != nil {
		return err
	} else {
		for {
			if dcpDevice, err := client.Read(); err != nil {
				log.Fatalf("dcp:Client.Read: %v\n", err)
			} else {
				dcpDevice.Print(os.Stdout)
				channel <- dcpDevice.Mode.Console.Channel
			}
		}
	}
}
