// bibleit provides terminal access to Bibleit through REST and SSH.
package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	bibleit "github.com/bibleit-labs/bibleit-go"
)

const (
	exitOK = iota
	exitFailure
	exitUsage
	exitUnauthenticated
	exitForbidden
	exitNotFound
	exitConflict
	exitRateLimited
)

type outputRecord struct {
	Type   string
	Fields map[string]string
	order  []string
}

const sshUser = "bibleit-cli"

const defaultWebURL = bibleit.ServerEndpoint

var (
	cliVersion       = "dev"
	defaultSSHServer = "bibleit.app:22"
	cliHTTPClient    = &http.Client{Timeout: 15 * time.Second}
)

func main() { os.Exit(run(os.Args[1:])) }
func run(args []string) int {
	// Completion arguments are shell words, not options for this invocation.
	if len(args) > 0 && args[0] == "__complete" {
		return completeCommand(args[1:])
	}
	previousFormat, previousCommand, previousTransport := activeFormat, activeCommand, activeTransport
	defer func() {
		activeFormat, activeCommand, activeTransport = previousFormat, previousCommand, previousTransport
	}()
	options, args, err := parseOptions(args, map[string]bool{"--format": true, "--identity": true, "--transport": true, "--cursor": true, "--limit": true, "--secret-output": true, "--yes": false})
	format := options["--format"]
	if format == "" {
		format = "table"
	}
	activeFormat = format
	if err != nil {
		return fail(exitUsage, "%v", err)
	}
	if !oneOf(format, "table", "raw", "json", "ndjson") {
		return fail(exitUsage, "unsupported format (use table, raw, json, or ndjson)")
	}
	if format == "ndjson" && !isLiveSubscribe(args) {
		return fail(exitUsage, "ndjson is for Live subscriptions; use json for finite commands")
	}
	if len(args) == 1 && oneOf(args[0], "version", "--version", "-v") {
		if format == "json" {
			return printLocal("version", map[string]string{"version": cliVersion}, format)
		}
		fmt.Printf("bibleit CLI version %s\n", cliVersion)
		return exitOK
	}
	if len(args) == 0 {
		return help(nil)
	}
	if oneOf(args[0], "help", "--help", "-h") {
		topics := args[1:]
		if len(topics) > 0 && oneOf(topics[len(topics)-1], "--help", "-h") {
			topics = topics[:len(topics)-1]
		}
		return help(topics)
	}
	if oneOf(args[len(args)-1], "--help", "-h") || (args[len(args)-1] == "help" && args[0] != "server") {
		return help(args[:len(args)-1])
	}
	if len(args) == 1 && oneOf(args[0], "auth", "live", "read", "search", "translation", "server", "account", "organization", "completion", "config") {
		return help(args)
	}
	if args[0] == "completion" {
		return completionCommand(args[1:], format)
	}
	if args[0] == "profile" {
		return fail(exitUsage, "named profiles are not supported")
	}
	if cursor := options["--cursor"]; cursor != "" {
		if len(args) == 2 && args[0] == "organization" && args[1] == "list" {
			args = append(args, cursor)
		} else if args[0] != "read" {
			return fail(exitUsage, "--cursor is only supported by organization list or a whole-book read")
		}
	}
	args, err = normalizeCommand(args)
	if err != nil {
		return fail(exitUsage, "%v", err)
	}
	activeCommand = args
	bookPage, err := bookPageOptions(args, options)
	if err != nil {
		return fail(exitUsage, "%v", err)
	}
	if path := options["--secret-output"]; path != "" {
		if len(args) != 4 || args[0] != "live" || args[2] != "secret" || !oneOf(args[3], "create", "rotate") {
			return fail(exitUsage, "--secret-output is only supported by secret create or secret rotate")
		}
		file, e := prepareSecretDestination(path)
		if e != nil {
			return fail(exitUsage, "%v", e)
		}
		oldFile, oldSaved := activeSecretFile, activeSecretSaved
		activeSecretFile = file
		activeSecretSaved = false
		defer func() {
			saved := activeSecretSaved
			file.Close()
			if !saved {
				os.Remove(file.Name())
			}
			activeSecretFile, activeSecretSaved = oldFile, oldSaved
		}()
	}
	if args[0] == "config" {
		return configCommand(args[1:], format)
	}
	cfg, err := loadConfig()
	if err != nil {
		return fail(exitFailure, "read configuration: %v", err)
	}
	cfg, format, err = applyPreferences(cfg, options)
	if err != nil {
		return fail(exitUsage, "%v", err)
	}
	activeFormat = format
	if format == "ndjson" && !isLiveSubscribe(args) {
		return fail(exitUsage, "ndjson is for Live subscriptions; use json for finite commands")
	}
	transport := cfg.Transport
	if !oneOf(transport, "http", "ssh") {
		return fail(exitUsage, "transport must be http or ssh")
	}
	if len(args) >= 3 && args[0] == "auth" && args[1] == "login" && args[2] == "ssh" {
		if options["--transport"] == "http" {
			return fail(exitUsage, "SSH login conflicts with --transport http")
		}
		transport = "ssh"
	}
	cfg.Transport = transport
	activeTransport = transport
	if bookPage != nil && transport != "http" {
		return fail(exitUsage, "whole-book pagination requires --transport http")
	}
	cfg.Endpoint = defaultWebURL
	if transport == "ssh" {
		cfg.Endpoint = defaultSSHServer
	}
	identity := options["--identity"]
	if identity != "" {
		if cfg.Transport != "ssh" {
			return fail(exitUsage, "--identity requires --transport ssh")
		}
		cfg.Identity = identity
	}
	if args[0] == "auth" {
		if len(args) > 1 && oneOf(args[1], "info", "whoami") {
			cfg, err = automationCredential(cfg)
			if err != nil {
				return fail(exitUsage, "%v", err)
			}
		}
		return authCommand(cfg, args[1:], format, identity != "")
	}
	cfg, err = automationCredential(cfg)
	if err != nil {
		return fail(exitUsage, "%v", err)
	}
	if cfg.Transport == "http" {
		return runHTTPPage(cfg, args, format, options["--yes"] == "true", bookPage)
	}
	command, err := commandFor(args)
	if err != nil {
		return fail(exitUsage, "%v", err)
	}
	if args[0] == "organization" || len(args) == 3 && args[0] == "server" && args[1] == "help" && args[2] == "organization" {
		lines, err := requestSSH(cfg, bibleit.ServerInfoCommand())
		if err != nil {
			return protocolFailure(err)
		}
		if err := requireOrganizationCapability(lines); err != nil {
			return fail(exitUsage, "%v", err)
		}
	}
	if action := destructiveAction(args); action != "" && options["--yes"] != "true" {
		if err := confirmAction(cfg, action, format, os.Stdin, os.Stderr, terminalInput(os.Stdin) && terminalInput(os.Stderr)); err != nil {
			var confirmation *confirmationError
			if errors.As(err, &confirmation) && confirmation.code == "confirmation_required" {
				return outputError(exitUsage, err)
			}
			return outputError(exitFailure, err)
		}
	}
	if isLiveSubscribe(args) {
		if cfg.Transport != "ssh" {
			return fail(exitUsage, "Live subscriptions require --transport ssh")
		}
		if format == "json" {
			return fail(exitUsage, "use --format ndjson for Live subscriptions")
		}
		if err := subscribeSSH(cfg, command, format); err != nil {
			return protocolFailure(err)
		}
		return exitOK
	}
	lines, err := requestSSH(cfg, command)
	if err != nil {
		return protocolFailure(err)
	}
	return printResponse(lines, format)
}
func automationCredential(cfg config) (config, error) {
	if token := os.Getenv("BIBLEIT_TOKEN"); token != "" {
		if cfg.Transport != "http" {
			return cfg, nil
		}
		if strings.ContainsAny(token, "\r\n\x00") {
			return cfg, errors.New("invalid BIBLEIT_TOKEN")
		}
		cfg.AccessToken = token
	}
	return cfg, nil
}

