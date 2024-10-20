
// r.20241020
// (c) 2023-2024 unix-world.org

package main

import (
	"log"

	"crypto"
	"crypto/x509"

	smart "github.com/unix-world/smartgo"

	"github.com/unix-world/smartgoext/xml-utils/c14n"
)

const (

// base 64 certificate (PEM) data goes here: MII...
theCertificate = `
-----BEGIN CERTIFICATE-----
MIIBvTCCASYCCQD55fNzc0WF7TANBgkqhkiG9w0BAQUFADAjMQswCQYDVQQGEwJK
UDEUMBIGA1UEChMLMDAtVEVTVC1SU0EwHhcNMTAwNTI4MDIwODUxWhcNMjAwNTI1
MDIwODUxWjAjMQswCQYDVQQGEwJKUDEUMBIGA1UEChMLMDAtVEVTVC1SU0EwgZ8w
DQYJKoZIhvcNAQEBBQADgY0AMIGJAoGBANGEYXtfgDRlWUSDn3haY4NVVQiKI9Cz
Thoua9+DxJuiseyzmBBe7Roh1RPqdvmtOHmEPbJ+kXZYhbozzPRbFGHCJyBfCLzQ
fVos9/qUQ88u83b0SFA2MGmQWQAlRtLy66EkR4rDRwTj2DzR4EEXgEKpIvo8VBs/
3+sHLF3ESgAhAgMBAAEwDQYJKoZIhvcNAQEFBQADgYEAEZ6mXFFq3AzfaqWHmCy1
ARjlauYAa8ZmUFnLm0emg9dkVBJ63aEqARhtok6bDQDzSJxiLpCEF6G4b/Nv/M/M
LyhP+OoOTmETMegAVQMq71choVJyOFE5BtQa6M/lCHEOya5QUfoRF2HF9EjRF44K
3OK+u3ivTSj3zwjtpudY5Xo=
-----END CERTIFICATE-----
`

// base 64 signature goes here: ...
theSignature string = `LVEBl4vQBuCV0qgSqDLFZlujYZ9lUXHMic7mXZkHSAKRMBrln4YkoYuG7s1bPJ1kP3wR5bkRtfWQiR+AdwEbs0HkJr7speLvHkZumR+6hHXCv1FozvQKg5HN42MMb0W3aorfyuPVYN4mwD8BvSppurS2pu09Kqt8jhE3JgC9T6I=`

// base 64 SHA256 checksum goes here
checksum string = "Ee8wai1D7m/NPi49N8DE3Gd6f8nhQv7GP8loHsABEL0="

xmlFile string = `<note><to>Test1</to><from>Test2</from><heading>Reminder</heading><body>This is a test</body></note>`

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

	// The XML File need to be canonicalized before using C14N standard ...

	var isValid bool
	var errValid error

	signature := theSignature

	data := xmlFile

	if(smart.StrTrimWhitespaces(data) == "") {
		log.Println("[ERROR]", "File is Empty")
		return
	}

	// if data is not already canonicalized, it should be, look at the sample-canonicalize.go ...

	cksum := smart.Sha256B64(data)
	if(cksum != checksum) {
		log.Println("[ERROR]", "Checksum does not match the file content", cksum, checksum)
		return
	}
	log.Println("[OK]", "Checksum SHA256")

	// simple check
	isValid, errValid = c14n.CheckRSASignature(x509.SHA1WithRSA, theCertificate, signature, data)
	if(errValid != nil) {
		log.Println("[ERROR]", "Check Data", errValid)
		return
	}
	if(!isValid) {
		log.Println("[ERROR]", "Check Data: Invalid !")
		return
	}
	log.Println("[OK]", "Check Data")

	// advanced check: need to know the mode: PKCS1 or PSS
	isValid, errValid = c14n.VerifyRSASignature(crypto.SHA1, theCertificate, signature, smart.Sha1B64(data), true, nil) // PKCS1
	if(errValid != nil) {
		log.Println("[ERROR]", "Verify Data Hash (PKCS1): FAILED: ", errValid)
		return
	}
	if(!isValid) {
		log.Println("[ERROR]", "Verify Data Hash (PKCS1): Invalid !")
		return
	}
	log.Println("[OK]", "Verify Data Hash (PKCS1)")

}

// #END
