package github

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// Github describe github settings
type Github struct {
	Owner      string
	Repository string
	Token      string
	URL        string
	Version    string
}

// GetVersion retrieves the version tag from releases
func (github *Github) GetVersion() string {

	url := fmt.Sprintf("https://%s/repos/%s/%s/releases/%s",
		github.URL,
		github.Owner,
		github.Repository,
		github.Version)

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		log.Println(err)
	}

	req.Header.Add("Authorization", "token "+github.Token)

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Println(err)
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		log.Println(err)
	}

	v := map[string]string{}
	json.Unmarshal(body, &v)

	if val, ok := v["name"]; ok {
		return val
	}
	log.Printf("\u2717 No tag founded from %s\n", url)
	return ""

}
