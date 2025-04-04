package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestLineParse(t *testing.T) {

	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host      : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Invalid header line
	headers = NewHeaders()
	data = []byte("Host :localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Invalid header line with no value
	headers = NewHeaders()
	data = []byte("Host: \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Invalid header line with no key
	headers = NewHeaders()
	data = []byte(": localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Invalid header line with no key and value
	headers = NewHeaders()
	data = []byte(": \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Valid header done line
	headers = NewHeaders()
	data = []byte("\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, 2, n)
	assert.True(t, done)

	// Test: Valid header with multiple lines
	headers = NewHeaders()
	headers["host"] = "localhost:42069"
	headers["user-Agent"] = "curl/7.81.0"
	data = []byte("Accept: */*\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, "curl/7.81.0", headers["user-Agent"])
	assert.Equal(t, "*/*", headers["accept"])
	assert.Equal(t, 13, n)
	assert.False(t, done)

	// Test: Valid header with multiple lines and extra spaces
	headers = NewHeaders()
	headers["host"] = "localhost:42069"
	headers["accept"] = "*/*"
	data = []byte("       User-Agent:       curl/7.81.0\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, "curl/7.81.0", headers["user-agent"])
	assert.Equal(t, "*/*", headers["accept"])
	assert.Equal(t, 25, n)
	assert.False(t, done)

	// Test: Valid header with same keys diffrent case sensitivity
	headers = NewHeaders()
	headers["host"] = "localhost:42069"
	headers["user-agent"] = "curl/7.81.0"
	data = []byte("hoSt: localhost:69420\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "localhost:42069, localhost:69420", headers["host"])
	assert.Equal(t, "curl/7.81.0", headers["user-agent"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Invalid character in header key
	headers = NewHeaders()
	data = []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}
