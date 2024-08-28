package main

import (
	"embed"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

//go:embed resources/*
var embeddedFiles embed.FS

func getBinaryName() (string, error) {
	os := runtime.GOOS
	arch := runtime.GOARCH

	switch {
	case os == "linux" && arch == "amd64":
		return "resources/apk-mitm-linux", nil
	case os == "windows" && arch == "amd64":
		return "resources/apk-mitm-windows.exe", nil
	default:
		return "", fmt.Errorf("unsupported OS or architecture: %s-%s", os, arch)
	}
}

func writeExecutable(tempDir string) (string, error) {
	binaryName, err := getBinaryName()
	if err != nil {
		return "", err
	}

	binaryData, err := embeddedFiles.ReadFile(binaryName)
	if err != nil {
		return "", fmt.Errorf("binary not found in resources: %v", err)
	}

	tempFilePath := filepath.Join(tempDir, "apk-mitm-executable")
	if runtime.GOOS == "windows" {
		tempFilePath += ".exe"
	}

	if err := ioutil.WriteFile(tempFilePath, binaryData, 0755); err != nil {
		return "", fmt.Errorf("failed to write temp file: %v", err)
	}

	return tempFilePath, nil
}

func executeCommand(tempFilePath string, args []string) error {
	fmt.Printf("[+] Executable file path: %s\n", tempFilePath)
	fmt.Printf("[+] Arguments: %v\n", args)

	time.Sleep(5 * time.Second)

	cmd := exec.Command(tempFilePath, args...)
	cmd.Stdin = nil
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute command: %v", err)
	}

	return nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("[+] Usage: %s <path-to-apk>\n", os.Args[0])
		os.Exit(1)
	}

	apkPath := os.Args[1]

	tempDir, err := ioutil.TempDir("", "apk-mitm-tempdir")
	if err != nil {
		fmt.Printf("[+] Error creating temp directory: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tempDir)

	tempFilePath, err := writeExecutable(tempDir)
	if err != nil {
		fmt.Printf("[+] Error: %v\n", err)
		os.Exit(1)
	}

	if err := executeCommand(tempFilePath, []string{apkPath}); err != nil {
		fmt.Printf("[+] Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("[+] Patching complete!")
}

