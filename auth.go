package main

import (
	"errors"
	"net/http"
	"strings"
)

func getAPIKey(headers http.Header) (string, error) {
	val := headers.Get("Authorization")

	if val == "" {
		return "", errors.New("no authentication info found")
	}

	vals := strings.Split(val, " ")

	if len(vals) != 2 {
		return "", errors.New("malformed authorization header")
	}

	if vals[0] != "ApiKey" {
		return "", errors.New("malformed authorization header")
	}

	return vals[1], nil
}
