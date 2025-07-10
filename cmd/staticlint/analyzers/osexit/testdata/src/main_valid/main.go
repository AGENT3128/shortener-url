package main

import (
	"fmt"
	"log"
)

func main() {
	if err := doSomething(); err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println("Success")
}

func doSomething() error {
	return nil
}
