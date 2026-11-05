package ltk

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// ensureLauncherStub builds ltk-manager.exe in the runtime directory.
// The stub launches ltk_patcher_host.exe, optionally with a signed
// ltk-manager-signed.exe as the spoofed parent to pass the DLL's
// VRF00000008 (image name) and VRF00000009 (signature) checks.
// The stub is cached after first build. Go must be on PATH.
func ensureLauncherStub(runtimeDir string) (string, error) {
	stubExe := filepath.Join(runtimeDir, "ltk-manager.exe")
	if _, err := os.Stat(stubExe); err == nil {
		return stubExe, nil
	}

	srcDir := filepath.Join(runtimeDir, "ltk-manager-stub")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		return "", err
	}
	src := filepath.Join(srcDir, "main.go")
	if err := os.WriteFile(src, []byte(managerStubSrc), 0644); err != nil {
		return "", err
	}

	cmd := exec.Command("go", "build", "-trimpath", "-ldflags=-s -w", "-o", stubExe, src)
	cmd.Dir = srcDir
	cmd.Env = append(os.Environ(), "GOOS=windows", "GOARCH=amd64", "CGO_ENABLED=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("build ltk-manager stub: %v\n%s", err, out)
	}
	return stubExe, nil
}

const managerStubSrc = `package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

const CREATE_SUSPENDED = 0x00000004

func main() {
	exe, _ := os.Executable()
	dir := filepath.Dir(exe)
	hostPath := filepath.Join(dir, "ltk_patcher_host.exe")
	signedManagerPath := filepath.Join(dir, "ltk-manager-signed.exe")

	if _, err := os.Stat(signedManagerPath); os.IsNotExist(err) {
		runDirect(hostPath, dir)
		return
	}

	// Create the signed ltk-manager-signed.exe SUSPENDED.
	// This process is never resumed — it exists only so the DLL's
	// WinVerifyTrust (VRF00000009) sees a signed parent on disk.
	managerPathPtr, err := syscall.UTF16PtrFromString(signedManagerPath)
	if err != nil {
		runDirect(hostPath, dir)
		return
	}

	var si syscall.StartupInfo
	var pi syscall.ProcessInformation

	err = syscall.CreateProcess(
		nil,
		managerPathPtr,
		nil,
		nil,
		false,
		CREATE_SUSPENDED,
		nil,
		nil,
		&si,
		&pi,
	)
	if err != nil {
		runDirect(hostPath, dir)
		return
	}

	parentHandle := syscall.Handle(pi.Process)

	// Launch the real host with the suspended signed manager as parent.
	// The Go runtime translates ParentProcess into
	// PROC_THREAD_ATTRIBUTE_PARENT_PROCESS automatically.
	hostCmd := exec.Command(hostPath, os.Args[1:]...)
	hostCmd.SysProcAttr = &syscall.SysProcAttr{
		ParentProcess: parentHandle,
	}
	hostCmd.Dir = dir
	hostCmd.Stdin = os.Stdin
	hostCmd.Stdout = os.Stdout
	hostCmd.Stderr = os.Stderr

	startErr := hostCmd.Start()

	// Terminate the suspended decoy — it is no longer needed once the
	// host is created (or failed). Close both handles.
	syscall.TerminateProcess(syscall.Handle(pi.Process), 0)
	syscall.CloseHandle(syscall.Handle(pi.Process))
	syscall.CloseHandle(syscall.Handle(pi.Thread))

	if startErr != nil {
		runDirect(hostPath, dir)
		return
	}

	// Wait for the host and propagate its exit code.
	err = hostCmd.Wait()
	if err != nil {
		if ex, ok := err.(*exec.ExitError); ok {
			os.Exit(ex.ExitCode())
		}
		os.Exit(1)
	}
}

func runDirect(hostPath, dir string) {
	cmd := exec.Command(hostPath, os.Args[1:]...)
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if ex, ok := err.(*exec.ExitError); ok {
			os.Exit(ex.ExitCode())
		}
		os.Exit(1)
	}
}
`