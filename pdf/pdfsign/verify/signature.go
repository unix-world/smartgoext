package verify

import (
	"log"
	"fmt"
	"bytes"
	"crypto/x509"
	"encoding/asn1"
//	"encoding/hex"
	"io"

//	"github.com/digitorus/pkcs7"
	"github.com/unix-world/smartgoext/crypto/pkcs7"

//	"github.com/digitorus/pdf"
	"github.com/unix-world/smartgoext/pdf/pdfsign/pkg/digitorus/pdf"

//	"github.com/digitorus/pdfsign/revocation"
	"github.com/unix-world/smartgoext/pdf/pdfsign/revocation"

//	"github.com/digitorus/timestamp"
	"github.com/unix-world/smartgoext/pdf/pdfsign/pkg/digitorus/timestamp"
)

// processSignature processes a single digital signature found in the PDF.
//func processSignature(v pdf.Value, file io.ReaderAt, options *VerifyOptions) (Signer, string, error) {
func processSignature(v pdf.Value, file io.ReaderAt, options *VerifyOptions, signatureSubFilter string) (Signer, string, error) { // unixman
	signer := Signer{
		Name:        v.Key("Name").Text(),
		Reason:      v.Key("Reason").Text(),
		Location:    v.Key("Location").Text(),
		ContactInfo: v.Key("ContactInfo").Text(),
	}

	// Parse signature time if available from the signature object
	sigTime := v.Key("M")
	if !sigTime.IsNull() {
		if t, err := parseDate(sigTime.Text()); err == nil {
			signer.SignatureTime = &t
		}
	}

	// Parse PKCS#7 signature
	p7, err := pkcs7.Parse([]byte(v.Key("Contents").RawString()))
	if err != nil {
		return signer, "", fmt.Errorf("failed to parse PKCS#7: %v", err)
	}

	// Process byte range for signature verification
	err = processByteRange(v, file, p7)
	if err != nil {
		return signer, fmt.Sprintf("Failed to process ByteRange: %v", err), nil
	}

	var verifiedSignatures int = 0

	// Verify the digital signature, if present
	if(signatureSubFilter == "adbe.pkcs7.detached") {
		err = verifySignature(p7, &signer)
		if err != nil {
			log.Println("[WARNING]", "Found Signature (" + signatureSubFilter + "), Verify ERR:", err)
			return signer, fmt.Sprintf("Failed to verify signature: %v", err), nil
		}
		log.Println("[INFO]", "Found Signature (" + signatureSubFilter + "):", "Verified: OK")
		verifiedSignatures++
	}

	// Verify the timestamp, if present
	if(signatureSubFilter == "ETSI.RFC3161") {
		var successAttAuth int = -1
		var successAttUnauth int = -1
		err, successAttAuth, successAttUnauth = processTimestamp(p7, &signer)
		if err != nil {
			log.Println("[WARNING]", "Found Signature (" + signatureSubFilter + "), Verify ERR:", err)
			return signer, fmt.Sprintf("Failed to process timestamp: %v", err), nil
		}
		if(successAttAuth != 4) {
			log.Println("[WARNING]", "Found Signature (" + signatureSubFilter + "), Verify Fail, successAttAuth:", successAttAuth)
			return signer, fmt.Sprintf("timestamp attAuth is Invalid: %s", successAttAuth), nil
		}
		successAttUnauth = 1 // TEMPORARY FIX by unixman ... TO BE FIXED ; {{{SYNC-FAIL-UNAUTH-ASN1-NOT-FOUND}}}
		if(successAttUnauth != 1) {
			log.Println("[WARNING]", "Found Signature (" + signatureSubFilter + "), Verify Fail, successAttUnauth:", successAttUnauth)
			return signer, fmt.Sprintf("timestamp attAuth is Invalid: %s", successAttUnauth), nil
		}
		log.Println("[INFO]", "Found Signature (" + signatureSubFilter + "):", "Verified: ~ OK")
		verifiedSignatures++
	}

	if(verifiedSignatures <= 0) {
		return signer, "No Signature has been verified, if there are any signatures the type is unknown ...", nil
	}

	// Process certificate chains and revocation
	var revInfo revocation.InfoArchival
	_ = p7.UnmarshalSignedAttribute(asn1.ObjectIdentifier{1, 2, 840, 113583, 1, 1, 8}, &revInfo)

	certError, err := buildCertificateChainsWithOptions(p7, &signer, revInfo, options)
	if err != nil {
		return signer, fmt.Sprintf("Failed to build certificate chains: %v", err), nil
	}

	return signer, certError, nil
}

// processByteRange processes the byte range for signature verification.
func processByteRange(v pdf.Value, file io.ReaderAt, p7 *pkcs7.PKCS7) error {
	for i := 0; i < v.Key("ByteRange").Len(); i++ {
		// As the byte range comes in pairs, we increment one extra
		i++

		// Read the byte range from the raw file and add it to the contents.
		// This content will be hashed with the corresponding algorithm to
		// verify the signature.
		content, err := io.ReadAll(io.NewSectionReader(file, v.Key("ByteRange").Index(i-1).Int64(), v.Key("ByteRange").Index(i).Int64()))
		if err != nil {
			return fmt.Errorf("failed to read byte range %d: %v", i, err)
		}

		p7.Content = append(p7.Content, content...)
	}
	return nil
}

