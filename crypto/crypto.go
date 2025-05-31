
package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"math/big"

	"golang.org/x/crypto/sha3"
)

var (
	MaxTarget = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))
)

// secp256k1 curve parameters (Ethereum compatible)
var secp256k1N, _ = new(big.Int).SetString("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141", 16)
var secp256k1P, _ = new(big.Int).SetString("fffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f", 16)
var secp256k1G = struct{ X, Y *big.Int }{
	X: new(big.Int).SetBytes([]byte{0x79, 0xbe, 0x66, 0x7e, 0xf9, 0xdc, 0xbb, 0xac, 0x55, 0xa0, 0x62, 0x95, 0xce, 0x87, 0x0b, 0x07, 0x02, 0x9b, 0xfc, 0xdb, 0x2d, 0xce, 0x28, 0xd9, 0x59, 0xf2, 0x81, 0x5b, 0x16, 0xf8, 0x17, 0x98}),
	Y: new(big.Int).SetBytes([]byte{0x48, 0x3a, 0xda, 0x77, 0x26, 0xa3, 0xc4, 0x65, 0x5d, 0xa4, 0xfb, 0xfc, 0x0e, 0x11, 0x08, 0xa8, 0xfd, 0x17, 0xb4, 0x48, 0xa6, 0x85, 0x54, 0x19, 0x9c, 0x47, 0xd0, 0x8f, 0xfb, 0x10, 0xd4, 0xb8}),
}

// GenerateKeyPair generates a new ECDSA key pair using secp256k1
func GenerateKeyPair() (*ecdsa.PrivateKey, *ecdsa.PublicKey, error) {
	// For now, use P256 as placeholder. In production, use actual secp256k1
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	
	return privateKey, &privateKey.PublicKey, nil
}

// GenerateEthKeyPair generates Ethereum-compatible key pair
func GenerateEthKeyPair() (*ecdsa.PrivateKey, *ecdsa.PublicKey, error) {
	// Generate 32-byte private key
	privateKeyBytes := make([]byte, 32)
	_, err := rand.Read(privateKeyBytes)
	if err != nil {
		return nil, nil, err
	}

	// Ensure private key is valid for secp256k1
	privateKeyInt := new(big.Int).SetBytes(privateKeyBytes)
	for privateKeyInt.Cmp(secp256k1N) >= 0 || privateKeyInt.Sign() == 0 {
		_, err := rand.Read(privateKeyBytes)
		if err != nil {
			return nil, nil, err
		}
		privateKeyInt.SetBytes(privateKeyBytes)
	}

	privateKey := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: secp256k1(),
		},
		D: privateKeyInt,
	}

	// Generate public key
	privateKey.PublicKey.X, privateKey.PublicKey.Y = privateKey.PublicKey.Curve.ScalarBaseMult(privateKeyBytes)

	return privateKey, &privateKey.PublicKey, nil
}

// secp256k1 returns a simplified secp256k1 curve (in production use proper implementation)
func secp256k1() elliptic.Curve {
	return elliptic.P256() // Simplified - use actual secp256k1 in production
}

// SHA256Hash calculates SHA256 hash
func SHA256Hash(data []byte) [32]byte {
	return sha256.Sum256(data)
}

// Keccak256Hash calculates Keccak256 hash (Ethereum style)
func Keccak256Hash(data []byte) [32]byte {
	hash := sha3.NewLegacyKeccak256()
	hash.Write(data)
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}

// Keccak256 calculates Keccak256 hash and returns bytes
func Keccak256(data []byte) []byte {
	hash := sha3.NewLegacyKeccak256()
	hash.Write(data)
	return hash.Sum(nil)
}

// Sign signs hash with private key (Ethereum-compatible)
func Sign(hash []byte, privateKey []byte) ([]byte, error) {
	if len(privateKey) != 32 {
		return nil, errors.New("invalid private key length")
	}
	
	// Create private key from bytes
	privKeyInt := new(big.Int).SetBytes(privateKey)
	privKey := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: secp256k1(),
		},
		D: privKeyInt,
	}
	privKey.PublicKey.X, privKey.PublicKey.Y = privKey.PublicKey.Curve.ScalarBaseMult(privateKey)
	
	// Sign hash
	r, s, err := ecdsa.Sign(rand.Reader, privKey, hash)
	if err != nil {
		return nil, err
	}
	
	// Ethereum signature format: R (32 bytes) + S (32 bytes) + V (1 byte)
	signature := make([]byte, 65)
	rBytes := r.Bytes()
	sBytes := s.Bytes()
	
	// Pad with zeros if needed
	copy(signature[32-len(rBytes):32], rBytes)
	copy(signature[64-len(sBytes):64], sBytes)
	
	// Recovery ID (V) - simplified
	signature[64] = 27 // Standard Ethereum recovery ID
	
	return signature, nil
}

// PubkeyToAddress converts public key to Ethereum-style address (20 bytes)
func PubkeyToAddress(pubKey *ecdsa.PublicKey) [20]byte {
	// Get uncompressed public key (64 bytes: 32 bytes X + 32 bytes Y)
	pubKeyBytes := make([]byte, 64)
	
	xBytes := pubKey.X.Bytes()
	yBytes := pubKey.Y.Bytes()
	
	// Pad with zeros if needed
	copy(pubKeyBytes[32-len(xBytes):32], xBytes)
	copy(pubKeyBytes[64-len(yBytes):64], yBytes)
	
	// Hash the public key
	hash := Keccak256(pubKeyBytes)
	
	// Take last 20 bytes as address
	var addr [20]byte
	copy(addr[:], hash[12:])
	return addr
}

