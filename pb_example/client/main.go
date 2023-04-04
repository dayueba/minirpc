package main

import (
	"fmt"

	"github.com/dayueba/minirpc"
	"github.com/dayueba/minirpc/pb_example/proto"
	"github.com/dayueba/minirpc/protocol"
)

func main() {
	client, err := minirpc.NewClient(":8972", minirpc.WithSerializeType(protocol.ProtoBuffer))
	if err != nil {
		panic(err)
	}
	defer client.Close()

	var reply proto.BenchmarkMessage

	err = client.Call("Hello", "Say", proto.PrepareArgs(), &reply)
	if err != nil {
		panic(err)
	}

	fmt.Println("reply: ", reply)
}
