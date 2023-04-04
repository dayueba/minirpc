package main

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/dayueba/minirpc"
	"github.com/dayueba/minirpc/pb_example/proto"
)

const delay = 0

type Hello int

func (t *Hello) Say(ctx context.Context, args *proto.BenchmarkMessage, reply *proto.BenchmarkMessage) error {
	args.Field1 = "OK"
	args.Field2 = 100
	*reply = *args
	fmt.Println(11111)
	if delay > 0 {
		time.Sleep(delay)
	} else {
		runtime.Gosched()
	}
	return nil
}

func main() {
	server := minirpc.NewServer(":8972", minirpc.WithTimeout(200*time.Millisecond))
	err := server.Register(new(Hello))
	if err != nil {
		panic(err)
	}

	if err = server.Start(); err != nil {
		panic(err)
	}
}
