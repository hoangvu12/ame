// Package ltk owns the bundled runtime, the skin package builder and the
// LTK Manager session that patches the game.
//
// The manager runs as a hidden background helper (see cdp.go) and is driven
// through the same Tauri IPC commands its own UI uses: ame installs the
// built skin package into its library, enables it, and starts the patcher.
// The manager builds its own overlay from the installed package, so ame
// builds only the package — no extracted overlay, no WAD work.
package ltk

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/json"
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

// ameModName is the library name every skin package ame installs carries.
// ame uninstalls exactly the mods holding this name before installing a new
// one, so leftovers from a crashed run cannot stack in the manager's library.
const ameModName = "ame skin set"

// gameDir is the game directory the last Build ran against; the manager's
// league path is configured from it. Guarded by managerMu.
var gameDir string

// Prepare installs immutable, content-addressed files; no download at runtime.
// The patcher DLL and host are excluded from hash validation and never
// overwritten: a runtime directory written by an older ame may hold the pair
// its delivery service installed (the bundled copies are the offline
// fallback). The manager runtime ame launches is delivered separately, into
// the shared manager cache (see delivery.go).
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
		// Legacy delivered pairs may sit beside the bundle; skip their hash.
		if strings.EqualFold(name, "ltk_patcher_dll.dll") || strings.EqualFold(name, "ltk_patcher_host.exe") {
			if _, err := os.Stat(filepath.Join(dest, filepath.FromSlash(name))); err != nil {
				valid = false
				break
			}
			continue
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
		// Legacy delivered pairs may sit beside the bundle; never overwrite them.
		if strings.EqualFold(entry.Name, "ltk_patcher_dll.dll") || strings.EqualFold(entry.Name, "ltk_patcher_host.exe") {
			if _, err := os.Stat(target); err == nil {
				continue
			}
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

// mergeFantomes combines the per-mod archives Build packed into the single
// package ame installs into the manager's library. Entries of earlier mods
// win — the same priority the overlay builder gives them — and the sources'
// metadata is replaced by ame's own, so the package carries a stable library
// name that identifies it on later runs.
func mergeFantomes(sources []string, dest string) error {
	file, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer file.Close()
	archive := zip.NewWriter(file)
	written := make(map[string]bool)
	for _, source := range sources {
		input, err := zip.OpenReader(source)
		if err != nil {
			return err
		}
		for _, entry := range input.File {
			if entry.FileInfo().IsDir() {
				continue
			}
			name := strings.ReplaceAll(entry.Name, `\`, "/")
			if strings.HasPrefix(strings.ToLower(name), "meta/") {
				continue
			}
			if written[name] {
				continue
			}
			if !entry.FileInfo().Mode().IsRegular() {
				return fmt.Errorf("unsupported mod entry: %s", name)
			}
			reader, err := entry.Open()
			if err != nil {
				input.Close()
				return err
			}
			header := &zip.FileHeader{Name: name, Method: zip.Store, Modified: entry.Modified}
			writer, err := archive.CreateHeader(header)
			if err != nil {
				reader.Close()
				input.Close()
				return err
			}
			_, err = io.Copy(writer, reader)
			reader.Close()
			if err != nil {
				input.Close()
				return err
			}
			written[name] = true
		}
		input.Close()
	}
	info, err := archive.CreateHeader(&zip.FileHeader{Name: "META/info.json", Method: zip.Store})
	if err != nil {
		return err
	}
	metadata := fmt.Sprintf(`{"Name":%q,"Author":"ame","Version":"1.0.0","Description":"Skins selected in ame"}`+"\n", ameModName)
	if _, err = io.WriteString(info, metadata); err != nil {
		return err
	}
	if err = archive.Close(); err != nil {
		return err
	}
	return file.Close()
}

// Build packs the selected mods and merges them into the skin package
// Start installs into the LTK Manager's library. The manager builds its own
// overlay from the package when it is installed, so ame performs no WAD work
// here: the previous ame-side overlay build (~12s per apply) produced an
// extracted overlay nothing consumed. The package is staged under a
// temporary name and published atomically, so a failed build leaves the
// previous complete package available to the armed runtime.
func Build(modsDir, overlay, game, names string) error {
	started := time.Now()
	managerMu.Lock()
	gameDir = game
	managerMu.Unlock()
	if _, err := Prepare(); err != nil {
		return err
	}
	scratch, err := os.MkdirTemp("", "ame-ltk-build-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(scratch)
	if err := os.MkdirAll(filepath.Dir(overlay), 0755); err != nil {
		return err
	}
	staged := overlay + ".fantome.new"
	var packages []string
	for i, name := range strings.Split(names, "/") {
		if name == "" || name == "." || name == ".." || strings.ContainsAny(name, `\:`) {
			return fmt.Errorf("invalid mod name")
		}
		path := filepath.Join(scratch, fmt.Sprintf("%d.fantome", i))
		if err = packMod(filepath.Join(modsDir, name), path); err != nil {
			return err
		}
		packages = append(packages, path)
	}
	if err := mergeFantomes(packages, staged); err != nil {
		os.Remove(staged)
		return fmt.Errorf("LTK skin package: %w", err)
	}
	if err := publishPackage(staged, overlay+".fantome"); err != nil {
		return err
	}
	display.Log(fmt.Sprintf("LTK build: merged %d mod(s) into the skin package in %s", len(packages), time.Since(started).Round(time.Millisecond)))
	return nil
}

// --- manager session ---

// managerSession is one LTK Manager process driven over the DevTools
// protocol. The manager stays resident across skin switches.
type managerSession struct {
	cmd     *exec.Cmd
	cdp     *cdpClient
	modID   string            // library id of the installed skin package
	modFile string            // the fantome file installed as modID
	modHash [sha256.Size]byte // package bytes successfully enabled as modID
	stop    chan struct{}
	exited  chan struct{}
}

// alive reports whether the session has neither been torn down nor died.
func (m *managerSession) alive() bool {
	select {
	case <-m.exited:
		return false
	default:
	}
	select {
	case <-m.stop:
		return false
	default:
	}
	return true
}

// managerMu guards the manager session, its installed mod, and gameDir.
var (
	managerMu sync.Mutex
	manager   *managerSession
)

// managerSnapshot is the polled view of the manager and its patcher. It is
// written by pollManager and read by the status accessors without taking
// managerMu, so status queries never block on IPC round-trips.
type managerSnapshot struct {
	alive    bool
	running  bool   // the manager's patcher session runs
	phase    string // idle | building | patching
	attached bool   // the injected DLL is in a game
	pid      uint64
	failure  string
}

var (
	snapshotMu sync.Mutex
	snapshot   managerSnapshot
)

func currentSnapshot() managerSnapshot {
	snapshotMu.Lock()
	defer snapshotMu.Unlock()
	return snapshot
}

func setSnapshot(update func(*managerSnapshot)) {
	snapshotMu.Lock()
	defer snapshotMu.Unlock()
	update(&snapshot)
}

func Running() bool {
	s := currentSnapshot()
	return s.alive && s.running
}

func Hooked() bool {
	s := currentSnapshot()
	return s.alive && s.running && s.attached && s.failure == ""
}

// WaitingForGame reports whether the patcher is scanning without a game
// attached, so the armed skin package can still be replaced.
func WaitingForGame() bool {
	s := currentSnapshot()
	return s.alive && s.running && !s.attached && s.failure == ""
}

func Failure() string { return currentSnapshot().failure }

// WaitForHook waits until the manager's injected DLL attaches to a game. A
// pid of 0 or a game the event did not identify accepts any game.
func WaitForHook(pid uint32, cancel <-chan struct{}, timeout time.Duration) error {
	if manager == nil && !currentSnapshot().alive {
		return fmt.Errorf("LTK manager is not running")
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()
	for {
		s := currentSnapshot()
		if s.failure != "" {
			return fmt.Errorf("LTK session failed: %s", s.failure)
		}
		if !s.alive {
			return fmt.Errorf("LTK manager stopped before the patcher attached")
		}
		if s.running && s.attached && (pid == 0 || s.pid == 0 || s.pid == uint64(pid)) {
			return nil
		}
		select {
		case <-cancel:
			return fmt.Errorf("LTK hook wait cancelled")
		case <-timer.C:
			return fmt.Errorf("LTK patcher did not attach to PID %d before the deadline", pid)
		case <-tick.C:
		}
	}
}

// armSuspendTimeout bounds how long ArmSuspended holds a suspended game.
var armSuspendTimeout = 25 * time.Second

// ArmSuspended blocks until the manager's patcher has built its overlay and
// is scanning for a game. The held game is released only once the host is
// armed, so its window cannot appear before the hook is installed.
func ArmSuspended(pid uint32) error {
	if manager == nil && !currentSnapshot().alive {
		return fmt.Errorf("LTK manager is not running")
	}
	start := time.Now()
	timer := time.NewTimer(armSuspendTimeout)
	defer timer.Stop()
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()
	for {
		s := currentSnapshot()
		if s.failure != "" {
			return fmt.Errorf("LTK session failed: %s", s.failure)
		}
		if !s.alive {
			return fmt.Errorf("LTK manager stopped before the patcher was ready")
		}
		if s.phase == "patching" {
			return nil
		}
		// The poll can lag a start by a few hundred milliseconds, so an idle
		// phase is only fatal once the session had time to appear.
		if !s.running && s.phase != "building" && time.Since(start) > 2*time.Second {
			return fmt.Errorf("LTK patcher stopped before the game was released")
		}
		select {
		case <-timer.C:
			return fmt.Errorf("LTK patcher did not finish building before the game was released")
		case <-tick.C:
		}
	}
}

// Start installs the skin package built for overlay into the LTK Manager's
// library and starts its patcher. The manager process stays resident for
// later skin switches.
func Start(overlay string) error {
	prefix, err := filepath.Abs(overlay)
	if err != nil {
		return err
	}
	pkg := prefix + ".fantome"
	digest, err := packageDigest(pkg)
	if err != nil {
		return fmt.Errorf("LTK skin package for %s (build the overlay first): %w", prefix, err)
	}
	managerMu.Lock()
	m := manager
	same := m != nil && m.alive() && m.matchesPackage(pkg, digest)
	managerMu.Unlock()
	if same {
		s := currentSnapshot()
		if s.failure != "" {
			return fmt.Errorf("LTK session failed: %s", s.failure)
		}
		if s.running {
			display.Log("LTK: preserving manager session for the armed skin package")
			return nil
		}
	}
	if _, err := os.Stat(pkg); err != nil {
		return fmt.Errorf("LTK skin package for %s is missing (build the overlay first): %w", prefix, err)
	}
	if _, err := Prepare(); err != nil {
		return err
	}
	return startSession(pkg)
}

// startSession arms pkg as the manager's skin package.
//
// When a session is already live the mod is swapped in place: the manager
// rebuilds its overlay in the background and the host process — whose
// session start the injected DLL compares against the game's creation —
// keeps running, so a game launched later in the session still passes the
// DLL's join check. Restarting the patcher here would reset that timeline
// and make every in-flight game join too late.
func startSession(pkg string) error {
	digest, err := packageDigest(pkg)
	if err != nil {
		return err
	}
	managerMu.Lock()
	defer managerMu.Unlock()
	m, err := ensureManagerLocked()
	if err != nil {
		return err
	}
	if status, err := m.cdp.patcherStatus(); err == nil && status.Running {
		if err := swapAmeMod(m, pkg); err != nil {
			return err
		}
		m.cdp.clearBridgeFailure()
		setSnapshot(func(s *managerSnapshot) { s.failure = "" })
		return awaitOverlayRebuild(m)
	}
	if err := uninstallAmeMods(m); err != nil {
		return err
	}
	modID, err := m.cdp.installMod(pkg)
	if err != nil {
		return fmt.Errorf("LTK install: %w", err)
	}
	if err := m.cdp.toggleMod(modID, true); err != nil {
		return fmt.Errorf("LTK enable: %w", err)
	}
	m.cdp.clearBridgeFailure()
	setSnapshot(func(s *managerSnapshot) { s.failure = "" })
	if err := m.cdp.startPatcher(); err != nil {
		return fmt.Errorf("LTK patcher start: %w", err)
	}
	m.modID, m.modFile, m.modHash = modID, pkg, digest
	return awaitPatcherStart(m)
}

// swapAmeMod installs pkg as the library's only ame mod while the patcher
// runs. The rebuild the edit triggers is deferred behind a running game and
// applied when it ends.
func swapAmeMod(m *managerSession, pkg string) error {
	digest, err := packageDigest(pkg)
	if err != nil {
		return err
	}
	if m.matchesPackage(pkg, digest) {
		return nil
	}
	if err := uninstallAmeMods(m); err != nil {
		return err
	}
	modID, err := m.cdp.installMod(pkg)
	if err != nil {
		return fmt.Errorf("LTK install: %w", err)
	}
	if err := m.cdp.toggleMod(modID, true); err != nil {
		return fmt.Errorf("LTK enable: %w", err)
	}
	m.modID, m.modFile, m.modHash = modID, pkg, digest
	return nil
}

// packageDigest identifies the published bytes, not Build's reusable pathname.
func packageDigest(pkg string) ([sha256.Size]byte, error) {
	data, err := os.ReadFile(pkg)
	if err != nil {
		return [sha256.Size]byte{}, fmt.Errorf("LTK package read: %w", err)
	}
	return sha256.Sum256(data), nil
}

func (m *managerSession) matchesPackage(pkg string, digest [sha256.Size]byte) bool {
	return m.modID != "" && m.modFile == pkg && m.modHash == digest
}

// awaitOverlayRebuild waits until the background rebuild a mid-session mod
// swap triggers has finished: the session thread picks the edit up by dipping
// the phase to "building" and returns to scanning ("patching") when the new
// overlay is in place. The manager picks a library edit up within about a
// second, but a rebuild fast enough to hide between polls — or an edit that
// reuses every WAD — leaves the phase at "patching" throughout; the
// no-building grace therefore stays short, so bursts of swaps (room party:
// every teammate pick) do not pay a stall each. A rebuild deferred behind a
// running game also leaves the phase at "patching" throughout.
func awaitOverlayRebuild(m *managerSession) error {
	buildingDeadline := time.Now().Add(3 * time.Second)
	deadline := time.Now().Add(30 * time.Second)
	seenBuilding := false
	for {
		status, err := m.cdp.patcherStatus()
		if err == nil {
			switch status.Phase {
			case "building":
				seenBuilding = true
			case "patching":
				if seenBuilding || time.Now().After(buildingDeadline) {
					return nil
				}
			case "idle":
				if failure := currentSnapshot().failure; failure != "" {
					return fmt.Errorf("LTK patcher start: %s", failure)
				}
				return fmt.Errorf("LTK patcher stopped while rebuilding the overlay")
			}
		} else {
			display.Log("LTK: patcher status: " + err.Error())
		}
		if failure := currentSnapshot().failure; failure != "" {
			return fmt.Errorf("LTK session failed: %s", failure)
		}
		if time.Now().After(deadline) {
			display.Log("LTK: overlay rebuild still running; the skin applies when it finishes")
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// uninstallAmeMods removes every skin package ame owns from the manager's
// library: the one this session installed and any leftovers from earlier
// runs, which outlive the manager process.
func uninstallAmeMods(m *managerSession) error {
	mods, err := m.cdp.installedMods()
	if err != nil {
		return fmt.Errorf("LTK library read: %w", err)
	}
	for _, mod := range mods {
		if !isAmeMod(mod, m.modID) {
			continue
		}
		if err := m.cdp.uninstallMod(mod.ID); err != nil {
			display.Log(fmt.Sprintf("LTK: uninstall of stale skin package %s failed: %v", mod.ID, err))
		}
	}
	m.modID = ""
	return nil
}

// awaitPatcherStart waits until the manager's session thread reports a phase
// beyond idle; the overlay build itself continues in the background. The
// session state is published to the snapshot directly so callers observe a
// running session the moment Start returns, without waiting for the poll.
func awaitPatcherStart(m *managerSession) error {
	deadline := time.Now().Add(20 * time.Second)
	for {
		status, err := m.cdp.patcherStatus()
		if err == nil {
			if status.Phase != "" && status.Phase != "idle" {
				setSnapshot(func(s *managerSnapshot) {
					s.alive = true
					s.running = status.Running
					s.phase = status.Phase
				})
				return nil
			}
		} else {
			display.Log("LTK: patcher status: " + err.Error())
		}
		if failure := currentSnapshot().failure; failure != "" {
			return fmt.Errorf("LTK patcher start: %s", failure)
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("LTK patcher did not enter its build phase")
		}
		time.Sleep(200 * time.Millisecond)
	}
}

// Stop tears the manager session down: the patcher is stopped, then the
// manager process is terminated. The installed mod stays in the library; the
// next run uninstalls it before installing a new package.
func Stop() {
	managerMu.Lock()
	defer managerMu.Unlock()
	stopLocked()
}

func stopLocked() {
	m := manager
	if m == nil {
		return
	}
	manager = nil
	close(m.stop)
	// Stop the patcher first so the injected host detaches cleanly, then
	// terminate the manager and its WebView2 children.
	if err := m.cdp.stopPatcher(); err != nil {
		display.Log("LTK: patcher stop during shutdown: " + err.Error())
	}
	killManagerTree(m.cmd)
	m.cdp.close()
	setSnapshot(func(s *managerSnapshot) { *s = managerSnapshot{} })
}

// killManagerTree terminates the manager process and its WebView2 children.
func killManagerTree(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	kill := exec.Command("taskkill", "/PID", strconv.Itoa(cmd.Process.Pid), "/T", "/F")
	hide(kill)
	if err := kill.Run(); err != nil {
		_ = cmd.Process.Kill()
	}
}

// ensureManagerLocked returns a live manager session, launching the manager
// when needed. The caller holds managerMu.
func ensureManagerLocked() (*managerSession, error) {
	if m := manager; m != nil && m.alive() {
		if _, err := m.cdp.patcherStatus(); err == nil {
			return m, nil
		}
		display.Log("LTK: manager session lost its DevTools connection; relaunching")
	}
	stopLocked()
	// Terminate any running manager before touching its files: the
	// single-instance plugin would otherwise focus the old window, and the
	// delivered pair cannot be replaced while it runs.
	killExistingManagers()
	// The manager runtime is delivered to the shared cache, one download for
	// the manager and its bundled injector. The manager delivers the patcher
	// DLL itself.
	exe, err := managerFetch()
	if err != nil {
		return nil, fmt.Errorf("LTK manager runtime: %w", err)
	}
	if managerPortInUse() {
		if err := waitForPortFree(3 * time.Second); err != nil {
			return nil, fmt.Errorf("LTK cannot use DevTools port %d: %w", cdpPort, err)
		}
	}
	if err := configureManagerSettings(gameDir); err != nil {
		display.Log("LTK: manager settings: " + err.Error())
	}
	cmd, err := launchManager(exe)
	if err != nil {
		return nil, fmt.Errorf("LTK manager launch: %w", err)
	}
	page, err := cdpFindPage(cdpHandshakeTimeout)
	if err != nil {
		killLaunchedManager(cmd)
		return nil, fmt.Errorf("LTK manager WebView: %w", err)
	}
	c, err := cdpDial(page.WS)
	if err != nil {
		killLaunchedManager(cmd)
		return nil, fmt.Errorf("LTK manager connect: %w", err)
	}
	if err := c.waitReady(cdpReadyTimeout); err != nil {
		c.close()
		killLaunchedManager(cmd)
		return nil, fmt.Errorf("LTK manager page: %w", err)
	}
	result, err := c.injectShim()
	if err != nil || (result != "installed" && result != "already") {
		c.close()
		killLaunchedManager(cmd)
		if err != nil {
			return nil, fmt.Errorf("LTK manager bridge: %w", err)
		}
		return nil, fmt.Errorf("LTK manager bridge: %s", result)
	}
	m := &managerSession{
		cmd:    cmd,
		cdp:    c,
		stop:   make(chan struct{}),
		exited: make(chan struct{}),
	}
	manager = m
	go func() {
		_ = cmd.Wait()
		close(m.exited)
	}()
	go pollManager(m)
	display.Log(fmt.Sprintf("LTK: manager ready pid=%d page=%s", cmd.Process.Pid, page.URL))
	return m, nil
}

func killLaunchedManager(cmd *exec.Cmd) {
	killManagerTree(cmd)
	go func() { _ = cmd.Wait() }()
}

// overlayFailureMessage translates the DLL's verdict about the overlay of
// an attached game into a failure. "Live" means the overlay went out;
// "None" means the DLL said nothing, which also happens mid-attach.
func overlayFailureMessage(outcome string) string {
	switch outcome {
	case "TooLate":
		return "the game started before the patcher session; start a fresh match after selecting the skin"
	case "EndOfLife":
		return "the patcher DLL reached end of life; update LTK Manager's patcher"
	case "Disabled":
		return "the overlay failed the DLL's verification and was disabled"
	case "HookFailed":
		return "the DLL could not install its overlay hook"
	}
	return ""
}

// pollManager keeps the session snapshot current: phase and running state
// from the manager's status command, attachment and failures from the event
// bridge. The loop ends when the session is torn down, the manager dies, or
// the DevTools connection breaks.
func pollManager(m *managerSession) {
	tick := time.NewTicker(300 * time.Millisecond)
	defer tick.Stop()
	failures := 0
	for {
		select {
		case <-m.stop:
			return
		case <-m.exited:
			setSnapshot(func(s *managerSnapshot) { s.alive = false })
			display.Log("LTK: manager process exited")
			return
		case <-tick.C:
		}
		state, err := m.cdp.bridgeState()
		if err != nil {
			failures++
			if failures >= 5 {
				setSnapshot(func(s *managerSnapshot) { s.alive = false })
				display.Log("LTK: manager connection lost: " + err.Error())
				return
			}
			continue
		}
		failures = 0
		if !state.Shim {
			// The page reloaded and dropped the bridge; re-install it.
			if _, err := m.cdp.injectShim(); err != nil {
				display.Log("LTK: bridge re-injection: " + err.Error())
			}
		}
		failure := state.Failure
		if failure != "" {
			failure = errorText(json.RawMessage(failure))
		}
		if overlay := overlayFailureMessage(state.Overlay); overlay != "" && state.Attached {
			// The game is attached but the DLL stayed inert: the skin the user
			// armed for this match did not apply, and that must be visible.
			failure = overlay
		}
		setSnapshot(func(s *managerSnapshot) {
			*s = managerSnapshot{
				alive:    true,
				attached: state.Attached,
				failure:  failure,
			}
			if state.Status != nil {
				s.running = state.Status.Running
				s.phase = state.Status.Phase
			}
			if state.PID != nil {
				s.pid = *state.PID
			}
		})
	}
}
