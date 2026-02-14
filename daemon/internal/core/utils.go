package core

import (
	"fmt"
	"os"
	"path/filepath"
)

func createDir(path string) (string, error) {
	if err := os.MkdirAll(path, 0o700); err != nil {
		return "", fmt.Errorf("mkdir %s: %w", path, err)
	}
	return path, nil
}

// UserHome Returns the home folder of the current user
func UserHome() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "", fmt.Errorf("unable to resolve home dir: %w", err)
	}
	return home, nil
}

// BackerHome Returns the default path for the
// backer folder where all configs, plans and registry
// are saved for the daemon to work.
func BackerHome() (string, error) {
	home, err := UserHome()
	if err != nil {
		return "", err
	}
	return createDir(filepath.Join(home, ".backer"))
}

// BackerLogHome Returns the default path where logs
// produced by runs are saved into
func BackerLogHome() (string, error) {
	home, err := BackerHome()
	if err != nil {
		return "", err
	}
	return createDir(filepath.Join(home, "log"))
}

// RegistryFile Returns the path to the registry file
func RegistryFile() (string, error) {
	home, err := BackerHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "registry.db"), nil
}

// Exist checks if a file exists in the current machine
func Exist(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
