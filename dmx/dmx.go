package dmx

import (
	"fmt"
	"io"
	"log"

	serial "go.bug.st/serial"
)

const (
	START_VAL = 0x7E
	END_VAL   = 0xE7
	// BAUD      = 57600
	// BAUD    = 115200
	TIMEOUT = 1
	// DEV             = "/dev/ttyUSB0"
	FRAME_SIZE      = 511
	FRAME_SIZE_LOW  = byte(FRAME_SIZE & 0xFF)
	FRAME_SIZE_HIGH = byte(FRAME_SIZE >> 8 & 0xFF)
)

type label byte

const (
	GET_WIDGET_PARAMETERS label = 3
	SET_WIDGET_PARAMETERS label = 4
	RX_DMX_PACKET         label = 5
	TX_DMX_PACKET         label = 6
	TX_RDM_PACKET_REQUEST label = 7
	RX_DMX_ON_CHANGE      label = 8
)

type DMX struct {
	serial io.ReadWriteCloser
}

func NewDMXConnection(dev string) (*DMX, error) {

	ports, err := serial.GetPortsList()
	if err != nil {
		log.Fatal(err)
	}
	if len(ports) == 0 {
		log.Fatal("No serial ports found!")
	}
	for _, port := range ports {
		fmt.Printf("Found port: %v\n", port)
	}

	cfg := &serial.Mode{
		BaudRate: 57600,
		Parity:   serial.NoParity,
		DataBits: 8,
		// StopBits: serial.TwoStopBits,
		StopBits: serial.TwoStopBits,

		// Name: dev,
		// Baud: BAUD,
	}
	serial, err := serial.Open(dev, cfg)
	if err != nil {
		return nil, err
	}
	return &DMX{
		serial: serial,
	}, nil
}

func (d *DMX) Send(data [512]byte) error {
	// _, err := d.serial.Write(data)
	b := make([]byte, 0, len(data)+1)
	b = append(b, 0)
	// b = append(b, START_VAL)
	// b = append(b, byte(TX_DMX_PACKET))
	// b = append(b, FRAME_SIZE_LOW)
	// b = append(b, FRAME_SIZE_HIGH)
	// b = append(b, data[:]...)
	// b = append(b, END_VAL)
	_, err := d.serial.Write(b)
	return err
}
