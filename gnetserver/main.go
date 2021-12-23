package main

import "github.com/panjf2000/gnet"

type Server struct {
	gnet.EventServer
}

func (this_ *Server) React(packet []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	return packet, gnet.None
}

func main() {
	s := &Server{}
	gnet.Serve(s, "tcp://0.0.0.0:9922", gnet.WithMulticore(true))
}
