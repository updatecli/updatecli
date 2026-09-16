module example.com/updatecli-replace-inactive-test

go 1.25.0

require (
    github.com/crewjam/saml v0.6.0
    github.com/rancher/saml v0.3.0 // indirect
    gopkg.in/yaml.v3 v3.0.1
)

replace (
    github.com/crewjam/saml v0.6.0 => github.com/crewjam/saml v0.5.0

    github.com/rancher/saml => github.com/rancher/saml v0.2.0

    gopkg.in/yaml.v3 v3.0.0 => gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c

    github.com/stretchr/testify => github.com/stretchr/testify v1.8.4
)
