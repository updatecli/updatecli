package sign

import (
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
)

type GPGSpec struct {
	SigningKey string
	Passphrase string
}

func GetCommitSignKey(armoredKeyRing string, keyPassphrase string) (*openpgp.Entity, error) {
	s := strings.NewReader(armoredKeyRing)
	es, err := openpgp.ReadArmoredKeyRing(s)

	if err != nil {
		return nil, err
	}

	key := es[0]
	err = key.PrivateKey.Decrypt([]byte(keyPassphrase))

	if err != nil {
		return nil, err
	}

	return key, nil
}