func isLiveSubscribe(args []string) bool {
	return len(args) == 3 && args[0] == "live" && args[2] == "subscribe"
}
func authCommand(cfg config, args []string, format string, explicitIdentity bool) int {
	if len(args) == 0 {
		return help([]string{"auth"})
	}
	switch args[0] {
	case "help", "--help", "-h":
		return help([]string{"auth"})
	case "login":
		if len(args) == 1 && cfg.Transport != "ssh" {
			actor, err := browserAuthenticate(&cfg)
			if err != nil {
				return protocolFailure(err)
			}
			if err := rememberTransport(cfg.Transport); err != nil {
				return fail(exitFailure, "signed in, but preferences were not saved: %v", err)
			}
			return printLocal("auth", map[string]string{"identity": actor, "auth": "true", "transport": "http"}, format)
		}
		if len(args) > 2 || (len(args) == 2 && args[1] != "ssh") {
			return usage()
		}
		// Saved preferences and legacy identity selection remain effective.
		// Use `config unset identity` to return to OpenSSH agent selection.
		lines, err := requestSSH(cfg, bibleit.IdentityCommand())
		if err != nil {
			return protocolFailure(err)
		}
		identityName := authenticatedIdentity(lines)
		if identityName == "" {
			return fail(exitFailure, "SSH authentication succeeded but the server did not identify the account")
		}
		if err := saveConfig(cfg); err != nil {
			return fail(exitFailure, "save configuration: %v", err)
		}
		if err := rememberTransport(cfg.Transport); err != nil {
			return fail(exitFailure, "signed in, but preferences were not saved: %v", err)
		}
		return printLocal("auth", map[string]string{"identity": identityName, "auth": "true", "transport": "ssh"}, format)
	case "ssh":
		return fail(exitUsage, "`bibleit auth ssh` was renamed; use `bibleit auth login ssh`")
	case "logout":
		if len(args) != 1 {
			return usage()
		}
		if err := logout(&cfg); err != nil {
			return protocolFailure(err)
		}
		return printLocal("auth", map[string]string{"event": "logged_out", "transport": cfg.Transport}, format)
	case "whoami", "info":
		if len(args) != 1 {
			return usage()
		}
		if cfg.Transport == "http" {
			return runHTTP(cfg, []string{"auth", args[0]}, format, false)
		}
		lines, err := requestSSH(cfg, bibleit.IdentityCommand())
		if err != nil {
			return protocolFailure(err)
		}
		return printResponse(lines, format)
	default:
		return usage()
	}
}
func commandFor(a []string) (bibleit.Command, error) {
	if len(a) == 0 {
		return bibleit.Command{}, errors.New("unsupported command")
	}
	switch a[0] {
	case "ping":
		if len(a) == 1 {
			return bibleit.PingCommand(), nil
		}
	case "server":
		if len(a) == 2 && a[1] == "info" {
			return bibleit.ServerInfoCommand(), nil
		}
		if len(a) == 2 && a[1] == "help" {
			return bibleit.HelpCommand("")
		}
		if len(a) == 3 && a[1] == "help" {
			return bibleit.HelpCommand(a[2])
		}
	case "account":
		if len(a) == 2 && a[1] == "tokens" {
			return bibleit.AccountTokensCommand(), nil
		}
		if len(a) == 2 && a[1] == "info" {
			return bibleit.AccountInfoCommand(), nil
		}
		if len(a) == 2 && a[1] == "quotas" {
			return bibleit.AccountQuotasCommand(), nil
		}
	case "organization":
		if len(a) == 2 && a[1] == "list" {
			return bibleit.OrganizationListCommand("")
		}
		if len(a) == 3 && a[1] == "list" {
			return bibleit.OrganizationListCommand(a[2])
		}
		if len(a) == 4 && a[1] == "application" && a[3] == "status" {
			return bibleit.OrganizationApplicationStatusCommand(a[2])
		}
	case "translation":
		if len(a) == 2 && a[1] == "list" {
			return bibleit.TranslationListCommand(), nil
		}
		if len(a) == 3 {
			switch a[1] {
			case "info":
				return bibleit.TranslationInfoCommand(a[2])
			case "catalog":
				return bibleit.TranslationCatalogCommand(a[2])
			case "add":
				return bibleit.AddTranslationCommand(a[2])
			case "remove":
				return bibleit.RemoveTranslationCommand(a[2])
			}
		}
	case "read":
		if len(a) >= 3 && len(a) <= 5 {
			ref, err := cliReference(a[1], a[2:])
			if err != nil {
				return bibleit.Command{}, err
			}
			return bibleit.ReadCommand(ref)
		}
	case "search":
		if len(a) == 3 {
			return bibleit.SearchCommand(a[1], a[2])
		}
	case "live":
		if len(a) == 2 && a[1] == "list" {
			return bibleit.LiveListCommand(), nil
		}
		if len(a) >= 2 && a[1] == "create" {
			return bibleit.CreateLiveCommand(strings.Join(a[2:], " "))
		}
		if len(a) == 3 && a[1] == "delete" && a[2] == "all" {
			return bibleit.DeleteAllLivesCommand(), nil
		}
		if len(a) == 3 && oneOf(a[2], "info", "stats", "start", "stop", "pause", "resume", "delete", "subscribe", "clear") {
			return bibleit.LiveCommand(a[1], bibleit.LiveAction(strings.ToUpper(a[2])))
		}
		if len(a) == 4 && a[2] == "secret" {
			if oneOf(a[3], "create", "rotate", "delete") {
				return bibleit.LiveSecretCommand(a[1], bibleit.SecretAction(strings.ToUpper(a[3])), "")
			}
			return bibleit.LiveSecretCommand(a[1], bibleit.SecretAuthenticate, a[3])
		}
		if len(a) == 5 && a[2] == "secret" && a[3] == "set" {
			return bibleit.Command{}, errors.New("custom Live secret setting is not supported by the server command API; use secret create or secret rotate")
		}
		if len(a) >= 5 && a[2] == "set" && oneOf(a[3], "name", "reference", "translations") {
			values := a[4:]
			if a[3] != "translations" {
				values = []string{strings.Join(values, " ")}
			}
			return bibleit.SetLiveCommand(a[1], bibleit.LiveField(strings.ToUpper(a[3])), values...)
		}
		if len(a) == 4 && a[2] == "stack" && oneOf(a[3], "info", "clear") {
			return bibleit.LiveStackCommand(a[1], bibleit.StackAction(strings.ToUpper(a[3])))
		}
		if len(a) >= 5 && len(a) <= 8 && a[2] == "stack" && a[3] == "push" {
			ref, err := stackPushReference(a[4:])
			if err != nil {
				return bibleit.Command{}, err
			}
			return bibleit.PushLiveCommand(a[1], ref)
		}
		if (len(a) == 4 || len(a) == 5) && a[2] == "stack" && a[3] == "pop" {
			count := 0
			if len(a) == 5 {
				var err error
				count, err = strconv.Atoi(a[4])
				if err != nil || count <= 0 {
					return bibleit.Command{}, errors.New("invalid pop count")
				}
			}
			return bibleit.PopLiveCommand(a[1], count)
		}
	}
	return bibleit.Command{}, errors.New("unsupported command")
}

