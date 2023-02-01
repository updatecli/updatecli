package checksum

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

func getChecksum(filename string) string {
	f, err := os.Open(filename)
	if err != nil {
		logrus.Debugln(err)
		return ""
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		logrus.Debugln(err)
		return ""
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}
