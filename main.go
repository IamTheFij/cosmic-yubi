package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"syscall"

	"git.iamthefij.com/iamthefij/cosmic-yubi/gocosmic"
	"git.iamthefij.com/iamthefij/slog"
	"github.com/yawn/ykoath"
	"github.com/zalando/go-keyring"
	"golang.org/x/term"
)

var (
	oath                *ykoath.OATH
	serviceName         = "com.iamthefij.cosmic-yubi"
	version             = "dev"
	errFailedValidation = errors.New("failed validation, password may be incorrect")
	logFile             = "cosmic-yubi-debug.log"
)

func setPassword(s *ykoath.Select) error {
	fmt.Print("Enter Password: ")

	bytePassword, err := term.ReadPassword(syscall.Stdin)
	if err != nil {
		return fmt.Errorf("failed reading password from input: %w", err)
	}

	password := string(bytePassword)

	// get key
	key := s.DeriveKey(password)

	// verify password is correct with a validate call
	ok, err := oath.Validate(s, key)
	if err != nil {
		return fmt.Errorf("error in validate: %w", err)
	}

	if !ok {
		return errFailedValidation
	}

	err = keyring.Set(
		serviceName,
		s.DeviceID(),
		password,
	)
	if err != nil {
		return fmt.Errorf("error saving password in keyring: %w", err)
	}

	return nil
}

func getPasskey(s *ykoath.Select) ([]byte, error) {
	password, err := keyring.Get(serviceName, s.DeviceID())
	if err != nil {
		return nil, fmt.Errorf("error retrieving key from keyring: %w", err)
	}

	return s.DeriveKey(password), nil
}

func handleRequest(results *gocosmic.PluginSearchResults, request gocosmic.Request) error {
	// Handle termination requests
	if request.Close || request.Exit || request.Interrupt {
		// We can exit because the runner should actually be handling this for us
		return nil
	}

	// Handle search for a requst
	if request.Search != "" {
		query := strings.TrimSpace(
			strings.TrimPrefix(strings.TrimPrefix(request.Search, "yubi"), ":"),
		)
		if results.IsEmpty() {
			slog.Debugf("populating initial results list")
			names, err := oath.List()
			if err != nil {
				return fmt.Errorf("failed listing names: %w", err)
			}

			// Generate list of search results
			for i, name := range names {
				slog.Debugf("adding result for %s", name.Name)
				results.Add(gocosmic.PluginSearchResult{
					Id:          uint32(i + 1),
					Name:        name.Name,
					Description: "Generate TOTP",
					Icon:        gocosmic.IconSource{Name: "password-manager"},
				})
			}
		}

		for _, result := range results.Search(query) {
			slog.Debugf("writing result for %s", result.Name)
			_ = gocosmic.WriteResponseToStdout(gocosmic.PluginResponse{
				Append: result,
			})
		}
		slog.Debugf("wrote all responses to stdout")
		_ = gocosmic.WriteResponseToStdout(gocosmic.PluginResponse{Finished: true})
	} else if request.Complete > 0 {
		result, found := results.Get(request.Complete)
		slog.Debugf("completion request: %v: %+v", request.Complete, result)
		if found {
			_ = gocosmic.WriteResponseToStdout(gocosmic.PluginResponse{
				Fill: result.Name,
			})
		}
		slog.Debugf("wrote completion responses to stdout")
	} else if request.Activate > 0 {
		result, found := results.Get(request.Activate)
		slog.Debugf("activate request: %v: %+v", request.Activate, result)
		slog.Debugf("clear results")
		_ = gocosmic.WriteResponseToStdout(gocosmic.PluginResponse{Clear: true})
		if found {
			code, err := oath.Calculate(result.Name, func(string) error {
				fmt.Fprintf(os.Stderr, "Touch key to generate code for %s...", result.Name)
				_ = gocosmic.WriteResponseToStdout(gocosmic.PluginResponse{Append: gocosmic.PluginSearchResult{
					Id:          0,
					Name:        result.Name,
					Description: "Touch key to generate code",
					Icon:        gocosmic.IconSource{Name: "password-manager"},
				}})
				_ = gocosmic.WriteResponseToStdout(gocosmic.PluginResponse{Finished: true})

				return nil
			})

			if err != nil {
				return fmt.Errorf("failed to retrieve credential: %w", err)
			}

			_ = gocosmic.WriteResponseToStdout(gocosmic.PluginResponse{
				Fill: code,
			})
		}
		slog.Debugf("wrote activate response to stdout")
	}

	return nil
}

func usage() {
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Usage:\n  Expects a request in the form of a JSON message via stdin and returns a result to stdout.\n\n")
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Format can be found here https://github.com/pop-os/launcher?tab=readme-ov-file#json-ipc")
	_, _ = fmt.Fprintf(flag.CommandLine.Output(), "If a touch is required for your code, the command will not return until the your key is touched.\n\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.BoolVar(&slog.DebugLevel, "debug", true, "enable debug logging")
	showVersion := flag.Bool("version", false, "print version and exit")
	shouldSetPassword := flag.Bool("set-password", false, "prompt for key password and store in system keychain")
	flag.Parse()

	logFile, err := os.OpenFile(logFile, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	slog.OnErrPanicf(err, "failed opening log file %v", err)
	slog.SetOutput(logFile)

	defer func() {
		_ = logFile.Close()
	}()

	if *showVersion {
		fmt.Printf("Version %s\n", version)
		os.Exit(0)
	}

	oath, err = ykoath.New()
	slog.OnErrFatalf(err, "failed to initialize new oath")

	defer func() {
		_ = oath.Close()
	}()

	if slog.DebugLevel {
		oath.Debug = slog.Debugf
	}

	// Select oath to begin
	s, err := oath.Select()
	slog.OnErrFatalf(err, "failed to select oath")

	// Check to see if we are trying to set a password
	if *shouldSetPassword {
		err = setPassword(s)
		slog.OnErrFatalf(err, "failed to save password")

		return
	}
	// If required, authenticate with password from keychain
	if s.Challenge != nil {
		passKey, err := getPasskey(s)
		slog.OnErrFatalf(err, "failed retrieving password key")

		ok, err := oath.Validate(s, passKey)
		slog.OnErrFatalf(err, "validation failed")

		if !ok {
			slog.Fatalf("failed validation, password is incorrect")
		}
	} else {
		slog.Debugf("no challenge required")
	}

	if flag.Arg(0) == "list" {
		// List names only
		names, err := oath.List()
		slog.OnErrFatalf(err, "failed to list names")

		for _, name := range names {
			fmt.Println(name.Name)
		}

		return
	}

	plugin := gocosmic.NewPlugin(handleRequest)
	plugin.Run()
}
