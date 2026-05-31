package client

import "testing"

const testDID = "did:plc:abc"

func TestSpecValidate(t *testing.T) {
	cases := []struct {
		name    string
		spec    Spec
		wantErr bool
	}{
		{"missing identity", Spec{AppPassword: "x"}, true},
		{"missing password", Spec{DID: testDID}, true},
		{"ok did", Spec{DID: testDID, AppPassword: "x"}, false},
		{"ok handle", Spec{Handle: "alice.tangled.sh", AppPassword: "x"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.spec.Validate()
			if (err != nil) != tc.wantErr {
				t.Fatalf("Validate() err=%v wantErr=%v", err, tc.wantErr)
			}
		})
	}
}

func TestSpecSanitizeDefaults(t *testing.T) {
	s := Spec{DID: testDID, AppPassword: "x"}
	s.Sanitize()
	if s.Appview != "https://tangled.sh" {
		t.Fatalf("default appview = %q", s.Appview)
	}
	if s.PDS != "https://bsky.social" {
		t.Fatalf("default pds = %q", s.PDS)
	}
}

func TestSpecSanitizeKeepsScheme(t *testing.T) {
	s := Spec{DID: testDID, AppPassword: "x", Appview: "tangled.local", PDS: "http://pds.local:3000/"}
	s.Sanitize()
	if s.Appview != "https://tangled.local" {
		t.Fatalf("appview = %q", s.Appview)
	}
	if s.PDS != "http://pds.local:3000" {
		t.Fatalf("pds = %q", s.PDS)
	}
}

func TestSpecSanitizeWithoutCredentials(t *testing.T) {
	// SCM-only flows (clone/push) must not require atproto credentials at
	// Sanitize time. Validate() is the gate for auth.
	s := Spec{}
	s.Sanitize()
	if s.PDS != "https://bsky.social" {
		t.Fatalf("default pds = %q", s.PDS)
	}
	if s.Appview != "https://tangled.sh" {
		t.Fatalf("default appview = %q", s.Appview)
	}
}
