package nonmainpackage

import (
	"fmt"
	"os"
)

func SomeFunction() {
	fmt.Println("This is in a non-main package")
	os.Exit(1) // This should not trigger the analyzer
}
