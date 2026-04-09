package registry

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/go-containerregistry/pkg/name"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sirupsen/logrus"
	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/errdef"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

// PushData is the data structure for the Push function.
type PushData struct {
	PolicyMetadataFile   string
	ManifestsFiles       []string
	ValuesFiles          []string
	SecretsFiles         []string
	AssetsFiles          []string
	PolicyReferenceNames []string
	DisableTLS           bool
	FileStore            string
	Overwrite            bool
}

// Push pushes updatecli manifest(s) as an OCI image to an OCI registry.
// func Push(policyMetadataFile string, manifests []string, values []string, secrets []string, assets []string, policyReferenceNames []string, disableTLS bool, fileStore string, overwrite bool) error {
func Push(p PushData) error {
	var err error

	policySpec, err := LoadPolicyFile(p.PolicyMetadataFile)
	if err != nil {
		return fmt.Errorf("load policy file: %w", err)
	}

	logrus.Infof("Pushing Updatecli policy:\n\t=> %s\n\n", strings.Join(p.PolicyReferenceNames, "\n\t=> "))

	if p.FileStore == "" {
		p.FileStore, err = os.Getwd()
		if err != nil {
			logrus.Errorln(err)
		}
	}

	// Create a file store
	fs, err := file.New(p.FileStore)
	if err != nil {
		return fmt.Errorf("create file store: %w", err)
	}

	defer fs.Close()
	ctx := context.Background()

	// Add files to the file store
	fileDescriptors := make([]v1.Descriptor, 0, len(p.ManifestsFiles))

	addfiles := func(files []string, mediaType string) error {
		for i := range files {
			fileDescriptor, err := fs.Add(ctx, files[i], mediaType, "")
			if err != nil {
				return fmt.Errorf("add file %s: %v", files[i], err)
			}

			fileDescriptors = append(fileDescriptors, fileDescriptor)
		}
		return nil
	}

	if err = addfiles(p.ManifestsFiles, updatecliManifestMediaType); err != nil {
		return fmt.Errorf("add manifests: %w", err)
	}

	if err = addfiles(p.ValuesFiles, updatecliValueMediaType); err != nil {
		return fmt.Errorf("add values: %w", err)
	}

	if err = addfiles(p.SecretsFiles, updatecliSecretMediaType); err != nil {
		return fmt.Errorf("add secrets: %w", err)
	}

	if err = addfiles(p.AssetsFiles, updatecliAssetMediaType); err != nil {
		return fmt.Errorf("add assets: %w", err)
	}

	// 2. Pack the files and tag the packed manifest
	opts := oras.PackManifestOptions{
		Layers: fileDescriptors,
	}

	// Set default OCI annotations which are understood by OCI registries
	// https://github.com/opencontainers/image-spec/blob/main/annotations.md
	opts.ManifestAnnotations = map[string]string{
		"org.opencontainers.image.created":       time.Now().Format(time.RFC3339),
		"org.opencontainers.image.authors":       strings.Join(policySpec.Authors, ", "),
		"org.opencontainers.image.url":           policySpec.URL,
		"org.opencontainers.image.documentation": policySpec.Documentation,
		"org.opencontainers.image.source":        policySpec.Source,
		"org.opencontainers.image.version":       policySpec.Version,
		"org.opencontainers.image.vendor":        policySpec.Vendor,
		"org.opencontainers.image.licenses":      strings.Join(policySpec.Licenses, ", "),
		// I don't understand why if set, it creates a file locally named with the value
		// To investigate...
		//"org.opencontainers.image.title":       policySpec.Title,
		"org.opencontainers.image.description": policySpec.Description,
		"org.opencontainers.image.changelog":   policySpec.Changelog,
	}

	manifestDescriptor, err := oras.PackManifest(ctx, fs, oras.PackManifestVersion1_1, ociArtifactType, opts)
	if err != nil {
		return fmt.Errorf("pack manifest: %w", err)
	}

	for i := range p.PolicyReferenceNames {

		refName, err := name.ParseReference(p.PolicyReferenceNames[i])
		if err != nil {
			logrus.Errorf("parse reference: %s", err)
			continue
		}

		tag := policySpec.Version

		if err = fs.Tag(ctx, manifestDescriptor, tag); err != nil {
			return fmt.Errorf("tag manifest: %w", err)
		}

		// 3. Connect to a remote repository
		repo, err := remote.NewRepository(refName.Name())
		if err != nil {
			return fmt.Errorf("connect to remote repository: %w", err)
		}
		ctx = auth.AppendRepositoryScope(ctx, repo.Reference, auth.ActionPush, auth.ActionPull)

		if p.DisableTLS {
			repo.PlainHTTP = true
		}

		// 2. Get credentials from the docker credential store
		if err := getCredentialsFromDockerStore(repo); err != nil {
			return fmt.Errorf("credstore from docker: %w", err)
		}

		_, _, err = repo.FetchReference(ctx, tag)
		if err == nil && !p.Overwrite {
			logrus.Infof("tag %q already published for policy %s", tag, p.PolicyReferenceNames[i])
			return nil
		} else if err == nil {
			logrus.Infof("overwriting, already published, tag %q for policy %s", tag, p.PolicyReferenceNames[i])

		} else if errors.Is(err, errdef.ErrNotFound) {
			logrus.Infof("publishing policy %s", tag)
		} else {
			return fmt.Errorf("check if %s is already published: %s", p.PolicyReferenceNames[i], err)
		}

		// 3. Copy from the file store to the remote repository
		_, err = oras.Copy(ctx, fs, tag, repo, tag, oras.DefaultCopyOptions)
		if err != nil {
			return fmt.Errorf("upload artifact to %s: %w", repo.Reference.Reference, err)
		}
	}

	return nil
}
