package gpio

import (
	"fmt"
	"github.com/depili/e2/tally"
	"github.com/kidoman/embd"
	_ "github.com/kidoman/embd/host/rpi" // This loads the RPi driver
	"github.com/qmsk/e2/hetec-dcp"
	"log"
	"os"
	"sync"
)

var nixie_numbers = []uint16{
	0x0008, // 0
	0x1000, // 1
	0x0800, // 2
	0x0400, // 3
	0x0200, // 4
	0x0100, // 5
	0x0080, // 6
	0x0040, // 7
	0x0020, // 8
	0x0010} // 9

const nixie_no_number = uint16(0x0000)
const nixie_number_mask = uint16(0x1ff8)
const nixie_color_mask = (0xe000)
const nixie_red = uint16(0xa000)
const nixie_blue = uint16(0x6000)
const nixie_green = uint16(0xc000)
const nixie_magenta = uint16(0x2000)
const nixie_cyan = uint16(0x4000)
const nixie_yellow = uint16(0x8000)
const nixie_white = uint16(0x0000)
const nixie_led_off = uint16(0xe000)

type Options struct {
	LivePin    string      `long:"gpio-live-pin"`
	KvmOptions dcp.Options `group:"Hetec DCP Serial client"`
}

func (options Options) Make() (*GPIO, error) {
	var gpio = GPIO{
		options: options,
	}

	if err := gpio.init(options); err != nil {
		return nil, err
	}

	return &gpio, nil
}

type GPIO struct {
	options Options

	// Livepin is HIGH if the kvm channel is live
	livePin    *Pin
	spiBus     embd.SPIBus
	kvmConsole int
	kvmTallies [4]bool

	kvmChan   chan int
	tallyChan chan tally.State
	closeChan chan bool
	waitGroup sync.WaitGroup
}

func (gpio *GPIO) init(options Options) error {
	fmt.Printf("Init GPIO\n")

	if err := embd.InitGPIO(); err != nil {
		return fmt.Errorf("embd.InitGPIO: %v", err)
	}

	if err := embd.InitSPI(); err != nil {
		panic(err)
	}

	fmt.Printf("Intialize SPI\n")
	gpio.spiBus = embd.NewSPIBus(embd.SPIMode0, 0, 50, 8, 100)

	gpio.send_nixie(nixie_no_number | nixie_white)
	gpio.send_nixie(nixie_no_number | nixie_white)
	gpio.send_nixie(nixie_no_number | nixie_white)

	if options.LivePin == "" {

	} else if pin, err := openPin("status:live", options.LivePin); err != nil {
		return err
	} else {
		gpio.livePin = pin
	}

	gpio.kvmConsole = 5
	gpio.kvmChan = make(chan int)

	gpio.closeChan = make(chan bool)

	return nil
}

func (gpio *GPIO) RegisterTally(t *tally.Tally) {
	gpio.tallyChan = make(chan tally.State)
	gpio.waitGroup.Add(1)

	go gpio.run()

	t.Register(gpio.tallyChan)
}

func (gpio *GPIO) close() {
	defer gpio.waitGroup.Done()

	log.Printf("GPIO: Close pins and SPI bus..")

	if gpio.livePin != nil {
		gpio.livePin.Close(&gpio.waitGroup)
	}

	// Turn off the nixie tube and release the spi bus
	gpio.send_nixie(nixie_no_number | nixie_led_off)
	gpio.spiBus.Close()
	embd.CloseSPI()
}

func (gpio *GPIO) updateTally(state tally.State) {
	log.Printf("GPIO: Update tally State:")
	fmt.Printf("KVM tallies: ")
	for id := tally.ID(1); id < 5; id++ {
		var pinState = false

		if status, exists := state.Tally[id]; !exists {
			// missing tally state for pin
		} else {
			if status.Status.Program {
				pinState = true
			}
		}
		fmt.Printf("%d: %t ", id, pinState)
		gpio.kvmTallies[id-1] = bool(pinState)
	}
	fmt.Printf("\n")
}

func (gpio *GPIO) run() {
	defer gpio.close()

	go gpio.listenKvm()

	// Initialize the nixie tube
	log.Printf("Initialize nixie tube")
	gpio.send_nixie(nixie_numbers[0] | nixie_yellow)

	log.Printf("Entering message loop")
	for {
		select {
		case gpio.kvmConsole = <-gpio.kvmChan:
			log.Printf("KVM console: %d", gpio.kvmConsole)
		case state := <-gpio.tallyChan:
			gpio.updateTally(state)
		case _ = <-gpio.closeChan:
			log.Printf("GPIO: Done")
			return
		}
		fmt.Printf("kvm tallies: ")
		for _, t := range gpio.kvmTallies {
			fmt.Printf("%t ", t)
		}
		fmt.Printf("\n")
		if gpio.kvmConsole < 4 && gpio.kvmConsole >= 0 {
			color := nixie_red
			if gpio.kvmTallies[gpio.kvmConsole] {
				log.Println("KVM console is LIVE")
				gpio.livePin.Set(true)
			} else {
				color = nixie_blue
				log.Println("KVM console is safe")
				gpio.livePin.Set(false)
			}
			gpio.send_nixie(nixie_numbers[gpio.kvmConsole+1] | color)
		}

	}

}

func (gpio *GPIO) listenKvm() error {
	if client, err := gpio.options.KvmOptions.Client(); err != nil {
		return err
	} else {
		for {
			if dcpDevice, err := client.Read(); err != nil {
				log.Fatalf("dcp:Client.Read: %v\n", err)
			} else {
				dcpDevice.Print(os.Stdout)
				gpio.kvmChan <- dcpDevice.Mode.Console.Channel
			}
		}
	}
}

func (gpio *GPIO) send_nixie(data uint16) {
	data_buf := []uint8{uint8(data >> 8), uint8(data)}
	fmt.Printf("Sending: %08b%08b\n", data_buf[0], data_buf[1])
	if err := gpio.spiBus.TransferAndReceiveData(data_buf); err != nil {
		panic(err)
	}
}

// Close and Wait..
func (gpio *GPIO) Close() {
	log.Printf("GPIO: Close..")

	gpio.closeChan <- true

	if gpio.tallyChan != nil {
		close(gpio.tallyChan)
	}

	gpio.waitGroup.Wait()
}
