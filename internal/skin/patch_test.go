package skin

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

// writePackage writes a package whose metadata states a version, the way the
// published packages do.
func writePackage(t *testing.T, dir, version string) string {
	t.Helper()
	path := filepath.Join(dir, "package.fantome")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create package: %v", err)
	}
	defer f.Close()
	w := zip.NewWriter(f)
	meta, err := w.Create("META/info.json")
	if err != nil {
		t.Fatalf("create metadata: %v", err)
	}
	meta.Write([]byte(`{"Name":"Test","Version":"` + version + `"}`))
	wad, err := w.Create("WAD/Test.wad.client")
	if err != nil {
		t.Fatalf("create wad: %v", err)
	}
	wad.Write(make([]byte, 64))
	if err := w.Close(); err != nil {
		t.Fatalf("close package: %v", err)
	}
	return path
}

func TestPackagePatchReadsTheVersion(t *testing.T) {
	dir := t.TempDir()
	path := writePackage(t, dir, "Patch 16.22")
	if got := packagePatch(path); got != "16.22" {
		t.Fatalf("packagePatch=%q, want 16.22", got)
	}
}

func TestPackagePatchWithoutAVersion(t *testing.T) {
	dir := t.TempDir()
	// The generator's own packages say "ame-local-8": no patch to read.
	path := writePackage(t, dir, "ame-local-8")
	if got := packagePatch(path); got != "" {
		t.Fatalf("packagePatch=%q, want empty", got)
	}
}

func TestPackagePatchOfAMissingPackage(t *testing.T) {
	if got := packagePatch(filepath.Join(t.TempDir(), "gone.fantome")); got != "" {
		t.Fatalf("packagePatch=%q, want empty", got)
	}
}

func TestPatchOfBuild(t *testing.T) {
	cases := map[string]string{
		"16.23.839.6405": "16.23",
		"16.22.1.2":      "16.22",
		"unknown":        "",
		"":               "",
		"16":             "",
	}
	for build, want := range cases {
		if got := patchOfBuild(build); got != want {
			t.Fatalf("patchOfBuild(%q)=%q, want %q", build, got, want)
		}
	}
}

func TestPackagePatchMatches(t *testing.T) {
	dir := t.TempDir()
	stale := writePackage(t, dir, "Patch 16.22")
	if packagePatchMatches(stale, "16.23") {
		t.Fatal("a package for another patch was accepted")
	}
	if !packagePatchMatches(stale, "16.22") {
		t.Fatal("a package for this patch was refused")
	}
	// A package that does not state its patch is left to the patcher DLL.
	if !packagePatchMatches(stale, "") {
		t.Fatal("a package was refused without an installed patch to compare")
	}
}

func TestPatchMismatchNamesBothPatches(t *testing.T) {
	dir := t.TempDir()
	path := writePackage(t, dir, "Patch 16.22")
	old := installedGamePatch
	installedGamePatch = func(string) string { return "16.23" }
	defer func() { installedGamePatch = old }()

	pkg, installed, mismatch := patchMismatch(path, "C:\\Riot Games\\League of Legends\\Game")
	if !mismatch {
		t.Fatal("a package for another patch was not seen as one")
	}
	if pkg != "16.22" || installed != "16.23" {
		t.Fatalf("patchMismatch=(%q,%q), want (16.22,16.23)", pkg, installed)
	}
}

func TestStalePackageNotice(t *testing.T) {
	setStalePackageNotice("The skin packages are built for patch 16.22.")
	if StalePackageNotice() == "" {
		t.Fatal("the patch mismatch was not remembered")
	}
	setStalePackageNotice("")
	if StalePackageNotice() != "" {
		t.Fatal("the patch mismatch was not cleared")
	}
}
