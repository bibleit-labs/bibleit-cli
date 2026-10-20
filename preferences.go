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
}

func preferencesPath() string {
	if p := os.Getenv("BIBLEIT_PREFERENCES"); p != "" {
		return p
	}
	if p := os.Getenv("BIBLEIT_CONFIG"); p != "" {
		if oneOf(strings.ToLower(filepath.Ext(p)), ".yaml", ".yml") {
			return p
		}
		return filepath.Join(filepath.Dir(p), "config.yaml")
	}
	return filepath.Join(homeDirectory(), ".bibleit", "config.yaml")
}
func loadPreferences() (preferences, error) {
	p := preferences{Version: 1}
	f, e := os.Open(preferencesPath())
	if os.IsNotExist(e) {
		return p, nil
	}
	if e != nil {
		return p, errors.New("cannot read preferences file; check `bibleit config path`")
	}
	defer f.Close()
	data, readErr := io.ReadAll(io.LimitReader(f, 16385))
	if readErr != nil || len(data) > 16384 {
		return p, errors.New("preferences file exceeds 16 KiB or could not be read")
	}
	d := yaml.NewDecoder(strings.NewReader(string(data)))
	d.KnownFields(true)
	if e = d.Decode(&p); e != nil {
		return p, preferenceYAMLError(e)
	}
	var nodes map[string]yaml.Node
	if e := yaml.Unmarshal(data, &nodes); e != nil {
		return p, preferenceYAMLError(e)
	}
	if _, ok := nodes["version"]; !ok {
		return p, errors.New("preferences must be a YAML mapping with version: 1")
	}
	for key, node := range nodes {
		if key == "version" {
			if node.Tag != "!!int" {
				return p, errors.New("preferences version must be an integer")
			}
		} else if node.Tag != "!!str" {
			return p, fmt.Errorf("preferences %s must be a string", key)
		}
	}
	_, p.IdentitySet = nodes["identity"]
	if d.Decode(new(any)) != io.EOF {
		return p, errors.New("preferences must contain one YAML document")
	}
	if p.Version != 1 {
		return p, errors.New("unsupported preferences version; use version: 1")
	}
	return p, validatePreferences(p)
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
	return fmt.Errorf("invalid preferences YAML%s; allowed keys: version, transport, identity, format", line)
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
	return nil
}
func savePreferences(p preferences) error {
	if e := validatePreferences(p); e != nil {
		return e
	}
	p.Version = 1
	values := map[string]any{"version": 1}
	if p.Transport != "" {
		values["transport"] = p.Transport
	}
	if p.Format != "" {
		values["format"] = p.Format
	}
	if p.IdentitySet || p.Identity != "" {
		values["identity"] = p.Identity
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
func effectivePreferences(p preferences, legacy config) (map[string]string, map[string]string) {
	v := map[string]string{"transport": "http", "format": "table", "identity": legacy.Identity}
	src := map[string]string{"transport": "default", "format": "default", "identity": "default"}
	if legacy.Identity != "" {
		src["identity"] = "legacy credentials"
	}
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
	if len(args) == 1 && args[0] == "path" {
		return printLocal("config", map[string]string{"path": preferencesPath()}, format)
	}
	p, e := loadPreferences()
	if e != nil {
		return fail(exitFailure, "%v", e)
	}
	if len(args) == 1 && args[0] == "show" {
		// Only the nonsecret legacy identity is projected; tokens never enter output.
		legacy, _ := loadConfig()
		v, src := effectivePreferences(p, legacy)
		fields := map[string]string{}
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
	legacy, _ := loadConfig()
	v, _ := effectivePreferences(p, legacy)
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
	values, _ := effectivePreferences(p, cfg)
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

// Preserve the service-bound version-2 credential document in its existing
// location. Preferences never duplicate tokens or rebind them to the REST host.
func credentialPath() string {
	if p := os.Getenv("BIBLEIT_CREDENTIALS"); p != "" {
		return p
	}
	if p := os.Getenv("BIBLEIT_CONFIG"); p != "" {
		if oneOf(strings.ToLower(filepath.Ext(p)), ".yaml", ".yml") {
			return filepath.Join(filepath.Dir(p), "config.json")
		}
		return p
	}
	return filepath.Join(homeDirectory(), ".bibleit", "config.json")
}

func rememberTransport(transport string) error {
	p, e := loadPreferences()
	if e != nil {
		return e
	}
	p.Transport = transport
	return savePreferences(p)
}
