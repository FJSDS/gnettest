package main

import (
	"github.com/FJSDS/gnettest/reuseport"
	"github.com/FJSDS/gnettest/utils"
	"net"
	"time"
)

type Server struct {
}

func NewServer() *Server {
	s := &Server{}
	return s
}

func (this_ *Server) Listen(addr string) {
	l, err := reuseport.Listen("tcp4", addr)
	utils.Must(err)
	defer l.Close()
	var tempDelay time.Duration
	for {
		c, err := l.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				time.Sleep(tempDelay)
				continue
			}
		}
		this_.handle(c)
	}
}

func (this_ *Server) handle(c net.Conn) {
	go func() {
		defer c.Close()
		b := make([]byte, 4096)
		for {
			n, err := c.Read(b)
			if err != nil {
				return
			}
			temp := b[:n]
			for {
				n, err := c.Write(temp)
				if err != nil {
					return
				}
				if n == len(temp) {
					break
				}
				temp = temp[n:]
			}
		}
	}()
}

func main() {
	s := NewServer()
	s.Listen(":9922")
}
