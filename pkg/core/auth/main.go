package auth

/*
	Package auth implements updatecli authentication with its backend
*/

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	cv "github.com/nirasan/go-oauth-pkce-code-verifier"
	"github.com/sirupsen/logrus"
	"github.com/skratchdot/open-golang/open"
	"github.com/spf13/viper"
)

// authorizeUser implements the PKCE OAuth2 flow.
func authorizeUser(clientID, authDomain, audience, redirectURL string) {
	// initialize the code verifier
	var CodeVerifier, _ = cv.CreateCodeVerifier()

	// Create code_challenge with S256 method
	codeChallenge := CodeVerifier.CodeChallengeS256()

	if authDomain == "" {
		authDomain = OauthIssuer
	}
	if audience == "" {
		audience = OauthAudience
	}
	if clientID == "" {
		clientID = OauthClientID
	}

	// construct the authorization URL (with Auth0 as the authorization provider)
	authorizationURL := fmt.Sprintf(
		"https://%s/authorize?audience=%s"+
			"&scope=openid"+
			"&response_type=code&client_id=%s"+
			"&code_challenge=%s"+
			"&code_challenge_method=S256&redirect_uri=%s",
		authDomain, audience, clientID, codeChallenge, redirectURL)

	// start a web server to listen on a callback URL
	server := &http.Server{
		Addr:              redirectURL,
		ReadHeaderTimeout: 60 * time.Second,
	}

	// define a handler that will get the authorization code, call the token endpoint, and close the HTTP server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// get the authorization code
		code := r.URL.Query().Get("code")
		if code == "" {
			fmt.Println("updatecli: Url Param 'code' is missing")
			_, err := io.WriteString(w, "Error: could not find 'code' URL parameter\n")
			if err != nil {
				logrus.Errorln(err)
				return
			}

			// close the HTTP server and return
			cleanup(server)
			return
		}

		// trade the authorization code and the code verifier for an access token
		codeVerifier := CodeVerifier.String()
		token, err := getAccessToken(authDomain, clientID, codeVerifier, code, redirectURL)
		if err != nil {
			fmt.Println("updatecli: could not get access token")
			_, err := io.WriteString(w, "Error: could not retrieve access token\n")
			if err != nil {
				logrus.Errorln(err)
				return
			}

			// close the HTTP server and return
			cleanup(server)
			return
		}

		updatecliConfigPath, err := initConfigFile()
		if err != nil {
			logrus.Errorln(err)
			return
		}

		encodedAudience := make([]byte, base64.StdEncoding.EncodedLen(len(audience)))
		base64.StdEncoding.Encode(encodedAudience, []byte(audience))

		fmt.Printf("%q - %q", audience, string(encodedAudience[:]))

		viper.Set(fmt.Sprintf("auths.%s.auth", string(encodedAudience[:])), token)
		viper.SetConfigFile(updatecliConfigPath)

		err = viper.WriteConfig()
		if err != nil {
			fmt.Println("updatecli: could not write config file")
			_, err := io.WriteString(w, "Error: could not store access token\n")
			if err != nil {
				logrus.Errorln(err)
				return
			}

			// close the HTTP server and return
			cleanup(server)
			return
		}

		// return an indication of success to the caller
		_, err = io.WriteString(w, `
		<html>
			<body>
				<h1>Login successful!</h1>
				<h2>You can close this window and return to the updatecli.</h2>
			</body>
		</html>`)

		if err != nil {
			logrus.Errorln(err)
			return
		}

		fmt.Println("Successfully logged into updatecli API.")

		// close the HTTP server
		cleanup(server)
	})

	// parse the redirect URL for the port number
	u, err := url.Parse(redirectURL)
	if err != nil {
		fmt.Printf("updatecli: bad redirect URL: %s\n", err)
		os.Exit(1)
	}

	// set up a listener on the redirect port
	port := fmt.Sprintf(":%s", u.Port())
	l, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Printf("updatecli: can't listen to port %s: %s\n", port, err)
		os.Exit(1)
	}

	// open a browser window to the authorizationURL
	err = open.Start(authorizationURL)
	if err != nil {
		fmt.Printf("updatecli: can't open browser to URL %s: %s\n", authorizationURL, err)
		os.Exit(1)
	}

	// start the blocking web server loop
	// this will exit when the handler gets fired and calls server.Close()
	err = server.Serve(l)
	if err != nil {
		logrus.Errorln(err)
		return
	}
}

// getAccessToken trades the authorization code retrieved from the first OAuth2 log for an access token
func getAccessToken(issuer, clientID, codeVerifier, authorizationCode, callbackURL string) (string, error) {
	// set the url and form-encoded data for the POST to the access token endpoint
	url := fmt.Sprintf("https://%s/oauth/token", issuer)

	data := fmt.Sprintf(
		"grant_type=authorization_code&client_id=%s"+
			"&code_verifier=%s"+
			"&code=%s"+
			"&redirect_uri=%s",
		clientID, codeVerifier, authorizationCode, callbackURL)
	payload := strings.NewReader(data)

	// create the request and execute it
	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Add("content-type", "application/x-www-form-urlencoded")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("updatecli: HTTP error: %s", err)
		return "", err
	}

	// process the response
	defer res.Body.Close()
	var responseData map[string]interface{}
	body, _ := io.ReadAll(res.Body)

	// unmarshal the json into a string map
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		fmt.Printf("updatecli: JSON error: %s", err)
		return "", err
	}

	// retrieve the access token out of the map, and return to caller
	accessToken := responseData["access_token"].(string)
	return accessToken, nil
}

// cleanup closes the HTTP server
func cleanup(server *http.Server) {
	// we run this as a goroutine so that this function falls through and
	// the socket to the browser gets flushed/closed before the server goes away
	go server.Close()
}

func getAvailablePort() (string, error) {
	logrus.Debugln("searching for available port on localhost")
	port := 8080
	maxPort := 8090

	foundOpenPort := false

	for port < maxPort {

		host := fmt.Sprintf("localhost:%d", port)

		logrus.Debugf("Trying %s\n", host)
		ln, err := net.Listen("tcp", host)
		if err != nil {
			//fmt.Fprintf(os.Stderr, "\t * Can't listen on port %d: %s\n", port, err)
			logrus.Debugf("can't listen on port %d: %s\n", port, err)
			// move to next port
			port = port + 1
			continue
		}

		_ = ln.Close()
		foundOpenPort = true
		break
	}

	if !foundOpenPort {
		return "", fmt.Errorf("available port not found")
	}
	return strconv.Itoa(port), nil
}

// initConfigFile creates Updatecli config directory
func initConfigFile() (string, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		logrus.Errorln(err)
		return "", err
	}

	updatecliConfigDir := filepath.Join(userConfigDir, "updatecli")

	if _, err := os.Stat(updatecliConfigDir); os.IsNotExist(err) {
		err := os.MkdirAll(updatecliConfigDir, 0755)
		if err != nil {
			return "", err
		}
	}

	return filepath.Join(updatecliConfigDir, "config.json"), nil
}
