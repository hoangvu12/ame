package ltk

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ensureLauncherStub builds ltk-manager.exe in the runtime directory.
// The stub launches ltk_patcher_host.exe as a child of explorer.exe
// to bypass the DLL's VRF00000008 (image name) and VRF00000009 (signature)
// checks. Explorer is signed by Microsoft and on the DLL's allowlist.
// The stub is recompiled whenever the source changes. Go must be on PATH.
func ensureLauncherStub(runtimeDir string) (string, error) {
	stubExe := filepath.Join(runtimeDir, "ltk-manager.exe")

	// Hash the current stub source. If the compiled exe was built from a
	// different source, rebuild it. This prevents stale stubs from running
	// after manager_stub.go is updated.
	srcHash := fmt.Sprintf("%x", sha256.Sum256([]byte(managerStubSrc)))
	hashFile := filepath.Join(runtimeDir, "ltk-manager-stub.hash")
	if existing, err := os.ReadFile(hashFile); err == nil {
		if _, statErr := os.Stat(stubExe); statErr == nil && strings.TrimSpace(string(existing)) == srcHash {
			return stubExe, nil
		}
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
	os.WriteFile(hashFile, []byte(srcHash), 0644)
	return stubExe, nil
}

const managerStubSrc = `package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"unsafe"
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	procCreateToolhelp32Snapshot = kernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32FirstW          = kernel32.NewProc("Process32FirstW")
	procProcess32NextW           = kernel32.NewProc("Process32NextW")
	procCloseHandle              = kernel32.NewProc("CloseHandle")
)

const (
	TH32CS_SNAPPROCESS = 0x00000002
	PROCESS_CREATE_PROCESS = 0x0080
)

type processEntry32 struct {
	Size              uint32
	Usage             uint32
	ProcessID         uint32
	DefaultHeapID     uintptr
	ModuleID          uint32
	Threads           uint32
	ParentProcessID   uint32
	PriorityClassBase int32
	Flags             uint32
	ExeFile           [260]uint16
}

// findExplorerPID returns the PID of the first explorer.exe process.
func findExplorerPID() (uint32, error) {
	snap, _, err := procCreateToolhelp32Snapshot.Call(TH32CS_SNAPPROCESS, 0)
	if snap == 0 || snap == uintptr(syscall.InvalidHandle) {
		return 0, err
	}
	defer procCloseHandle.Call(snap)

	var pe processEntry32
	pe.Size = uint32(unsafe.Sizeof(pe))

	ret, _, _ := procProcess32FirstW.Call(snap, uintptr(unsafe.Pointer(&pe)))
	for ret != 0 {
		name := syscall.UTF16ToString(pe.ExeFile[:])
		if name == "explorer.exe" {
			return pe.ProcessID, nil
		}
		ret, _, _ = procProcess32NextW.Call(snap, uintptr(unsafe.Pointer(&pe)))
	}
	return 0, syscall.ERROR_FILE_NOT_FOUND
}

func main() {
	exe, _ := os.Executable()
	dir := filepath.Dir(exe)
	hostPath := filepath.Join(dir, "ltk_patcher_host.exe")

	// Find running explorer.exe to use as the spoofed parent.
	// Explorer is signed by Microsoft and is on the DLL's allowlist,
	// bypassing both VRF00000008 (image name) and VRF00000009 (signature).
	explorerPID, err := findExplorerPID()
	if err != nil {
		runDirect(hostPath, dir)
		return
	}

	explorerHandle, err := syscall.OpenProcess(PROCESS_CREATE_PROCESS, false, explorerPID)
	if err != nil {
		runDirect(hostPath, dir)
		return
	}
	defer syscall.CloseHandle(explorerHandle)

	hostCmd := exec.Command(hostPath, os.Args[1:]...)
	hostCmd.SysProcAttr = &syscall.SysProcAttr{
		ParentProcess: explorerHandle,
	}
	hostCmd.Dir = dir
	hostCmd.Stdin = os.Stdin
	hostCmd.Stdout = os.Stdout
	hostCmd.Stderr = os.Stderr

	if err := hostCmd.Start(); err != nil {
		runDirect(hostPath, dir)
		return
	}

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