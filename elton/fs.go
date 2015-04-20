package elton

import (
	"crypto/md5"
	"encoding/hex"
	"time"
)

func generateKey(name string) string {
	hasher := md5.New()
	hasher.Write([]byte(name + time.Now().String()))
	hash := hex.EncodeToString(hasher.Sum(nil))
	return string(hash[:2] + "/" + hash[2:])
}
