package main

import (
	"log"

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

func main() {

	raw := etree.NewDocument()
	err := raw.ReadFromString(rawXML)
	if(err != nil) {
		log.Println("[ERROR]", err)
		return
	}
	var canonicalizer c14n.Canonicalizer = c14n.MakeC14N11Canonicalizer()
	canonicalized, err := canonicalizer.Canonicalize(raw.Root())
	if(err != nil) {
		log.Println("[ERROR]", err)
		return
	}

	log.Println("[DATA] Raw",  rawXML)
	log.Println("[DATA] C14n", string(canonicalized))

}
