package chart

import "testing"

func TestCondition(t *testing.T) {
	c := &Chart{
		URL:     "https://kubernetes-charts.storage.googleapis.com",
		Name:    "jenkins",
		Version: "1.21.1",
	}

	got, _ := c.Condition()

	expected := true

	if got != expected {
		t.Errorf("%s Version %v is published! expected %v, got %v", c.Name, c.Version, expected, got)
	}

	c = &Chart{
		URL:     "https://kubernetes-charts.storage.googleapis.com",
		Name:    "jenkins",
		Version: "999",
	}

	got, _ = c.Condition()

	expected = false

	if got != expected {
		t.Errorf("%s Version %v is not published! expected %v, got %v", c.Name, c.Version, expected, got)
	}

	c = &Chart{
		URL:  "https://example.com",
		Name: "tor-prox",
	}

	got, _ = c.Condition()

	expected = false

	if got != expected {
		t.Errorf("repository %v doesn't exist! expected version %v, got %v", c.URL, expected, got)
	}
}

func TestSource(t *testing.T) {
	c := &Chart{
		URL:  "https://charts.jetstack.io",
		Name: "tor-proxy",
	}

	got, _ := c.Source()

	expected := "0.1.1"

	if got != expected {
		t.Errorf("%v is published! latest expected version %v, got %v", c.Name, expected, got)
	}

	c = &Chart{
		URL:  "https://charts.jetstack.io",
		Name: "tor-prox",
	}

	got, _ = c.Source()

	expected = ""

	if got != expected {
		t.Errorf("%v doesn't exist! latest expected version %v, got %v", c.Name, expected, got)
	}
	c = &Chart{
		URL:  "https://example.com",
		Name: "tor-prox",
	}

	got, _ = c.Source()

	expected = ""

	if got != expected {
		t.Errorf("repository %v doesn't exist! expected version %v, got %v", c.URL, expected, got)
	}
}
