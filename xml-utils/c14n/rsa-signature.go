
// (c) 2023 unix-world.org

// ex: signature
// [B64...]

// ex: certificate:
// -----BEGIN CERTIFICATE-----
// [PEM/B64....]
// -----END CERTIFICATE-----

package c14n

import (
	"encoding/pem"
	"crypto"
	"crypto/rsa"
	"crypto/x509"

	smart "github.com/unix-world/smartgo"
)


func CheckRSASignature(algo x509.SignatureAlgorithm, certificatePEM string, signatureB64 string, data string) (bool, error) {

	// ex: algo: x509.SHA256WithRSA

	certificatePEM = smart.StrTrimWhitespaces(certificatePEM)
	if(certificatePEM == "") {
		return false, nil
	}
	signatureB64 = smart.StrTrimWhitespaces(signatureB64)
	if(signatureB64 == "") {
		return false, nil
	}

	block, _ := pem.Decode([]byte(certificatePEM))
	var cert *x509.Certificate
	cert, errParse := x509.ParseCertificate(block.Bytes)
	if(errParse != nil) {
		return false, errParse
	}

	signature := smart.Base64Decode(smart.StrTrimWhitespaces(signatureB64))

	errVfy := cert.CheckSignature(algo, []byte(data), []byte(signature))
	if(errVfy != nil) {
		return false, errVfy
	}

	return true, nil

}


func VerifyRSASignature(algo crypto.Hash, certificatePEM string, signatureB64 string, hashB64 string, modePKCS1 bool, optionsPSS *rsa.PSSOptions) (bool, error) {

	// ex: algo: crypto.SHA256

	certificatePEM = smart.StrTrimWhitespaces(certificatePEM)
	if(certificatePEM == "") {
		return false, nil
	}
	signatureB64 = smart.StrTrimWhitespaces(signatureB64)
	if(signatureB64 == "") {
		return false, nil
	}
	hashB64 = smart.StrTrimWhitespaces(hashB64)
	if(hashB64 == "") {
		return false, nil
	}

	block, _ := pem.Decode([]byte(certificatePEM))
	var cert *x509.Certificate
	cert, errParse := x509.ParseCertificate(block.Bytes)
	if(errParse != nil) {
		return false, errParse
	}

	rsaPublicKey := cert.PublicKey.(*rsa.PublicKey)
	pubKey := rsa.PublicKey{
		N: rsaPublicKey.N,
		E: rsaPublicKey.E,
	}

	signature := smart.Base64Decode(signatureB64)
	hashed := smart.Base64Decode(hashB64)

	var errVfy error
	if(modePKCS1 == true) {
		errVfy = rsa.VerifyPKCS1v15(&pubKey, algo, []byte(hashed), []byte(signature))
	} else {
		errVfy = rsa.VerifyPSS(&pubKey, algo, []byte(hashed), []byte(signature), optionsPSS)
	}
	if(errVfy != nil) {
		return false, errVfy
	}

	return true, nil

}

