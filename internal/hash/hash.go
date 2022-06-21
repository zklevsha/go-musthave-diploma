package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func Sign(key, msg string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(msg))
	return hex.EncodeToString(h.Sum(nil))

}

func Verify(msg, key, hash string) (bool, error) {
	sig, err := hex.DecodeString(hash)
	if err != nil {
		return false, err
	}

	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(msg))

	return hmac.Equal(sig, mac.Sum(nil)), nil
}
