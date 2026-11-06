package ltk

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPackageIdentitySwap(t *testing.T) {
	for _, failure := range []string{"", "install_mod", "toggle_mod"} {
		t.Run("failure="+failure, func(t *testing.T) {
			isolateManagerState(t)
			pkg := filepath.Join(t.TempDir(), "overlay.fantome")
			write := func(data string) {
				t.Helper()
				if err := os.WriteFile(pkg, []byte(data), 0600); err != nil {
					t.Fatal(err)
				}
			}
			write("old")
			old, err := packageDigest(pkg)
			if err != nil {
				t.Fatal(err)
			}
			installs := 0
			fail := failure
			fake := startFakeCDP(t, func(expr string) (any, error) {
				switch {
				case strings.Contains(expr, "get_patcher_status"):
					return fakeInvokeReply(map[string]any{"running": true, "phase": "patching"}), nil
				case strings.Contains(expr, "__ame"):
					return nil, nil
				case strings.Contains(expr, "get_installed_mods"):
					return fakeInvokeReply([]any{}), nil
				case strings.Contains(expr, "install_mod"):
					installs++
					if fail == "install_mod" {
						return nil, fmt.Errorf("install failed")
					}
					return fakeInvokeReply(map[string]any{"kind": "installed", "mod": map[string]any{"id": "new", "name": ameModName}}), nil
				case strings.Contains(expr, "toggle_mod"):
					if fail == "toggle_mod" {
						return nil, fmt.Errorf("enable failed")
					}
					return fakeInvokeReply(nil), nil
				}
				return nil, fmt.Errorf("unexpected IPC: %s", expr)
			})
			c, err := cdpDial(wsURL(fake.server.URL) + "/ws")
			if err != nil {
				t.Fatal(err)
			}
			defer c.close()
			m := &managerSession{cdp: c, modID: "old", modFile: pkg, modHash: old}
			if err := swapAmeMod(m, pkg); err != nil {
				t.Fatal(err)
			}
			if installs != 0 {
				t.Fatal("unchanged bytes installed again")
			}
			write("new")
			digest, err := packageDigest(pkg)
			if err != nil {
				t.Fatal(err)
			}
			if m.matchesPackage(pkg, digest) {
				t.Fatal("Start fast path accepted changed bytes")
			}
			if failure == "" {
				previous, snap := manager, currentSnapshot()
				m.stop, m.exited = make(chan struct{}), make(chan struct{})
				manager = m
				setSnapshot(func(s *managerSnapshot) {
					*s = managerSnapshot{alive: true, running: true, phase: "patching", observed: time.Now()}
				})
				err = Start(strings.TrimSuffix(pkg, ".fantome"))
				manager = previous
				setSnapshot(func(s *managerSnapshot) { *s = snap })
			} else {
				err = swapAmeMod(m, pkg)
			}
			if failure != "" {
				if err == nil {
					t.Fatal("expected install/enable failure")
				}
				if m.matchesPackage(pkg, digest) {
					t.Fatal("failed install published identity")
				}
				fail = ""
				if err := swapAmeMod(m, pkg); err != nil {
					t.Fatal(err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if installs == 0 || !m.matchesPackage(pkg, digest) {
				t.Fatal("changed bytes were not installed and remembered")
			}
			count := installs
			if err := swapAmeMod(m, pkg); err != nil {
				t.Fatal(err)
			}
			if installs != count {
				t.Fatal("unchanged replacement installed again")
			}
		})
	}
}

func TestStartReusesUnchangedPackage(t *testing.T) {
	isolateManagerState(t)
	pkg := filepath.Join(t.TempDir(), "overlay.fantome")
	if err := os.WriteFile(pkg, []byte("package"), 0600); err != nil {
		t.Fatal(err)
	}
	digest, err := packageDigest(pkg)
	if err != nil {
		t.Fatal(err)
	}
	previous, previousSnapshot := manager, currentSnapshot()
	defer func() { manager = previous; setSnapshot(func(s *managerSnapshot) { *s = previousSnapshot }) }()
	manager = &managerSession{modID: "installed", modFile: pkg, modHash: digest, stop: make(chan struct{}), exited: make(chan struct{})}
	setSnapshot(func(s *managerSnapshot) {
		*s = managerSnapshot{alive: true, running: true, phase: "patching", observed: time.Now()}
	})
	if err := Start(strings.TrimSuffix(pkg, ".fantome")); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(pkg); err != nil {
		t.Fatal(err)
	}
	if err := Start(strings.TrimSuffix(pkg, ".fantome")); err == nil {
		t.Fatal("missing package reused by pathname")
	}
}

func TestRebuildTimeoutRetainsReadinessFence(t *testing.T) {
	isolateManagerState(t)
	old, snap := overlayRebuildTimeout, currentSnapshot()
	overlayRebuildTimeout = 20 * time.Millisecond
	defer func() { overlayRebuildTimeout = old; setSnapshot(func(s *managerSnapshot) { *s = snap }) }()
	fake := startFakeCDP(t, func(expr string) (any, error) {
		return fakeInvokeReply(map[string]any{"running": true, "phase": "building"}), nil
	})
	c, err := cdpDial(wsURL(fake.server.URL) + "/ws")
	if err != nil {
		t.Fatal(err)
	}
	defer c.close()
	setSnapshot(func(s *managerSnapshot) { *s = managerSnapshot{pending: true} })
	if err := awaitOverlayRebuild(&managerSession{cdp: c}); !errors.Is(err, ErrRebuildPending) {
		t.Fatalf("timeout reported as %v", err)
	}
	s := currentSnapshot()
	if !s.pending || s.phase != "building" || s.observed.IsZero() {
		t.Fatalf("status/fence lost: %+v", s)
	}
}

func TestStartRetriesReadinessAfterRebuildTimeout(t *testing.T) {
	isolateManagerState(t)
	pkg := filepath.Join(t.TempDir(), "overlay.fantome")
	if err := os.WriteFile(pkg, []byte("package"), 0600); err != nil {
		t.Fatal(err)
	}
	digest, err := packageDigest(pkg)
	if err != nil {
		t.Fatal(err)
	}
	oldManager, snap, oldTimeout := manager, currentSnapshot(), overlayRebuildTimeout
	defer func() {
		manager = oldManager
		overlayRebuildTimeout = oldTimeout
		setSnapshot(func(s *managerSnapshot) { *s = snap })
	}()
	phase := "building"
	reads := 0
	fake := startFakeCDP(t, func(expr string) (any, error) {
		switch {
		case strings.Contains(expr, "get_patcher_status"):
			reads++
			return fakeInvokeReply(map[string]any{"running": true, "phase": phase}), nil
		case strings.Contains(expr, "__AME_BRIDGE__"):
			return nil, nil
		}
		return nil, fmt.Errorf("unexpected IPC while retrying unchanged package: %s", expr)
	})
	c, err := cdpDial(wsURL(fake.server.URL) + "/ws")
	if err != nil {
		t.Fatal(err)
	}
	defer c.close()
	manager = &managerSession{cdp: c, modID: "installed", modFile: pkg, modHash: digest, stop: make(chan struct{}), exited: make(chan struct{})}
	setSnapshot(func(s *managerSnapshot) {
		*s = managerSnapshot{alive: true, running: true, phase: "building", pending: true}
	})
	overlayRebuildTimeout = time.Millisecond
	if err := Start(strings.TrimSuffix(pkg, ".fantome")); err == nil {
		t.Fatal("pending package reported ready on timeout")
	}
	if Ready() {
		t.Fatal("timed-out package remained ready")
	}
	phase = "patching"
	overlayRebuildTimeout = 4 * time.Second
	before := reads
	if err := Start(strings.TrimSuffix(pkg, ".fantome")); err != nil {
		t.Fatal(err)
	}
	if reads <= before || !Ready() {
		t.Fatal("retry did not confirm readiness through fresh status reads")
	}
}

// Readiness needs a fresh status of a running scan with no edit in flight.
func TestReadyRejectsStalePendingAndStoppedStatus(t *testing.T) {
	isolateManagerState(t)
	for _, s := range []managerSnapshot{
		{alive: true, running: true, phase: "patching"},
		{alive: true, running: true, phase: "patching", observed: time.Now(), pending: true},
		{alive: true, phase: "patching", observed: time.Now()},
		{alive: true, running: true, phase: "building", observed: time.Now()},
		{alive: true, running: true, phase: "patching", observed: time.Now().Add(-5 * time.Second)},
	} {
		setSnapshot(func(dst *managerSnapshot) { *dst = s })
		if Ready() {
			t.Fatalf("accepted unsafe status: %+v", s)
		}
	}
	setSnapshot(func(s *managerSnapshot) {
		*s = managerSnapshot{alive: true, running: true, phase: "patching", observed: time.Now()}
	})
	if !Ready() {
		t.Fatal("fresh scanning status not ready")
	}
}
