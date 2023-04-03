package main

import (
	"time"

	"github.com/dayueba/minirpc"
	data "github.com/dayueba/minirpc/examples"
)

func main() {
	arith := new(data.Arith)
	server := minirpc.NewServer(":8080", minirpc.WithTimeout(200*time.Millisecond))
	err := server.Register(arith)
	if err != nil {
		panic(err)
	}

	if err = server.Start(); err != nil {
		panic(err)
	}
}
