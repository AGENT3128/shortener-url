package main

import "os"

func main() {
	helper()
}

func helper() {
	os.Exit(1) // this should not be flagged as it's not in main function
}
