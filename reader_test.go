//
// Copyright © 2021 Yuxiang Zhang
//
package main

import (
	"fmt"
	"testing"
	"time"
)

func TestReaderFunction(t *testing.T) {
	r := new(Reader)
	err := r.Open()
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	r.PowerOn()
	for i := 0; i < 10; i++ {
		cardType, uid, err := r.PollA()
		if err != nil {
			fmt.Println(err)
			time.Sleep(time.Millisecond * 500)
			continue
		}
		fmt.Printf("%s %s\n", cardType, uid)
		r.Beep()
		resp, err := r.ExchangeApduStr("00A40000")
		fmt.Printf("%s %v\n", resp, err)
		break
	}
	r.Close()
}