func cliReference(translation string, args []string) (bibleit.Reference, error) {
	if len(args) < 1 || len(args) > 3 {
		return bibleit.Reference{}, errors.New("invalid reference")
	}
	ref := bibleit.Reference{Translation: translation, Book: args[0]}
	parts := args[1:]
	if len(parts) == 1 && strings.Contains(parts[0], ":") {
		parts = strings.Split(parts[0], ":")
	}
	if len(parts) > 2 {
		return ref, errors.New("invalid reference")
	}
	for i, part := range parts {
		value, err := strconv.Atoi(part)
		if err != nil || value <= 0 || value > 255 {
			return ref, errors.New("invalid reference")
		}
		if i == 0 {
			ref.Chapter = value
		} else {
			ref.Verse = value
		}
	}
	return ref, nil
}
func oneOf(s string, values ...string) bool {
	for _, v := range values {
		if s == v {
			return true
		}
	}
	return false
}

type authenticationError struct{ message string }

func (e authenticationError) Error() string { return e.message }

func requestSSH(cfg config, command bibleit.Command) ([]string, error) {
	ssh, identity, err := sshCommand(cfg, command)
	if err != nil {
		return nil, err
	}
	var stdout, stderr bytes.Buffer
	ssh.Stdout, ssh.Stderr = &stdout, &stderr
	if err := ssh.Run(); err != nil {
		return nil, sshCommandError(identity, stderr.String(), err)
	}
	result, err := bibleit.DecodeCommandResponse(command, &stdout)
	return result.Lines, err
}

