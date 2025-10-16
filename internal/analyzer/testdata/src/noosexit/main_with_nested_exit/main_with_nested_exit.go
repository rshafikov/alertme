package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("This main function calls a function that has os.Exit")
	exitFunction()
}

func exitFunction() {
	os.Exit(1)
}
