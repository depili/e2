package discovery

import (
	"fmt"
	"log"
	"net"
	"time"
)

const DISCOVERY_ADDR = "255.255.255.255"
const DISCOVERY_PORT = "40961"
const DISCOVERY_SEND = "\x3f\x00"

type Options struct {
	Address   string `long:"discovery-address" default:""`
	Interface string `long:"discovery-interface"`

	Interval time.Duration `long:"discovery-interval" default:"10s"`
}

func (options Options) Discovery() (*Discovery, error) {
	discovery := &Discovery{
		options: options,

		recvChan: make(chan Packet),
	}

	if udpConn, err := net.ListenUDP("udp4", nil); err != nil {
		return nil, err
	} else {
		discovery.udpConn = udpConn
	}

	addr := options.Address

	if addr != "" {

	} else if options.Interface != "" {
		if ip, err := lookupInterfaceBroadcast(options.Interface); err != nil {
			return nil, err
		} else {
			addr = ip.String()

			log.Printf("Discovery: using interface %v broadcast address: %v\n", options.Interface, addr)
		}
	} else {
		addr = DISCOVERY_ADDR
	}

	if udpAddr, err := net.ResolveUDPAddr("udp4", net.JoinHostPort(addr, DISCOVERY_PORT)); err != nil {
		return nil, err
	} else {
		discovery.udpAddr = udpAddr
	}

	return discovery, nil
}

type Discovery struct {
	options Options
	udpConn *net.UDPConn
	udpAddr *net.UDPAddr

	recvChan  chan Packet
	recvError error
}

func (discovery *Discovery) String() string {
	return fmt.Sprintf("%v", discovery.udpAddr)
}

func (discovery *Discovery) send() error {
	pkt := ([]byte)(DISCOVERY_SEND)

	if _, err := discovery.udpConn.WriteToUDP(pkt, discovery.udpAddr); err != nil {
		return err
	}

	return nil
}

func (discovery *Discovery) recv(packet *Packet) error {
	buf := make([]byte, 1500)

	if n, recvAddr, err := discovery.udpConn.ReadFromUDP(buf); err != nil {
		return err
	} else if err := packet.unpack(recvAddr, buf[:n]); err != nil {
		return err
	} else {
		return nil
	}
}

func (discovery *Discovery) receiver() {
	defer close(discovery.recvChan)

	for {
		var packet Packet

		if err := discovery.recv(&packet); err != nil {
			log.Printf("Discovery.receiver: %v\n", err)

			discovery.recvError = err

			return
		}

		discovery.recvChan <- packet
	}
}

func (discovery *Discovery) run(outChan chan Packet) {
	defer close(outChan)
	defer log.Printf("Discovery.run: stopped")

	intervalChan := time.Tick(discovery.options.Interval)

	// initial discover
	discovery.send()

	for {
		select {
		case <-intervalChan:
			if err := discovery.send(); err != nil {
				log.Printf("Discovery.Send: %v\n", err)

				discovery.recvError = err

				return

			} else {
				//log.Printf("Discovery.Send...\n")
			}

		case packet, ok := <-discovery.recvChan:
			if !ok {
				return
			}

			outChan <- packet
		}
	}
}

func (discovery *Discovery) Run() chan Packet {
	outChan := make(chan Packet)

	go discovery.receiver()
	go discovery.run(outChan)

	return outChan
}

// Get error set after the Run() chan is closed
func (discovery *Discovery) Error() error {
	return discovery.recvError
}

func (discovery *Discovery) Stop() {
	discovery.udpConn.Close()
}
