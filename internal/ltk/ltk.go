// Package ltk owns the bundled signed host and package overlay builder.
package ltk

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"embed"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hoangvu12/ame/internal/config"
	"github.com/hoangvu12/ame/internal/display"
	"github.com/hoangvu12/ame/internal/zipname"
)

//go:embed assets/*
var assets embed.FS

var prepareMu sync.Mutex

// Prepare installs immutable, content-addressed files; no download at runtime.
func Prepare() (string, error) {
	prepareMu.Lock()
	defer prepareMu.Unlock()
	data, err := assets.ReadFile("assets/runtime.zip")
	if err != nil {
		return "", fmt.Errorf("LTK bundle missing: run tools/ltk-runtime/build.ps1 before building AME")
	}
	id := fmt.Sprintf("%x", sha256.Sum256(data))
	dest := filepath.Join(config.AmeDir, "runtimes", id[:16])
	archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", err
	}
	valid := true
	for _, entry := range archive.File {
		if entry.FileInfo().IsDir() {
			continue
		}
		name := strings.ReplaceAll(entry.Name, `\`, "/")
		if zipname.Unsafe(entry.Name) {
			return "", fmt.Errorf("unsafe runtime entry")
		}
		existing, err := os.ReadFile(filepath.Join(dest, filepath.FromSlash(name)))
		if err != nil || uint64(len(existing)) != entry.UncompressedSize64 {
			valid = false
			break
		}
		input, err := entry.Open()
		if err != nil {
			return "", err
		}
		h := sha256.New()
		_, err = io.Copy(h, input)
		input.Close()
		previous := sha256.Sum256(existing)
		if err != nil || !bytes.Equal(h.Sum(nil), previous[:]) {
			valid = false
			break
		}
	}
	if valid {
		return dest, nil
	}
	if err = os.MkdirAll(dest, 0755); err != nil {
		return "", err
	}
	for _, entry := range archive.File {
		if zipname.Unsafe(entry.Name) {
			return "", fmt.Errorf("unsafe runtime entry")
		}
		target := filepath.Join(dest, filepath.FromSlash(strings.ReplaceAll(entry.Name, `\`, "/")))
		if entry.FileInfo().IsDir() {
			if err = os.MkdirAll(target, 0755); err != nil {
				return "", err
			}
			continue
		}
		if err = os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return "", err
		}
		input, err := entry.Open()
		if err != nil {
			return "", err
		}
		data, err := io.ReadAll(input)
		input.Close()
		if err != nil {
			return "", err
		}
		if err = os.WriteFile(target, data, 0600); err != nil {
			return "", err
		}
	}
	return dest, nil
}

// Archive an extracted AME mod without including cache markers or other files.
func packMod(dir, path string) error {
	root, err := filepath.EvalSymlinks(dir)
	if err != nil {
		return err
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	archive := zip.NewWriter(file)
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		name := filepath.ToSlash(rel)
		top := strings.SplitN(name, "/", 2)[0]
		if top != "WAD" && top != "RAW" && top != "META" {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("nested mod symlink unsupported: %s", name)
		}
		if info.IsDir() {
			return nil
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("unsupported mod entry")
		}
		header := &zip.FileHeader{Name: name, Method: zip.Store}
		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}
		input, err := os.Open(path)
		if err != nil {
			return err
		}
		defer input.Close()
		_, err = io.Copy(writer, input)
		return err
	})
	closeErr := archive.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return file.Close()
}

func Build(modsDir, overlay, game, names string) error {
	lifecycle.Lock()
	owner := active
	lifecycle.Unlock()
	dir, err := Prepare()
	if err != nil {
		return err
	}
	scratch, err := os.MkdirTemp("", "ame-ltk-build-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(scratch)
	// Build away from the live prefix. A failed build must leave the previous
	// complete overlay available to the preserved scanner.
	if err := os.MkdirAll(filepath.Dir(overlay), 0755); err != nil {
		return err
	}
	staged, err := os.MkdirTemp(filepath.Dir(overlay), ".ame-overlay-stage-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(staged)
	args := []string{game, staged, filepath.Join(scratch, "state")}
	for i, name := range strings.Split(names, "/") {
		if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `\:`) {
			return fmt.Errorf("invalid mod name")
		}
		path := filepath.Join(scratch, fmt.Sprintf("%d.fantome", i))
		if err = packMod(filepath.Join(modsDir, name), path); err != nil {
			return err
		}
		args = append(args, path)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, filepath.Join(dir, "ame-ltk-overlay.exe"), args...)
	hide(cmd)
	output, err := cmd.CombinedOutput()
	if len(output) > 8192 {
		output = output[len(output)-8192:]
	}
	display.Log("LTK build: " + string(output))
	if err != nil {
		return fmt.Errorf("LTK overlay: %w", err)
	}
	return publishOverlay(staged, overlay, owner)
}

type host struct {
	cmd             *exec.Cmd
	stdin           io.WriteCloser
	done            chan struct{}
	ready           chan error
	mu              sync.Mutex
	running, hooked bool
	scanning        bool
	failure         string
	hookPID         uint32
	releaseHooks    func()
	overlay         string
	resumeScanner   func() error
}

var lifecycle sync.Mutex
var active *host

// hostStatus is one consistent snapshot of the active host's state.
type hostStatus struct {
	running, hooked, scanning bool
	failure                   string
	hookPID                   uint32
}

// status samples the active host under both locks. All read-style
// accessors (Running, Hooked, WaitingForGame, Failure) go through it so
// they cannot disagree about lock order or field defaults.
func status() hostStatus {
	lifecycle.Lock()
	defer lifecycle.Unlock()
	if active == nil {
		return hostStatus{}
	}
	active.mu.Lock()
	defer active.mu.Unlock()
	return hostStatus{
		running:  active.running,
		hooked:   active.hooked,
		scanning: active.scanning,
		failure:  active.failure,
		hookPID:  active.hookPID,
	}
}

func Running() bool { return status().running }

func Hooked() bool {
	s := status()
	return s.running && s.hooked && s.failure == ""
}

// WaitingForGame permits replacing an armed overlay only before attachment starts.
func WaitingForGame() bool {
	s := status()
	return s.running && s.scanning && s.hookPID == 0 && s.failure == ""
}

func Failure() string { return status().failure }

// WaitForHook waits for SetWindowsHookEx installation, not DLL execution: a
// suspended game's window thread cannot execute the DLL until it is resumed.
func WaitForHook(pid uint32, cancel <-chan struct{}, timeout time.Duration) error {
	lifecycle.Lock()
	h := active
	lifecycle.Unlock()
	if h == nil {
		return fmt.Errorf("LTK host is not running")
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	tick := time.NewTicker(10 * time.Millisecond)
	defer tick.Stop()
	for {
		h.mu.Lock()
		ready, failure, running := h.hookPID == pid, h.failure, h.running
		h.mu.Unlock()
		if failure != "" {
			return fmt.Errorf("LTK hook failed: %s", failure)
		}
		if !running {
			return fmt.Errorf("LTK host stopped before hook installation")
		}
		if ready {
			return nil
		}
		select {
		case <-cancel:
			return fmt.Errorf("LTK hook wait cancelled")
		case <-h.done:
			return fmt.Errorf("LTK host exited before hook installation")
		case <-timer.C:
			return fmt.Errorf("LTK hook not installed for PID %d before release deadline", pid)
		case <-tick.C:
		}
	}
}

func (h *host) recordHook(line string) {
	if !strings.Contains(line, "game found; hook installed") {
		return
	}
	for _, field := range strings.Fields(line) {
		if strings.HasPrefix(field, "pid=") {
			pid, err := strconv.ParseUint(strings.TrimPrefix(field, "pid="), 10, 32)
			if err == nil {
				h.mu.Lock()
				h.hookPID = uint32(pid)
				h.scanning = false
				h.mu.Unlock()
			}
		}
	}
}

func stopLocked() {
	if active == nil {
		return
	}
	h := active
	active = nil
	if h.resumeScanner != nil {
		if err := h.resumeScanner(); err != nil {
			display.Log("LTK: scanner resume during stop: " + err.Error())
		}
		h.resumeScanner = nil
	}
	if h.releaseHooks != nil {
		h.releaseHooks()
		h.releaseHooks = nil
	}
	fmt.Fprintln(h.stdin, "stop")
	h.stdin.Close()
	select {
	case <-h.done:
	case <-time.After(3 * time.Second):
		h.cmd.Process.Kill()
		<-h.done
	}
}
func Stop() { lifecycle.Lock(); defer lifecycle.Unlock(); stopLocked() }
func Start(overlay string) error {
	dir, err := Prepare()
	if err != nil {
		return err
	}
	return startHost(dir, overlay, exec.Command, true)
}

// StartSuspended is an experimental test entry point, not the AME launch path.
// It requires initialized target GUI state and can block on a frozen target.
func StartSuspended(overlay string, pid uint32) error {
	dir, err := Prepare()
	if err != nil {
		return err
	}
	if err = startHostMode(dir, overlay, exec.Command, "passive"); err != nil {
		return err
	}
	lifecycle.Lock()
	defer lifecycle.Unlock()
	if active == nil {
		return fmt.Errorf("passive host stopped before early hook")
	}
	release, err := installProcessHooks(dir, pid)
	if err != nil {
		stopLocked()
		return err
	}
	active.releaseHooks = release
	active.mu.Lock()
	active.hookPID = pid
	active.mu.Unlock()
	display.Log(fmt.Sprintf("LTK: early thread hooks installed for held PID %d; awaiting resume", pid))
	return nil
}

// ArmSuspended checks that the prepared host can continue after resume. A held
// process may not have initialized user32, so installing a hook on its oldest
// thread is not safe. Keep the scanner alive to discover the actual window.
func ArmSuspended(pid uint32) error {
	lifecycle.Lock()
	defer lifecycle.Unlock()
	h := active
	if h == nil {
		return fmt.Errorf("LTK host is not running")
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.failure != "" {
		return fmt.Errorf("LTK already failed: %s", h.failure)
	}
	if !h.running {
		return fmt.Errorf("LTK host stopped before game release")
	}
	if h.hookPID == pid {
		return nil
	}
	if h.hookPID != 0 || h.hooked || !h.scanning {
		return fmt.Errorf("LTK host is not scanning for held PID %d", pid)
	}
	display.Log(fmt.Sprintf("LTK: prepared scanner retained for PID %d; window discovery requires game resume", pid))
	return nil
}

func startHost(dir, overlay string, command func(string, ...string) *exec.Cmd, scan bool) error {
	mode := ""
	if scan {
		mode = "scan"
	}
	return startHostMode(dir, overlay, command, mode)
}

func startHostMode(dir, overlay string, command func(string, ...string) *exec.Cmd, mode string) error {
	lifecycle.Lock()
	defer lifecycle.Unlock()
	prefix, err := filepath.Abs(overlay)
	if err != nil {
		return err
	}
	if strings.ContainsAny(prefix, "\r\n") {
		return fmt.Errorf("invalid overlay path")
	}
	if mode == "scan" && active != nil {
		active.mu.Lock()
		running, failure := active.running, active.failure
		same := strings.EqualFold(active.overlay, prefix)
		active.mu.Unlock()
		if running {
			if failure != "" {
				return fmt.Errorf("LTK session failed: %s", failure)
			}
			if !same {
				return fmt.Errorf("cannot change an active LTK session's overlay prefix")
			}
			display.Log(fmt.Sprintf("LTK: preserving scan session pid=%d", active.cmd.Process.Pid))
			return nil
		}
	}
	stopLocked()
	cmd := command(filepath.Join(dir, "ltk_patcher_host.exe"))
	cmd.Dir = dir
	hide(cmd)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		stdin.Close()
		return err
	}
	if err = cmd.Start(); err != nil {
		stdin.Close()
		return err
	}
	h := &host{cmd: cmd, stdin: stdin, done: make(chan struct{}), ready: make(chan error, 1), running: true, overlay: prefix}
	active = h
	var readers sync.WaitGroup
	readers.Add(2)
	read := func(stream io.Reader) {
		defer readers.Done()
		scanner := bufio.NewScanner(stream)
		scanner.Buffer(make([]byte, 4096), 256*1024)
		for scanner.Scan() {
			line := scanner.Text()
			h.recordHook(line)
			display.Log("LTK: " + line)
			if strings.Contains(line, "joined too late, not overlaying") || (strings.HasPrefix(line, "dll ") && strings.Contains(line, " ERROR ")) {
				h.mu.Lock()
				h.failure = line
				h.hooked = false
				h.scanning = false
				h.mu.Unlock()
			}
			fields := strings.Fields(line)
			if len(fields) >= 3 && fields[0] == "status" {
				h.mu.Lock()
				h.scanning = fields[2] == "injecting" && strings.Contains(line, "scanning for game")
				switch fields[2] {
				case "injected", "waiting":
					h.hooked = true
				case "injecting", "idle", "failed":
					h.hooked = false
					if fields[2] == "failed" {
						h.failure = line
					}
				}
				h.mu.Unlock()
				if fields[2] == "injecting" {
					select {
					case h.ready <- nil:
					default:
					}
				}
				if fields[2] == "failed" {
					select {
					case h.ready <- fmt.Errorf("%s", line):
					default:
					}
				}
			}
			if mode == "" && strings.HasPrefix(line, "ok ") && strings.HasSuffix(line, "config prefix set") {
				select {
				case h.ready <- nil:
				default:
				}
			}
			if strings.HasPrefix(line, "error ") {
				select {
				case h.ready <- fmt.Errorf("%s", line):
				default:
				}
			}
		}
		if err := scanner.Err(); err != nil {
			display.Log("LTK output: " + err.Error())
		}
	}
	go read(stdout)
	go read(stderr)
	go func() {
		readers.Wait()
		err := cmd.Wait()
		h.mu.Lock()
		h.running = false
		h.hooked = false
		h.mu.Unlock()
		display.Log(fmt.Sprintf("LTK host exited: %v", err))
		close(h.done)
	}()
	_, err = fmt.Fprintf(stdin, "config loglevel 16\nconfig flags 0\nconfig prefix %s\\\n", strings.TrimRight(prefix, `/\`))
	if err == nil && mode != "" {
		_, err = fmt.Fprintln(stdin, "start "+mode)
	}
	if err == nil {
		select {
		case err = <-h.ready:
		case <-h.done:
			err = fmt.Errorf("LTK host exited during startup")
		case <-time.After(5 * time.Second):
			err = fmt.Errorf("LTK host startup timeout")
		}
	}
	if err != nil {
		stopLocked()
	} else if mode == "scan" {
		display.Log(fmt.Sprintf("LTK: scan session established pid=%d; preserve this session through preparation", h.cmd.Process.Pid))
	}
	return err
}
