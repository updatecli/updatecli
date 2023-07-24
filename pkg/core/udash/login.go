package udash

import "fmt"

func Login(serviceURL, clientID, issuer, audience, accessToken string) error {
	port, err := getAvailablePort()
	if err != nil {
		return fmt.Errorf("get available port: %w", err)
	}

	err = authorizeUser(
		serviceURL,
		clientID,
		issuer,
		audience,
		fmt.Sprintf("http://localhost:%s", port),
		accessToken,
	)

	if err != nil {
		return err
	}

	return nil
}
