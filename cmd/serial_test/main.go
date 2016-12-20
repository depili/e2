package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
	"github.com/tarm/serial"
	"log"
	"time"
)

var options = struct {
	SerialName string `short:"s" long:"serial-name" value-name:"/dev/tty*"`
	SerialBaud int    `short:"b" long:"serial-baud" value-name:"BAUD" default:"115200"`
	// SerialTimeout   time.Duration   `long:"serial-timeout" value-name:"DURATION" default:"1s"`
}{}

var parser = flags.NewParser(&options, flags.Default)

func main() {
	if _, err := parser.Parse(); err != nil {
		log.Fatalf("%v\n", err)
	}

	serialConfig := serial.Config{
		Name: options.SerialName,
		Baud: options.SerialBaud,
		// ReadTimeout:    options.SerialTimeout,
	}

	var sp *serial.Port

	if serialPort, err := serial.OpenPort(&serialConfig); err != nil {
		fmt.Errorf("serial.OpenPort: %v", err)
		panic(err)
	} else {
		sp = serialPort
	}

	time.Sleep(5 * time.Second)

	// tallies := "123456789ABCDEF"
	tallies := "6789ABC"
	states := "01234"

	for {
		for _, t := range []byte(tallies) {
			for _, s := range []byte(states) {
				msg := fmt.Sprintf("<%c%c>", t, s)
				fmt.Printf("T: %c S: %c Sending: %s", t, s, msg)
				n, err := sp.Write([]byte(msg))
				fmt.Printf(" Wrote %d bytes\n", n)
				if err != nil {
					log.Fatal(err)
				}
				time.Sleep(500 * time.Millisecond)
			}
		}
	}
}
