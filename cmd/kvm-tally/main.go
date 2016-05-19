package main

import (
	"github.com/depili/e2/tally"
	"github.com/jessevdk/go-flags"
	"github.com/qmsk/e2/client"
	"github.com/qmsk/e2/discovery"
	"github.com/qmsk/e2/hetec-dcp"
	"github.com/wujiang/embd"
	_ "github.com/wujiang/embd/host/rpi" // This loads the RPi drive
	"log"
	"os"
)

var options = struct {
	DiscoveryOptions discovery.Options `group:"E2 Discovery"`
	ClientOptions    client.Options    `group:"E2 XML"`
	TallyOptions     tally.Options     `group:"Tally"`
	KvmOptions       dcp.Options       `group:"Hetec DCP Serial client"`
}{}

const OP_NOOP = byte(0)
const OP_DIGIT0 = byte(1)
const OP_DIGIT1 = byte(2)
const OP_DIGIT2 = byte(3)
const OP_DIGIT3 = byte(4)
const OP_DIGIT4 = byte(5)
const OP_DIGIT5 = byte(6)
const OP_DIGIT6 = byte(7)
const OP_DIGIT7 = byte(8)
const OP_DECODEMODE = byte(9)
const OP_INTENSITY = byte(10)
const OP_SCANLIMIT = byte(11)
const OP_SHUTDOWN = byte(12)
const OP_DISPLAYTEST = byte(15)

var led_numbers = [4][5]uint8{
	{0, 0x81, 0xff, 0x81, 0},       //1
	{0x47, 0x89, 0x89, 0x89, 0x71}, //2
	{0x42, 0x81, 0x99, 0x99, 0x66}, //3
	{0x18, 0x28, 0x48, 0xff, 0x08},
}

var led_numbers2 = [4][8]uint8{
	{0x08, 0x18, 0x38, 0x18, 0x18, 0x18, 0x18, 0x3c},
	{0x3c, 0x66, 0x06, 0x0c, 0x18, 0x30, 0x60, 0x7e},
	{0x7c, 0xc6, 0x06, 0x1c, 0x06, 0x06, 0xc6, 0x7c},
	{0x0e, 0x1e, 0x36, 0x66, 0xff, 0x06, 0x06, 0x06},
}

var spiBus embd.SPIBus

var parser = flags.NewParser(&options, flags.Default)

func main() {
	if err := embd.InitSPI(); err != nil {
		panic(err)
	}
	defer embd.CloseSPI()

	spiBus = embd.NewSPIBus(embd.SPIMode0, 0, 50000, 8, 50)
	defer spiBus.Close()

	init_matrix()

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

	kvm := int(5)
	tallies := [4]bool{false, false, false, false}

	for {
		select {
		case kvm = <-kvm_chan:
			log.Printf("KVM console: %d", kvm)
		case tallies = <-e2_chan:
			log.Printf("E2 tallies: 1: %t 2: %t 3: %t 4: %t", tallies[0], tallies[1], tallies[2], tallies[3])
		}
		if kvm < 4 && kvm >= 0 {
			send_number(kvm, tallies[kvm])
		}
	}
}

func init_matrix() {
	sendOpcode(OP_DISPLAYTEST, 0)
	sendOpcode(OP_SCANLIMIT, 7)
	sendOpcode(OP_DECODEMODE, 0)
	sendOpcode(OP_SHUTDOWN, 1)
	sendOpcode(OP_INTENSITY, 2)
	sendOpcode(OP_DIGIT0, 0)
	sendOpcode(OP_DIGIT1, 0)
	sendOpcode(OP_DIGIT2, 0)
	sendOpcode(OP_DIGIT3, 0x18)
	sendOpcode(OP_DIGIT4, 0x18)
	sendOpcode(OP_DIGIT5, 0)
	sendOpcode(OP_DIGIT6, 0)
	sendOpcode(OP_DIGIT7, 0)
}

func sendOpcode(opcode byte, data byte) {
	databuf := [2]byte{opcode, data}
	if err := spiBus.TransferAndReceiveData(databuf[:]); err != nil {
		panic(err)
	}
}

func send_number(n int, invert bool) {
	number := led_numbers2[n]
	for j, row := range number {
		if invert {
			row = row ^ 0xff
		}
		sendOpcode(OP_DIGIT0+byte(j), row)
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
