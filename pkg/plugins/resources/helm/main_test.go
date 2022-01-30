package helm

import "testing"

func TestCondition(t *testing.T) {
	type dataSet struct {
		chart    Chart
		expected bool
	}

	set := []dataSet{
		{
			chart: Chart{
				spec: Spec{
					URL:     "https://charts.jenkins.io",
					Name:    "jenkins",
					Version: "2.19.0",
				},
			},
			expected: true,
		},
		{
			chart: Chart{
				spec: Spec{
					URL:     "https://kubernetes-charts.storage.googleapis.com",
					Name:    "jenkins",
					Version: "999",
				},
			},
			expected: false,
		},
		{
			chart: Chart{
				spec: Spec{
					URL:     "https://example.com",
					Name:    "jenkins",
					Version: "999",
				},
			},
			expected: false,
		},
	}

	for _, d := range set {
		got, _ := d.chart.Condition("")

		if got != d.expected {
			t.Errorf("%s Version %v is published! expected %v, got %v", d.chart.spec.Name, d.chart.spec.Version, d.expected, got)
		}

	}
}

func TestSource(t *testing.T) {

	type dataSet struct {
		chart    Chart
		expected string
	}

	set := []dataSet{
		{
			chart: Chart{
				spec: Spec{
					URL:  "https://stenic.github.io/helm-charts",
					Name: "proxy",
				},
			},
			expected: "1.0.3",
		},
		{
			chart: Chart{
				spec: Spec{
					URL:  "https://charts.jetstack.io",
					Name: "tor-prox",
				},
			},
			expected: "",
		},
		{
			chart: Chart{
				spec: Spec{
					URL:     "https://example.com",
					Name:    "jenkins",
					Version: "999",
				},
			},
			expected: "",
		},
	}

	for _, d := range set {
		got, _ := d.chart.Source("")

		if got != d.expected {
			t.Errorf("%v is published! latest expected version %v, got %v", d.chart.spec.Name, d.expected, got)
		}

	}
}
