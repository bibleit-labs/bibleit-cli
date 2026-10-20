package main

import (
	"os"
	"strings"
)

// Secret material is never accepted as an option value. This optional destination
// avoids shell history, terminal scrollback and incidental pipeline capture.
var activeSecretFile *os.File
var activeSecretSaved bool

func prepareSecretDestination(path string) (*os.File, error) {
	f, e := os.OpenFile(expandHome(path), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return nil, userMessage{"cannot create secret destination; choose a new file in a private directory"}
	}
	if e = restrictConfigFile(f); e != nil {
		f.Close()
		os.Remove(f.Name())
		return nil, userMessage{"cannot protect secret destination"}
	}
	return f, nil
}
func deliverSecret(value string) (string, error) {
	if activeSecretFile == nil {
		return value, nil
	}
	if len(value) == 0 || len(value) > 4096 || strings.ContainsAny(value, "\r\n\x00") {
		return "", userMessage{"invalid one-time secret response; do not retry the mutation automatically"}
	}
	if activeSecretSaved {
		return "", userMessage{"unexpected repeated one-time secret; do not retry the mutation automatically"}
	}
	if _, e := activeSecretFile.WriteString(value + "\n"); e != nil {
		return "", userMessage{"secret could not be saved; mutation may have completed; rotate the secret before relying on it"}
	}
	if e := activeSecretFile.Sync(); e != nil {
		return "", userMessage{"secret could not be saved durably; rotate it before relying on it"}
	}
	activeSecretSaved = true
	return "written_to_private_file", nil
}
