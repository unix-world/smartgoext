package main

import (
	"fmt"
	"os"

	"github.com/unix-world/smartgoext/arithmetic-parser/lexer"
	"github.com/unix-world/smartgoext/arithmetic-parser/parser"
)

const (
	DEBUG bool      = false

	mathExpr string = "1 + 3*4"
)

func main() {

	lxr := lexer.Lex(mathExpr)
	psr := parser.NewParser(lxr)

	tree := psr.Parse()

	if DEBUG {
		fmt.Println("******** [DEBUG] ********")
		tree.GraphNode(os.Stdout) // prints a GraphViz dot format representation of the parse tree
		fmt.Println("******** /[DEBUG] ********")
	}

	fmt.Printf("Original expression: %q\n", mathExpr)
	fmt.Printf("Reconstituted expression: %q\n", tree)
	fmt.Printf("Result (calculated): %s\n", tree.Eval())

}
