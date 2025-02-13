package sign

import (
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
)

// GPGSpec defines the specification for manipulating gpg keys in the context of git commits.
type GPGSpec struct {
	/*
		signingKey defines the gpg key used to sign the commit message

		default:
			none
	*/
	SigningKey string `yaml:",omitempty"`
	/*
		passphrase defines the gpg passphrase used to sign the commit message
	*/
	Passphrase string `yaml:",omitempty"`
}

// GetCommitSignKey returns the gpg key used to sign the commit message
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

// IsZero returns true if the GPGSpec is empty
func (g GPGSpec) IsZero() bool {
	return g.SigningKey == "" && g.Passphrase == ""
}
