package dust_sensor

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/jacobsa/go-serial/serial"
)

var (
	//mode
	setPassiveMode        = []byte{0x42, 0x4d, 0xe1, 0x00, 0x00, 0x01, 0x70}
	setPassiveModeCheck   = []byte{66, 77, 0, 4, 225, 0, 1, 116}
	passiveModeQuery      = []byte{0x42, 0x4d, 0xe2, 0x00, 0x00, 0x01, 0x71}
	passiveModeQueryCheck = []byte{66, 77, 0, 28}
)

type Sensor struct {
	Port io.ReadWriteCloser
}

type Config struct {
	PortName        string
	BaudRate        uint
	DataBits        uint
	StopBits        uint
	MinimumReadSize uint
}

func Connect(c *Config) (*Sensor, error) {
	options := serial.OpenOptions{
		PortName:        c.PortName,
		BaudRate:        c.BaudRate,
		DataBits:        c.DataBits,
		StopBits:        c.StopBits,
		MinimumReadSize: c.MinimumReadSize,
	}

	port, err := serial.Open(options)
	if err != nil {
		return nil, err
	}

	return &Sensor{port}, err
}

func (s *Sensor) SetDeviceMod() (err error) {
	fmt.Println("1")
	n, err := s.Port.Write(setPassiveMode)
	if err != nil {
		err = fmt.Errorf("port.write: %s", err)
		return err
	}

	if n == len(setPassiveMode) {
		buf := make([]byte, 8)
		n, err = s.Port.Read(buf)
		if err != nil {
			err = fmt.Errorf("port read: %s", err)
			return err
		}

		if bytes.Equal(buf[:n], setPassiveModeCheck) {
			fmt.Println(time.Now().Format("2006-01-02 15:04"), "😀", "Set succeed:", buf[:n])
			return nil
		}
	}
	err = fmt.Errorf("check fail")
	return err
}

func (s *Sensor) QueryDust() (result int, err error) {
	fmt.Println("2")
	n, err := s.Port.Write(passiveModeQuery)
	if err != nil {
		err = fmt.Errorf("port.Write: %s", err)
		return 0, err
	}

	var results []byte

	if n == len(passiveModeQuery) {
		for len(results) < 32 {
			buf := make([]byte, 32-len(results))
			n, err = s.Port.Read(buf)
			if err != nil {
				err = fmt.Errorf("port.Read: %s", err)
				return 0, err
			}

			results = append(results, buf[:n]...)
			if len(results) >= len(passiveModeQueryCheck) {
				tempSlice := results[:len(passiveModeQueryCheck)]
				if !bytes.Equal(tempSlice, passiveModeQueryCheck) {
					results = nil
					err := errors.New("😢 Read check fail , reread")
					return 0, err
				}
			}
		}
	}

	lrcCheckValue := 0
	lrcQueryValue := uint16(results[30])<<8 | uint16(results[31])

	for _, r := range results[:29] {
		lrcCheckValue += int(r)
	}

	if uint16(lrcCheckValue) == lrcQueryValue {
		result = ((int(results[6]) * 256) + int(results[7]))
		return result, nil
	}

	err = errors.New("😢 Read check fail, reread")
	return 0, err
}
