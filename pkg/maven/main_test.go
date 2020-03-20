package maven

import "testing"

func TestCondition(t *testing.T) {
	// Test if existing image tag return true
	m := &Maven{
		URL:        "repo.jenkins-ci.org",
		Repository: "releases",
		GroupID:    "org.eclipse.mylyn.wikitext",
		ArtifactID: "wikitext.core",
		Version:    "1.7.4.v20130429",
	}

	got, _ := m.Condition()
	expected := true
	if got != expected {
		t.Errorf("ArtifactID %v is published! expected %v, got %v", m.ArtifactID, expected, got)
	}

	// Test if none existing image tag return false
	m = &Maven{
		URL:        "repo.jenkins-ci.org",
		Repository: "releases",
		GroupID:    "org.eclipse.mylyn.wikitext",
		ArtifactID: "wikitext.core",
		Version:    "0.3",
	}

	got, _ = m.Condition()
	expected = false
	if got != expected {
		t.Errorf("ArtifactID %v is not published! expected %v, got %v", m.ArtifactID, expected, got)
	}
}

func TestSource(t *testing.T) {
	// Test if existing image tag return true
	m := &Maven{
		URL:        "repo.jenkins-ci.org",
		Repository: "releases",
		GroupID:    "org.eclipse.mylyn.wikitext",
		ArtifactID: "wikitext.core",
	}

	got, _ := m.Source()
	expected := "1.7.4.v20130429"
	if got != expected {
		t.Errorf("Latest version published expected is %v, got %v", expected, got)
	}

	// Test if none existing image tag return false
	m = &Maven{
		URL:        "repo.jenkins-ci.org",
		Repository: "releases",
		GroupID:    "org.eclipse.mylyn.wikitext",
		ArtifactID: "wikitext.core",
		Version:    "0.3",
	}

	got, _ = m.Source()
	expected = "2.21"
	if got == expected {
		t.Errorf("Latest version published expected is %v, got %v", expected, got)
	}
}
