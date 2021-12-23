package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/FJSDS/gnettest/utils"
)

type Client struct {
	connS map[net.Conn]struct{}
}

func NewClient() *Client {
	c := &Client{
		connS: map[net.Conn]struct{}{},
	}
	return c
}

func (this_ *Client) Dail(addr string) {
	c, err := net.DialTimeout("tcp", addr, time.Second*3)
	utils.Must(err)
	this_.connS[c] = struct{}{}
	this_.handle(c)
}

const packetSize = 128

var (
	initCount = uint64(0)
)

func (this_ *Client) handle(c net.Conn) {
	index := atomic.AddUint64(&initCount, 100000000)
	expectedIndex := index
	go func() {
		defer c.Close()
		b := make([]byte, packetSize)
		time.Sleep(time.Millisecond * 100)
		for {
			binary.LittleEndian.PutUint64(b, index)
			index++
			_, err := c.Write(b)
			if err != nil {
				panic(err)
			}
			time.Sleep(time.Nanosecond * 10)
			//runtime.Gosched()
		}
	}()

	go func() {
		defer c.Close()
		b := make([]byte, packetSize)
		for {
			_, err := io.ReadFull(c, b)
			if err != nil {
				panic(err)
			}
			index := binary.LittleEndian.Uint64(b)
			if index != expectedIndex {
				panic(fmt.Sprintf("index[%d] != expectedIndex[%d]", index, expectedIndex))
			}
			atomic.AddUint64(&count, 1)
			expectedIndex++
		}
	}()
}

func main() {
	for i := 0; i < 12; i++ {
		Start("10.23.20.53:9922", 10)
	}

	oldCount := uint64(0)
	for {
		time.Sleep(time.Second)
		c := atomic.LoadUint64(&count)
		fmt.Println(c, c-oldCount)
		oldCount = c
	}
}

var (
	m     sync.Map
	count = uint64(0)
)

func Start(addr string, connCount int) {
	c := NewClient()
	for i := 0; i < connCount; i++ {
		c.Dail(addr)
	}
	m.Store(c, struct{}{})
}
