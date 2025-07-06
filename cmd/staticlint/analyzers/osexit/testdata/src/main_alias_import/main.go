package main

import myos "os"

func main() {
	myos.Exit(1) // want "direct call to os.Exit in main function of main package is prohibited"
}
