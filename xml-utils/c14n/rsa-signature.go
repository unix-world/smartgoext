
// r.20241020
// (c) 2023-2024 unix-world.org

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

	// ex: algo: x509.SHA1WithRSA

	defer smart.PanicHandler()

	certificatePEM = smart.StrTrimWhitespaces(certificatePEM)
	if(certificatePEM == "") {
		return false, nil
	}

	signatureB64 = smart.StrNormalizeSpaces(signatureB64)
	signatureB64 = smart.StrReplaceAll(signatureB64, " ", "")
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

	signature := smart.Base64Decode(signatureB64)
	if(smart.StrTrimWhitespaces(signature) == "") {
		return false, smart.NewError("Signature is Empty after Base64Decode")
	}

	errVfy := cert.CheckSignature(algo, []byte(data), []byte(signature))
	if(errVfy != nil) {
		return false, errVfy
	}

	return true, nil

}


func VerifyRSASignature(algo crypto.Hash, certificatePEM string, signatureB64 string, digestB64 string, modePKCS1 bool, optionsPSS *rsa.PSSOptions) (bool, error) {

	// ex: algo: crypto.SHA1
	// and the hash must be Sha1B64(data)

	defer smart.PanicHandler()

	certificatePEM = smart.StrTrimWhitespaces(certificatePEM)
	if(certificatePEM == "") {
		return false, nil
	}

	signatureB64 = smart.StrNormalizeSpaces(signatureB64)
	signatureB64 = smart.StrReplaceAll(signatureB64, " ", "")
	signatureB64 = smart.StrTrimWhitespaces(signatureB64)
	if(signatureB64 == "") {
		return false, nil
	}

	digestB64 = smart.StrTrimWhitespaces(digestB64)
	if(digestB64 == "") {
		return false, nil
	}

	block, _ := pem.Decode([]byte(certificatePEM))
	var cert *x509.Certificate
	cert, errParse := x509.ParseCertificate(block.Bytes)
	if(errParse != nil) {
		return false, errParse
	}
	rsaPublicKey := cert.PublicKey.(*rsa.PublicKey)
	/*
	pubKey := rsa.PublicKey{
		N: rsaPublicKey.N,
		E: rsaPublicKey.E,
	}
	*/

	signature := smart.Base64Decode(signatureB64)
	if(smart.StrTrimWhitespaces(signature) == "") {
		return false, smart.NewError("Signature is Empty after Base64Decode")
	}

	digest := smart.Base64Decode(digestB64)
	if(smart.StrTrimWhitespaces(digest) == "") {
		return false, smart.NewError("Hash is Empty after Base64Decode")
	}

	var errVfy error
	if(modePKCS1 == true) {
	//	errVfy = rsa.VerifyPKCS1v15(&pubKey, algo, []byte(digest), []byte(signature))
		errVfy = rsa.VerifyPKCS1v15(rsaPublicKey, algo, []byte(digest), []byte(signature))
	} else {
	//	errVfy = rsa.VerifyPSS(&pubKey, algo, []byte(digest), []byte(signature), optionsPSS)
		errVfy = rsa.VerifyPSS(rsaPublicKey, algo, []byte(digest), []byte(signature), optionsPSS)
	}
	if(errVfy != nil) {
		return false, errVfy
	}

	return true, nil

}

// #END
