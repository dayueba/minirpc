package main

import (
	"sync"

	"github.com/dayueba/minirpc"
	data "github.com/dayueba/minirpc/examples"
)

func main() {
	client, err := minirpc.NewClient(":8080")
	defer client.Close()
	if err != nil {
		panic(err)
	}

	//reply := new(int)
	//err = client.Call("Arith", "Multiply", data.Args{
	//	A: 1,
	//	B: 2,
	//}, reply)
	//if err != nil {
	//	panic(err)
	//}
	//
	//fmt.Println("reply: ", *reply)

	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()

			reply := new(int)
			err = client.Call("Arith", "Multiply", data.Args{
				A: 1,
				B: 2,
			}, reply)
			if err != nil {
				panic(err)
			}

			//fmt.Println("reply: ", *reply)
		}()
	}
	wg.Wait()
}
