package docker

import (
	"fmt"
	"net/http"
	"strings"
)

// Docker contains various information to interact with a docker registry
type Docker struct {
	Image        string
	Tag          string
	URL          string
	Architecture string
}

// Check verify if Docker parameters are correctly set
func (d *Docker) Check() (bool, error) {
	if d.Image == "" {
		err := fmt.Errorf("Docker Image is required")
		return false, err
	}

	if d.URL == "" {
		d.URL = "hub.docker.com"
	}

	if d.Tag == "" {
		d.Tag = "latest"
	}

	if d.Architecture == "" {
		d.Architecture = "amd64"
	}

	image := strings.Split(d.Image, "/")

	if len(image) == 1 && d.isDockerHub() {
		d.Image = "library/" + d.Image
	}

	return true, nil
}

func (d *Docker) isDockerHub() bool {
	return strings.Contains(d.URL, "hub.docker.com")
}

// IsDockerRegistry validates that we are on docker registry api
// https://docs.docker.com/registry/spec/api/#api-version-check
func (d *Docker) IsDockerRegistry() (bool, error) {

	if ok, err := d.Check(); !ok {
		return false, err
	}

	if d.isDockerHub() {
		return false, fmt.Errorf("DockerHub Api is not docker registry api compliant")
	}

	URL := fmt.Sprintf("https://%s/v2/", d.URL)

	req, err := http.NewRequest("GET", URL, nil)

	if err != nil {
		return false, err
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return false, err
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return false, nil
	}
	return true, nil
}

//// ConditionFromSCM returns an error because it's not supported
//func (d *Docker) ConditionFromSCM(source string, scm scm.Scm) (bool, error) {
//	return false, fmt.Errorf("SCM configuration is not supported for dockerRegistry condition, aborting")
//}
//
//// Condition checks if a docker image with a specific tag is published
//func (d *Docker) Condition(source string) (bool, error) {
//	URL := ""
//
//	if d.Tag != "" {
//		fmt.Printf("Tag %v, already defined from configuration file\n", d.Tag)
//	} else {
//		d.Tag = source
//	}
//
//	if ok, err := d.Check(); !ok {
//		return false, err
//	}
//
//	if d.isDockerHub() {
//		URL = fmt.Sprintf("https://%s/v2/repositories/%s/tags/%s/",
//			d.URL,
//			d.Image,
//			d.Tag)
//
//	} else {
//		if ok, err := d.IsDockerRegistry(); !ok {
//			return false, err
//		}
//		URL = fmt.Sprintf("https://%s/v2/%s/manifests/%s",
//			d.URL,
//			d.Image,
//			d.Tag)
//	}
//
//	req, err := http.NewRequest("GET", URL, nil)
//
//	if err != nil {
//		return false, err
//	}
//
//	res, err := http.DefaultClient.Do(req)
//
//	if err != nil {
//		return false, err
//	}
//
//	defer res.Body.Close()
//
//	body, err := ioutil.ReadAll(res.Body)
//
//	if err != nil {
//		return false, err
//	}
//
//	if res.StatusCode == 200 && !d.isDockerHub() {
//		fmt.Printf("\u2714 %s:%s available on the Docker Registry\n", d.Image, d.Tag)
//		return true, nil
//
//	} else if d.isDockerHub() {
//
//		data := map[string]string{}
//
//		json.Unmarshal(body, &data)
//
//		if val, ok := data["message"]; ok && strings.Contains(val, "not found") {
//			fmt.Printf("\u2717 %s:%s doesn't exist on the Docker Registry \n", d.Image, d.Tag)
//			return false, nil
//		}
//
//		if val, ok := data["name"]; ok && val == d.Tag {
//			fmt.Printf("\u2714 %s:%s available on the Docker Registry\n", d.Image, d.Tag)
//			return true, nil
//		}
//
//	} else {
//
//		fmt.Printf("\u2717Something went wrong on URL: %s\n", URL)
//	}
//
//	return false, fmt.Errorf("something went wrong %s", URL)
//}

//// Source retrieve docker image tag digest from a registry
//func (d *Docker) Source() (string, error) {
//
//	if ok, err := d.Check(); !ok {
//		return "", err
//	}
//
//	// https://hub.docker.com/v2/repositories/olblak/updatecli/tags/latest
//	URL := ""
//
//	if d.isDockerHub() {
//		URL = fmt.Sprintf("https://%s/v2/repositories/%s/tags/%s/",
//			d.URL,
//			d.Image,
//			d.Tag)
//
//	} else {
//		if ok, err := d.IsDockerRegistry(); !ok {
//			return "", err
//		}
//		URL = fmt.Sprintf("https://%s/v2/%s/manifests/%s",
//			d.URL,
//			d.Image,
//			d.Tag)
//	}
//
//	req, err := http.NewRequest("GET", URL, nil)
//
//	if err != nil {
//		return "", err
//	}
//
//	if ok, err := d.IsDockerRegistry(); ok && err == nil {
//		// Retrieve v2 manifest
//		// application/vnd.docker.distribution.manifest.v1+prettyjws v1 manifest
//		req.Header.Add("Accept", "application/vnd.docker.distribution.manifest.v2+json")
//
//	}
//
//	res, err := http.DefaultClient.Do(req)
//
//	if err != nil {
//		return "", err
//	}
//
//	defer res.Body.Close()
//
//	body, err := ioutil.ReadAll(res.Body)
//
//	if err != nil {
//		return "", err
//	}
//
//	if d.isDockerHub() {
//
//		type respond struct {
//			ID     string
//			Images []map[string]string
//		}
//
//		data := respond{}
//
//		json.Unmarshal(body, &data)
//
//		for _, image := range data.Images {
//			if image["architecture"] == d.Architecture {
//				digest := strings.TrimPrefix(image["digest"], "sha256:")
//				fmt.Printf("\u2714 Digest '%v' found for docker image %s:%s available from Docker Registry\n", digest, d.Image, d.Tag)
//				fmt.Printf("\nRemark: Do not forget to add @sha256 after your the docker image name\n")
//				fmt.Printf("Example: %v@sha256:%v\n", d.Image, digest)
//				return digest, nil
//			}
//		}
//
//		fmt.Printf("\u2717 No Digest found for docker image %s:%s on the Docker Registry \n", d.Image, d.Tag)
//
//		return "", nil
//	}
//
//	digest := res.Header.Get("Docker-Content-Digest")
//	digest = strings.TrimPrefix(digest, "sha256:")
//
//	fmt.Printf("\u2714 Digest '%v' found for docker image %s:%s available from Docker Registry\n", digest, d.Image, d.Tag)
//	fmt.Printf("\nRemark: Do not forget to add @sha256 after your the docker image name\n")
//	fmt.Printf("Example: %v/%v@sha256:%v\n", d.URL, d.Image, digest)
//
//	return digest, nil
//
//}
