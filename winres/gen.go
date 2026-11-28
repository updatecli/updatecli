//go:build ignore

// Generate COFF resource objects from the application manifest.
package main

import (
	"debug/pe"
	"encoding/binary"
	"fmt"
	"log"
	"os"
)

func object(manifest []byte, machine, relocation uint16) []byte {
	// Each resource directory contains a 16-byte header and one 8-byte entry.
	const leaf = 3 * 24
	const payload = leaf + 16
	resource := make([]byte, (payload+len(manifest)+3)&^3)
	put16 := binary.LittleEndian.PutUint16
	put32 := binary.LittleEndian.PutUint32
	for i, id := range []uint32{24, 1, 1033} {
		dir := resource[i*24:]
		put16(dir[14:], 1)
		put32(dir[16:], id)
		target := uint32((i + 1) * 24)
		if i < 2 {
			target |= 1 << 31
		}
		put32(dir[20:], target)
	}
	put32(resource[leaf:], payload)
	put32(resource[leaf+4:], uint32(len(manifest)))
	copy(resource[payload:], manifest)

	// File and section headers precede the data, one relocation and one symbol.
	const raw = 20 + 40
	reloc := raw + len(resource)
	sym := reloc + 10
	data := make([]byte, sym+18+4)
	put16(data, machine)
	put16(data[2:], 1)
	put32(data[8:], uint32(sym))
	put32(data[12:], 1)
	section := data[20:]
	copy(section, ".rsrc")
	put32(section[16:], uint32(len(resource)))
	put32(section[20:], raw)
	put32(section[24:], uint32(reloc))
	put16(section[32:], 1)
	put32(section[36:], 0x40300040) // readable, aligned initialized data
	put32(data[reloc:], leaf)
	put16(data[reloc+8:], relocation)
	copy(data[sym:], ".rsrc")
	put16(data[sym+12:], 1)
	data[sym+16] = 3        // static section symbol
	put32(data[sym+18:], 4) // empty string table
	copy(data[raw:], resource)
	return data
}

func main() {
	manifest, err := os.ReadFile("winres/manifest.xml")
	if err != nil {
		log.Fatal(err)
	}
	for _, arch := range []struct {
		name                string
		machine, relocation uint16
	}{
		{"amd64", pe.IMAGE_FILE_MACHINE_AMD64, 3}, // IMAGE_REL_AMD64_ADDR32NB
		{"arm64", pe.IMAGE_FILE_MACHINE_ARM64, 2}, // IMAGE_REL_ARM64_ADDR32NB
	} {
		if err := os.WriteFile(fmt.Sprintf("rsrc_windows_%s.syso", arch.name), object(manifest, arch.machine, arch.relocation), 0644); err != nil {
			log.Fatal(err)
		}
	}
}
