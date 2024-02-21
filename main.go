package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Not enough arguments")
	}
	filePath1, filePath2 := os.Args[1], os.Args[2]

	zr1, err := zip.OpenReader(filePath1)
	if err != nil {
		log.Fatal(err)
	}
	defer closeAndIgnoreError(zr1)

	zr2, err := zip.OpenReader(filePath2)
	if err != nil {
		log.Fatal(err)
	}
	defer closeAndIgnoreError(zr2)

	if err := diffZipHeader(zr1.File, zr2.File); err != nil {
		log.Fatal(err)
	}
}

func diffZipHeader(headers1, headers2 []*zip.File) error {
	for _, header1 := range headers1 {
		index2 := slices.IndexFunc(headers2, func(header2 *zip.File) bool {
			return header1.Name == header2.Name
		})
		if index2 < 0 {
			return fmt.Errorf("file %s not found in the second zip", header1.Name)
		}
		header2 := headers2[index2]
		if err := compareFiles(header1, header2); err != nil {
			return fmt.Errorf("file %s is different: %s", header1.Name, err)
		}
	}
	return nil
}

func closeAndIgnoreError(c io.Closer) {
	_ = c.Close()
}

func compareFiles(file1, file2 *zip.File) error {
	if file1.UncompressedSize64 != file2.UncompressedSize64 {
		return fmt.Errorf("file %s has different size", file1.Name)
	}
	if !file1.Modified.Equal(file2.Modified) {
		return fmt.Errorf("file %s has different modified time", file1.Name)
	}
	if file1.Method != file2.Method {
		return fmt.Errorf("file %s has different compression method", file1.Name)
	}
	if file1.Comment != file2.Comment {
		return fmt.Errorf("file %s has different comment", file1.Name)
	}
	if !bytes.Equal(file1.Extra, file2.Extra) {
		return fmt.Errorf("file %s has different extra data", file1.Name)
	}
	if file1.NonUTF8 != file2.NonUTF8 {
		return fmt.Errorf("file %s has different NonUTF8 flag", file1.Name)
	}
	if file1.CreatorVersion != file2.CreatorVersion {
		return fmt.Errorf("file %s has different creator version", file1.Name)
	}
	if file1.ReaderVersion != file2.ReaderVersion {
		return fmt.Errorf("file %s has different reader version", file1.Name)
	}
	if file1.Flags != file2.Flags {
		return fmt.Errorf("file %s has different flags", file1.Name)
	}
	if file1.CRC32 != file2.CRC32 {
		return fmt.Errorf("file %s has different CRC32", file1.Name)
	}
	if file1.CompressedSize64 != file2.CompressedSize64 {
		return fmt.Errorf("file %s has different compressed size", file1.Name)
	}
	if file1.ExternalAttrs != file2.ExternalAttrs {
		return fmt.Errorf("file %s has different external attributes", file1.Name)
	}

	checksum1, err := zipChecksum(file1)
	if err != nil {
		return fmt.Errorf("file %s cannot be opened from file 1: %s", file1.Name, err)
	}
	checksum2, err := zipChecksum(file2)
	if err != nil {
		return fmt.Errorf("file %s cannot be opened from file 1: %s", file1.Name, err)
	}

	if checksum1 != checksum2 {
		return fmt.Errorf("file %s has different content %q != %q", file1.Name, checksum1, checksum2)
	}

	return nil
}

func zipChecksum(file *zip.File) (string, error) {
	r, err := file.Open()
	if err != nil {
		return "", err
	}
	defer closeAndIgnoreError(r)
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
