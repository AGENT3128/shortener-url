package main

import "os"

func main() {
	os.Getenv("PATH")      // this should not be flagged
	os.Setenv("TEST", "1") // this should not be flagged
}
