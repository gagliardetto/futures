package main

import (
	"fmt"
	"time"

	"github.com/gagliardetto/futures"
)

func main() {
	f := futures.New()

	go func() {
		time.Sleep(time.Second * 2)
		howManyReceived := f.Answer("one", 1, nil)
		fmt.Println("howManyReceived", howManyReceived)
	}()

	v, err := f.Ask("one")
	fmt.Println(v, err)

	///

	v, err = f.AskWithTimeout("two", time.Second)
	fmt.Println(v, err)

	///

	go func() {
		v, err := f.Ask("key")
		fmt.Println("key", v, err)

		time.Sleep(time.Second * 3)
		v, err = f.Ask("key")
		fmt.Println("key 2", v, err)
	}()
	// it takes time to register the listener;
	// if you set before registering a listener, the listener will not receive the response.
	time.Sleep(time.Second)

	fmt.Println("keyyyy")
	howMany := f.Answer("key", 33, nil)
	fmt.Println("/keyyyy", howMany)

	time.Sleep(time.Second * 5)
	howMany = f.Answer("key", 77, nil)
	fmt.Println("/keyyyy 2", howMany)

	time.Sleep(time.Minute)
}
