package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type config struct{ Transport, Endpoint, Identity, AccessToken string }
type configDocument struct {
	Version     int    `json:"version"`
	Endpoint    string `json:"endpoint"`
	SSHServer   string `json:"ssh_server"`
	Identity    string `json:"identity,omitempty"`
	AccessToken string `json:"access_token,omitempty"`
}

func loadConfig() (config, error) {
	data, err := os.ReadFile(configPath())
	if os.IsNotExist(err) {
		return config{Transport: "http", Endpoint: defaultWebURL}, nil
	}
	if err != nil {
		return config{}, err
	}
	var d configDocument
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err = dec.Decode(&d); err != nil {
		return config{}, errors.New("unsupported configuration; back up the old file and sign in using a fresh config")
	}
	if dec.Decode(new(any)) != io.EOF || d.Version != 2 {
		return config{}, errors.New("unsupported configuration version; back up the old file and sign in using a fresh config")
	}
	if d.Endpoint != defaultWebURL || d.SSHServer != defaultSSHServer {
		return config{}, errors.New("configuration does not match the project-managed service; back up the old file and sign in using a fresh config")
	}
	if strings.ContainsAny(d.AccessToken, "\r\n\x00") || strings.ContainsAny(d.Identity, "\r\n\x00") {
		return config{}, errors.New("invalid configuration credential")
	}
	return config{Transport: "http", Endpoint: d.Endpoint, Identity: d.Identity, AccessToken: d.AccessToken}, nil
}
func saveConfig(c config) error {
	if c.Transport != "" && c.Transport != "http" && c.Transport != "ssh" {
		return errors.New("transport must be http or ssh")
	}
	expected := defaultWebURL
	if c.Transport == "ssh" {
		expected = defaultSSHServer
	}
	if c.Endpoint != "" && c.Endpoint != expected {
		return errors.New("cannot rebind credentials to another service")
	}
	if strings.ContainsAny(c.AccessToken, "\r\n\x00") || strings.ContainsAny(c.Identity, "\r\n\x00") {
		return errors.New("invalid configuration credential")
	}
	if _, err := loadConfig(); err != nil {
		return err
	}
	d := configDocument{Version: 2, Endpoint: defaultWebURL, SSHServer: defaultSSHServer, Identity: c.Identity, AccessToken: c.AccessToken}
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	path := configPath()
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".config-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if err = restrictConfigFile(file); err == nil {
		_, err = file.Write(append(data, '\n'))
	}
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}
