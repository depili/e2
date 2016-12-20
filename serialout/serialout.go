package serialout

import (
	"fmt"
	"github.com/qmsk/e2/tally"
	"github.com/tarm/serial"
	"log"
	"sync"
	"time"
)

// Protocol constants
const tallySymbols = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const maxTallies = 35

// Tally states
const stateNone = 0
const stateSafe = 1
const statePVM = 2
const statePGM = 3
const statePGMPVM = 4

type Options struct {
	SerialCount uint   `long:"serial-count" value-name:"COUNT" description:"Number of tallies"`
	SerialName  string `long:"serial-out-name" value-name:"/dev/tty*" default:"/dev/ttyAMA0"`
	SerialBaud  int    `long:"serial-out-baud" value-name:"BAUD" default:"115200"`
	Debug       bool   `long:"serial-debug" description:"Enable debug output"`
}

func (options Options) Make() (*SerialOut, error) {
	var serialout = SerialOut{
		options: options,
	}

	if err := serialout.init(options); err != nil {
		return nil, err
	}

	return &serialout, nil
}

type SerialOut struct {
	options   Options
	count     uint
	sp        *serial.Port
	tallyChan chan tally.State
	waitGroup sync.WaitGroup
	states    []byte
}

func (serialout *SerialOut) init(options Options) error {
	if options.SerialCount > maxTallies {
		return fmt.Errorf("Invalid --serial-count=%v. Max is %v", options.SerialCount, maxTallies)
	}

	// Open the serial port
	serialConfig := serial.Config{
		Name: options.SerialName,
		Baud: options.SerialBaud,
	}

	if serialPort, err := serial.OpenPort(&serialConfig); err != nil {
		return fmt.Errorf("serial.OpenPort\n%v", err)
	} else {
		serialout.sp = serialPort
	}

	log.Printf("SerialOut: Sleeping due to the arduino bootloader...")
	time.Sleep(5 * time.Second)

	// Number of connected tallies
	serialout.count = options.SerialCount
	serialout.states = make([]byte, serialout.count)

	log.Printf("SerialOut: Open %v (%dbps) %s tallies", options.SerialName, options.SerialBaud, options.SerialCount)
	return nil
}

func (serialout *SerialOut) write() error {
	for tally, state := range serialout.states {
		msg := fmt.Sprintf("<%c%1d>", []byte(tallySymbols)[tally], state)
		if serialout.options.Debug {
			log.Printf("SerialOut: Sending: %v\n", msg)
		}
		if _, err := serialout.sp.Write([]byte(msg)); err != nil {
			return err
		}
	}
	return nil
}

func (serialout *SerialOut) close() {
	defer serialout.waitGroup.Done()

	log.Printf("SerialOut: closing...")

	if err := serialout.sp.Close(); err != nil {
		log.Printf("SerialOut.Close: serialPort.close: %v", err)
	}
}

func (serialout *SerialOut) updateTally(tallyState tally.State) {
	if serialout.options.Debug {
		log.Printf("SerialOut: Update tally State:")
	}

	states := make([]byte, serialout.count)

	var found int

	for i, state := range states {
		id := tally.ID(i)

		if tally, exists := tallyState.Tally[id]; !exists {
			state = stateNone
		} else {
			found++

			if tally.Status.Program && tally.Status.Preview {
				state = statePGMPVM
			} else if tally.Status.Preview {
				state = statePVM
			} else if tally.Status.Program {
				state = statePGM
			} else {
				state = stateSafe
			}
			if serialout.options.Debug {
				log.Printf("SerialOut %v: id=%v status=%v errors=%v state=%v", i, id, tally.Status, len(tally.Errors), state)
			}
		}

		states[i] = state
	}

	/*
		errors = len(tallyState.Errors)

		// status LED
		var statusLED LED

		if found > 0 && errors > 0 {
			statusLED = spiled.options.StatusWarn
		} else if errors > 0 {
			statusLED = spiled.options.StatusError
		} else if found > 0 {
			statusLED = spiled.options.StatusOK
		} else {
			statusLED = spiled.options.StatusIdle
		}

		statusLED.Intensity = spiled.options.Intensity

	*/
	if serialout.options.Debug {
		log.Printf("SerialOut: found=%v", found)
	}
	// refresh
	serialout.states = states
	serialout.write()
}

func (serialout *SerialOut) run() {
	defer serialout.close()
	refreshTimer := time.Tick(time.Duration(5 * time.Second))
	for {
		select {
		case tallyState, ok := <-serialout.tallyChan:
			if ok {
				serialout.updateTally(tallyState)
			} else {
				return
			}
		case <-refreshTimer:
			if err := serialout.write(); err != nil {
				log.Printf("SerialOut: Write error: %v", err)
			}
		}
	}

	log.Printf("SerialOut: Done")
}

func (serialout *SerialOut) RegisterTally(t *tally.Tally) {
	serialout.tallyChan = make(chan tally.State)
	serialout.waitGroup.Add(1)

	go serialout.run()

	t.Register(serialout.tallyChan)
}

// Close and Wait..
func (serialout *SerialOut) Close() {
	log.Printf("SerialOut: Close..")

	if serialout.tallyChan != nil {
		close(serialout.tallyChan)
	}

	serialout.waitGroup.Wait()
}
