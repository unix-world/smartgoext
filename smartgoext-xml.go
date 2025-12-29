
// GO Lang :: SmartGo Extra :: Smart.Go.Framework
// (c) 2020-present unix-world.org
// r.20251229.2358 :: STABLE
// [ XML ]

// REQUIRE: go 1.19 or later
package smartgoext

import (
	"strings"

	smart "github.com/unix-world/smartgo"
	"github.com/unix-world/smartgoext/xml-utils/xml2json"
)


//-----


func XmlConvertToJson(xmlData string) (string, error) {
	//--
	defer smart.PanicHandler() // for YAML Parser
	//--
	xml := strings.NewReader(xmlData) // xml is an io.Reader
	json, err := xml2json.Convert(xml)
	if(err != nil) {
		return "", err // returns empty string and the conversion error
	} //end if
	//--
	return json.String(), nil // returns the json as string, no error
	//--
} //END FUNCTION


//-----


// #END
