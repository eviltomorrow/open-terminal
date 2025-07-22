package fs

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

func Sha256Sum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var (
		buf  = make([]byte, 1024*1024)
		hash = sha256.New()
	)
	for {
		n, err := file.Read(buf[0:])
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
		hash.Write(buf[:n])
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