func sshCommand(cfg config, command bibleit.Command) (*exec.Cmd, string, error) {
	identity := ""
	if cfg.Identity != "" {
		var err error
		identity, err = validateSSHIdentity(cfg.Identity)
		if err != nil {
			return nil, "", err
		}
	}
	host, port, err := net.SplitHostPort(sshEndpoint(cfg))
	if err != nil || host == "" || port == "" {
		return nil, "", fmt.Errorf("invalid SSH server address %q", sshEndpoint(cfg))
	}
	targetHost := host
	if strings.Contains(host, ":") {
		targetHost = "[" + host + "]"
	}
	args := []string{"-T", "-p", port}
	if identity != "" {
		args = append(args, "-i", identity, "-o", "IdentitiesOnly=yes")
	}
	args = append(args,
		"-o", "PreferredAuthentications=publickey",
		"-o", "PasswordAuthentication=no", "-o", "KbdInteractiveAuthentication=no",
		"-o", "StrictHostKeyChecking=accept-new", "-o", "ConnectTimeout=5")
	args = append(args, "--", sshUser+"@"+targetHost, command.String())
	ssh := exec.Command("ssh", args...)
	ssh.Stdin = os.Stdin
	return ssh, identity, nil
}

func sshCommandError(identity, stderr string, commandErr error) error {
	if strings.Contains(strings.ToLower(stderr), "permission denied") {
		return authenticationError{message: "SSH public-key authentication was rejected. Register the matching public key in the dashboard and check the configured identity or SSH agent."}
	}
	if strings.Contains(strings.ToLower(stderr), "host key") {
		return userMessage{"SSH host verification failed. Verify the server fingerprint through a trusted channel before updating known_hosts."}
	}
	return userMessage{"SSH connection failed. Check your connection, registered key and OpenSSH configuration; a mutation may already have completed."}
}

