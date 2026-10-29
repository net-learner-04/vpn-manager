package app

import (
	"fmt"
	"os"
	"os/exec"
)

func RestartApp() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}

	cmd := exec.Command(exePath, os.Args[1:]...)
	cmd.Dir = getWorkingDirectory()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("restart application: %w", err)
	}

	os.Exit(0)

	return nil
}

func getWorkingDirectory() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	return dir
}
