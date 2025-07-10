package other

import "os"

func main() {
	os.Exit(1) // this should not be flagged as it's not in main package
}