func subscribeSSH(cfg config, command bibleit.Command, format string) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	ssh, identity, err := sshCommand(cfg, command)
	if err != nil {
		return err
	}
	ssh = exec.CommandContext(ctx, ssh.Path, ssh.Args[1:]...)
	ssh.Stdin = os.Stdin
	stdout, err := ssh.StdoutPipe()
	if err != nil {
		return err
	}
	var stderr bytes.Buffer
	ssh.Stderr = &stderr
	if err := ssh.Start(); err != nil {
		return err
	}

	// Always reap the OpenSSH child, including protocol errors and Ctrl-C.
	defer func() { _ = ssh.Process.Kill(); _ = ssh.Wait() }()
	var streamOutputErr error
	display := subscriptionTableDisplay{}
	streamErr := consumeSubscription(stdout, func(line string, initial bool) {
		if streamOutputErr != nil {
			return
		}
		if format != "table" {
			streamOutputErr = writeSubscriptionLine(line, format)
			if streamOutputErr != nil {
				stop()
			}
		} else {
			var message string
			message, streamOutputErr = display.line(line, initial)
			if streamOutputErr != nil {
				stop()
				return
			}
			if message != "" {
				fmt.Println(message)
			}
			if initial {
				fmt.Print("Watching Live changes. Press Ctrl-C to stop.\n\n")
			}
		}
	})
	if streamOutputErr != nil {
		return streamOutputErr
	}
	if ctx.Err() != nil {
		return nil
	}
	if errors.Is(streamErr, io.ErrUnexpectedEOF) {
		if err := ssh.Wait(); err != nil {
			return sshCommandError(identity, stderr.String(), err)
		}
	}
	return streamErr
}

func consumeSubscription(reader io.Reader, emit func(string, bool)) error {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 4096), 2*1024*1024)
	scanner.Split(func(data []byte, atEOF bool) (int, []byte, error) {
		if atEOF && len(data) > 0 && !bytes.Contains(data, []byte{'\n'}) {
			return 0, nil, io.ErrUnexpectedEOF
		}
		return bufio.ScanLines(data, atEOF)
	})
	initial := true
	var pending string
	for scanner.Scan() {
		line := scanner.Text()
		if pending != "" {
			line = pending + "\n" + line
		}
		if len(line) > 2*1024*1024 {
			return errors.New("Live subscription record exceeds 2 MiB")
		}
		record, err := bibleit.ParseRecord(line)
		if errors.Is(err, bibleit.ErrIncompleteTerm) {
			pending = line
			continue
		}
		if err != nil {
			return fmt.Errorf("read Live subscription: %w", err)
		}
		pending = ""
		if record.Type == "err" {
			_, err := bibleit.DecodeResponse(strings.NewReader(line + "\n"))
			return err
		}
		if initial {
			if record.Type != "ok" {
				return errors.New("subscription must start with OK or ERR")
			}
		} else if record.Type != "event" {
			return errors.New("expected a Live subscription EVENT")
		}
		emit(line, initial)
		initial = false
		if line == "EVENT closed" {
			return nil
		}
		if line == "EVENT revoked" {
			return &bibleit.ServerError{Code: "revoked"}
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read Live subscription: %w", err)
	}
	if pending != "" {
		return fmt.Errorf("read Live subscription: %w", io.ErrUnexpectedEOF)
	}
	return io.ErrUnexpectedEOF // Disconnect is not a successful subscription completion.
}

