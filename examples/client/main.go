package main

import (
	"fmt"

	"github.com/dayueba/minirpc"
	data "github.com/dayueba/minirpc/examples"
)

func main() {
	client, err := minirpc.NewClient(":8080")
	if err != nil {
		panic(err)
	}
	defer client.Close()

	reply := new(int)
	err = client.Call("Arith", "Multiply", data.Args{
		A: 1,
		B: 2,
	}, reply)
	if err != nil {
		panic(err)
	}

	fmt.Println("reply: ", *reply)
}
