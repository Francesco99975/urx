package helpers

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"net"
)

func GenerateNonce() (string, error) {
	bytes := make([]byte, 16) // 16 bytes nonce
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// Internal function — can be tested with fake reader
func generateNonceWithReader(reader io.Reader) (string, error) {
	b := make([]byte, 16)
	if _, err := io.ReadFull(reader, b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func IsPrivateIP(ip net.IP) bool {
	if ip == nil {
		return false
	}

	// Normalize to 16-byte form for consistent checks
	ip = ip.To16()
	if ip == nil {
		return false
	}

	// Built-in checks (Go handles many cases already)
	if ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsMulticast() ||
		ip.IsUnspecified() {
		return true
	}

	// Additional explicit IPv6 unique local addresses (fc00::/7)
	_, fc00, _ := net.ParseCIDR("fc00::/7")
	return fc00.Contains(ip)
}
