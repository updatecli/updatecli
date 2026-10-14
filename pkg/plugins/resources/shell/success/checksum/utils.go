package checksum

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	"github.com/updatecli/updatecli/pkg/core/log"
)

func getChecksum(filename string) string {
	f, err := os.Open(filename)
	if err != nil {
		log.Debugln(err)
		return ""
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Debugln(err)
		return ""
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}
