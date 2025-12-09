package md5

import (
	"crypto/md5"
	"fmt"
	"io"
)

func StringHex(content string) string {
	hash := md5.New()

	_, err := io.WriteString(hash, content)
	if err != nil {
		return "a"
	}

	md5Sum := hash.Sum(nil)
	return fmt.Sprintf("%x", md5Sum)
}
