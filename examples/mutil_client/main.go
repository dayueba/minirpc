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
			go func(client minirpc.Client) {
				defer wg.Done()

				reply := new(int)
				err := client.Call("Arith", "Multiply", data.Args{
					A: 1,
					B: 2,
				}, reply)
				if err != nil || *reply != 2 {
					panic(err)
				}
				fmt.Println(*reply)
			}(client)
		}
	}

	wg.Wait()
}
