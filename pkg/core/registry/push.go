package registry

import (
	"context"
	"fmt"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

// Push pushes updatecli manifest(s) as an OCI image to an OCI registry.
func Push(manifests []string, values []string, secrets []string, ociName string, disableTLS bool) error {
	// Create a file store
	fs, err := file.New(fileStore)
	if err != nil {
		return fmt.Errorf("create file store: %w", err)
	}

	defer fs.Close()
	ctx := context.Background()

	// Add files to the file store
	fileDescriptors := make([]v1.Descriptor, 0, len(manifests))
	//for _, name := range manifests {
	//	// TODO add a way to specify the artifact name
	//	fileDescriptor, err := fs.Add(ctx, name, updatecliManifestMediaType, "")
	//	if err != nil {
	//		return fmt.Errorf("add file %s: %w", name, err)
	//	}

	//	fileDescriptors = append(fileDescriptors, fileDescriptor)
	//	fmt.Printf("file descriptor for %s: %v\n", name, fileDescriptor)
	//}

	addfiles := func(files []string, mediaType string) {
		for i := range files {
			fileDescriptor, err := fs.Add(ctx, files[i], mediaType, "")
			if err != nil {
				fmt.Printf("add file %s: %v\n", files[i], err)
			} else {
				fileDescriptors = append(fileDescriptors, fileDescriptor)
				fmt.Printf("file descriptor for %s: %v\n", files[i], fileDescriptor)
			}
		}
	}

	addfiles(manifests, updatecliManifestMediaType)
	addfiles(values, updatecliValueMediaType)
	addfiles(secrets, updatecliSecretMediaType)

	// 2. Pack the files and tag the packed manifest
	opts := oras.PackManifestOptions{
		Layers: fileDescriptors,
	}
	manifestDescriptor, err := oras.PackManifest(ctx, fs, oras.PackManifestVersion1_1_RC4, ociArtifactType, opts)
	if err != nil {
		return fmt.Errorf("pack manifest: %w", err)
	}
	fmt.Println("manifest descriptor:", manifestDescriptor)

	tag := "latest"
	if err = fs.Tag(ctx, manifestDescriptor, tag); err != nil {
		return fmt.Errorf("tag manifest: %w", err)
	}

	// 3. Connect to a remote repository
	repo, err := remote.NewRepository(ociName)
	if err != nil {
		return fmt.Errorf("connect to remote repository: %w", err)
	}

	if disableTLS {
		repo.PlainHTTP = true
	}

	// Note: The below code can be omitted if authentication is not required
	repo.Client = &auth.Client{
		Client: retry.DefaultClient,
		Cache:  auth.DefaultCache,
		Credential: auth.StaticCredential(repo.Reference.Host(), auth.Credential{
			Username: "username",
			Password: "password",
		}),
	}

	// 3. Copy from the file store to the remote repository
	_, err = oras.Copy(ctx, fs, tag, repo, tag, oras.DefaultCopyOptions)
	if err != nil {
		return fmt.Errorf("upload artifact to %s: %w", repo.Reference.Reference, err)
	}

	return nil
}
