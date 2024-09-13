package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	pb "github.com/happycrud/golib/demo/helloworld/helloworld"
	"github.com/happycrud/golib/disc"
	"github.com/happycrud/golib/net/rpc/pjsonc"
)

const (
	defaultName = "world"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
	name = flag.String("name", defaultName, "Name to greet")
)

func main() {
	flag.Parse()

	conn, _ := disc.NewConn("live.live.helloworld")
	// Set up a connection to the server.
	// conn, err := grpc.Dial("etcd:///"+"helloworld", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultCallOptions(grpc.CallContentSubtype(protojson.JSON{}.Name())))
	// if err != nil {
	// 	log.Fatalf("did not connect: %v", err)
	// }
	defer conn.Close()
	fmt.Println(conn.Target())
	c := pb.NewGreeterClient(conn)
	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	for i := 0; i < 10; i++ {
		r, err := c.SayHello(ctx, &pb.HelloRequest{Name: "xx"})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		log.Printf("Greeting: %s", r.GetMessage())
	}

	req := `{"name":"world"}`
	resp := &pjsonc.Response{}
	err := conn.Invoke(ctx, "/helloworld.Greeter/SayHello", req, resp)
	log.Printf("Greeting: %+v err:%v", resp, err)
}
