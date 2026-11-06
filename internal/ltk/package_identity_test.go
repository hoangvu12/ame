package ltk

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPackageIdentitySwap(t *testing.T) {
	for _, failure := range []string{"", "install_mod", "toggle_mod"} {
		t.Run("failure="+failure, func(t *testing.T) {
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
				setSnapshot(func(s *managerSnapshot) { *s = managerSnapshot{alive: true, running: true} })
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
	setSnapshot(func(s *managerSnapshot) { *s = managerSnapshot{alive: true, running: true} })
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
