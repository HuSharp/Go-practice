package repl

import (
	"bufio"
	"fmt"
	"io"
	"src/evaluator"
	"src/lexer"
	"src/object"
	"src/parser"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()

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

		evaluated := evaluator.Eval(program, env)
		if evaluated != nil {
			io.WriteString(out, evaluated.Inspect())
			io.WriteString(out, "\n")
		}
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
