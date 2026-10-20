package main

import (
	"errors"
	"strings"
)

type config struct{ Transport, Endpoint, Identity, AccessToken string }

func loadConfig() (config, error) {
	p, err := loadPreferences()
	if err != nil {
		return config{}, err
	}
	return config{Transport: "http", Endpoint: defaultWebURL, Identity: p.Identity, AccessToken: p.AccessToken}, nil
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
	p, err := loadPreferences()
	if err != nil {
		return err
	}
	p.AccessToken = c.AccessToken
	p.Identity = c.Identity
	p.IdentitySet = true
	if c.Transport != "" {
		p.Transport = c.Transport
	}
	return savePreferences(p)
}
