package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type preferences struct {
	IdentitySet bool   `yaml:"-"`
	Version     int    `yaml:"version"`
	Transport   string `yaml:"transport,omitempty"`
	Identity    string `yaml:"identity,omitempty"`
	Format      string `yaml:"format,omitempty"`
	AccessToken string `yaml:"access_token,omitempty"`
	Endpoint    string `yaml:"endpoint,omitempty"`
	SSHServer   string `yaml:"ssh_server,omitempty"`
}

// BIBLEIT_CONFIG selects the single YAML document for every CLI command.
func preferencesPath() string {
	if p := os.Getenv("BIBLEIT_CONFIG"); p != "" {
		return p
	}
	return filepath.Join(homeDirectory(), ".bibleit", "config.yaml")
}
func loadPreferences() (preferences, error) { p, _, err := readPreferencesFile(); return p, err }
func readPreferencesFile() (preferences, bool, error) {
	p := preferences{Version: 2}
	if e := validateCredentialFile(preferencesPath()); e != nil {
		return p, false, e
	}
	f, e := os.Open(preferencesPath())
	if os.IsNotExist(e) {
		return p, false, nil
	}
	if e != nil {
		return p, false, errors.New("cannot read configuration; check BIBLEIT_CONFIG or ~/.bibleit/config.yaml")
	}
	defer f.Close()
	data, e := io.ReadAll(io.LimitReader(f, 16385))
	if e != nil || len(data) > 16384 {
		return p, true, errors.New("configuration exceeds 16 KiB or could not be read")
	}
	d := yaml.NewDecoder(strings.NewReader(string(data)))
	d.KnownFields(true)
	if e = d.Decode(&p); e != nil {
		return p, true, preferenceYAMLError(e)
	}
	var nodes map[string]yaml.Node
	if e = yaml.Unmarshal(data, &nodes); e != nil {
		return p, true, preferenceYAMLError(e)
	}
	if _, ok := nodes["version"]; !ok {
		return p, true, errors.New("configuration must be a YAML mapping with version: 2")
	}
	for key, node := range nodes {
		if key == "version" {
			if node.Tag != "!!int" {
				return p, true, errors.New("configuration version must be an integer")
			}
		} else if node.Tag != "!!str" {
			return p, true, fmt.Errorf("configuration %s must be a string", key)
		}
	}
	_, p.IdentitySet = nodes["identity"]
	if d.Decode(new(any)) != io.EOF {
		return p, true, errors.New("configuration must contain one YAML document")
	}
	if p.Version != 1 && p.Version != 2 {
		return p, true, errors.New("unsupported configuration version; use version: 2")
	}
	if p.Version == 1 && (p.AccessToken != "" || p.Endpoint != "" || p.SSHServer != "") {
		return p, true, errors.New("legacy preferences cannot contain credentials; use version: 2")
	}
	return p, true, validatePreferences(p)
}
func preferenceYAMLError(err error) error {
	// YAML diagnostics can quote arbitrary user values, including accidentally
	// pasted credentials. Keep only line numbers and the key schema.
	line := ""
	words := strings.Fields(err.Error())
	for i, w := range words {
		if w == "line" && i+1 < len(words) {
			line = " near line " + strings.TrimRight(words[i+1], ":")
			break
		}
	}
	return fmt.Errorf("invalid preferences YAML%s; allowed keys: version, transport, identity, format, access_token, endpoint, ssh_server", line)
}
func validatePreferences(p preferences) error {
	if p.Transport != "" && !oneOf(p.Transport, "http", "ssh") {
		return errors.New("preferences transport must be http or ssh")
	}
	if p.Format != "" && !oneOf(p.Format, "table", "raw", "json") {
		return errors.New("preferences format must be table, raw, or json; set ndjson explicitly for a watch")
	}
	if strings.ContainsAny(p.Identity, "\r\n\x00") {
		return errors.New("preferences identity must be a single-line path")
	}
	if strings.ContainsAny(p.AccessToken, "\r\n\x00") {
		return errors.New("invalid configuration credential")
	}
	if p.Endpoint != "" && p.Endpoint != defaultWebURL || p.SSHServer != "" && p.SSHServer != defaultSSHServer {
		return errors.New("configuration does not match the project-managed service")
	}
	if p.AccessToken != "" && (p.Endpoint != defaultWebURL || p.SSHServer != defaultSSHServer) {
		return errors.New("saved credentials require the project-managed service binding")
	}
	return nil
}
func savePreferences(p preferences) error {
	if p.AccessToken != "" {
		if p.Endpoint == "" {
			p.Endpoint = defaultWebURL
		}
		if p.SSHServer == "" {
			p.SSHServer = defaultSSHServer
		}
	}
	if e := validatePreferences(p); e != nil {
		return e
	}
	p.Version = 2
	values := map[string]any{"version": 2}
	if p.Transport != "" {
		values["transport"] = p.Transport
	}
	if p.Format != "" {
		values["format"] = p.Format
	}
	if p.IdentitySet || p.Identity != "" {
		values["identity"] = p.Identity
	}
	if p.AccessToken != "" {
		values["access_token"] = p.AccessToken
		values["endpoint"] = p.Endpoint
		values["ssh_server"] = p.SSHServer
	}
	b, e := yaml.Marshal(values)
	if e != nil {
		return errors.New("could not encode preferences")
	}
	return writePrivateDocument(preferencesPath(), b)
}
func writePrivateDocument(path string, data []byte) error {
	if info, e := os.Lstat(path); e == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			return errors.New("refusing to replace a symlink or nonregular configuration file")
		}
	} else if !os.IsNotExist(e) {
		return e
	}
	if e := os.MkdirAll(filepath.Dir(path), 0700); e != nil {
		return errors.New("cannot create configuration directory")
	}
	f, e := os.CreateTemp(filepath.Dir(path), ".bibleit-*")
	if e != nil {
		return errors.New("cannot create private configuration file")
	}
	defer os.Remove(f.Name())
	if e = restrictConfigFile(f); e == nil {
		_, e = f.Write(data)
	}
	if e == nil {
		e = f.Sync()
	}
	if closeErr := f.Close(); e == nil {
		e = closeErr
	}
	if e != nil {
		return errors.New("cannot save private configuration file")
	}
	if e = os.Rename(f.Name(), path); e != nil {
		return errors.New("cannot replace configuration file")
	}
	return nil
}
func effectivePreferences(p preferences) (map[string]string, map[string]string) {
	v := map[string]string{"transport": "http", "format": "table", "identity": ""}
	src := map[string]string{"transport": "default", "format": "default", "identity": "default"}
	for k, x := range map[string]string{"transport": p.Transport, "format": p.Format, "identity": p.Identity} {
		if x != "" || k == "identity" && p.IdentitySet {
			v[k] = x
			src[k] = "config.yaml"
		}
	}
	for k, env := range map[string]string{"transport": "BIBLEIT_TRANSPORT", "format": "BIBLEIT_FORMAT", "identity": "BIBLEIT_IDENTITY"} {
		if x := os.Getenv(env); x != "" {
			v[k] = x
			src[k] = env
		}
	}
	return v, src
}
func configCommand(args []string, format string) int {
	p, e := loadPreferences()
	if e != nil {
		return fail(exitFailure, "%v", e)
	}
	if len(args) == 1 && args[0] == "show" {
		// Show preferences and file location without projecting stored secrets.
		v, src := effectivePreferences(p)
		fields := map[string]string{"path": preferencesPath()}
		for k, x := range v {
			fields[k] = x
			fields[k+"_source"] = src[k]
		}
		return printLocal("config", fields, format)
	}
	if len(args) < 2 || !oneOf(args[1], "transport", "identity", "format") {
		return fail(exitUsage, "config key must be transport, identity, or format; credentials cannot be read with config commands")
	}
	key := args[1]
	v, _ := effectivePreferences(p)
	if args[0] == "get" {
		return printLocal("config", map[string]string{key: v[key]}, format)
	}
	value := ""
	if args[0] == "set" {
		value = args[2]
	}
	switch key {
	case "transport":
		p.Transport = value
	case "format":
		p.Format = value
	case "identity":
		p.Identity = value
		p.IdentitySet = true
	}
	if e = savePreferences(p); e != nil {
		return fail(exitFailure, "save preferences: %v", e)
	}
	return printLocal("config", map[string]string{"key": key, "event": "saved"}, format)
}
func applyPreferences(cfg config, options map[string]string) (config, string, error) {
	p, e := loadPreferences()
	if e != nil {
		return cfg, "", e
	}
	values, _ := effectivePreferences(p)
	for k := range values {
		if x := options["--"+k]; x != "" {
			values[k] = x
		}
	}
	if e = validatePreferences(preferences{Transport: values["transport"], Identity: values["identity"]}); e != nil {
		return cfg, "", e
	}
	if !oneOf(values["format"], "table", "raw", "json", "ndjson") {
		return cfg, "", errors.New("format must be table, raw, json, or ndjson")
	}
	cfg.Transport = values["transport"]
	cfg.Identity = values["identity"]
	return cfg, values["format"], nil
}

// Kept internally for existing callers; the sole persisted document is YAML.
func credentialPath() string { return preferencesPath() }
func rememberTransport(transport string) error {
	p, e := loadPreferences()
	if e != nil {
		return e
	}
	p.Transport = transport
	return savePreferences(p)
}
