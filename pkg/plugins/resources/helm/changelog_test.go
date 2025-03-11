package helm

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func TestChangelog(t *testing.T) {
	tests := []struct {
		name                string
		from                string
		to                  string
		spec                Spec
		requiresGithubToken bool
		expected            *result.Changelogs
	}{
		{
			name: "Valid chart with single changelog information in artifacthub.io/changes annotation",
			spec: Spec{
				URL:     "https://charts.jenkins.io",
				Name:    "jenkins",
				Version: "5.8.0",
			},
			from: "5.8.15",
			to:   "5.8.16",
			expected: &result.Changelogs{
				{
					Title:       "5.8.16",
					Body:        "\n* Update `docker.io/kiwigrid/k8s-sidecar` to version `1.30.1`\n",
					PublishedAt: "2025-02-21 19:29:15.858414421 +0000 UTC",
				},
			},
		},
		{
			name: "Another valid chart with single changelog information in artifacthub.io/changes annotation",
			spec: Spec{
				URL:     "https://kubernetes.github.io/ingress-nginx",
				Name:    "ingress-nginx",
				Version: "4.11.3",
			},
			from: "4.11.3",
			to:   "4.11.4",
			expected: &result.Changelogs{
				{
					Title:       "4.11.4",
					Body:        "\n* CI: Fix chart testing. (#12259)\n* Update Ingress-Nginx version controller-v1.11.4\n",
					PublishedAt: "2024-12-30 17:36:51.265913014 +0000 UTC",
				},
			},
		},
		{
			name: "Another valid chart with multiple changelog information in artifacthub.io/changes annotation",
			spec: Spec{
				URL:     "https://kubernetes.github.io/ingress-nginx",
				Name:    "ingress-nginx",
				Version: "4.11.3",
			},
			from: "4.11.3",
			to:   "4.12.0",
			expected: &result.Changelogs{
				{
					Title:       "4.12.0",
					Body:        "\n* CI: Fix chart testing. (#12258)\n* Update Ingress-Nginx version controller-v1.12.0\n",
					PublishedAt: "2024-12-30 17:42:14.794948649 +0000 UTC",
				},
				{
					Title:       "4.12.0-beta.0",
					Body:        "\n* Update Ingress-Nginx version controller-v1.12.0-beta.0\n",
					PublishedAt: "2024-10-15 09:49:01.496212197 +0000 UTC",
				},
				{
					Title:       "4.11.4",
					Body:        "\n* CI: Fix chart testing. (#12259)\n* Update Ingress-Nginx version controller-v1.11.4\n",
					PublishedAt: "2024-12-30 17:36:51.265913014 +0000 UTC",
				},
			},
		},
		{
			name: "Valid chart with rich changelog string in annotation",
			spec: Spec{
				URL:     "https://kyverno.github.io/kyverno/",
				Name:    "kyverno",
				Version: "3.3.4",
			},
			from: "3.3.4",
			to:   "3.3.5",
			expected: &result.Changelogs{
				{
					Title:       "3.3.5",
					PublishedAt: "2025-02-06 11:06:30.639777967 +0000 UTC",
					Body:        "\n## Added\n\n* added a new option .reportsController.sanityChecks to disable checks for policy reports crds\n\n## Fixed\n\n* fix validation error in validate.yaml\n* fixed global image registry config by introducing *.image.defaultRegistry.\n",
				},
			},
		},
		{
			name: "Valid chart with multiple changelog information in artifacthub.io/changes annotation",
			spec: Spec{
				URL:     "https://charts.jenkins.io",
				Name:    "jenkins",
				Version: "5.8.0",
			},
			from: "5.8.14",
			to:   "5.8.16",
			expected: &result.Changelogs{
				{
					Title:       "5.8.16",
					Body:        "\n* Update `docker.io/kiwigrid/k8s-sidecar` to version `1.30.1`\n",
					PublishedAt: "2025-02-21 19:29:15.858414421 +0000 UTC",
				},
				{
					Title:       "5.8.15",
					Body:        "\n* Update `kubernetes` to version `4313.va_9b_4fe2a_0e34`\n",
					PublishedAt: "2025-02-20 08:48:55.363415299 +0000 UTC",
				},
			},
		},
		{
			name: "Chart with changelog information in github releases and artifacthub.io/links annotation",
			spec: Spec{
				URL:     "https://prometheus-community.github.io/helm-charts",
				Name:    "kube-prometheus-stack",
				Version: "69.7.0",
			},
			from: "69.7.0",
			to:   "69.7.1",
			expected: &result.Changelogs{
				{
					Title:       "kube-prometheus-stack-69.7.1",
					Body:        "kube-prometheus-stack collects Kubernetes manifests, Grafana dashboards, and Prometheus rules combined with documentation and scripts to provide easy to operate end-to-end Kubernetes cluster monitoring with Prometheus using the Prometheus Operator.\n\n## What's Changed\n* [kube-prometheus-stack] Fix indentation for the nameValidationScheme field in Prometheus CR by @sviatlo in https://github.com/prometheus-community/helm-charts/pull/5400\n\n\n**Full Changelog**: https://github.com/prometheus-community/helm-charts/compare/prometheus-pingdom-exporter-3.0.2...kube-prometheus-stack-69.7.1",
					PublishedAt: "2025-03-03 10:56:56 +0000 UTC",
					URL:         "https://github.com/prometheus-community/helm-charts/releases/tag/kube-prometheus-stack-69.7.1",
				},
			},
			requiresGithubToken: true,
		},
		{
			name: "Chart with changelog information in github releases and artifacthub.io/links annotation. Multiple releases",
			spec: Spec{
				URL:     "https://prometheus-community.github.io/helm-charts",
				Name:    "prometheus-operator-crds",
				Version: "17.0.2",
			},
			from: "17.0.2",
			to:   "18.0.1",
			expected: &result.Changelogs{
				{
					Title:       "prometheus-operator-crds-18.0.1",
					Body:        "A Helm chart that collects custom resource definitions (CRDs) from the Prometheus Operator, allowing for seamless integration with GitOps tools \n\n## What's Changed\n* [prometheus-operator-crds] bump to v0.80.1 by @DrFaust92 in https://github.com/prometheus-community/helm-charts/pull/5355\n\n\n**Full Changelog**: https://github.com/prometheus-community/helm-charts/compare/kube-prometheus-stack-69.5.0...prometheus-operator-crds-18.0.1",
					URL:         "https://github.com/prometheus-community/helm-charts/releases/tag/prometheus-operator-crds-18.0.1",
					PublishedAt: "2025-02-25 08:36:34 +0000 UTC",
				},
				{
					Title:       "prometheus-operator-crds-18.0.0",
					Body:        "A Helm chart that collects custom resource definitions (CRDs) from the Prometheus Operator, allowing for seamless integration with GitOps tools \n\n## What's Changed\n* [prometheus-operator-crds] bump prometheus-operator to 'v0.80.0' by @sebastiangaiser in https://github.com/prometheus-community/helm-charts/pull/5289\n\n\n**Full Changelog**: https://github.com/prometheus-community/helm-charts/compare/kube-prometheus-stack-69.1.0...prometheus-operator-crds-18.0.0",
					PublishedAt: "2025-02-06 16:13:44 +0000 UTC",
					URL:         "https://github.com/prometheus-community/helm-charts/releases/tag/prometheus-operator-crds-18.0.0",
				},
			},
			requiresGithubToken: true,
		},
		{
			name: "Chart without changes annotation",
			spec: Spec{
				URL:     "https://kubernetes.github.io/autoscaler",
				Name:    "cluster-autoscaler",
				Version: "9.46.1",
			},
			from: "9.46.1",
			to:   "9.46.2",
			expected: &result.Changelogs{
				{
					Title:       "9.46.1",
					Body:        "\nRemark: We couldn't identify a way to automatically retrieve changelog information.\nPlease use following information to take informed decision\n\nHelm Chart: cluster-autoscaler\nScales Kubernetes worker nodes within autoscaling groups.\nProject Home: https://github.com/kubernetes/autoscaler\n\nVersion created on the 2025-02-25 01:26:59.960231386 &#43;0000 UTC\n\nSources:\n\n* https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler\n\n\n\nURL:\n\n* https://github.com/kubernetes/autoscaler/releases/download/cluster-autoscaler-chart-9.46.2/cluster-autoscaler-9.46.2.tgz\n\n\n",
					PublishedAt: "2025-02-25 01:26:59.960231386 +0000 UTC",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.requiresGithubToken {
				if os.Getenv("GITHUB_TOKEN") == "" && os.Getenv("UPDATECLI_GITHUB_TOKEN") == "" {
					t.Skip("Skipping test because GITHUB_TOKEN is not set")
				}
			}
			chart, err := New(tt.spec)
			assert.NoError(t, err)

			changelog := chart.Changelog(tt.from, tt.to)
			if tt.expected == nil && changelog == nil {
				return
			}

			assert.EqualExportedValues(t, tt.expected, changelog)

		})
	}
}
