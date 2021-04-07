module github.com/olblak/updateCli

go 1.16

require (
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/fatih/color v1.7.0
	github.com/go-git/go-git/v5 v5.3.0
	github.com/mitchellh/hashstructure v1.1.0
	github.com/mitchellh/mapstructure v1.4.1
	github.com/moby/buildkit v0.7.2
	github.com/pkg/errors v0.9.1
	github.com/shurcooL/githubv4 v0.0.0-20200802174311-f27d2ca7f6d5
	github.com/shurcooL/graphql v0.0.0-20181231061246-d48a9a75455f //indirect
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.0
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
	helm.sh/helm/v3 v3.2.4
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.4.0-beta.2.0.20200730150746-fa1220fce33f
	github.com/docker/docker => github.com/docker/docker v17.12.0-ce-rc1.0.20200310163718-4634ce647cf2+incompatible
)
