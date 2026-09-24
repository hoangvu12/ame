package celestial

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type HostConfig struct {
	Executable     string
	DLL            string
	Overlay        string
	LogFile        string
	HookTimeoutMS  uint32
	Flags          uint32
	StartupTimeout time.Duration
}

type HostStatus struct {
	Running     bool
	Initialized bool
	Hooked      bool
	ThreadID    uint32
	Phase       string
}

// Host owns one child process. No process-name-wide termination is used.
type Host struct {
	mu     sync.Mutex
	status HostStatus
	cancel context.CancelFunc
	done   chan struct{}
	err    error
}

// StartHost requires the opaque 140-byte authorization record supplied by an
// external authorized producer. It does not sign, forge, persist, or log that
// record. A correct length is not proof of authorization: INIT_OK is still
// required, and game-hook success remains a separate state.
func StartHost(ctx context.Context, config HostConfig, session []byte) (*Host, error) {
	return startHost(ctx, config, session, exec.CommandContext)
}

func startHost(ctx context.Context, config HostConfig, session []byte, command func(context.Context, string, ...string) *exec.Cmd) (*Host, error) {
	if len(session) != 140 {
		return nil, fmt.Errorf("Celestial requires a 140-byte record from a session authorization provider")
	}
	if !filepath.IsAbs(config.Executable) || !filepath.IsAbs(config.DLL) || !filepath.IsAbs(config.Overlay) {
		return nil, fmt.Errorf("host, DLL, and overlay paths must be absolute")
	}
	if config.LogFile != "" && !filepath.IsAbs(config.LogFile) {
		return nil, fmt.Errorf("log-file path must be absolute")
	}
	if config.StartupTimeout <= 0 {
		config.StartupTimeout = 10 * time.Second
	}
	childCtx, cancel := context.WithCancel(ctx)
	h := &Host{cancel: cancel, done: make(chan struct{}), status: HostStatus{Phase: "starting"}}
	args := []string{config.DLL, config.Overlay, strconv.FormatUint(uint64(config.HookTimeoutMS), 10), strconv.FormatUint(uint64(config.Flags), 10)}
	if config.LogFile != "" {
		args = append(args, config.LogFile)
	}
	cmd := command(childCtx, config.Executable, args...)
	cmd.Dir = filepath.Dir(config.Executable)
	configureHostProcess(cmd)
	// Copy the record so subsequent caller mutation cannot alter the packet.
	packet := append([]byte(nil), session...)
	cmd.Stdin = bytes.NewReader(packet)
	reader, writer := io.Pipe()
	cmd.Stdout = writer
	// The host reports protocol errors on stdout. A bounded stderr tail is
	// retained only for failures before that protocol is available.
	var stderr hostErrorTail
	cmd.Stderr = &stderr
	cmd.WaitDelay = 2 * time.Second
	if err := cmd.Start(); err != nil {
		cancel()
		reader.Close()
		writer.Close()
		return nil, err
	}
	h.mu.Lock()
	h.status.Running = true
	h.mu.Unlock()
	ready := make(chan error, 1)
	var readyOnce sync.Once
	announce := func(err error) { readyOnce.Do(func() { ready <- err }) }
	outputDone := make(chan error, 1)
	go func() {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "EVT INIT_ERR ") {
				announce(fmt.Errorf("host initialization: %s", strings.TrimPrefix(line, "EVT INIT_ERR ")))
			}
			h.consume(line)
			if line == "EVT INIT_OK" {
				announce(nil)
			}
		}
		readErr := scanner.Err()
		reader.Close()
		if readErr != nil {
			cancel()
		}
		outputDone <- readErr
	}()
	go func() {
		defer cancel()
		err := cmd.Wait()
		writer.Close()
		readErr := <-outputDone
		for i := range packet {
			packet[i] = 0
		}
		if err == nil {
			err = readErr
		}
		if err != nil && len(stderr.data) > 0 {
			err = fmt.Errorf("%w: %s", err, strings.TrimSpace(string(stderr.data)))
		}
		h.mu.Lock()
		h.err = err
		h.status.Running = false
		h.status.Hooked = false
		h.status.ThreadID = 0
		h.status.Phase = "stopped"
		h.mu.Unlock()
		if err == nil {
			announce(fmt.Errorf("host exited before initialization"))
		} else {
			announce(fmt.Errorf("host exited before initialization: %w", err))
		}
		close(h.done)
	}()
	timer := time.NewTimer(config.StartupTimeout)
	defer timer.Stop()
	select {
	case err := <-ready:
		if err != nil {
			h.Stop()
			return nil, err
		}
		return h, nil
	case <-ctx.Done():
		h.Stop()
		return nil, ctx.Err()
	case <-timer.C:
		h.Stop()
		return nil, fmt.Errorf("host initialization timed out")
	}
}

func (h *Host) consume(line string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	switch line {
	case "EVT INIT_OK":
		h.status.Initialized = true
		h.status.Phase = "initialized"
		return
	case "EVT WAITING_FOR_GAME":
		h.status.Hooked = false
		h.status.ThreadID = 0
		h.status.Phase = "waiting"
		return
	}
	fields := strings.Fields(line)
	if len(fields) < 3 || fields[0] != "EVT" || !strings.HasPrefix(fields[2], "tid=") {
		return
	}
	tid, err := strconv.ParseUint(strings.TrimPrefix(fields[2], "tid="), 10, 32)
	if err != nil || tid == 0 {
		return
	}
	switch fields[1] {
	case "GAME_FOUND":
		h.status.ThreadID = uint32(tid)
		h.status.Hooked = false
		h.status.Phase = "found"
	case "HOOK_APPLIED":
		if h.status.Initialized && h.status.ThreadID == uint32(tid) {
			h.status.Hooked = true
			h.status.Phase = "hooked"
		}
	case "GAME_EXIT":
		if h.status.ThreadID == uint32(tid) {
			h.status.Hooked = false
			h.status.ThreadID = 0
			h.status.Phase = "waiting"
		}
	}
}

func (h *Host) Status() HostStatus { h.mu.Lock(); defer h.mu.Unlock(); return h.status }

func (h *Host) Wait() error { <-h.done; h.mu.Lock(); defer h.mu.Unlock(); return h.err }

// Stop terminates this host only and waits until it has been reaped.
func (h *Host) Stop() { h.cancel(); <-h.done }

type hostErrorTail struct{ data []byte }

func (b *hostErrorTail) Write(p []byte) (int, error) {
	n := len(p)
	if len(p) > 4096 {
		p = p[len(p)-4096:]
	}
	b.data = append(b.data, p...)
	if len(b.data) > 4096 {
		b.data = b.data[len(b.data)-4096:]
	}
	return n, nil
}
