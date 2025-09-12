package sign

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/go-git/go-git/v5/plumbing/object"
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

// SignCommit signs a git commit object using the provided GPG key
func SignCommit(commit *object.Commit, signer *openpgp.Entity) (*object.Commit, error) {
	// Create the commit data to be signed (same format git uses)
	commitData := createCommitDataForSigning(commit)

	// Create a buffer to hold the signature
	var sigBuf bytes.Buffer

	// Sign the commit data using the GPG entity
	err := openpgp.ArmoredDetachSign(&sigBuf, signer, strings.NewReader(commitData), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create GPG signature: %w", err)
	}

	// Create a new commit with the signature
	signedCommit := *commit // Copy the original commit
	signedCommit.PGPSignature = sigBuf.String()

	return &signedCommit, nil
}

// createCommitDataForSigning creates the commit data in the format that git uses for signing
func createCommitDataForSigning(commit *object.Commit) string {
	var buf strings.Builder

	// Tree line
	buf.WriteString(fmt.Sprintf("tree %s\n", commit.TreeHash.String()))

	// Parent lines
	for _, parent := range commit.ParentHashes {
		buf.WriteString(fmt.Sprintf("parent %s\n", parent.String()))
	}

	// Author line with timestamp
	buf.WriteString(fmt.Sprintf("author %s <%s> %d %s\n",
		commit.Author.Name,
		commit.Author.Email,
		commit.Author.When.Unix(),
		commit.Author.When.Format("-0700")))

	// Committer line with timestamp
	buf.WriteString(fmt.Sprintf("committer %s <%s> %d %s\n",
		commit.Committer.Name,
		commit.Committer.Email,
		commit.Committer.When.Unix(),
		commit.Committer.When.Format("-0700")))

	// Empty line before message
	buf.WriteString("\n")

	// Commit message
	buf.WriteString(commit.Message)

	return buf.String()
}

// ValidateGPGKey validates that the GPG key can be used for signing
func ValidateGPGKey(armoredKeyRing, passphrase string) error {
	entity, err := GetCommitSignKey(armoredKeyRing, passphrase)
	if err != nil {
		return fmt.Errorf("failed to load GPG key: %w", err)
	}

	// Check if the key can sign
	if entity.PrivateKey == nil {
		return fmt.Errorf("GPG key does not have a private key for signing")
	}

	if entity.PrivateKey.Encrypted {
		return fmt.Errorf("GPG key is still encrypted after decryption attempt")
	}

	return nil
}
