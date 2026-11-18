package skin

// Which patch of League a package was built for.
//
// The patcher DLL refuses an overlay whose champion skin database comes
// from another game version ("skin database for karma is from another game
// version"), and it disables the whole overlay when it does: the match then
// runs without any skin. The published packages lag a patch behind after a
// Riot update, so ame compares the patch a package states with the patch of
// the installed game before it hands the package to the manager, and falls
// back to generating the package from the installed files.

import (
	"archive/zip"
	"encoding/json"
	"io"
	"regexp"
	"strings"
	"sync"
)

// patchPattern reads the patch out of a package's version string
// ("Patch 16.22").
var patchPattern = regexp.MustCompile(`(\d+)\.(\d+)`)

// installedGamePatch is installedPatch, replaceable in tests.
var installedGamePatch = installedPatch

// installedPatch is the League patch the installed game is on, "16.23", or
// "" when it cannot be read.
func installedPatch(gameDir string) string {
	return patchOfBuild(gameBuild(gameDir))
}

// patchOfBuild is the patch part of a game build, "16.23.839.6405" ->
// "16.23".
func patchOfBuild(build string) string {
	if build == "" || build == "unknown" {
		return ""
	}
	parts := strings.Split(build, ".")
	if len(parts) < 2 {
		return ""
	}
	return parts[0] + "." + parts[1]
}

// packagePatch is the League patch a package was built for, read from its
// metadata ("Version": "Patch 16.22"), or "" when it does not say.
func packagePatch(path string) string {
	z, err := zip.OpenReader(path)
	if err != nil {
		return ""
	}
	defer z.Close()
	for _, f := range z.File {
		if strings.ReplaceAll(f.Name, `\`, "/") != "META/info.json" {
			continue
		}
		r, err := f.Open()
		if err != nil {
			return ""
		}
		data, err := io.ReadAll(io.LimitReader(r, 64<<10))
		r.Close()
		if err != nil {
			return ""
		}
		var info struct {
			Version string `json:"Version"`
		}
		if json.Unmarshal(data, &info) != nil {
			return ""
		}
		match := patchPattern.FindStringSubmatch(info.Version)
		if match == nil {
			return ""
		}
		return match[1] + "." + match[2]
	}
	return ""
}

// packagePatchMatches reports whether a package may be used on the installed
// game. A package that does not state its patch is accepted: the patcher
// DLL, not ame, has the last word about such a package.
func packagePatchMatches(path, installed string) bool {
	pkg := packagePatch(path)
	if pkg == "" || installed == "" {
		return true
	}
	return pkg == installed
}

// patchMismatch names the patch a package was built for and the patch of the
// installed game when they differ. The patcher DLL refuses such an overlay,
// so the package must not be used as it is.
func patchMismatch(path, gameDir string) (pkg, installed string, mismatch bool) {
	installed = installedGamePatch(gameDir)
	if installed == "" {
		return "", "", false
	}
	pkg = packagePatch(path)
	if pkg == "" || pkg == installed {
		return pkg, installed, false
	}
	return pkg, installed, true
}

// stalePackages remembers the patch mismatch seen last, so the message the
// user gets when the patcher refuses an overlay can name it.
var stalePackages struct {
	sync.Mutex
	notice string
}

func setStalePackageNotice(notice string) {
	stalePackages.Lock()
	stalePackages.notice = notice
	stalePackages.Unlock()
}

// StalePackageNotice is the last patch mismatch ame ran into, or "".
func StalePackageNotice() string {
	stalePackages.Lock()
	defer stalePackages.Unlock()
	return stalePackages.notice
}
