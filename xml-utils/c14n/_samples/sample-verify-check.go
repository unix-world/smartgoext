package main

import (
	"log"

	"crypto"
	"crypto/x509"

	smart "github.com/unix-world/smartgo"
	"github.com/unix-world/smartgoext/xml-utils/c14n"
)

const (

theCertificate = `
-----BEGIN CERTIFICATE-----
base 64 certificate (PEM) data goes here: MII...
-----END CERTIFICATE-----
`

theSignature string = `
base 64 signature goes here: ...
`

checksum string = "base 64 SHA256 checksum goes here"

xmlFile string = "xml/test-signed.xml"

)

func main() {

	// The XML File need to be canonicalized before using C14N standard ...

	var isValid bool
	var errValid error

//	isValid, errValid = c14n.VerifyRSASignature(crypto.SHA256, theCertificate, theSignature, checksum, false, nil) // PSS
	isValid, errValid = c14n.VerifyRSASignature(crypto.SHA256, theCertificate, theSignature, checksum, true, nil) // PKCS1
	if(errValid != nil) {
		log.Println("[ERROR]", "Verify", errValid)
		return
	}
	if(!isValid) {
		log.Println("[ERROR]", "Verify: Invalid !")
		return
	}
	log.Println("OK", "Verify")

	data, errRd := smart.SafePathFileRead(xmlFile, false)
	if(errRd != "") {
		log.Println("[ERROR]", "File Read", errRd)
		return
	}
	cksum := smart.Sha256B64(data)
	if(cksum != checksum) {
		log.Println("[ERROR]", "Checksum does not match the file content", cksum)
		return
	}

	isValid, errValid = c14n.CheckRSASignature(x509.SHA256WithRSA, theCertificate, theSignature, data)
	if(errValid != nil) {
		log.Println("[ERROR]", "Check Data", errValid)
		return
	}
	if(!isValid) {
		log.Println("[ERROR]", "Check Data: Invalid !")
		return
	}
	log.Println("OK", "Check Data")

}
