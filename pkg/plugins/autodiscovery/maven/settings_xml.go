package maven

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

var (

	// mavenMirrorURLFromEnv is the maven mirror URL from the environment variable MAVEN_MIRROR_URL
	mavenMirrorURLFromEnv = os.Getenv("MAVEN_MIRROR_URL")
	//mavenEnvVariableRegex is the regex to match maven environment variables
	// Example: ${env.MAVEN_MIRROR_URL}
	mavenEnvVariableRegex = regexp.MustCompile(`\${env.([a-zA-Z0-9_]+)}`)
)

type Settings struct {
	XMLName        xml.Name       `xml:"settings"`
	Mirrors        []Mirror       `xml:"mirrors>mirror"`
	Profiles       []Profile      `xml:"profiles>profile"`
	Servers        []Server       `xml:"servers>server"`
	ActiveProfiles ActiveProfiles `xml:"activeProfiles"`
}

type Server struct {
	ID       string `xml:"id"`
	Username string `xml:"username"`
	Password string `xml:"password"`
}

type Mirror struct {
	ID       string `xml:"id"`
	Name     string `xml:"name"`
	URL      string `xml:"url"`
	MirrorOf string `xml:"mirrorOf"`
}

type Profile struct {
	ID           string       `xml:"id"`
	Repositories []Repository `xml:"repositories>repository"`
}

type Repository struct {
	ID        string      `xml:"id"`
	URL       string      `xml:"url"`
	Releases  EnabledFlag `xml:"releases"`
	Snapshots EnabledFlag `xml:"snapshots"`
}

type EnabledFlag struct {
	Enabled string `xml:"enabled"`
}

type ActiveProfiles struct {
	ActiveProfile []string `xml:"activeProfile"`
}

// readSettingsXML loads all identified settings.xml file and return a map of settingsXML
func readSettingsXML(path string) *Settings {

	var settings Settings

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	b, err := os.ReadFile(path)
	if err != nil {
		logrus.Errorf("Reading file %q failed: %s", path, err)
	}

	err = xml.Unmarshal(b, &settings)
	if err != nil {
		logrus.Errorf("Failed to unmarshal %q: %s", path, err)
		return nil
	}

	return &settings
}

// getMatchingServerCredentials returns the username and password of server, given an id
func (s Settings) getMatchingServerCredentials(id string) (username string, password string) {
	for _, server := range s.Servers {
		if server.ID == id {

			username = interpolateMavenEnvVariable(server.Username)
			password = interpolateMavenEnvVariable(server.Password)

			return username, password
		}
	}
	return "", ""
}

func (s Settings) getCentralMatchingMirrorOf() (mirrorURL string, mirrorID string) {
	for _, mirror := range s.Mirrors {
		if mirror.MirrorOf == "central" {
			return mirror.URL, mirror.ID
		}
	}

	return "", ""
}

// getMatchingMirrorOf returns true if the given repository id matches the mirrorOf field of any mirror in the settings.xml
func (s Settings) getMatchingMirrorOf(id, url string) (mirrorURL string, mirrorID string, foundMirror bool) {

	for _, mirror := range s.Mirrors {
		matching := false
		excluded := false

		mirrorURL = mirror.URL
		mirrorID = mirror.ID

		for _, rule := range strings.Split(mirror.MirrorOf, ",") {

			switch rule {
			// matches all repo ids.
			case "*":
				return mirrorURL, mirrorID, true

			// matches all repositories using HTTP except those using localhost.
			case "external:http:*":
				if strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "http://localhost") {
					return mirrorURL, mirrorID, true
				}

			// matches all repositories except those using localhost or file based repositories.
			case "external:*":
				if !strings.Contains(url, "localhost") && !strings.HasPrefix(url, "file:") {
					return mirrorURL, mirrorID, true
				}

			// multiple repositories may be specified using a comma as the delimiter
			default:
				switch strings.HasPrefix(rule, "!") {
				case true:
					if strings.TrimPrefix(rule, "!") == id {
						excluded = true
					}
				case false:
					if rule == id {
						matching = true
					}
				}
			}
		}

		if matching && !excluded {
			return mirrorURL, mirrorID, true
		}

	}
	return "", "", false
}

// interpolateMavenEnvVariable updates values from the settings.xml using env variables
func interpolateMavenEnvVariable(input string) string {
	return mavenEnvVariableRegex.ReplaceAllStringFunc(input, func(s string) string {
		envVar := mavenEnvVariableRegex.FindStringSubmatch(s)[1]
		return os.Getenv(envVar)
	})
}

// getSettingsXMLPath returns the path to the settings.xml file
func getSettingsXMLPath(projectDirname string) []string {

	result := []string{}

	if _, err := os.Stat(filepath.Join(projectDirname, "settings.xml")); err == nil {
		result = append(result, filepath.Join(projectDirname, "settings.xml"))
	}

	// Order matter here
	// By using a join on the different path component, we ensure that the generated path is correct
	// regardless of the OS
	defaultSettingsXMLPath := []string{
		filepath.Join(os.Getenv("HOME"), ".m2", "settings.xml"),
		filepath.Join(os.Getenv("MAVEN_HOME"), "conf", "settings.xml"),
		filepath.Join(os.Getenv("M2_HOME"), "conf", "settings.xml"),
	}

	for _, path := range defaultSettingsXMLPath {
		if _, err := os.Stat(path); err == nil {
			result = append(result, path)
		}
	}

	return result
}

// getRepositoriesFromSettingsXML returns the repositories from the settings.xml
func (s Settings) getRepositoriesFromSettingsXML() []repository {

	repositories := []repository{}

	for _, activeactiveProfile := range s.ActiveProfiles.ActiveProfile {
		for _, profile := range s.Profiles {
			if profile.ID == activeactiveProfile {
				for _, repo := range profile.Repositories {
					repositories = append(repositories, repository{
						ID:  repo.ID,
						URL: repo.URL,
					})
				}
			}
		}
	}

	return repositories
}
