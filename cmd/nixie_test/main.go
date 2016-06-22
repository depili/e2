package main

import (
	"fmt"
	"github.com/wujiang/embd"
	_ "github.com/wujiang/embd/host/rpi" // This loads the RPi drive
	"time"
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

var spiBus embd.SPIBus

func main() {
	if err := embd.InitSPI(); err != nil {
		panic(err)
	}
	defer embd.CloseSPI()

	spiBus = embd.NewSPIBus(embd.SPIMode0, 0, 50000, 8, 50)
	defer spiBus.Close()

	send_nixie(nixie_numbers[0] | nixie_yellow)
	time.Sleep(5 * time.Second)
	for _, number := range nixie_numbers {
		send_nixie(number | nixie_led_off)
		time.Sleep(5 * time.Second)
	}

	send_nixie(nixie_no_number | nixie_led_off)

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

func send_nixie(data uint16) {
	data_buf := [2]uint8{uint8(data >> 8), uint8(data)}
	fmt.Printf("Sending: %08b%08b\n", data_buf[0], data_buf[1])
	if err := spiBus.TransferAndReceiveData(data_buf[:]); err != nil {
		panic(err)
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

func color_test() {
	fmt.Printf("Sending red\n")
	send_nixie(nixie_red)
	time.Sleep(5 * time.Second)

	fmt.Printf("Sending green\n")
	send_nixie(nixie_green)
	time.Sleep(5 * time.Second)

	fmt.Printf("Sending blue\n")
	send_nixie(nixie_blue)
	time.Sleep(5 * time.Second)

	fmt.Printf("Sending yellow\n")
	send_nixie(nixie_yellow)
	time.Sleep(5 * time.Second)

	fmt.Printf("Sending cyan\n")
	send_nixie(nixie_cyan)
	time.Sleep(5 * time.Second)

	fmt.Printf("Sending magenta\n")
	send_nixie(nixie_magenta)
	time.Sleep(5 * time.Second)

	fmt.Printf("Sending white\n")
	send_nixie(nixie_white)
	time.Sleep(5 * time.Second)
}
