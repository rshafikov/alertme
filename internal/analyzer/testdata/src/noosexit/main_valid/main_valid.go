package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("This is a valid main function without os.Exit")
	defer func() {
	}()

	nestedFunc()
}

func nestedFunc() {
	os.Exit(1)
}