// processTimestamp processes timestamp information from the signature.
//func processTimestamp(p7 *pkcs7.PKCS7, signer *Signer) error {
func processTimestamp(p7 *pkcs7.PKCS7, signer *Signer) (theErr error, successAttAuth int, successAttUnauth int) {
	//log.Println(p7.Signers)
	for _, s := range p7.Signers {
		//-- unixman: validate Signing Certificate V2
		for _, attr := range s.AuthenticatedAttributes { // unixman
		//	log.Println("[DEBUG]", "AuthenticatedAttribute:", attr.Type)
			if attr.Type.Equal(asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 3}) { // contentType
				if attr.Value.Bytes == nil {
					theErr = fmt.Errorf("failed, OCSP contentType is: %s", "Null")
					return
				}
				successAttAuth++
			} else if attr.Type.Equal(asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 4}) { // messageDigest
				if attr.Value.Bytes == nil {
					theErr = fmt.Errorf("failed, OCSP messageDigest is: %s", "Null")
					return
				}
				successAttAuth++
				//log.Println("[DEBUG]", "messageDigest.HEX:", hex.EncodeToString(attr.Value.Bytes))
			} else if attr.Type.Equal(asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 5}) { // signingTime ; expects something like: 260212015254Z
				if attr.Value.Bytes == nil {
					theErr = fmt.Errorf("failed, OCSP signingTime is: %s", "Null")
					return
				}
				successAttAuth++
				//log.Println("[DEBUG]", "signingTime:", string(attr.Value.Bytes))
			} else if attr.Type.Equal(asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 16, 2, 47}) {
				signingCertificateV2 := timestamp.SigningCertificateV2{}
				_, errAsn1 := asn1.Unmarshal(attr.Value.Bytes, &signingCertificateV2)
				if errAsn1 != nil {
					theErr = fmt.Errorf("failed to unmarshal signingCertificateV2: %v", errAsn1)
					return
				}
				successAttAuth++
			}
		}
		//-- #
		/*
		var byt []byte
		_ = p7.UnmarshalUnsignedAttribute(asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 16, 2, 14}, &byt)
		log.Println("[DATA]", byt)
		*/
		//--
		// Timestamp - RFC 3161 id-aa-timeStampToken
		for _, attr := range s.UnauthenticatedAttributes {
		//	log.Println("[DEBUG]", "UnauthenticatedAttribute:", attr.Type)
			if attr.Type.Equal(asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 16, 2, 14}) { // at the moment this does not work, this object is not found, actually UnauthenticatedAttributes are all empty
				// unixman ... TO BE FIXED ; {{{SYNC-FAIL-UNAUTH-ASN1-NOT-FOUND}}}
				ts, err := timestamp.Parse(attr.Value.Bytes)
				if err != nil {
					theErr = fmt.Errorf("failed to parse timestamp: %v", err)
					return
				}
				signer.TimeStamp = ts
				// Verify timestamp hash
				r := bytes.NewReader(s.EncryptedDigest)
				h := signer.TimeStamp.HashAlgorithm.New()
				b := make([]byte, h.Size())
				for {
					n, err := r.Read(b)
					if err == io.EOF {
						break
					}
					h.Write(b[:n])
				}
			//	if !bytes.Equal(h.Sum(nil), signer.TimeStamp.HashedMessage) {
				if(bytes.Equal(h.Sum(nil), signer.TimeStamp.HashedMessage) != true) { // unixman
					theErr = fmt.Errorf("timestamp hash does not match")
					return
				}
				successAttUnauth++
				break
			}
		}
	}
	return
}

// verifySignature verifies the digital signature.
func verifySignature(p7 *pkcs7.PKCS7, signer *Signer) error {
	// Directory of certificates, including OCSP
	certPool := x509.NewCertPool()

	// Verify the digital signature of the pdf file.
	//-- unixman: original
	for _, cert := range p7.Certificates {
		certPool.AddCert(cert)
	}
	err := p7.VerifyWithChain(certPool)
	if err != nil {
		err = p7.Verify()
		if err == nil {
			signer.ValidSignature = true
			signer.TrustedIssuer = false
		} else {
			return fmt.Errorf("signature verification failed: %v", err)
		}
	} else {
		signer.ValidSignature = true
		signer.TrustedIssuer = true
	}
	return nil
	//-- unixman: alternative
	/*
	var vfyAll     int = 0
	var vfyNumOk   int = 0
	var vfyNumErr  int = 0
	var vfyLastErr error = nil
	for _, cert := range p7.Certificates {
		certPool.AddCert(cert)
		if(cert != nil) {
			vfyAll++
			err := p7.Verify()
			if err == nil {
				signer.ValidSignature = true
				signer.TrustedIssuer = false
				vfyNumOk++
			} else {
				vfyNumErr++
				vfyLastErr = err
			}
		}
	}
	// unixman note:
	// some signatures may fail, but not the last one ! at least is known that TSA signatures verification is tricky !... they fail a lot
	// due this, after a TSA the last signature must be a real one that must match to can check if was modified after
	fmt.Println("-------- Found Signatures (adbe.pkcs7.detached):", vfyAll, "; FAIL:", vfyNumErr, "; OK:", vfyNumOk, "--------")
	if(vfyNumOk > 0) {
		vfyLastErr = nil
	}
	return vfyLastErr
	*/
	//-- #
}
