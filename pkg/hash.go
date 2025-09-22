package pkg

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type HashConfig struct {
	Memory  uint32
	Time    uint32
	Thread  uint8
	KeyLen  uint32
	SaltLen uint32
}

func NewHashConfig() *HashConfig {
	return &HashConfig{}
}

func (h *HashConfig) SetConfig(memory, time, keylen, saltlen uint32, thread uint8) {
	h.KeyLen = keylen
	h.SaltLen = saltlen
	h.Memory = memory
	h.Time = time
	h.Thread = thread
}

func (h *HashConfig) UseRecommended() {
	h.KeyLen = 32
	h.SaltLen = 16
	h.Memory = 64 * 1024
	h.Time = 2
	h.Thread = 1
}

func (h *HashConfig) GenHash(password string) (string, error) {
	salt, err := h.genSalt()
	if err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, h.Time, h.Memory, h.Thread, h.KeyLen)
	// dalam penulisan hash ada format
	// $jenisKey$versiKey$konfigurasi(memory, time, thread)$salt$hash
	version := argon2.Version
	saltStr := base64.RawStdEncoding.EncodeToString(salt)
	hashStr := base64.RawStdEncoding.EncodeToString(hash)
	hashedPwd := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", version, h.Memory, h.Time, h.Thread, saltStr, hashStr)
	return hashedPwd, nil
}

func (h *HashConfig) genSalt() ([]byte, error) {
	salt := make([]byte, h.SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	return salt, nil
}

// func (h *HashConfig) CompareHashAndPassword(password, hashedPassword string) (bool, error) {
// 	result := strings.Split(hashedPassword, "$")
// 	fmt.Println(len(result), result)
// 	// cek panjang hasil split, kalau bukan 6 maka format hash invalid
// 	if len(result) != 6 {
// 		return false, errors.New("invalid hash format")
// 	}
// 	// cek kriptografi yang digunakan
// 	if result[1] != "argon2id" {
// 		return false, errors.New("invalid crypto method")
// 	}
// 	// cek versi nya
// 	var version int
// 	fmt.Sscanf(result[2], "v=%d", &version)
// 	if version != argon2.Version {
// 		return false, errors.New("invalid argon2id version")
// 	}
// 	// ambil konfigurasi memory, time dan thread
// 	if _, err := fmt.Sscanf(result[3], "m=%d,t=%d,p=%d", &h.Memory, &h.Time, &h.Thread); err != nil {
// 		return false, errors.New("invalid format")
// 	}
// 	// ambil nilai salt
// 	salt, err := base64.RawStdEncoding.DecodeString(result[4])
// 	if err != nil {
// 		return false, err
// 	}
// 	h.SaltLen = uint32(len(salt))
// 	// ambil nilai hash
// 	hash, err := base64.RawStdEncoding.DecodeString(result[5])
// 	if err != nil {
// 		return false, err
// 	}
// 	h.KeyLen = uint32(len(hash))

// 	// Comparison
// 	// Generate Hash dari password
// 	hashPwd := argon2.IDKey([]byte(password), salt, h.Time, h.Memory, h.Thread, h.KeyLen)
// 	// komparasi hasil hash dengan waktu tidak konstan
// 	// if slices.Compare(hash, hashPwd) != 0 {
// 	// 	return false, nil
// 	// }
// 	// komparasi hasil hash dengan waktu konstan (lebih aman dari timing attack di hash)
// 	if subtle.ConstantTimeCompare(hash, hashPwd) == 0 {
// 		return false, nil
// 	}
// 	return true, nil
// }

func (h *HashConfig) CompareHashAndPassword(password, hashedPassword string) (bool, error) {
	if hashedPassword == "" {
		return false, errors.New("empty hashed password")
	}

	result := strings.Split(hashedPassword, "$")
	if len(result) != 6 {
		return false, fmt.Errorf("invalid hash format: got %d parts → %v", len(result), result)
	}

	// cek crypto method
	if result[1] != "argon2id" {
		return false, errors.New("unsupported crypto method (expected argon2id)")
	}

	// cek versi argon2
	var version int
	if _, err := fmt.Sscanf(result[2], "v=%d", &version); err != nil {
		return false, fmt.Errorf("invalid version format: %s", result[2])
	}
	if version != argon2.Version {
		return false, fmt.Errorf("argon2 version mismatch: expected %d got %d", argon2.Version, version)
	}

	// parsing konfigurasi memory, time, thread
	if _, err := fmt.Sscanf(result[3], "m=%d,t=%d,p=%d", &h.Memory, &h.Time, &h.Thread); err != nil {
		return false, fmt.Errorf("invalid argon2 config format: %s", result[3])
	}

	// decode salt
	salt, err := base64.RawStdEncoding.DecodeString(result[4])
	if err != nil {
		return false, fmt.Errorf("invalid salt encoding: %v", err)
	}
	h.SaltLen = uint32(len(salt))

	// decode hash
	hash, err := base64.RawStdEncoding.DecodeString(result[5])
	if err != nil {
		return false, fmt.Errorf("invalid hash encoding: %v", err)
	}
	h.KeyLen = uint32(len(hash))

	// generate ulang hash dari password input
	hashPwd := argon2.IDKey([]byte(password), salt, h.Time, h.Memory, h.Thread, h.KeyLen)

	if subtle.ConstantTimeCompare(hash, hashPwd) == 0 {
		return false, nil
	}

	return true, nil
}
