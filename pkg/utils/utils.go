package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/jinzhu/copier"
	"strconv"
)

func Sha256String(data string) string {
	return Sha256Bytes([]byte(data))
}

func Sha256Bytes(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func DeepCopy(dst interface{}, src interface{}) error {
	return copier.CopyWithOption(dst, src, copier.Option{
		DeepCopy: true,
	})
}

func FindStrInSlice(a []string, s string) int {
	for i, v := range a {
		if v == s {
			return i
		}
	}
	return -1
}

func ParseBoolOrFalse(s *string) bool {
	if s == nil {
		return false
	}
	b, err := strconv.ParseBool(*s)
	if err != nil {
		return false
	}
	return b
}

func StrPtr(s string) *string {
	return &s
}
