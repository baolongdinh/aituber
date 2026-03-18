package siwe

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// GenerateNonce creates a random hex nonce for SIWE challenge
func GenerateNonce() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// BuildMessage returns the SIWE message that the frontend must sign
func BuildMessage(address, nonce string) string {
	return fmt.Sprintf(
		"Chào mừng bạn đến với ViralCraft!\n\nĐịa chỉ ví: %s\nNonce: %s",
		strings.ToLower(address),
		nonce,
	)
}

// VerifySignature checks that `sig` was produced by signing `message` with the private key
// corresponding to `expectedAddress`. Returns nil if valid.
func VerifySignature(expectedAddress, message, sig string) error {
	// Prefix the message per eth_sign standard (EIP-191)
	prefixed := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
	hash := crypto.Keccak256Hash([]byte(prefixed))

	// Decode hex signature
	sigBytes, err := hex.DecodeString(strings.TrimPrefix(sig, "0x"))
	if err != nil {
		return fmt.Errorf("invalid signature hex: %w", err)
	}
	if len(sigBytes) != 65 {
		return fmt.Errorf("invalid signature length: %d", len(sigBytes))
	}

	// Fix recovery id (v): Ethereum uses 27/28 instead of 0/1
	if sigBytes[64] >= 27 {
		sigBytes[64] -= 27
	}

	// Recover public key from signature
	pubKey, err := crypto.SigToPub(hash.Bytes(), sigBytes)
	if err != nil {
		return fmt.Errorf("failed to recover public key: %w", err)
	}

	// Derive address from public key and compare
	recovered := crypto.PubkeyToAddress(*pubKey)
	expected := common.HexToAddress(expectedAddress)

	if !strings.EqualFold(recovered.Hex(), expected.Hex()) {
		return fmt.Errorf("signature mismatch: got %s, expected %s", recovered.Hex(), expected.Hex())
	}
	return nil
}
