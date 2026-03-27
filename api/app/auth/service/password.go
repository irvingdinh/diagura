package service

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonMemory  uint32 = 19456
	argonTime    uint32 = 2
	argonThreads uint8  = 1
	argonKeyLen  uint32 = 32
	argonSaltLen        = 16
)

// HashPassword hashes a password using argon2id with OWASP-recommended
// parameters and returns a PHC-formatted string.
func HashPassword(password string) (string, error) {
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argonMemory,
		argonTime,
		argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

// VerifyPassword checks a password against a PHC-formatted argon2id hash.
func VerifyPassword(phc, password string) (bool, error) {
	p, err := parsePHC(phc)
	if err != nil {
		return false, err
	}

	derived := argon2.IDKey([]byte(password), p.salt, p.time, p.memory, p.threads, uint32(len(p.hash)))

	return subtle.ConstantTimeCompare(derived, p.hash) == 1, nil
}

// NeedsRehash returns true if the PHC string was produced with parameters
// different from the current defaults, indicating the hash should be
// recomputed on the next successful login.
func NeedsRehash(phc string) bool {
	p, err := parsePHC(phc)
	if err != nil {
		return true
	}
	return p.memory != argonMemory || p.time != argonTime || p.threads != argonThreads
}

type phcParams struct {
	memory  uint32
	time    uint32
	threads uint8
	salt    []byte
	hash    []byte
}

func parsePHC(phc string) (*phcParams, error) {
	// $argon2id$v=19$m=19456,t=2,p=1$<salt>$<hash>
	parts := strings.Split(phc, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return nil, fmt.Errorf("invalid PHC string")
	}

	var p phcParams
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return nil, fmt.Errorf("parse version: %w", err)
	}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.memory, &p.time, &p.threads); err != nil {
		return nil, fmt.Errorf("parse params: %w", err)
	}

	var err error
	p.salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, fmt.Errorf("decode salt: %w", err)
	}
	p.hash, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, fmt.Errorf("decode hash: %w", err)
	}

	return &p, nil
}
