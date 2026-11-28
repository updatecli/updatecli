package main

import (
	"bytes"
	"debug/pe"
	"encoding/binary"
	"encoding/xml"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"go.yaml.in/yaml/v3"
)

func TestReleasePackage(t *testing.T) {
	data, err := os.ReadFile(".goreleaser.yml")
	if err != nil {
		t.Fatal(err)
	}
	var config struct{ Builds []struct{ ID, Main string } }
	if err := yaml.Unmarshal(data, &config); err != nil {
		t.Fatal(err)
	}
	for _, build := range config.Builds {
		if build.ID == "updatecli" && build.Main == "." {
			return
		}
	}
	t.Fatal("missing updatecli package build")
}

func TestWindowsManifest(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "winres"), 0755); err != nil {
		t.Fatal(err)
	}
	manifest, err := os.ReadFile("winres/manifest.xml")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "winres", "manifest.xml"), manifest, 0644); err != nil {
		t.Fatal(err)
	}
	generator, err := filepath.Abs("winres/gen.go")
	if err != nil {
		t.Fatal(err)
	}
	generate := exec.Command("go", "run", generator)
	generate.Dir = dir
	if out, err := generate.CombinedOutput(); err != nil {
		t.Fatalf("%v: %s", err, out)
	}
	for _, arch := range []string{"amd64", "arm64"} {
		t.Run(arch, func(t *testing.T) {
			name := "rsrc_windows_" + arch + ".syso"
			generated, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				t.Fatal(err)
			}
			committed, err := os.ReadFile(name)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(committed, generated) {
				t.Fatal("run go generate to refresh", name)
			}
			if testing.Short() {
				return
			}
			exe := filepath.Join(t.TempDir(), "updatecli.exe")
			build := exec.Command("go", "build", "-o", exe, ".")
			build.Env = append(os.Environ(), "GOOS=windows", "GOARCH="+arch, "CGO_ENABLED=0")
			if out, err := build.CombinedOutput(); err != nil {
				t.Fatalf("%v: %s", err, out)
			}
			file, err := pe.Open(exe)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := file.Close(); err != nil {
					t.Errorf("closing %s: %v", exe, err)
				}
			}()
			section := file.Section(".rsrc")
			if section == nil {
				t.Fatal("missing .rsrc")
			}
			data, err := section.Data()
			if err != nil {
				t.Fatal(err)
			}
			offset := uint32(0)
			for i, id := range []uint32{24, 1, 1033} {
				entry := data[offset+16:]
				if binary.LittleEndian.Uint32(entry) != id {
					t.Fatalf("missing resource ID %d", id)
				}
				target := binary.LittleEndian.Uint32(entry[4:])
				if (target>>31 != 0) != (i < 2) {
					t.Fatal("invalid resource directory entry")
				}
				offset = target &^ (1 << 31)
			}
			start := binary.LittleEndian.Uint32(data[offset:]) - section.VirtualAddress
			size := binary.LittleEndian.Uint32(data[offset+4:])
			var assembly struct {
				Execution struct {
					Level    string `xml:"level,attr"`
					UIAccess string `xml:"uiAccess,attr"`
				} `xml:"trustInfo>security>requestedPrivileges>requestedExecutionLevel"`
			}
			if err := xml.Unmarshal(data[start:start+size], &assembly); err != nil {
				t.Fatal(err)
			}
			if assembly.Execution.Level != "asInvoker" || assembly.Execution.UIAccess != "false" {
				t.Fatalf("unexpected execution level %+v", assembly.Execution)
			}
		})
	}
}