func printSubscriptionLine(line, format string) { printResponse([]string{line}, format) }

func browserAuthenticate(cfg *config) (string, error) {
	if _, err := validatedWebURL(defaultWebURL); err != nil {
		return "", err
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", fmt.Errorf("start local authentication callback: %w", err)
	}
	defer listener.Close()
	verifier, err := randomURLSafe(48)
	if err != nil {
		return "", err
	}
	clientState, err := randomURLSafe(32)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(digest[:])
	redirectURI := "http://" + listener.Addr().String() + "/callback"
	authorizeURL, err := url.Parse(endpoint(defaultWebURL, "/cli/auth"))
	if err != nil {
		return "", err
	}
	query := authorizeURL.Query()
	query.Set("redirect_uri", redirectURI)
	query.Set("state", clientState)
	query.Set("code_challenge", challenge)
	authorizeURL.RawQuery = query.Encode()

	result := make(chan browserAuthorizationResult, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", browserAuthorizationCallback(clientState, result))
	server := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() { _ = server.Serve(listener) }()

	fmt.Fprintln(os.Stderr, "Opening your browser to authorize Bibleit CLI…")
	if err := openBrowser(authorizeURL.String()); err != nil {
		fmt.Fprintf(os.Stderr, "Could not open a browser automatically. Open this URL:\n%s\n", authorizeURL.String())
	}
	var received browserAuthorizationResult
	select {
	case received = <-result:
	case <-time.After(30 * time.Minute):
		_ = server.Shutdown(context.Background())
		return "", errors.New("web authentication timed out")
	}
	_ = server.Shutdown(context.Background())
	if received.err != nil {
		return "", received.err
	}
	displayName, token, err := exchangeBrowserCode(received.code, verifier)
	if err != nil {
		return "", err
	}
	cfg.AccessToken = token
	if err := saveConfig(*cfg); err != nil {
		return "", fmt.Errorf("save configuration: %w", err)
	}
	return displayName, nil
}

func authenticatedIdentity(lines []string) string {
	for _, record := range parseRecords(lines) {
		if identityName := record.Fields["identity_name"]; identityName != "" {
			return identityName
		}
		if displayName := record.Fields["display_name"]; displayName != "" {
			return displayName
		}
		if actor := record.Fields["actor"]; actor != "" {
			return actor
		}
	}
	return ""
}

func exchangeBrowserCode(code, verifier string) (string, string, error) {
	deviceName, _ := os.Hostname()
	if deviceName == "" {
		deviceName = "Bibleit CLI"
	}
	client, err := bibleit.NewAuthClientWithHTTPClient(cliHTTPClient)
	if err != nil {
		return "", "", err
	}
	payload, err := client.ExchangeCode(context.Background(), bibleit.CodeExchange{Code: code, Verifier: verifier, DeviceName: "Bibleit CLI on " + deviceName})
	if err != nil {
		var httpErr *bibleit.HTTPError
		if errors.As(err, &httpErr) && httpErr.Code == "quota_exceeded" {
			return "", "", errors.New("CLI credential limit reached; revoke an old CLI credential in account settings and try again")
		}
		return "", "", fmt.Errorf("exchange browser authorization: %w", err)
	}
	name := payload.IdentityName
	if name == "" {
		name = payload.DisplayName
	}
	if name == "" {
		name = payload.Actor
	}
	return name, payload.AccessToken, nil
}

func logout(cfg *config) error {
	if cfg.AccessToken != "" {
		client, err := bibleit.NewClient(cfg.AccessToken, cliHTTPClient)
		if err != nil {
			return err
		}
		if err := client.RevokeCredential(context.Background()); err != nil {
			var httpErr *bibleit.HTTPError
			// An unauthorized credential is already unusable. Other failures
			// leave the local credentials intact so the user can retry logout.
			if !errors.As(err, &httpErr) || httpErr.StatusCode != http.StatusUnauthorized {
				return fmt.Errorf("revoke cached web authentication: %w", err)
			}
		}
	}
	next := *cfg
	next.AccessToken = ""
	next.Identity = ""
	if err := saveConfig(next); err != nil {
		return err
	}
	*cfg = next
	return nil
}

func validatedWebURL(raw string) (*url.URL, error) {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return nil, fmt.Errorf("invalid web server URL %q", raw)
	}
	if parsed.Scheme == "http" && !oneOf(parsed.Hostname(), "127.0.0.1", "localhost", "::1") {
		return nil, fmt.Errorf("insecure web server URL %q; remote web authentication requires HTTPS", raw)
	}
	return parsed, nil
}

