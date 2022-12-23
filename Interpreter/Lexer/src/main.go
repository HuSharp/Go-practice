package main

import (
	"fmt"
	"os"
	"src/repl"
)

func main() {
	fmt.Println("Welcome to Monkey Language!")
	repl.Start(os.Stdin, os.Stdout)
}
