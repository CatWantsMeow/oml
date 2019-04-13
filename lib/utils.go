package lib

import (
	"bytes"
	"io"
	"os"
	"regexp"
	"unicode"
)

const (
	chunkSize     = 512
	maxBufferSize = 1024 * 1024
)

func Get(opts map[string]string, key string, def string) string {
	val, ok := opts[key]
	if !ok {
		return def
	}
	return val
}

func In(elem string, arr []string) bool {
	for _, e := range arr {
		if e == elem {
			return true
		}
	}
	return false
}

func SplitToWords(s string) []string {
	return regexp.MustCompile(`[^\p{L}]`).Split(s, -1)
}

func IsSpace(c byte) bool {
	switch c {
	case '\t', '\v', '\f', '\r', ' ':
		return true
	}
	return false
}

func ContainsSpacesOnly(s []byte) bool {
	idx := bytes.IndexFunc(s, func(r rune) bool {
		return !unicode.IsSpace(r)
	})
	return idx == -1
}

func CompressSpaces(s []byte) []byte {
	buf := new(bytes.Buffer)

	var prev byte
	for _, c := range s {
		if IsSpace(c) && IsSpace(prev) {
			continue
		}
		buf.WriteByte(c)
		prev = c
	}
	return buf.Bytes()
}

func ReadFile(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	chunk := make([]byte, chunkSize)
	for buf.Len() < maxBufferSize {
		n, err := file.Read(chunk)
		if err != nil {
			if err == io.EOF {
				buf.Write(chunk[:n])
				break
			}
			return nil, err
		}
		buf.Write(chunk[:n])
	}
	return buf.Bytes(), nil
}
