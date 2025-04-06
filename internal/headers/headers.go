package headers

import (
	"fmt"
	"strings"
)

const allowedFieldNameChars = "abcdefghijklmnopqrstuvwxyz0123456789!#$%&'*+-.^_`|~"

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Get(key string) (string, bool) {
	key = strings.ToLower(key)
	value, ok := h[key]
	return value, ok
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	bytesConsumed := 0
	fieldLine := string(data)

	// Check if the field line starts with "\r\n" - meaning the end of headers (field-lines)
	if strings.HasPrefix(fieldLine, "\r\n") {
		return 2, true, nil
	}

	CRLFCount := strings.Count(fieldLine, "\r\n")
	//check if request has at least "\r\n" in it and trim accordingly
	if CRLFCount == 2 {
		parts := strings.Split(fieldLine, "\r\n\r\n")
		fieldLine = parts[0]
	} else if CRLFCount == 1 {
		parts := strings.Split(fieldLine, "\r\n")
		fieldLine = parts[0]
	} else {
		return bytesConsumed, false, nil
	}

	// Split the header line into key and value
	parts := strings.SplitN(fieldLine, ":", 2)
	if len(parts) != 2 {
		return bytesConsumed, false, fmt.Errorf("invalid field line, illegal colon count: %q", fieldLine)
	}

	// Check if parts[0] has space before ":"
	if strings.HasSuffix(parts[0], " ") {
		return bytesConsumed, false, fmt.Errorf("invalid field line, space before colon: %q", fieldLine)
	}

	// Trim whitespace from key and value
	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])
	if key == "" || value == "" {
		return bytesConsumed, false, fmt.Errorf("invalid field line, empty field-name or field-value: %q", fieldLine)
	}

	// Before storing convert key to lower-case
	key = strings.ToLower(key)

	// Check if key contains only allowed characters using a lookup string
	for _, char := range key {
		if !strings.ContainsRune(allowedFieldNameChars, char) {
			return bytesConsumed, false, fmt.Errorf("invalid field line, field-name contains illegal character: %q", fieldLine)
		}
	}

	// Store the header in the map
	// Check if key is already present in the map and append the new value to the existing value, separated by a comma.
	if existingValue, ok := h[key]; ok {
		// Check if the existing value already equals the new value
		if existingValue == value {
			return bytesConsumed, false, fmt.Errorf("invalid field line, field-value already present: %q", fieldLine)
		}
		h[key] = existingValue + ", " + value
	} else {
		// If the key is not present, simply store the new value.
		h[key] = value
	}

	// Calculate the number of bytes consumed
	bytesConsumed = len(key) + len(value) + 4 // 4 for ": " and "\r\n"

	//print h for debugging
	//fmt.Println("Headers: ", h)

	return bytesConsumed, false, nil
}
