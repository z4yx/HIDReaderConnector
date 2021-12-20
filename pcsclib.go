//
// Copyright © 2021 Yuxiang Zhang
//

package main

import "C"
import (
	"fmt"
	"unsafe"
)

var reader_ctx map[int]*Reader
var reader_ctx_cnt int

func main() {

}

//export ReaderConnectorInit
func ReaderConnectorInit() {
	reader_ctx = make(map[int]*Reader)
	reader_ctx_cnt = 1
}

//export MakeReaderConnector
func MakeReaderConnector() (handle int) {
	r := new(Reader)
	reader_ctx[reader_ctx_cnt] = r
	handle = reader_ctx_cnt
	reader_ctx_cnt++
	return
}

//export FreeReaderConnector
func FreeReaderConnector(handle int) {
	delete(reader_ctx, handle)
}

//export ReaderOpen
func ReaderOpen(handle int) {
	r := reader_ctx[handle]

	err := r.Open()
	if err != nil {
		fmt.Printf("err=%v\n", err)
	}
}

//export ReaderClose
func ReaderClose(handle int) {
	r := reader_ctx[handle]
	r.Close()
}

//export ReaderPwrCtl
func ReaderPwrCtl(handle int, on bool) {
	r := reader_ctx[handle]
	if on {
		r.PowerOn()
	} else {
		r.PowerOff()
	}
}

//export ReaderBeep
func ReaderBeep(handle int) {
	r := reader_ctx[handle]
	r.Beep()
}

//export ReaderPollA
func ReaderPollA(handle int) (rc int, cardType *C.char, uid *C.char, errstr *C.char) {
	r := reader_ctx[handle]
	var err error
	for i := 0; i < 10; i++ {
		gCardType, gUID, err := r.PollA()
		if err != nil {
			continue
		}
		return 0, C.CString(gCardType), C.CString(gUID), nil
	}
	return -1, nil, nil, C.CString(fmt.Sprintf("%v", err))
}

//export ReaderExchangeApduStr
func ReaderExchangeApduStr(handle int, apdu *C.char) (rc int, resp *C.char, errstr *C.char) {
	r := reader_ctx[handle]
	rapdu, err := r.ExchangeApduStr(C.GoString(apdu))
	if err != nil {
		return -1, nil, C.CString(fmt.Sprintf("%v", err))
	}
	return 0, C.CString(rapdu), nil
}

//export ReaderExchangeApdu
func ReaderExchangeApdu(handle int, apdu unsafe.Pointer, apduLen C.int) (rc int, respLen C.int, resp unsafe.Pointer, errstr *C.char) {
	r := reader_ctx[handle]
	rapdu, err := r.ExchangeApdu(C.GoBytes(apdu, apduLen))
	if err != nil {
		return -1, 0, nil, C.CString(fmt.Sprintf("%v", err))
	}
	// fmt.Println(rapdu)
	return 0, C.int(len(rapdu)), C.CBytes(rapdu), nil
}
