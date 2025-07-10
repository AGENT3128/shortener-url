package main

import "os"

func main() {
	if true {
		os.Exit(1) // want "direct call to os.Exit in main function of main package is prohibited"
	}

	if false {
		os.Exit(2) // want "direct call to os.Exit in main function of main package is prohibited"
	}
}
