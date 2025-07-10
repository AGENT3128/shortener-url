package main

import "os"

func main() {
	func() {
		os.Exit(1) // want "direct call to os.Exit in main function of main package is prohibited"
	}()

	go func() {
		os.Exit(2) // want "direct call to os.Exit in main function of main package is prohibited"
	}()
}
