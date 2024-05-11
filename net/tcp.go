package net

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
)

type Server struct {
}

func (s *Server) Listen() error {
	li, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		panic(err)
	}
	for {
		conn, err := li.Accept()
		if err != nil {
			panic(err)
		}
		fmt.Println("accept")
		go s.processConn(conn)
	}

}

type Hello struct {
	Name  string
	Greet string
}

func (s *Server) processConn(conn net.Conn) {

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	for {
		// read http request
		// parse
		// hander
		// response
		head := [4]byte{}
		_, err := io.ReadFull(reader, head[:])
		if err != nil {
			fmt.Println(1, err)
			return
		}
		bodylen := int32(binary.BigEndian.Uint32(head[:]))

		body := make([]byte, bodylen)
		n, err := io.ReadFull(reader, body[:])
		if err != nil {
			fmt.Println(2, err)
			return
		}
		if int32(n) != bodylen {
			fmt.Println(3, err)
			return
		}
		h := &Hello{}
		if err := json.Unmarshal(body, &h); err != nil {
			fmt.Println(4, err)
			return
		}
		fmt.Println(string(body), *h)

		writer.Write(body)
		writer.Flush()

	}
}