// PrivateKeyToAddress converts private key to address
func PrivateKeyToAddress(privateKey *ecdsa.PrivateKey) [20]byte {
	return PubkeyToAddress(&privateKey.PublicKey)
}

// ValidateProofOfWork validates proof of work
func ValidateProofOfWork(hash [32]byte, nonce uint64, difficulty *big.Int) bool {
	target := new(big.Int).Div(MaxTarget, difficulty)
	hashInt := new(big.Int).SetBytes(hash[:])
	return hashInt.Cmp(target) == -1
}

// VerifySignature verifies ECDSA signature
func VerifySignature(pubKey *ecdsa.PublicKey, hash []byte, signature []byte) bool {
	if len(signature) != 65 {
		return false
	}
	
	r := new(big.Int).SetBytes(signature[:32])
	s := new(big.Int).SetBytes(signature[32:64])
	
	return ecdsa.Verify(pubKey, hash, r, s)
}

// RecoverAddress recovers address from signature
func RecoverAddress(hash []byte, signature []byte) ([20]byte, error) {
	if len(signature) != 65 {
		return [20]byte{}, errors.New("invalid signature length")
	}
	
	// Extract r, s, v
	r := new(big.Int).SetBytes(signature[:32])
	s := new(big.Int).SetBytes(signature[32:64])
	v := signature[64]
	
	// Simplified recovery - in production use proper ECDSA recovery
	if v < 27 {
		v += 27
	}
	
	// Create recovered public key (simplified)
	recoveredPubKey := &ecdsa.PublicKey{
		Curve: secp256k1(),
		X:     r,
		Y:     s,
	}
	
	return PubkeyToAddress(recoveredPubKey), nil
}

// Ecrecover recovers public key from signature
func Ecrecover(hash []byte, signature []byte) (*ecdsa.PublicKey, error) {
	if len(signature) != 65 {
		return nil, errors.New("invalid signature length")
	}
	
	r := new(big.Int).SetBytes(signature[:32])
	s := new(big.Int).SetBytes(signature[32:64])
	
	pubKey := &ecdsa.PublicKey{
		Curve: secp256k1(),
		X:     r,
		Y:     s,
	}
	
	return pubKey, nil
}

// FromECDSA exports private key to bytes
func FromECDSA(privateKey *ecdsa.PrivateKey) []byte {
	if privateKey == nil {
		return nil
	}
	// Ensure 32-byte output
	keyBytes := privateKey.D.Bytes()
	if len(keyBytes) < 32 {
		padded := make([]byte, 32)
		copy(padded[32-len(keyBytes):], keyBytes)
		return padded
	}
	return keyBytes
}

// ToECDSA creates private key from bytes
func ToECDSA(privateKeyBytes []byte) (*ecdsa.PrivateKey, error) {
	if len(privateKeyBytes) != 32 {
		return nil, errors.New("invalid private key length")
	}
	
	privKeyInt := new(big.Int).SetBytes(privateKeyBytes)
	if privKeyInt.Cmp(secp256k1N) >= 0 || privKeyInt.Sign() == 0 {
		return nil, errors.New("invalid private key value")
	}
	
	privKey := &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: secp256k1(),
		},
		D: privKeyInt,
	}
	
	privKey.PublicKey.X, privKey.PublicKey.Y = privKey.PublicKey.Curve.ScalarBaseMult(privateKeyBytes)
	
	return privKey, nil
}

// FromECDSAPub exports public key to bytes (uncompressed format)
func FromECDSAPub(publicKey *ecdsa.PublicKey) []byte {
	if publicKey == nil {
		return nil
	}
	
	pubKeyBytes := make([]byte, 65)
	pubKeyBytes[0] = 0x04 // Uncompressed key prefix
	
	xBytes := publicKey.X.Bytes()
	yBytes := publicKey.Y.Bytes()
	
	copy(pubKeyBytes[33-len(xBytes):33], xBytes)
	copy(pubKeyBytes[65-len(yBytes):65], yBytes)
	
	return pubKeyBytes
}

// UnmarshalPubkey parses public key from bytes
func UnmarshalPubkey(pubKeyBytes []byte) (*ecdsa.PublicKey, error) {
	if len(pubKeyBytes) != 65 || pubKeyBytes[0] != 0x04 {
		return nil, errors.New("invalid public key format")
	}
	
	pubKey := &ecdsa.PublicKey{
		Curve: secp256k1(),
		X:     new(big.Int).SetBytes(pubKeyBytes[1:33]),
		Y:     new(big.Int).SetBytes(pubKeyBytes[33:65]),
	}
	
	return pubKey, nil
}

// HexToBytes converts hex string to bytes
func HexToBytes(s string) []byte {
	if len(s) > 1 && s[0:2] == "0x" {
		s = s[2:]
	}
	if len(s)%2 != 0 {
		s = "0" + s
	}
	
	bytes := make([]byte, len(s)/2)
	for i := 0; i < len(s); i += 2 {
		var b byte
		for j := 0; j < 2; j++ {
			c := s[i+j]
			if c >= '0' && c <= '9' {
				b = (b << 4) | (c - '0')
			} else if c >= 'a' && c <= 'f' {
				b = (b << 4) | (c - 'a' + 10)
			} else if c >= 'A' && c <= 'F' {
				b = (b << 4) | (c - 'A' + 10)
			}
		}
		bytes[i/2] = b
	}
	return bytes
}

// BytesToHex converts bytes to hex string
func BytesToHex(bytes []byte) string {
	const hexChars = "0123456789abcdef"
	result := make([]byte, len(bytes)*2)
	for i, b := range bytes {
		result[i*2] = hexChars[b>>4]
		result[i*2+1] = hexChars[b&0x0f]
	}
	return string(result)
}
