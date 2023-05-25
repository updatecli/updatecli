package auth

import "fmt"

func Login(clientID, authDomain, audience string) error {
	port, err := getAvailablePort()

	if err != nil {
		return fmt.Errorf("get available port: %w", err)
	}

	authorizeUser(
		clientID,
		authDomain,
		audience,
		fmt.Sprintf("http://localhost:%s", port),
	)

	return nil
}
