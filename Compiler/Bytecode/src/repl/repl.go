package repl

import (
	"bufio"
	"fmt"
	"io"
	"src/compiler"
	"src/lexer"
	"src/parser"
	"src/vm"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)

	for {
		fmt.Printf(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.New(line)
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Error()) != 0 {
			printParserErrors(out, p.Error())
			continue
		}

		comp := compiler.New()
		if err := comp.Compile(program); err != nil {
			fmt.Fprintf(out, "Woops! Compilation failed:\n %s\n", err)
			continue
		}

		machine := vm.New(comp.Bytecode())
		if err := machine.Run(); err != nil {
			fmt.Fprintf(out, "Woops! Executing bytecode failed:\n %s\n", err)
			continue
		}

		lastPopped := machine.LastPoppedStackElem()
		io.WriteString(out, lastPopped.Inspect())
		io.WriteString(out, "\n")
	}
}

const Rabbit_FACE = `                 
       /\_/\
       || ||  	
      ( o.o )
        >^<
       /   \
`

func printParserErrors(out io.Writer, errors []string) {
	io.WriteString(out, Rabbit_FACE)
	io.WriteString(out, "Woops! We jumped down the rabbit hole!\n")
	io.WriteString(out, " parser errors:\n")
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
