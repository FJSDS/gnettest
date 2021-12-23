package main

import (
	"encoding/binary"
	"fmt"
	"github.com/FJSDS/gnettest/utils"
	"github.com/panjf2000/gnet"
	"sync/atomic"
	"time"
)

type BytesINFO struct {
	Data  []byte
	Index uint64
}

type Client struct {
	gnet.EventServer
	connS map[gnet.Conn]struct{}
	c     *gnet.Client
}

func NewClient() *Client {
	c := &Client{
		connS: map[gnet.Conn]struct{}{},
	}
	gc, err := gnet.NewClient(c)
	utils.Must(err)
	c.c = gc
	gc.Start()
	return c
}

func (this_ *Client) Start(addr string, count int) {
	go func() {
		for i := 0; i < count; i++ {
			c, err := this_.c.Dial("tcp4", addr)
			utils.Must(err)
			this_.connS[c] = struct{}{}
		}
	}()
}

const packetSize = 128

func (this_ *Client) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	// startIndex := uint64(0) //atomic.AddUint64(&initIndex, 10000000)
	startIndex := atomic.AddUint64(&initIndex, 10000000)
	bi := &BytesINFO{Index: startIndex}
	c.SetContext(bi)
	go func() {

		index := startIndex
		for {
			b := make([]byte, packetSize)
			binary.LittleEndian.PutUint64(b, index)
			index++
			err := c.AsyncWrite(b)
			if err != nil {
				panic(err)
			}
			time.Sleep(time.Millisecond * 10)
		}
	}()
	return nil, gnet.None
}

var (
	initIndex = uint64(0)
)

func (this_ *Client) React(packet []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	bi := c.Context().(*BytesINFO)
	bi.Data = append(bi.Data, packet...)
	temp := bi.Data
	for len(temp) >= packetSize {
		index := binary.LittleEndian.Uint64(temp[:8])
		if bi.Index != index {
			panic(fmt.Sprintf("bi.Index[%d] != index[%d]", bi.Index, index))
		}
		bi.Index++
		atomic.AddUint64(&count, 1)
		temp = temp[packetSize:]
	}
	if len(temp) > 0 {
		copy(bi.Data, temp)
		bi.Data = bi.Data[:len(temp)]
	} else {
		bi.Data = bi.Data[:0]
	}

	return nil, gnet.None
}

func (this_ *Client) OnClosed(c gnet.Conn, err error) (action gnet.Action) {
	fmt.Println("close:", c.RemoteAddr(), err)
	return 0
}

func main() {
	Start("10.23.20.53:9922", 40)
	oldCount := uint64(0)
	for {
		time.Sleep(time.Second)
		c := atomic.LoadUint64(&count)
		fmt.Println(c, c-oldCount)
		oldCount = c
	}
}

var (
	m     = map[*Client]struct{}{}
	count uint64
)

func Start(addr string, count int) {
	for i := 0; i < count; i++ {
		c := NewClient()
		c.Start(addr, 20)
		m[c] = struct{}{}
	}
}
