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
	"errors"
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
	"github.com/hoangvu12/ame/internal/game"
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

// findGameDir locates the install when no Build has run yet, so the first
// manager launch is pointed at the right League. A variable for tests.
var findGameDir = game.FindGameDir

// ErrGameRunning refuses a library edit while a League game process exists.
// Every edit stops the manager's scan and starts a new one after the
// rebuild, and the injected DLL refuses a game created before that scan
// ("joined too late"), so an edit now can only take the skins away from this
// match. The armed set stays for it.
var ErrGameRunning = errors.New("League is already running; the skin set cannot change until the match ends")

// GameRunning reports whether a League game process exists.
func GameRunning() bool { return gameProcessRunning() }

// ErrRebuildPending: the manager holds the new skin set, but its rebuild
// outlasted the wait (a slow disk rewriting map WADs). It is armed for the
// match only if the rebuild ends before League starts.
var ErrRebuildPending = errors.New("the manager is still rebuilding the skin set")

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
// name that identifies it on later runs. The extra entries (the map
// keep-alive files) follow the mods; one whose WAD a mod already ships is
// left out, since the builder reads a WAD that has directory entries from
// those alone and would ignore a packed WAD of the same name.
func mergeFantomes(sources []string, dest string, extra []packageEntry) error {
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
	for _, entry := range extra {
		if shipsWAD(written, entry.name) {
			continue
		}
		writer, err := archive.CreateHeader(&zip.FileHeader{Name: entry.name, Method: zip.Store})
		if err != nil {
			return err
		}
		if _, err = writer.Write(entry.data); err != nil {
			return err
		}
		written[entry.name] = true
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

// shipsWAD reports whether the written entries already hold the WAD the
// directory-layout entry name (WAD/<wad>/<file>) belongs to, packed or as
// directory entries.
func shipsWAD(written map[string]bool, name string) bool {
	parts := strings.SplitN(name, "/", 3)
	if len(parts) < 3 {
		return false
	}
	wad := strings.ToLower(parts[0] + "/" + parts[1])
	for existing := range written {
		lower := strings.ToLower(existing)
		if lower == wad || strings.HasPrefix(lower, wad+"/") {
			return true
		}
	}
	return false
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
	keepAlive := keepAliveEntries(game)
	if err := mergeFantomes(packages, staged, keepAlive); err != nil {
		os.Remove(staged)
		return fmt.Errorf("LTK skin package: %w", err)
	}
	if err := publishPackage(staged, overlay+".fantome"); err != nil {
		return err
	}
	display.Log(fmt.Sprintf("LTK build: merged %d mod(s) into the skin package in %s (%d map keep-alive entries)", len(packages), time.Since(started).Round(time.Millisecond), len(keepAlive)))
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
	verdict  string    // the DLL's word on the attached game (overlayVerdict)
	deferred bool      // a library edit waits for the running game to end
	prefix   string    // the session's overlay root
	pending  bool      // library edit has not yet passed the rebuild fence
	observed time.Time // last successful authoritative status read
}

var (
	snapshotMu sync.Mutex
	snapshot   managerSnapshot
	// Guarded by snapshotMu: what ame's skin package is doing in the manager.
	rebuild     *rebuildTracker // edit still on its way into the scan
	armedAt     time.Time       // the scan last restarted with ame's package
	gameEndedAt time.Time       // the last League game process ended
)

// beginRebuild fences readiness from the moment ame edits the library (or
// starts the patcher) until the scan runs with the edit.
func beginRebuild(at time.Time, phase string) {
	ready := mapsReadyNow() // file checks, kept out of snapshotMu
	snapshotMu.Lock()
	defer snapshotMu.Unlock()
	snapshot.pending = true
	snapshot.observed = time.Time{}
	rebuild = newRebuildTracker(at, phase)
	rebuild.mapsReady = ready
}

// observeRebuild folds a status read into the pending edit and reports
// whether the scan now runs with it. The first caller to see completion
// clears the fence and records how long this PC took.
func observeRebuild(status patcherStatus, now time.Time) bool {
	snapshotMu.Lock()
	r := rebuild
	if r == nil {
		ready := status.Running && status.Phase == "patching" && !status.ChangesAfterGame
		if ready {
			snapshot.pending = false
		}
		snapshotMu.Unlock()
		return ready
	}
	if !status.Running || !r.observe(status.Phase, status.ChangesAfterGame, now) {
		snapshotMu.Unlock()
		return false
	}
	rebuild = nil
	snapshot.pending = false
	armedAt = r.armedAt
	snapshotMu.Unlock()
	if d := r.duration(); d > 0 {
		wroteMap := rewroteMapWAD(status.prefix(), r.editAt, r.armedAt)
		heavy := wroteMap && !r.mapsReady
		recordRebuild(d, heavy, r.mapsReady, now)
		kind := "champion WADs only"
		switch {
		case heavy:
			kind = "map WADs copied"
		case wroteMap:
			kind = "prepared map WADs updated in place"
		}
		display.Log(fmt.Sprintf("LTK: skin set reached the patcher scan %s after the edit (%s); rebuild budget now %s", d.Round(100*time.Millisecond), kind, RebuildBudget().Round(time.Second)))
	} else {
		display.Log("LTK: skin set reached the patcher scan")
	}
	return true
}

// Prepared is what the manager holds of ame's skin package for the match
// about to start.
type Prepared struct {
	// Armed: the scan runs with ame's latest package, and it started after
	// the last game ended, so it belongs to the next match.
	Armed bool
	// Pending: an edit is still on its way into the scan.
	Pending    bool
	PendingFor time.Duration
}

// PreparedState reports whether a skin set is armed or rebuilding for the
// coming match. A set left from a finished match is neither: it is what the
// previous champion select chose.
func PreparedState() Prepared {
	snapshotMu.Lock()
	defer snapshotMu.Unlock()
	s := snapshot
	if !s.alive || !s.running {
		return Prepared{}
	}
	if r := rebuild; r != nil {
		return Prepared{Pending: true, PendingFor: time.Since(r.editAt)}
	}
	armed := s.phase == "patching" && s.failure == "" && !armedAt.IsZero() && armedAt.After(gameEndedAt)
	return Prepared{Armed: armed}
}

// OnOverlayVerdict is told what the DLL said about each game it attached to.
var OnOverlayVerdict func(pid uint64, verdict string)

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

// Ready reports a recent scanning status, not merely a live patcher process.
func Ready() bool {
	s := currentSnapshot()
	return s.alive && s.running && s.phase == "patching" && !s.pending && s.failure == "" && !s.observed.IsZero() && time.Since(s.observed) < 2*time.Second
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
		if s.running && s.phase == "patching" && !s.pending && time.Since(s.observed) < 2*time.Second {
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
// A live session keeps its host process, but not its scan: every library
// edit stops the scan, rebuilds the overlay and starts a new scan session,
// and the injected DLL refuses a game created before that session ("joined
// too late"). So no edit may happen once League's process exists — it could
// only take the skins away from the match — and an edit made earlier helps
// only if its rebuild ends before League starts.
func startSession(pkg string) error {
	digest, err := packageDigest(pkg)
	if err != nil {
		return err
	}
	// A map preparation writing the copies goes first. One can only start
	// while managerMu is free, so the check is repeated under the lock.
	for {
		if err := waitForMapPrep(); err != nil {
			return err
		}
		managerMu.Lock()
		if !MapPrepRunning() {
			break
		}
		managerMu.Unlock()
	}
	defer managerMu.Unlock()
	// Checked before the manager is (re)launched too: a relaunch replaces
	// the host serving a running game.
	if gameProcessRunning() {
		return ErrGameRunning
	}
	m, err := ensureManagerLocked()
	if err != nil {
		return err
	}
	status, statusErr := m.cdp.patcherStatus()
	running := statusErr == nil && status.Running
	if !m.matchesPackage(pkg, digest) {
		// The last look before the write: from here the manager restarts its
		// scan whatever happens.
		if gameProcessRunning() {
			return ErrGameRunning
		}
		if err := replaceAmeMod(m, pkg, digest, running, status.Phase); err != nil {
			return err
		}
	}
	m.cdp.clearBridgeFailure()
	setSnapshot(func(s *managerSnapshot) { s.failure = "" })
	if !running {
		if gameProcessRunning() {
			return ErrGameRunning
		}
		beginRebuild(time.Now(), "idle")
		if err := m.cdp.startPatcher(); err != nil {
			return fmt.Errorf("LTK patcher start: %w", err)
		}
		if err := awaitPatcherStart(m); err != nil {
			return err
		}
	}
	return awaitOverlayRebuild(m)
}

// swapAmeMod makes pkg the library's ame skin package while the patcher
// runs, unless the manager already holds these exact bytes.
func swapAmeMod(m *managerSession, pkg string) error {
	digest, err := packageDigest(pkg)
	if err != nil {
		return err
	}
	if m.matchesPackage(pkg, digest) {
		return nil
	}
	return replaceAmeMod(m, pkg, digest, true, "patching")
}

// replaceAmeMod puts pkg into the manager's library as its one ame skin
// package. The new package is installed and enabled first, then every older
// ame package is uninstalled, all in one burst inside the manager's 750ms
// settle window, so a refresh in between never sees an empty set.
//
// update_mod is never used: manager v1.33.0 clears its overlay build-version
// marker on every update (mods/archive/install/update.rs ->
// invalidate_overlay_for), and the next build then deletes every cached
// overlay WAD of every profile and copies them all again, the prepared map
// copies included. A fresh install under a new library id keeps them.
func replaceAmeMod(m *managerSession, pkg string, digest [sha256.Size]byte, running bool, phase string) error {
	mods, err := m.cdp.installedMods()
	if err != nil {
		return fmt.Errorf("LTK library read: %w", err)
	}
	var previous []managerMod
	for _, mod := range mods {
		if isAmeMod(mod, m.modID) {
			previous = append(previous, mod)
		}
	}
	if running {
		beginRebuild(time.Now(), phase)
	}
	// Until the write succeeds the manager may hold either package.
	m.modID = ""
	modID, err := m.cdp.installMod(pkg)
	if err != nil {
		return fmt.Errorf("LTK install: %w", err)
	}
	if err := m.cdp.toggleMod(modID, true); err != nil {
		return fmt.Errorf("LTK enable: %w", err)
	}
	for _, mod := range previous {
		// The manager answers a package it already holds with that mod.
		if mod.ID == modID {
			continue
		}
		if err := m.cdp.uninstallMod(mod.ID); err != nil {
			display.Log(fmt.Sprintf("LTK: uninstall of stale skin package %s failed: %v", mod.ID, err))
		}
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

// awaitOverlayRebuild waits until the scan runs with ame's latest edit (see
// rebuildTracker). It is bounded: a slow disk can take well over a minute to
// rewrite the map WADs, and the caller must not sit on the preparation lock
// that long. A timeout leaves the fence up; pollManager keeps following the
// rebuild and lowers the fence when the scan restarts.
var overlayRebuildTimeout = 30 * time.Second

func publishStatus(status patcherStatus) {
	setSnapshot(func(s *managerSnapshot) {
		s.alive = true
		s.running = status.Running
		s.phase = status.Phase
		s.deferred = status.ChangesAfterGame
		s.prefix = status.prefix()
		s.observed = time.Now()
	})
}

func awaitOverlayRebuild(m *managerSession) error {
	deadline := time.Now().Add(overlayRebuildTimeout)
	for {
		status, err := m.cdp.patcherStatus()
		if err == nil {
			publishStatus(status)
			if status.Phase == "idle" || !status.Running {
				if failure := currentSnapshot().failure; failure != "" {
					return fmt.Errorf("LTK patcher start: %s", failure)
				}
				return fmt.Errorf("LTK patcher stopped while rebuilding the overlay")
			}
			if observeRebuild(status, time.Now()) {
				return nil
			}
		} else {
			display.Log("LTK: patcher status: " + err.Error())
		}
		if failure := currentSnapshot().failure; failure != "" {
			return fmt.Errorf("LTK session failed: %s", failure)
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("LTK overlay rebuild readiness deadline exceeded: %w", ErrRebuildPending)
		}
		time.Sleep(100 * time.Millisecond)
	}
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
					s.observed = time.Now()
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
	snapshotMu.Lock()
	snapshot = managerSnapshot{}
	rebuild, armedAt = nil, time.Time{}
	snapshotMu.Unlock()
}

// disarmStopTimeout bounds the wait for the patcher to report itself
// stopped after Disarm asked it to stop.
var disarmStopTimeout = 5 * time.Second

// Disarm stops the patcher, so no later game is hooked with the set it held,
// but keeps the manager: its process, DevTools connection and the installed
// package bookkeeping survive, so the next arm is one library edit and a
// patcher start, and a preparation can build the map copies in between.
// Never while League runs: the patcher serves that match. A patcher that
// will not stop gets the full Stop.
func Disarm() error {
	managerMu.Lock()
	defer managerMu.Unlock()
	if gameProcessRunning() {
		return ErrGameRunning
	}
	m := manager
	if m == nil || !m.alive() {
		return nil
	}
	if err := stopPatcherLocked(m); err != nil {
		display.Log("LTK: patcher stop failed, stopping the manager: " + err.Error())
		stopLocked()
		return err
	}
	snapshotMu.Lock()
	snapshot.running, snapshot.phase, snapshot.pending = false, "idle", false
	rebuild, armedAt = nil, time.Time{}
	snapshotMu.Unlock()
	display.Log("LTK: patcher stopped; the manager stays ready for the next skin set")
	scheduleIdleStop()
	return nil
}

// stopPatcherLocked stops the patcher and waits until the manager reports
// it stopped, so the next arm finds an idle patcher.
func stopPatcherLocked(m *managerSession) error {
	if err := m.cdp.stopPatcher(); err != nil {
		return err
	}
	deadline := time.Now().Add(disarmStopTimeout)
	for {
		status, err := m.cdp.patcherStatus()
		if err == nil && !status.Running {
			return nil
		}
		if time.Now().After(deadline) {
			if err != nil {
				return err
			}
			return fmt.Errorf("the patcher still runs %s after the stop", disarmStopTimeout)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// managerIdleTimeout: a manager with nothing armed and no preparation
// running is shut down after this long. The next use relaunches it.
var managerIdleTimeout = 30 * time.Minute

var (
	idleStopMu    sync.Mutex
	idleStopTimer *time.Timer
)

// scheduleIdleStop (re)starts the idle countdown.
func scheduleIdleStop() {
	idleStopMu.Lock()
	defer idleStopMu.Unlock()
	if idleStopTimer != nil {
		idleStopTimer.Stop()
	}
	idleStopTimer = time.AfterFunc(managerIdleTimeout, stopIfIdle)
}

// stopIfIdle shuts the manager down unless something uses it. An armed
// patcher restarts the countdown when it is disarmed.
func stopIfIdle() {
	if MapPrepRunning() {
		scheduleIdleStop()
		return
	}
	managerMu.Lock()
	defer managerMu.Unlock()
	m := manager
	if m == nil {
		return
	}
	if gameProcessRunning() {
		scheduleIdleStop()
		return
	}
	if status, err := m.cdp.patcherStatus(); err == nil && status.Running {
		return
	}
	display.Log(fmt.Sprintf("LTK: stopping the idle manager (nothing armed for %s)", managerIdleTimeout))
	stopLocked()
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
	// The first launch of a run precedes any Build, so gameDir may still be
	// empty: without a path the manager auto-detects, which can pick another
	// install than the one ame patches.
	league := gameDir
	if league == "" {
		league = findGameDir()
	}
	if err := configureManagerSettings(league); err != nil {
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
// an attached game into a failure. "live" means the overlay went out; "none"
// means the DLL said nothing, which also happens mid-attach.
func overlayFailureMessage(outcome string) string {
	switch overlayVerdict(outcome) {
	case "too-late":
		return "the game started before the patcher scan; this match runs without the new skins, the next match will have them"
	case "end-of-life":
		return "the patcher DLL does not support this game build yet; LTK Manager must deliver a newer patcher"
	case "disabled":
		return "the overlay failed the DLL's verification and was disabled"
	case "hook-failed":
		return "the DLL could not install its overlay hook"
	}
	return ""
}

// managerView is what one poll of the manager saw.
type managerView struct {
	running  bool
	phase    string
	attached bool
	pid      uint64
	verdict  string
	failure  string
	deferred bool
	game     bool // a League game process exists
}

// managerWatch remembers the previous poll, so the log records transitions
// with their timing rather than repeating states.
type managerWatch struct {
	seen       bool
	last       managerView
	phaseSince time.Time
	scanSince  time.Time // the patcher scan last (re)started
}

// note compares a poll with the previous one and describes what changed.
// pendingFor is how long ame's latest edit has been on its way into the
// scan, zero when none is.
func (w *managerWatch) note(v managerView, now time.Time, pendingFor time.Duration) []string {
	var lines []string
	prev := w.last
	if !w.seen {
		w.seen = true
		w.phaseSince = now
		if v.phase == "patching" {
			w.scanSince = now
		}
		lines = append(lines, fmt.Sprintf("LTK: patcher phase %s", phaseName(v.phase)))
		if v.game {
			lines = append(lines, "LTK: a League game was already running when ame connected to the manager")
		}
		prev = managerView{phase: v.phase, game: v.game}
	}
	if v.phase != prev.phase {
		line := fmt.Sprintf("LTK: patcher phase %s -> %s after %s", phaseName(prev.phase), phaseName(v.phase), now.Sub(w.phaseSince).Round(100*time.Millisecond))
		switch v.phase {
		case "building":
			line += " (scan stopped to build the overlay)"
		case "patching":
			line += " (scan started)"
			w.scanSince = now
		}
		w.phaseSince = now
		lines = append(lines, line)
	}
	if v.game && !prev.game {
		switch {
		case pendingFor > 0:
			lines = append(lines, fmt.Sprintf("LTK: League started while ame's skin set was still rebuilding (edit %s ago); the patcher will refuse this match as joined too late, so it runs without the new skins and the next match will have them", pendingFor.Round(100*time.Millisecond)))
		case v.phase == "patching" && !w.scanSince.IsZero():
			lines = append(lines, fmt.Sprintf("LTK: League started %s after the patcher scan began", now.Sub(w.scanSince).Round(100*time.Millisecond)))
		default:
			lines = append(lines, fmt.Sprintf("LTK: League started while the patcher was %s; this match runs without ame's skins", phaseName(v.phase)))
		}
	} else if !v.game && prev.game {
		lines = append(lines, "LTK: League game process ended")
	}
	if v.attached && !prev.attached {
		lines = append(lines, fmt.Sprintf("LTK: patcher DLL attached pid=%d", v.pid))
	}
	if v.verdict != "" && v.verdict != prev.verdict {
		lines = append(lines, verdictLine(v.pid, v.verdict))
	}
	if v.failure != "" && v.failure != prev.failure && overlayFailureMessage(v.verdict) != v.failure {
		lines = append(lines, "LTK: patcher failure: "+v.failure)
	}
	if v.deferred && !prev.deferred {
		lines = append(lines, "LTK: a library edit waits for the running game to end (changesAfterGame)")
	}
	w.last = v
	return lines
}

func phaseName(phase string) string {
	if phase == "" {
		return "unknown"
	}
	return phase
}

// verdictLine is the log line for the DLL's verdict on an attached game.
// Only "live" confirms that skins reached a match.
func verdictLine(pid uint64, verdict string) string {
	switch verdict {
	case "live":
		return fmt.Sprintf("LTK: overlay verdict pid=%d: live, the patcher DLL confirmed the skins in this match", pid)
	case "too-late":
		return fmt.Sprintf("LTK: overlay verdict pid=%d: too-late, League started before the patcher scan; this match runs without the new skins, the next match will have them", pid)
	}
	if message := overlayFailureMessage(verdict); message != "" {
		return fmt.Sprintf("LTK: overlay verdict pid=%d: %s, %s", pid, verdict, message)
	}
	return fmt.Sprintf("LTK: overlay verdict pid=%d: %s", pid, verdict)
}

// pollManager keeps the session snapshot current: phase and running state
// from the manager's status command, attachment and failures from the event
// bridge. It also follows ame's pending edit into the scan, notices League
// starting and ending, and logs every transition. The loop ends when the
// session is torn down, the manager dies, or the DevTools connection breaks.
func pollManager(m *managerSession) {
	tick := time.NewTicker(300 * time.Millisecond)
	defer tick.Stop()
	failures := 0
	var watch managerWatch
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
		now := time.Now()
		view := managerView{attached: state.Attached, game: gameProcessRunning()}
		if state.Attached {
			view.verdict = overlayVerdict(state.Overlay)
		}
		if state.PID != nil {
			view.pid = *state.PID
		}
		view.failure = state.Failure
		if view.failure != "" {
			view.failure = errorText(json.RawMessage(view.failure))
		}
		if overlay := overlayFailureMessage(view.verdict); overlay != "" {
			// The game is attached but the DLL stayed inert: the skin the user
			// armed for this match did not apply, and that must be visible.
			view.failure = overlay
		}
		if state.Status != nil {
			view.running = state.Status.Running
			view.phase = state.Status.Phase
			view.deferred = state.Status.ChangesAfterGame
		}
		pendingFor := time.Duration(0)
		snapshotMu.Lock()
		next := managerSnapshot{
			pending:  snapshot.pending,
			alive:    true,
			attached: view.attached,
			pid:      view.pid,
			failure:  view.failure,
			verdict:  view.verdict,
		}
		if state.Status != nil {
			next.running = view.running
			next.phase = view.phase
			next.deferred = view.deferred
			next.prefix = state.Status.prefix()
			next.observed = now
		} else {
			next.running, next.phase, next.prefix = snapshot.running, snapshot.phase, snapshot.prefix
		}
		snapshot = next
		if watch.last.game && !view.game {
			gameEndedAt = now
		}
		if rebuild != nil {
			pendingFor = now.Sub(rebuild.editAt)
		}
		snapshotMu.Unlock()
		verdictChanged := view.verdict != "" && view.verdict != watch.last.verdict
		for _, line := range watch.note(view, now, pendingFor) {
			display.Log(line)
		}
		if state.Status != nil {
			observeRebuild(*state.Status, now)
		}
		if notify := OnOverlayVerdict; verdictChanged && notify != nil {
			notify(view.pid, view.verdict)
		}
	}
}
