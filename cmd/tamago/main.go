// tamago-go installer and runner
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// The tamago command downloads, compiles, and runs the go command from the
// TamaGo Go distribution matching the tamago module version.
//
// Add github.com/usbarmory/tamago to go.mod, and then either use
//
//	go run github.com/usbarmory/tamago/cmd/tamago
//
// as the go command, or add a line like
//
//	tool github.com/usbarmory/tamago/cmd/tamago
//
// to go.mod and use "go tool tamago" as the go command.
package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
)

func moduleVersion() (string, error) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "", fmt.Errorf("failed to read build info")
	}
	for _, dep := range info.Deps {
		if dep.Path == "github.com/usbarmory/tamago" {
			return dep.Version, nil
		}
	}
	if info.Main.Path == "github.com/usbarmory/tamago" {
		return info.Main.Version, nil
	}
	return "", fmt.Errorf("tamago module not found in build info")
}

func main() {
	log.SetFlags(0)

	version, err := moduleVersion()
	if err != nil {
		log.Fatalf("tamago: %v", err)
	}
	if !strings.HasPrefix(version, "v1.") {
		log.Fatalf("tamago: unsupported tamago module version %q", version)
	}
	// A version of e.g. v1.25.7-0.20260130090423-5d846371fd71+dirty means the
	// latest tag was v1.25.6. If present, strip the prerelease suffix and
	// rollback the patch.
	if v, _, ok := strings.Cut(version, "-"); ok {
		parts := strings.SplitN(v, ".", 3)
		if len(parts) != 3 {
			log.Fatalf("tamago: unsupported tamago module version %q", version)
		}
		patch, err := strconv.Atoi(parts[2])
		if err != nil || patch == 0 {
			log.Fatalf("tamago: unsupported tamago module version %q", version)
		}
		parts[2] = strconv.Itoa(patch - 1)
		version = strings.Join(parts, ".")
	}
	version = strings.TrimPrefix(version, "v")
	version = "tamago-go" + version

	root, err := goroot(version)
	if err != nil {
		log.Fatalf("tamago: %v", err)
	}

	gobin := filepath.Join(root, "bin", "go"+exe())
	if _, err := os.Stat(gobin); err != nil {
		fmt.Printf("tamago: installing %s...\n", version)
		if err := install(root, version); err != nil {
			log.Fatalf("tamago: %v", err)
		}
	}

	runGo(root)
}

func install(root, tag string) error {
	if err := os.MkdirAll(root, 0755); err != nil {
		return fmt.Errorf("failed to create repository: %v", err)
	}

	cmd := exec.Command("git", "clone", "--depth=1", "--branch="+tag, "https://github.com/usbarmory/tamago-go", root)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "PWD="+cmd.Dir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone git repository: %v", err)
	}

	cmd = exec.Command(filepath.Join(root, "src", makeScript()))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = filepath.Join(root, "src")
	// Add new GOROOT/bin to PATH to silence path warning at end of make.bash.
	// Add PWD to environment to fix future calls to os.Getwd.
	newPath := filepath.Join(root, "bin")
	if p := os.Getenv("PATH"); p != "" {
		newPath += string(filepath.ListSeparator) + p
	}
	cmd.Env = append(os.Environ(), "PATH="+newPath, "PWD="+cmd.Dir)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build go: %v", err)
	}

	return nil
}

func makeScript() string {
	switch runtime.GOOS {
	case "plan9":
		return "make.rc"
	case "windows":
		return "make.bat"
	default:
		return "make.bash"
	}
}

func runGo(root string) {
	gobin := filepath.Join(root, "bin", "go"+exe())
	cmd := exec.Command(gobin, os.Args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	newPath := filepath.Join(root, "bin")
	if p := os.Getenv("PATH"); p != "" {
		newPath += string(filepath.ListSeparator) + p
	}
	cmd.Env = append(os.Environ(), "GOROOT="+root, "PATH="+newPath)

	ignoreSignals()

	err := cmd.Run()
	if eerr, ok := err.(*exec.ExitError); ok {
		os.Exit(eerr.ExitCode())
	} else if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func exe() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

func goroot(version string) (string, error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("failed to get cache directory: %v", err)
	}
	return filepath.Join(cache, "tamago-go", version), nil
}

func ignoreSignals() {
	// Ensure that signals intended for the child process are not handled by
	// this process' runtime (e.g. SIGQUIT). See issue #36976.
	signal.Ignore(signalsToIgnore...)
}
