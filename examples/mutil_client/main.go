package main

import (
	"fmt"
	"sync"

	"github.com/dayueba/minirpc"
	data "github.com/dayueba/minirpc/examples"
)

func main() {
	var wg sync.WaitGroup
	clientsNum := 3
	clients := make([]minirpc.Client, clientsNum)
	for i := 0; i < clientsNum; i++ {
		client, err := minirpc.NewClient(":8080")
		defer client.Close()
		if err != nil {
			panic(err)
		}
		clients[i] = client
	}
	wg.Add(clientsNum * 10)
	for _, client := range clients {
		for i := 0; i < 10; i++ {
			go func(client minirpc.Client, i int) {
				defer wg.Done()

				reply := new(int)
				args := data.Args{
					A: i + 1,
					B: i + 2,
				}
				err := client.Call("Arith", "Multiply", args, reply)
				if err != nil || *reply != args.A*args.B {
					panic(err)
				}
				fmt.Println(*reply)
			}(client, i)
		}
	}

	wg.Wait()
}
