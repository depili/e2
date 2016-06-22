package main

import (
	//	"fmt"
	"github.com/wujiang/embd"
	_ "github.com/wujiang/embd/host/rpi" // This loads the RPi drive
	"time"
)

var nixie_numbers = []uint16{
	0x0200, // 0
	0x0001, // 1
	0x0002, // 2
	0x0004, // 3
	0x0008, // 4
	0x0010, // 5
	0x0020, // 6
	0x0040, // 7
	0x0080, // 8
	0x0100} // 9

const nixie_no_number = uint16(0x0000)
const nixie_red = uint16(0x5000)
const nixie_green = uint16(0x3000)
const nixie_blue = uint16(0x6000)
const nixie_yellow = uint16(0x1000)
const nixie_cyan = uint16(0x2000)
const nixie_magenta = uint16(0x4000)
const nixie_white = uint16(0x0000)
const nixie_led_off = uint16(0x7000)

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

func main() {
	if err := embd.InitSPI(); err != nil {
		panic(err)
	}
	defer embd.CloseSPI()

	spiBus = embd.NewSPIBus(embd.SPIMode0, 0, 50000, 8, 50)
	defer spiBus.Close()

	init_matrix()

	for i, _ := range led_numbers2 {
		send_number(i, false)
		time.Sleep(5 * time.Second)
		send_number(i, true)
		time.Sleep(5 * time.Second)
	}

	// databuf := [2]uint8{0, 1}
	// fmt.Printf("%08b %08b\n", databuf[1], databuf[0])
	//
	// if err := spiBus.TransferAndReceiveData(databuf[:]); err != nil {
	// 	panic(err)
	// }
	// time.Sleep(10 * time.Second)
	//
	// for i := uint8(0); i < 9; i++ {
	// 	databuf[1] = uint8(0x01 << i)
	// 	databuf[0] = uint8(0)
	// 	fmt.Printf("%08b %08b\n", databuf[1], databuf[0])
	// 	if err := spiBus.TransferAndReceiveData(databuf[:]); err != nil {
	// 		panic(err)
	// 	}
	//
	// 	time.Sleep(10 * time.Second)
	// }
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
	sendOpcode(OP_DIGIT3, 0)
	sendOpcode(OP_DIGIT4, 0)
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

func reverseBits(b byte) byte {
	var d byte
	for i := 0; i < 8; i++ {
		d <<= 1
		d |= b & 1
		b >>= 1
	}
	return d
}
