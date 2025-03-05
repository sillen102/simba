package simbaAuth

import (
	"encoding/base64"
	"strings"
)

// BasicAuthDecode decodes the basic auth string and returns the username and password.
func BasicAuthDecode(auth string) (string, string, bool) {
	if auth == "" {
		return "", "", false
	}

	// Check if the auth starts with "Basic "
	const prefix = "Basic "
	if len(auth) < len(prefix) || !strings.HasPrefix(auth, prefix) {
		return "", "", false
	}

	// Decode the Base64-encoded credentials
	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return "", "", false
	}

	// Split the credentials by colon
	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return "", "", false
	}

	return cs[:s], cs[s+1:], true
}