func endpoint(webURL, path string) string { return strings.TrimRight(webURL, "/") + path }

func randomURLSafe(size int) (string, error) {
	value := make([]byte, size)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func openBrowser(target string) error {
	var command *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		command = exec.Command("open", target)
	case "windows":
		command = exec.Command("rundll32", "url.dll,FileProtocolHandler", target)
	default:
		command = exec.Command("xdg-open", target)
	}
	return command.Start()
}

func validateSSHIdentity(identity string) (string, error) {
	if identity == "" {
		return "", authenticationError{message: "no SSH identity configured; pass --identity /path/to/id_ed25519 or set identity in the Bibleit config"}
	}
	identity = expandHome(identity)
	if strings.HasSuffix(identity, ".pub") {
		return "", authenticationError{message: "--identity must point to the private key, not its .pub file"}
	}
	info, err := os.Stat(identity)
	if err != nil {
		return "", authenticationError{message: fmt.Sprintf("read SSH private key: %v", err)}
	}
	if !info.Mode().IsRegular() {
		return "", authenticationError{message: fmt.Sprintf("SSH private key %q is not a regular file", identity)}
	}
	contents, err := os.ReadFile(identity + ".pub")
	if err != nil {
		return "", authenticationError{message: fmt.Sprintf("read SSH public key: %v", err)}
	}
	parts := strings.Fields(string(contents))

	if len(parts) < 2 || !oneOf(parts[0], "ssh-ed25519", "ssh-rsa",
		"ecdsa-sha2-nistp256", "ecdsa-sha2-nistp384", "ecdsa-sha2-nistp521") {
		return "", authenticationError{message: "Bibleit supports Ed25519, RSA (2048 bits or larger), and ECDSA P-256/P-384/P-521 public keys"}
	}
	// OpenSSH validates the encoded key rather than trusting its text label.
	fingerprint, err := exec.Command("ssh-keygen", "-l", "-f", identity+".pub").Output()
	if err != nil {
		return "", authenticationError{message: "invalid SSH public key (ssh-keygen could not validate it)"}
	}
	fields := strings.Fields(string(fingerprint))
	if len(fields) < 2 {
		return "", authenticationError{message: "invalid SSH public key fingerprint"}
	}
	bits, err := strconv.Atoi(fields[0])
	if err != nil || (parts[0] == "ssh-rsa" && bits < 2048) {
		return "", authenticationError{message: "Bibleit requires RSA keys of at least 2048 bits"}
	}
	return identity, nil
}

func expandHome(path string) string {
	if path == "~" {
		return homeDirectory()
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(homeDirectory(), strings.TrimPrefix(path, "~/"))
	}
	return path
}
func response(r *bufio.Reader) ([]string, error) {
	result, err := bibleit.DecodeResponse(r)
	return result.Lines, err
}
func parseRecords(lines []string) []outputRecord {
	records := make([]outputRecord, 0, len(lines))
	for _, line := range lines {
		record, err := bibleit.ParseRecord(line)
		if err != nil {
			continue
		}
		records = append(records, outputRecord{Type: record.Type, Fields: record.Fields, order: record.Order})
	}
	return records
}
func printTable(records []outputRecord) {
	records = humanRecords(records)
	if presentHuman(records) {
		return
	}
	if len(records) == 0 {
		return
	}
	rows := records
	if len(records) > 1 && records[0].Type == "ok" {
		rows = records[1:]
	}
	if len(rows) == 1 && rows[0].Type == "ok" {
		if _, event := rows[0].Fields["event"]; event {
			return
		}
		printKeyValue(rows[0])
		return
	}
	if len(rows) == 1 && len(activeCommand) > 1 && oneOf(activeCommand[len(activeCommand)-1], "info", "stats") {
		printKeyValue(rows[0])
		return
	}
	headers := orderedHeaders(rows)
	if len(headers) == 0 {
		return
	}
	if len(headers) == 1 && headers[0] == "text" {
		for _, row := range rows {
			fmt.Println(humanValue(row.Fields["text"]))
		}
		return
	}
	widths := make([]int, len(headers))
	for index, header := range headers {
		widths[index] = cellWidth(strings.ToUpper(header))
	}
	for _, row := range rows {
		for index, header := range headers {
			widths[index] = max(widths[index], cellWidth(tableValue(row, header)))
		}
	}
	printTableRow(headers, widths, true)
	for _, row := range rows {
		values := make([]string, len(headers))
		for index, header := range headers {
			values[index] = tableValue(row, header)
		}
		printTableRow(values, widths, false)
	}
}
func orderedHeaders(records []outputRecord) []string {
	seen := map[string]bool{}
	for _, r := range records {
		for key := range r.Fields {
			seen[key] = true
		}
	}
	headers := []string{}
	for _, key := range []string{"position", "id", "name", "translation", "reference", "text", "status", "mode", "paused", "count"} {
		if seen[key] {
			headers = append(headers, key)
			delete(seen, key)
		}
	}
	rest := []string{}
	for key := range seen {
		rest = append(rest, key)
	}
	sort.Strings(rest)
	return append(headers, rest...)
}

