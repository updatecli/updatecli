package chart

import "testing"

func TestCondition(t *testing.T) {
	type dataSet struct {
		chart    Chart
		expected bool
	}

	set := []dataSet{
		{
			chart: Chart{
				URL:     "https://charts.jenkins.io",
				Name:    "jenkins",
				Version: "2.19.0",
			},
			expected: true,
		},
		{
			chart: Chart{
				URL:     "https://kubernetes-charts.storage.googleapis.com",
				Name:    "jenkins",
				Version: "999",
			},
			expected: false,
		},
		{
			chart: Chart{
				URL:     "https://example.com",
				Name:    "jenkins",
				Version: "999",
			},
			expected: false,
		},
	}

	for _, d := range set {
		got, _ := d.chart.Condition("")

		if got != d.expected {
			t.Errorf("%s Version %v is published! expected %v, got %v", d.chart.Name, d.chart.Version, d.expected, got)
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
				URL:  "https://charts.jetstack.io",
				Name: "tor-proxy",
			},
			expected: "0.1.1",
		},
		{
			chart: Chart{
				URL:  "https://charts.jetstack.io",
				Name: "tor-prox",
			},
			expected: "",
		},
		{
			chart: Chart{
				URL:     "https://example.com",
				Name:    "jenkins",
				Version: "999",
			},
			expected: "",
		},
	}

	for _, d := range set {
		got, _ := d.chart.Source("")

		if got != d.expected {
			t.Errorf("%v is published! latest expected version %v, got %v", d.chart.Name, d.expected, got)
		}

	}
}
