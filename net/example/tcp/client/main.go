package main

import (
	"encoding/binary"
	"encoding/json"
	"net"

	net2 "github.com/happycrud/golib/net"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		panic(err)
	}
	x := net2.Hello{Name: "Jack", Greet: "sdfsdfasfdfadfasdfasdfasdf"}
	data, err := json.Marshal(x)
	if err != nil {
		panic(err)
	}
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(len(data)))

	conn.Write(append(buf, data...))

	//conn.Close()
	c := make(chan struct{})
	<-c
}
