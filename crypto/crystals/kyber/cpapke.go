package kyber

// modified by unixman
// v.20231228
// includes security patches (kyberslash1, kyberslash2) by kudelskisecurity: r.20240929
// (c) 2023-2024 unix-world.org

import (
	"errors"
	"crypto/rand"

//	"golang.org/x/crypto/sha3"
	"github.com/unix-world/smartgo/crypto/sha3"
)

//PKEKeyGen creates a public and private key pair.
//A 32 byte long seed can be given as argument. If a nil seed is given, the seed is generated using Go crypto's random number generator.
//The keys returned are packed into byte arrays.
//unixman: returns err, publicKey, privateKey
//func (k *Kyber) PKEKeyGen(seed []byte) ([]byte, []byte) { // modified by unixman
func (k *Kyber) PKEKeyGen(seed []byte) (error, []byte, []byte) {
	//-- fix by unixman
//	if seed == nil || len(seed) != SEEDBYTES {
//		seed = make([]byte, SEEDBYTES)
//		rand.Read(seed)
//	}
	if(len(seed) != SEEDBYTES) {
		if(seed == nil) {
			seed = make([]byte, SEEDBYTES)
			rand.Read(seed)
		} else {
			return errors.New("PKE Seed must be exactly 32 bytes"), nil, nil
		}
	}
	//-- #end fix

	K := k.params.K
	ETA1 := k.params.ETA1

	var rho, sseed [SEEDBYTES]byte
	state := sha3.New512()
	state.Write(seed)
	hash := state.Sum(nil)
	copy(rho[:], hash[:32])
	copy(sseed[:], hash[32:])

	Ahat := expandSeed(rho[:], false, K)

	shat := make(Vec, K)
	for i := 0; i < K; i++ {
		shat[i] = polyGetNoise(ETA1, sseed[:], byte(i))
		shat[i].ntt()
		shat[i].reduce()
	}

	ehat := make(Vec, K)
	for i := 0; i < K; i++ {
		ehat[i] = polyGetNoise(ETA1, sseed[:], byte(i+K))
		ehat[i].ntt()
	}

	t := make(Vec, K)
	for i := 0; i < K; i++ {
		t[i] = vecPointWise(Ahat[i], shat, K)
		t[i].toMont()
		t[i] = add(t[i], ehat[i])
		t[i].reduce()
	}

	return nil, k.PackPK(&PublicKey{T: t, Rho: rho[:]}), k.PackPKESK(&PKEPrivateKey{S: shat})
}

//Encrypt generates the encryption of a message using a public key.
//A 32 byte long seed can be given as argument (r). If a nil seed is given, the seed is generated using Go crypto's random number generator.
//The ciphertext returned is packed into a byte array.
//If an error occurs during the encrpytion process, a nil array is returned.
//unixman: this encrypts just one block of 32 bytes
//func (k *Kyber) Encrypt(packedPK, msg, r []byte) []byte { // modified by unixman
func (k *Kyber) Encrypt(packedPK, msg, r []byte) (error, []byte) {

//	if len(msg) < n/8 {
//		return errors.New("Message is too short to be encrypted"), nil
	if len(msg) != n/8 { // bug fix by unixman ; the algo does not support longer messages than 32 bytes per block
		return errors.New("Message to be encrypted must be exactly 32 bytes"), nil
	}

	if len(packedPK) != k.SIZEPK() {
		return errors.New("Cannot encrypt with this public key, size does not match"), nil
	}

	if len(r) != SEEDBYTES {
		if(r == nil) { // fix by unixman
			r = make([]byte, SEEDBYTES)
			rand.Read(r[:])
		} else { // fix by unixman
			return errors.New("The seed must be exactly 32 bytes"), nil
		}
	}

	K := k.params.K
	errKU, pk := k.UnpackPK(packedPK)
	if(errKU != nil) {
		return errKU, nil
	}
	Ahat := expandSeed(pk.Rho[:], true, K)

	sp := make(Vec, K)
	for i := 0; i < K; i++ {
		sp[i] = polyGetNoise(k.params.ETA1, r[:], byte(i))
		sp[i].ntt()
		sp[i].reduce()
	}
	ep := make(Vec, K)
	for i := 0; i < K; i++ {
		ep[i] = polyGetNoise(eta2, r[:], byte(i+K))
		ep[i].ntt()
	}
	epp := polyGetNoise(eta2, r[:], byte(2*K))
	epp.ntt()

	u := make(Vec, K)
	for i := 0; i < K; i++ {
		u[i] = vecPointWise(Ahat[i], sp, K)
		u[i].toMont()
		u[i] = add(u[i], ep[i])
		u[i].invntt()
		u[i].reduce()
		u[i].fromMont()
	}

	m := polyFromMsg(msg)
	m.ntt()

	v := vecPointWise(pk.T, sp, K)
	v.toMont()
	v = add(v, epp)
	v = add(v, m)
	v.invntt()
	v.reduce()
	v.fromMont()

	c := make([]byte, k.params.SIZEC)
	copy(c[:], u.compress(k.params.DU, K))
	copy(c[K*k.params.DU*n/8:], v.compress(k.params.DV))
	return nil, c[:]
}

//Decrypt decrypts a ciphertext given a secret key.
//The secret key and ciphertext must be give as packed byte array.
//The recovered message is returned as byte array.
//If an error occurs durirng the decryption process (wrong key format for example), a nil message is returned.
//unixman: this decrypts just one block of 32 bytes
//func (k *Kyber) Decrypt(packedSK, c []byte) []byte { // modified by unixman
func (k *Kyber) Decrypt(packedSK, c []byte) (error, []byte) {
	if len(c) != k.SIZEC() || len(packedSK) != k.SIZEPKESK() {
		return errors.New("Cannot decrypt, inputs do not have correct size"), nil
	}
	errKU, sk := k.UnpackPKESK(packedSK)
	if(errKU != nil) {
		return errKU, nil
	}
	K := k.params.K
	uhat := decompressVec(c[:K*k.params.DU*n/8], k.params.DU, K)
	uhat.ntt(K)
	v := decompressPoly(c[K*k.params.DU*n/8:], k.params.DV)
	v.ntt()

	m := vecPointWise(sk.S, uhat, K)
	m.toMont()
	m = sub(v, m)
	m.invntt()
	m.reduce()
	m.fromMont()

	return nil, polyToMsg(m)
}
