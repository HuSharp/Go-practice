package main

import (
	"fmt"
	"os"
	"src/repl"
)

func main() {
	fmt.Println("Welcome to be friends with HuSharp's Rabbit!")
	repl.Start(os.Stdin, os.Stdout)
}
