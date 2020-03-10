package maven

import "testing"

func TestIsPublished(t *testing.T) {
	// Test if existing image tag return true
	m := &Maven{
		URL:        "repo.jenkins-ci.org",
		Repository: "releases",
		GroupID:    "org.eclipse.mylyn.wikitext",
		ArtifactID: "wikitext.core",
		Version:    "1.7.4.v20130429",
	}

	got := m.IsTagPublished()
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

	got = m.IsTagPublished()
	expected = false
	if got != expected {
		t.Errorf("ArtifactID %v is not published! expected %v, got %v", m.ArtifactID, expected, got)
	}
}

func TestGetVersion(t *testing.T) {
	// Test if existing image tag return true
	m := &Maven{
		URL:        "repo.jenkins-ci.org",
		Repository: "releases",
		GroupID:    "org.eclipse.mylyn.wikitext",
		ArtifactID: "wikitext.core",
	}

	got := m.GetVersion()
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

	got = m.GetVersion()
	expected = "2.21"
	if got == expected {
		t.Errorf("Latest version published expected is %v, got %v", expected, got)
	}
}
