package main

import (
	"fmt"
	"os"
)

func Greet(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}

func main() {
	fmt.Println(Greet(os.Args[1]))
}