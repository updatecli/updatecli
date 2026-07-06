module example.com/updatecli-replace-test

go 1.25.0

require (
	github.com/rancher/saml v0.4.14
	github.com/crewjam/saml v0.6.0
	github.com/stretchr/testify v1.11.1
)

replace (
	github.com/rancher/saml => github.com/rancher/saml v0.4.14

	github.com/crewjam/saml v0.6.0 => github.com/crewjam/saml v0.5.0

	github.com/stretchr/testify => ../local/testify
)
