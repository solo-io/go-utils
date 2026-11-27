package cliutils

import (
	"math/rand"
	"strings"
	"time"
)

const (
	lcAlpha        = "abcdefghijklmnopqrstuvwxyz"
	lcAlphaNumeric = "abcdefghijklmnopqrstuvwxyz0123456789"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RandStringBytes produces a random string of length n using the characters present in the basis string
func RandStringBytes(n int, basis string) string {
	if basis == "" {
		return ""
	}
	b := make([]byte, n)
	for i := range b {
		b[i] = basis[rand.Intn(len(basis))]
	}
	return string(b)
}

// RandDNS1035 generates a random string of length n that meets the DNS-1035 standard used by Kubernetes names
//
// Typical kubernetes error message for invalid names: a DNS-1035 label must consist of lower case alphanumeric characters or '-', start with an alphabetic character, and end with an alphanumeric character (e.g. 'my-name',  or 'abc-123', regex used for validation is '[a-z]([-a-z0-9]*[a-z0-9])?')
func RandKubeNameBytes(n int) string {
	if n < 1 {
		return ""
	}
	firstChar := RandStringBytes(1, lcAlpha)
	suffix := ""
	if n > 1 {
		suffix = RandStringBytes(n-1, lcAlphaNumeric)
	}
	return strings.Join([]string{firstChar, suffix}, "")
}

// Contains indicates if a string slice 'a' contains the string s
func Contains(a []string, s string) bool {
	for _, n := range a {
		if s == n {
			return true
		}
	}
	return false
}

// Contains indicates if a string slice 'a' contains a string that encompasses the string s
func ContainsSubstring(a []string, substring string) bool {
	for _, n := range a {
		if strings.Contains(n, substring) {
			return true
		}
	}
	return false
}
