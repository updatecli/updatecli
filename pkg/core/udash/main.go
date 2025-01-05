package udash

/*
	Package auth implements updatecli authentication with its backend
*/

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	cv "github.com/nirasan/go-oauth-pkce-code-verifier"
	"github.com/sirupsen/logrus"
	"github.com/skratchdot/open-golang/open"
)

var (
	Audience string
)

// authorizeUser implements the PKCE OAuth2 flow.
func authorizeUser(frontURL, clientID, authDomain, audience, redirectURL, accessToken string) error {
	// initialize the code verifier
	var CodeVerifier, _ = cv.CreateCodeVerifier()
	var err error

	// Create code_challenge with S256 method
	codeChallenge := CodeVerifier.CodeChallengeS256()

	if frontURL != "" {
		authDomain, audience, clientID, err = getOauthInfo(frontURL)
		// We don't want to exit on error if we fail retrieving oauth config from the endpoint
		if err != nil {
			logrus.Errorln(err)
		}
	}

	authDomain = setDefaultHTTPSScheme(authDomain)
	audience = setDefaultHTTPSScheme(audience)

	// If accessToken is provided, we skip the oauth flow
	if accessToken != "" {
		logrus.Infof("Token detected via flag --oauth-access-token, skipping the oauth PKCE flow")
		err = updateConfigFile(authData{
			Token: accessToken,
			API:   audience,
			URL:   frontURL,
		})
		if err != nil {
			logrus.Errorln(err)
			return nil
		}
		logrus.Println("Successfully logged into updatecli service.")

		return nil
	}

	authorizationURL, err := url.Parse(authDomain)
	if err != nil {
		logrus.Errorln(err)
		return err
	}

	authorizationURL = authorizationURL.JoinPath("authorize")

	query := authorizationURL.Query()
	query.Add("audience", audience)
	query.Add("scope", "openid")
	query.Add("response_type", "code")
	query.Add("client_id", clientID)
	query.Add("code_challenge", codeChallenge)
	query.Add("code_challenge_method", "S256")
	query.Add("redirect_uri", redirectURL)
	authorizationURL.RawQuery = query.Encode()

	// start a web server to listen on a callback URL
	server := &http.Server{
		Addr:              redirectURL,
		ReadHeaderTimeout: 60 * time.Second,
	}

	// define a handler that will get the authorization code, call the token endpoint, and close the HTTP server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		// get the authorization code
		code := r.URL.Query().Get("code")
		error := r.URL.Query().Get("error")
		errorDescription := r.URL.Query().Get("error_description")

		if error != "" {
			logrus.Errorf("Error:\n\t%s\n", error)
			logrus.Errorf("\t%s", errorDescription)

			// close the HTTP server and return
			cleanup(server)
			return
		}

		if code == "" {
			errmsg := "could not find 'code' URL parameter\n"
			_, err := io.WriteString(w, fmt.Sprintf("Error: %s", errmsg))
			if err != nil {
				logrus.Errorln(err)
				return
			}

			logrus.Debugln(errmsg)

			// close the HTTP server and return
			cleanup(server)
			return
		}

		// trade the authorization code and the code verifier for an access token
		codeVerifier := CodeVerifier.String()
		accessToken, err = getAccessToken(authDomain, clientID, codeVerifier, code, redirectURL)
		if err != nil {

			errmsg := "could not retrieve access token\n"
			_, err := io.WriteString(w, fmt.Sprintf("Error: %s", errmsg))
			if err != nil {
				logrus.Errorln(err)
				return
			}

			logrus.Debugln(errmsg)
			// close the HTTP server and return
			cleanup(server)
			return
		}

		err = updateConfigFile(authData{
			Token: accessToken,
			API:   audience,
			URL:   frontURL,
		})
		if err != nil {
			if !os.IsNotExist(err) {

				logrus.Errorln(err)
				// close the HTTP server and return
				cleanup(server)
				return
			}
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

		logrus.Println("Successfully logged into updatecli service.")

		// close the HTTP server
		cleanup(server)
	})

	// parse the redirect URL for the port number
	u, err := url.Parse(redirectURL)
	if err != nil {
		logrus.Errorf("updatecli: bad redirect URL: %s\n", err)
		os.Exit(1)
	}

	// set up a listener on the redirect port
	port := fmt.Sprintf(":%s", u.Port())
	l, err := net.Listen("tcp", port)
	if err != nil {
		logrus.Errorf("updatecli: can't listen to port %s: %s\n", port, err)
		os.Exit(1)
	}

	// open a browser window to the authorizationURL
	logrus.Debugf("Opening: %q", authorizationURL.String())

	err = open.Start(authorizationURL.String())
	if err != nil {
		logrus.Printf("updatecli: can't open browser to URL %s: %s\n",
			authorizationURL.String(),
			err,
		)
		os.Exit(1)
	}

	// start the blocking web server loop
	// this will exit when the handler gets fired and calls server.Close()
	err = server.Serve(l)
	if err != nil && !strings.Contains(err.Error(), "Server closed") {
		logrus.Errorln(err)
		return err
	}

	return nil
}

// cleanup closes the HTTP server
func cleanup(server *http.Server) {
	// we run this as a goroutine so that this function falls through and
	// the socket to the browser gets flushed/closed before the server goes away
	go server.Close()
}

// getAvailablePort returns an available port on localhost
func getAvailablePort() (string, error) {
	logrus.Debugln("searching available port on localhost")
	port := 8080
	maxPort := 8090

	foundOpenPort := false

	for port < maxPort {
		host := fmt.Sprintf("localhost:%d", port)

		logrus.Debugf("Trying %s\n", host)
		ln, err := net.Listen("tcp", host)
		if err != nil {
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