func printKeyValue(record outputRecord) {
	record.order = orderedHeaders([]outputRecord{record})
	width := 0
	for _, key := range record.order {
		width = max(width, cellWidth(strings.ToUpper(key)))
	}
	for _, key := range record.order {
		fmt.Printf("%-*s  %s\n", width, strings.ToUpper(key), tableValue(record, key))
	}
}
func printTableRow(values []string, widths []int, header bool) {
	for index, value := range values {
		if index > 0 {
			fmt.Print("  │ ")
		}
		if header {
			value = strings.ToUpper(value)
		}
		fmt.Print(value + strings.Repeat(" ", max(0, widths[index]-cellWidth(value))))
	}
	fmt.Println()
}
func protocolFailure(err error) int {
	// Classify through the shared HTTP error codes without losing the original
	// REST headers needed for safe Retry-After guidance in describeError.
	displayErr := err
	var problem *bibleit.ProblemError
	if errors.As(err, &problem) {
		err = &bibleit.HTTPError{StatusCode: problem.Status, Code: problem.Code, Fields: map[string]any{"request_id": problem.RequestID, "retry_after": problem.Header.Get("Retry-After")}}
	}

	code := exitFailure
	var authentication authenticationError
	if errors.As(err, &authentication) {
		code = exitUnauthenticated
	}
	serverCode := ""
	var serverErr *bibleit.ServerError
	if errors.As(err, &serverErr) {
		serverCode = serverErr.Code
	}
	var httpErr *bibleit.HTTPError
	if errors.As(err, &httpErr) {
		serverCode = httpErr.Code
		switch httpErr.StatusCode {
		case 400, 422:
			code = exitUsage
		case 401:
			code = exitUnauthenticated
		case 403:
			code = exitForbidden
		case 404:
			code = exitNotFound
		case 409, 412, 428:
			code = exitConflict
		case 429:
			code = exitRateLimited
		}
	}
	switch serverCode {
	case "unauthorized", "invalid_token", "revoked":
		code = exitUnauthenticated
	case "forbidden", "organization_pending_verification", "organization_suspended":
		code = exitForbidden
	case "not_found", "actor_not_found", "user_not_found", "live_not_found", "translation_not_found", "book_not_found", "chapter_not_found", "verse_not_found":
		code = exitNotFound
	case "rate_limited":
		code = exitRateLimited
	case "actor_exists", "quota_exceeded", "conflict", "precondition_failed", "precondition_required":
		code = exitConflict
	}
	return outputError(code, displayErr)
}
func takeOption(a []string, name string) (string, []string) {
	for i := 0; i < len(a); i++ {
		if a[i] == name && i+1 < len(a) {
			value := a[i+1]
			remaining := append([]string{}, a[:i]...)
			remaining = append(remaining, a[i+2:]...)
			return value, remaining
		}
	}
	return "", a
}
func configPath() string { return credentialPath() }

func homeDirectory() string {
	if home, err := os.UserHomeDir(); err == nil {
		return home
	}
	return "."
}
func usage() int {
	return fail(exitUsage, "usage: bibleit [--transport http|ssh] [--identity path] [--format table|raw|json|ndjson] <command>\n\nRun \"bibleit help\" to list commands.")
}
func help(topic []string) int { return registryHelp(topic) }

func fail(code int, format string, args ...any) int {
	return outputError(code, userMessage{fmt.Sprintf(format, args...)})
}

func sshEndpoint(c config) string {
	if c.Transport == "ssh" && c.Endpoint != "" {
		return c.Endpoint
	}
	return defaultSSHServer
}
