// Copyright (c) 2021 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	// defaultApp is the default application assumed when either no arguments
	// are specified or the first argument starts with a -.
	defaultApp = "exccd"
)

// argN either returns the arguments at the provided position within the given
// args array when it exists or an empty string otherwise.
func argN(args []string, n int) string {
	if len(args) > n {
		return args[n]
	}
	return ""
}

// prepend return a new slice that consists of the provided value followed by
// the given args.
func prepend(args []string, val string) []string {
	newArgs := make([]string, 0, len(args)+1)
	newArgs = append(newArgs, val)
	newArgs = append(newArgs, args...)
	return newArgs
}

// fileExists reports whether the named file or directory exists.
func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func main() {
	// Name of the invoking executable.  This should typically be "entrypoint".
	exeName := filepath.Base(os.Args[0])

	// Local copy of supplied arguments without the invoking process.  This
	// allows the params to be modified independently below as needed.
	args := make([]string, len(os.Args)-1)
	copy(args, os.Args[1:])

	// Assume the provided arguments are for default app when the first
	// parameter starts with a dash.
	if arg0 := argN(args, 0); arg0 == "" || arg0[0] == '-' {
		fmt.Printf("%s: assuming arguments for %s\n", exeName, defaultApp)
		args = prepend(args, defaultApp)
	}

	// Additional setup when running in a container.
	arg0 := argN(args, 0)
	args = args[1:]
	switch arg0 {
	case "exccd":
		// Determine the app data directory based on environment variable.
		exccData := os.Getenv("EXCC_DATA")
		exccdAppData := filepath.Join(exccData, ".exccd")

		// TODO: Recognize t/true/1, f/false/0
		if os.Getenv("EXCCD_NO_FILE_LOGGING") != "false" {
			args = append(args, "--nofilelogging")
		}
		args = append(args, fmt.Sprintf("--appdata=%s", exccdAppData))
		args = append(args, "--rpclisten=")

	case "exccctl":
		// Determine the app data directories based on environment variable.
		exccData := os.Getenv("EXCC_DATA")
		exccdAppData := filepath.Join(exccData, ".exccd")
		rpcCert := filepath.Join(exccdAppData, "rpc.cert")
		exccctlAppData := filepath.Join(exccData, ".exccctl")
		exccctlConfig := filepath.Join(exccctlAppData, "exccctl.conf")

		// TODO: These all unconditionally override config file settings.
		// Detect if already there and don't do it?
		//
		// Prepend the arguments in case the caller wants to override them.
		args = prepend(args, fmt.Sprintf("--rpccert=%s", rpcCert))
		args = prepend(args, fmt.Sprintf("--configfile=%s", exccctlConfig))

		// Change the home directory to match the data path since dcrctl
		// relies in it to discover the dcrd config file in order to extract
		// the rpc credentials.
		if !fileExists(filepath.Join(exccctlAppData, "exccctl.conf")) {
			os.Setenv("HOME", exccData)
		}
	}

	// Run the command with the given arguments while redirecting stdin, stdout,
	// and stderr to the parent process.
	cmd := exec.Command(arg0, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(cmd.ProcessState.ExitCode())
	}
}
