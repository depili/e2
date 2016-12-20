package main

import (
	"github.com/depili/e2/serialout"
	"github.com/qmsk/e2/tally"
	"log"
)

type SerialModule struct {
	serialout.Options

	serialout *serialout.SerialOut

	Enabled bool `long:"serial" description:"Enable serial output"`
}

func init() {
	registerModule("SerialOut", &SerialModule{})
}

func (module *SerialModule) start(tally *tally.Tally) error {
	if !module.Enabled {
		return nil
	}

	if serialout, err := module.Options.Make(); err != nil {
		return err
	} else {
		module.serialout = serialout
	}

	log.Printf("SerialOut: Register tally")

	module.serialout.RegisterTally(tally)

	return nil
}

func (module *SerialModule) stop() error {
	if module.serialout != nil {
		module.serialout.Close()
	}

	return nil
}
