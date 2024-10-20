
// r.20241020
// (c) 2023-2024 unix-world.org

package main

import (
	"log"

	smart "github.com/unix-world/smartgo"

	"github.com/unix-world/smartgoext/xml-utils/c14n"
	"github.com/unix-world/smartgoext/xml-utils/c14n/etree"
)

const (
rawXML string = `
<xml>
	<abc>xyz</abc><!-- comment -->
	<def></def>
</xml>
`
)


func LogToConsoleWithColors() {
	//--
	smart.ClearPrintTerminal()
	//--
	smart.LogToConsole("DEBUG", true) // colored, terminal
	//--
} //END FUNCTION


func main() {

	defer smart.PanicHandler()

	LogToConsoleWithColors()

	raw := etree.NewDocument()
	err := raw.ReadFromString(rawXML)
	if(err != nil) {
		log.Println("[ERROR]", err)
		return
	}
//	var canonicalizer c14n.Canonicalizer = c14n.MakeC14N11Canonicalizer()
	var canonicalizer c14n.Canonicalizer = c14n.MakeC14N10RecCanonicalizer()
	canonicalized, err := canonicalizer.Canonicalize(raw.Root())
	if(err != nil) {
		log.Println("[ERROR]", err)
		return
	}

	log.Println("[DATA]", "Raw",  rawXML)
	log.Println("[DATA]", "C14n", string(canonicalized))

}

// #END
