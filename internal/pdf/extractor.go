package pdf

import (
	"bytes"
	"io"
	"os"
	"strings"

	"github.com/ledongthuc/pdf"
)

func ExtractContent(r io.ReaderAt) (string, error) {
	// Create a temporary file to handle the ReaderAt interface
	tmpFile, err := os.CreateTemp("", "pdf-extract-*.pdf")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Copy content from ReaderAt to the temp file
	if readSeeker, ok := r.(io.ReadSeeker); ok {
		// If it's also a ReadSeeker, use that for efficiency
		_, err = io.Copy(tmpFile, readSeeker)
	} else {
		// Otherwise read from the ReaderAt
		data := make([]byte, 1024)
		var offset int64
		for {
			n, err := r.ReadAt(data, offset)
			if n > 0 {
				if _, err := tmpFile.Write(data[:n]); err != nil {
					return "", err
				}
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				return "", err
			}
			offset += int64(n)
		}
	}

	// Rewind the file
	if _, err := tmpFile.Seek(0, 0); err != nil {
		return "", err
	}

	// Open the PDF file
	f, reader, err := pdf.Open(tmpFile.Name())
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Extract text
	var buf bytes.Buffer
	b, err := reader.GetPlainText()
	if err != nil {
		return "", err
	}
	buf.ReadFrom(b)

	return normalizeText(buf.String()), nil
}

func normalizeText(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}
