//
// Copyright © 2020 Fan Dang
//
package main

import (
	"encoding/hex"
	"errors"
	"strings"

	"github.com/karalabe/hid"
)

type Reader struct {
	d      *hid.Device
	opened bool
}

func (r *Reader) buildFrame(cmd byte, data []byte) []byte {
	frame := make([]byte, 64)
	frame[0] = 0x55
	frame[1] = cmd
	frame[2] = byte(len(data))
	copy(frame[3:], data)
	parity := byte(0)
	for j := 0; j < len(data)+3; j++ {
		parity = parity ^ frame[j]
	}
	frame[len(data)+3] = parity
	return frame
}

func (r *Reader) reshape(data []byte, frameLen int) [][]byte {
	if len(data) == 0 {
		return [][]byte{{}}
	}

	var packets [][]byte
	for i := 0; i < len(data); i += frameLen {
		upper := i + frameLen
		if upper > len(data) {
			upper = len(data)
		}
		packets = append(packets, data[i:upper])
	}
	if len(packets[len(packets)-1]) == frameLen {
		packets = append(packets, []byte{})
	}
	return packets
}

func (r *Reader) transceive(cmd byte, data []byte) ([]byte, error) {
	if r.opened == false {
		return nil, errors.New("reader not opened")
	}

	packets := r.reshape(data, 60)
	for i, p := range packets {
		packets[i] = r.buildFrame(cmd, p)
	}
	for i, p := range packets {
		r.d.Write(p)
		if i != len(packets)-1 {
			r.d.Read(nil)
		}
	}

	var response []byte
	for {
		buf := make([]byte, 64)
		r.d.Read(buf)
		if buf[3] != 0 || buf[4] != 0 {
			return nil, errors.New("execution failed")
		}
		if buf[2] == 60 {
			response = append(response, buf[5:63]...)
			r.d.Write(r.buildFrame(buf[1], nil))
		} else {
			response = append(response, buf[5:buf[2]+3]...)
			break
		}
	}

	return response, nil
}

func (r *Reader) Open() error {
	if r.opened {
		r.Close()
	}
	devices := hid.Enumerate(0x5351, 0x4452)
	if len(devices) == 0 {
		return errors.New("no reader found")
	}
	d, err := devices[0].Open()
	r.d = d
	r.opened = err == nil
	return err
}

func (r *Reader) Close() {
	if r.d != nil {
		r.d.Close()
	}
	r.opened = false
}

func (r *Reader) Beep() {
	r.transceive(0xE4, []byte{0, 0, 200, 0})
}

func (r *Reader) PowerOn() {
	r.transceive(0x0A, nil)
	r.transceive(0x0D, nil)
}

func (r *Reader) PowerOff() {
	r.transceive(0x0C, nil)
	r.transceive(0x0B, nil)
}

func (r *Reader) PollA() (string, string, error) {
	_, err := r.transceive(0xA1, nil)
	if err != nil {
		return "", "", errors.New("no card detected")
	}
	resp, err := r.transceive(0xA3, nil)
	if err != nil {
		return "", "", err
	}
	cardType := "UID"
	uid := strings.ToUpper(hex.EncodeToString(resp[1 : resp[0]+1]))
	sak := resp[len(resp)-1]
	if sak&0x20 > 0 {
		_, err = r.transceive(0xA4, nil)
		if err == nil {
			cardType = "CPU"
		}
	}
	return cardType, uid, nil
}

func (r *Reader) PollB() (string, string, error) {
	resp, err := r.transceive(0xB1, []byte{0, 0})
	if err != nil {
		return "", "", errors.New("no card detected")
	}
	_, err = r.transceive(0xB3, resp[1:5])
	if err != nil {
		return "", "", err
	}
	protocol := resp[10]
	if protocol&0x01 > 0 {
		return "CPU", "", nil
	} else {
		resp, err = r.transceive(0xB5, nil)
		if err == nil {
			return "ID", strings.ToUpper(hex.EncodeToString(resp)), nil
		}
		return "", "", err
	}
}

func (r *Reader) Insert() (cardType string, uid string, err error) {
	cardType, uid, err = r.PollA()
	if err == nil {
		return
	}
	return r.PollB()
}

func (r *Reader) ExchangeApdu(capdu []byte) ([]byte, error) {
	capdu = append([]byte{0}, capdu...)
	resp, err := r.transceive(0xC2, capdu)
	if err == nil {
		buf := make([]byte, 64)
		r.d.Read(buf)
	}
	return resp, err
}

func (r *Reader) ExchangeApduStr(capdu string) (string, error) {
	bytes, err := hex.DecodeString(capdu)
	if err != nil {
		return "", nil
	}
	resp, err := r.ExchangeApdu(bytes)
	return strings.ToUpper(hex.EncodeToString(resp)), err
}
