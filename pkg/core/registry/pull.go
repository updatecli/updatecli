package registry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	spec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sirupsen/logrus"

	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
)

// PullResult lists the files of a policy pulled from a registry.
type PullResult struct {
	Manifests []string
	Values    []string
	Secrets   []string
	// Assets are the extra files shipped with the policy, stored next to its manifests.
	Assets []string
}

// Pull pulls an OCI image from a registry.
func Pull(ociName string, disableTLS bool) (result PullResult, err error) {

	ref, err := registry.ParseReference(ociName)
	if err != nil {
		return PullResult{}, fmt.Errorf("parse reference: %w", err)
	}

	if ref.Reference == ociLatestTag || ref.Reference == "" {
		ref.Reference, err = getLatestTagSortedBySemver(ref.Registry+"/"+ref.Repository, disableTLS)
		if err != nil {
			return PullResult{}, fmt.Errorf("get latest tag sorted by semver: %w", err)
		}
	}

	// 1. Connect to a remote repository
	ctx := context.Background()

	repo, err := remote.NewRepository(ociName)
	if err != nil {
		return PullResult{}, fmt.Errorf("new repository: %w", err)
	}

	ctx = auth.AppendRepositoryScope(ctx, repo.Reference, auth.ActionPull)

	if disableTLS {
		logrus.Debugln("TLS connection is disabled")
		repo.PlainHTTP = true
	}

	// 2. Get credentials from the docker credential store
	if err := getCredentialsFromDockerStore(repo); err != nil {
		return PullResult{}, fmt.Errorf("credstore from docker: %w", err)
	}

	// 2.5 Get remote manifest digest
	remoteManifestSpec, remoteManifestReader, err := oras.Fetch(ctx, repo, ref.String(), oras.DefaultFetchOptions)
	if err != nil {
		return PullResult{}, fmt.Errorf("fetch: %w", err)
	}

	remoteManifestData, err := content.ReadAll(remoteManifestReader, remoteManifestSpec)
	if err != nil {
		return PullResult{}, fmt.Errorf("fetch remote content: %w", err)
	}

	// Create the policy root directory
	policyRootDir := filepath.Join(getReferencePath(remoteManifestSpec.Digest.String())...)

	fs, err := file.New(policyRootDir)
	if err != nil {
		return PullResult{}, fmt.Errorf("create file store: %w", err)
	}
	defer fs.Close()

	// 3. Copy from the remote repository to the file store

	// Fetch the remote manifest

	result, err = getUpdatecliFilesFromManifestLayers(remoteManifestData, policyRootDir)
	if err != nil {
		return PullResult{}, fmt.Errorf("get media types from remote layers: %w", err)
	}

	if isPolicyFilesExistLocally(policyRootDir, result) {
		logrus.Debugf("Policy %q already available in:\n\t* %s\n", ociName, policyRootDir)
	} else {
		logrus.Infof("Pulling Updatecli policy %q\n", ociName)

		manifestDescriptor, err := oras.Copy(ctx, repo, ref.Reference, fs, ref.Reference, oras.DefaultCopyOptions)
		if err != nil {
			return PullResult{}, fmt.Errorf("copy: %w", err)
		}

		manifestData, err := content.FetchAll(ctx, fs, manifestDescriptor)
		if err != nil {
			return PullResult{}, fmt.Errorf("fetch manifest: %w", err)
		}

		result, err = getUpdatecliFilesFromManifestLayers(manifestData, policyRootDir)
		if err != nil {
			return PullResult{}, fmt.Errorf("get media types from layers: %w", err)
		}
	}

	logrus.Debugf("Manifests:\n")
	for _, manifest := range result.Manifests {
		logrus.Debugf("\t*%q\n", manifest)
	}

	if len(result.Values) > 0 {
		logrus.Debugf("Values:\n")
		for _, value := range result.Values {
			logrus.Debugf("\t*%q\n", value)
		}
	}

	if len(result.Secrets) > 0 {
		logrus.Debugf("Secrets:\n")
		for _, secret := range result.Secrets {
			logrus.Debugf("\t*%q\n", secret)
		}
	}

	if len(result.Assets) > 0 {
		logrus.Debugf("Assets:\n")
		for _, asset := range result.Assets {
			logrus.Debugf("\t*%q\n", asset)
		}
	}

	logrus.Debugf("policy successfully pulled in %s", policyRootDir)

	return result, nil
}

// getReferencePath returns the path to the file store for a given reference.
func getReferencePath(ref string) []string {
	refPath := []string{
		os.TempDir(),
		"updatecli",
		"store",
	}
	refPath = append(refPath, strings.TrimPrefix(ref, "sha256:"))

	return refPath
}

// getUpdatecliFilesFromManifestLayers returns the list of manifests, values, secrets and assets from an OCI manifest
func getUpdatecliFilesFromManifestLayers(manifestData []byte, policyRootDir string) (result PullResult, err error) {

	spec := spec.Manifest{}
	err = json.Unmarshal(manifestData, &spec)
	if err != nil {
		return PullResult{}, fmt.Errorf("unmarshal manifest: %w", err)
	}

	for _, layer := range spec.Layers {
		var files *[]string
		switch layer.MediaType {
		case updatecliManifestMediaType:
			files = &result.Manifests
		case updatecliValueMediaType:
			files = &result.Values
		case updatecliSecretMediaType:
			files = &result.Secrets
		case updatecliAssetMediaType:
			files = &result.Assets
		default:
			logrus.Warningf("unknown media type: %q\n", layer.MediaType)
			continue
		}

		if title, ok := layer.Annotations["org.opencontainers.image.title"]; ok && title != "" {
			*files = append(*files, filepath.Join(policyRootDir, title))
		}
	}

	return result, nil
}

/*
isPolicyFilesExistLocally returns true if the policy files are already available locally.
note it does not check if the files are up to date, only if they exist locally.
*/
func isPolicyFilesExistLocally(policyRootDir string, files PullResult) bool {

	errs := []error{}

	policyRootDirFileInfo, err := os.Stat(policyRootDir)
	// If store path exist and is a directory then we do nothing
	if err == nil {
		if !policyRootDirFileInfo.IsDir() {
			errs = append(errs, fmt.Errorf("policy root dir %s already exist and is not a directory", policyRootDir))
		}
	} else {
		if !errors.Is(err, os.ErrNotExist) {
			errs = append(errs, fmt.Errorf("getting information about %s: %w", policyRootDir, err))
		}
	}

	isFileExist := func(files []string) {
		for _, f := range files {
			_, err := os.Stat(f)
			if err == nil {
				continue
			}
			if errors.Is(err, os.ErrNotExist) {
				errs = append(errs, fmt.Errorf("%s does not exist locally", f))
			} else {
				errs = append(errs, fmt.Errorf("something went wrong while checking if %s exist: %w", f, err))
			}
		}
	}

	isFileExist(files.Manifests)
	isFileExist(files.Values)
	isFileExist(files.Secrets)
	isFileExist(files.Assets)

	if len(errs) > 0 {
		for i := range errs {
			logrus.Debugln(errs[i])
		}
		return false
	}

	return true

}
