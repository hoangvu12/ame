package skin

// The local generator builds a skin package from the installed League files.
//
// It is the answer to the moment after a Riot patch: the published packages
// are built for the previous patch, and the patcher DLL refuses an overlay
// whose champion skin database comes from another game version ("skin
// database for karma is from another game version"), which leaves the match
// without any skin. A package generated from the installed files always
// matches the installed game, so it is used whenever a published package
// turns out to be for another patch.

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hoangvu12/ame/internal/config"
)

const generatorVersion = "ame-local-8"

// The generator bundle is a release asset, pinned to the release that
// published it. The bundle is versioned by its bytes and installed under
// %LOCALAPPDATA%\ame\generators, so a newer bundle can replace it later.
const generatorAssetURL = "https://github.com/hoangvu12/ame/releases/download/v1.16.1/skin-generator.zip"

// A bundle on disk replaces the download (development builds).
const generatorAssetEnv = "AME_SKIN_GENERATOR_ZIP"

const generatorBundleLimit = 256 << 20

var generatorClient = &http.Client{Timeout: 5 * time.Minute}

// generatorInstall records which installed bundle the current build uses.
type generatorInstall struct {
	Dir     string `json:"dir"`
	SHA     string `json:"sha"`
	Version string `json:"version"`
}

// inputFingerprint describes one installed file the generator reads. The
// champion WAD is large, so only its header is hashed.
type inputFingerprint struct {
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	Modified int64  `json:"modified"`
	Header   string `json:"headerSHA256"`
}

// localMetadata is the cache record of a generated package.
type localMetadata struct {
	Request           PackageRequest     `json:"request"`
	Generator         string             `json:"generator"`
	Bundle            string             `json:"bundle"`
	Build             string             `json:"gameBuild"`
	Inputs            []inputFingerprint `json:"inputs"`
	Artifact          string             `json:"artifact"`
	Digest            string             `json:"sha256"`
	MeshCompatibility bool               `json:"meshCompatibility,omitempty"`
	LTKCompatibility  bool               `json:"ltkCompatibility,omitempty"`
}

// helperRequest is what the generator expects on its standard input.
type helperRequest struct {
	PackageRequest
	GameDir           string `json:"gameDir"`
	Output            string `json:"output"`
	MeshCompatibility bool   `json:"meshCompatibility"`
	LTKCompatibility  bool   `json:"ltkCompatibility"`
}

// helperResult is what the generator answers on its standard output.
type helperResult struct {
	OK                bool     `json:"ok"`
	Error             string   `json:"error"`
	Kind              string   `json:"kind"`
	Generator         string   `json:"generator"`
	Inputs            []string `json:"inputs"`
	MeshCompatibility bool     `json:"meshCompatibility"`
	LTKCompatibility  bool     `json:"ltkCompatibility"`
}

var (
	helperMu          sync.Mutex
	preparedHelper    string
	preparedRoot      string
	helperLifecycleMu sync.Mutex
	helperProcesses   sync.WaitGroup
)

// PrepareLocalGenerator installs the generator bundle in the background, so
// it is ready when a published package turns out to be for another patch.
// The bundle is only needed for that fallback: a failure here never stops
// ame, and the install is retried on demand.
func PrepareLocalGenerator() {
	go func() {
		if _, err := helperDir(); err != nil {
			logSkin(err.Error() + "; skin generation will retry on demand")
		}
	}()
}

// RemoveLocalGenerator drops the generator bundles of other versions. The
// bundle in use stays: it is the fallback for packages built for another
// patch, and downloading it again costs megabytes.
func RemoveLocalGenerator() {
	dir := filepath.Join(config.AmeDir, "generators")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	install, err := readGeneratorInstall(generatorManifestPath())
	keep := ""
	if err == nil && install.Dir != "" {
		keep = filepath.Clean(install.Dir)
	}
	for _, entry := range entries {
		if entry.Name() == "install.json" {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		if keep != "" && filepath.Clean(path) == keep {
			continue
		}
		_ = os.RemoveAll(path)
	}
	if keep == "" {
		// No bundle in use: the folder goes too.
		_ = os.Remove(generatorManifestPath())
		_ = os.Remove(dir)
	}
}

// helperDir is the installed generator bundle. Bundles are immutable and
// versioned by their bytes: a failed unpack cannot be mistaken for an
// installed helper on the next launch.
func helperDir() (string, error) {
	helperMu.Lock()
	defer helperMu.Unlock()
	if preparedHelper != "" && preparedRoot == config.AmeDir {
		return preparedHelper, nil
	}
	manifest := generatorManifestPath()
	install, err := readGeneratorInstall(manifest)
	if err == nil && install.Version == generatorVersion && readyMarker(install.Dir) == install.SHA {
		preparedHelper, preparedRoot = install.Dir, config.AmeDir
		return preparedHelper, nil
	}
	data, err := generatorBundle()
	if err != nil {
		return "", err
	}
	id := fmt.Sprintf("%x", sha256.Sum256(data))
	dir := filepath.Join(config.AmeDir, "generators", id[:16])
	if readyMarker(dir) != id {
		if err := installGeneratorBundle(data, dir, id); err != nil {
			return "", err
		}
	}
	if err := writeGeneratorInstall(manifest, &generatorInstall{Dir: dir, SHA: id, Version: generatorVersion}); err != nil {
		return "", err
	}
	preparedHelper, preparedRoot = dir, config.AmeDir
	return dir, nil
}

func generatorManifestPath() string {
	return filepath.Join(config.AmeDir, "generators", "install.json")
}

func readGeneratorInstall(path string) (*generatorInstall, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var install generatorInstall
	if err := json.Unmarshal(data, &install); err != nil {
		return nil, err
	}
	if install.Dir == "" || !strings.HasPrefix(filepath.Clean(install.Dir), filepath.Join(config.AmeDir, "generators")) {
		return nil, fmt.Errorf("invalid generator install")
	}
	return &install, nil
}

func writeGeneratorInstall(path string, install *generatorInstall) error {
	data, err := json.MarshalIndent(install, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return atomicWrite(path, data)
}

// readyMarker is the content of a bundle's ready file, or "" when it is not
// installed. The marker holds the bundle's content hash.
func readyMarker(dir string) string {
	data, err := os.ReadFile(filepath.Join(dir, "ready"))
	if err != nil {
		return ""
	}
	return string(data)
}

func installGeneratorBundle(data []byte, dir, id string) error {
	if err := os.MkdirAll(filepath.Dir(dir), 0755); err != nil {
		return err
	}
	tmp, err := os.MkdirTemp(filepath.Dir(dir), ".prepare-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	archive := filepath.Join(tmp, "helper.zip")
	if err := os.WriteFile(archive, data, 0600); err != nil {
		return err
	}
	if err := Extract(archive, tmp); err != nil {
		return err
	}
	_ = os.Remove(archive)
	if err := os.WriteFile(filepath.Join(tmp, "ready"), []byte(id), 0600); err != nil {
		return err
	}
	return os.Rename(tmp, dir)
}

func generatorBundle() ([]byte, error) {
	if path := os.Getenv(generatorAssetEnv); path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("local generator bundle: %w", err)
		}
		return data, nil
	}
	logSkin("downloading skin generator bundle")
	req, err := http.NewRequest("GET", generatorAssetURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "ame")
	resp, err := generatorClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("generator bundle download: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("generator bundle download: status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, generatorBundleLimit))
	if err != nil {
		return nil, fmt.Errorf("generator bundle download: %w", err)
	}
	logSkin(fmt.Sprintf("downloaded skin generator bundle (%d bytes)", len(data)))
	return data, nil
}

// fingerprint the WAD files as well as size and timestamp. This avoids
// rereading hundreds of MB of texture payloads on every lobby event.
func fingerprint(path string) (inputFingerprint, error) {
	f, err := os.Open(path)
	if err != nil {
		return inputFingerprint{}, err
	}
	defer f.Close()
	s, err := f.Stat()
	if err != nil {
		return inputFingerprint{}, err
	}
	h := sha256.New()
	if _, err := io.Copy(h, io.LimitReader(f, 4<<20)); err != nil {
		return inputFingerprint{}, err
	}
	return inputFingerprint{path, s.Size(), s.ModTime().UnixNano(), hex.EncodeToString(h.Sum(nil))}, nil
}

// localProvenance describes what the generator would read for a selection:
// the champion alias and the installed files it builds from.
func localProvenance(r PackageRequest, gameDir, dir string) (*localMetadata, error) {
	data, err := os.ReadFile(filepath.Join(dir, "aliases.json"))
	if err != nil {
		return nil, err
	}
	aliases := map[string]string{}
	if err := json.Unmarshal(data, &aliases); err != nil {
		return nil, err
	}
	alias := aliases[r.ChampionID]
	if alias == "" || strings.ContainsAny(alias, `/\.:`) {
		return nil, fmt.Errorf("unsupported champion alias")
	}
	m := &localMetadata{Request: r, Generator: generatorVersion, Bundle: filepath.Base(dir)}
	// The generated package re-identifies the skin as the base skin and
	// rewrites the mesh paths, which is what the LTK runtime accepts.
	m.MeshCompatibility = true
	m.LTKCompatibility = true
	inputPaths := []string{filepath.Join(gameDir, "League of Legends.exe"), filepath.Join(gameDir, "DATA", "FINAL", "Champions", alias+".wad.client")}
	voices, err := filepath.Glob(filepath.Join(gameDir, "DATA", "FINAL", "Champions", alias+".*.wad.client"))
	if err != nil {
		return nil, err
	}
	inputPaths = append(inputPaths, voices...)
	for _, path := range inputPaths {
		fp, err := fingerprint(path)
		if err != nil {
			return nil, fmt.Errorf("local input unavailable: %w", err)
		}
		m.Inputs = append(m.Inputs, fp)
	}
	m.Build = installedBuild(m.Inputs[0].Path)
	return m, nil
}

func sameInputs(a, b *localMetadata) bool {
	if a == nil || b == nil {
		return false
	}
	x, y := *a, *b
	x.Artifact, x.Digest, y.Artifact, y.Digest = "", "", "", ""
	xb, _ := json.Marshal(x)
	yb, _ := json.Marshal(y)
	return bytes.Equal(xb, yb)
}

// localCached is a package generated earlier from the same installed files.
func localCached(r PackageRequest, expected *localMetadata) string {
	m, err := readCacheMetadata(r.ChampionID, r.SkinID)
	if err != nil || m.Source != "local" || m.Local == nil || !sameInputs(m.Local, expected) {
		return ""
	}
	if filepath.Base(m.Local.Artifact) != m.Local.Artifact {
		return ""
	}
	path := filepath.Join(skinDir(r.ChampionID, r.SkinID), m.Local.Artifact)
	data, err := os.ReadFile(path)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(data)) != m.Local.Digest {
		return ""
	}
	return path
}

// localPackage is localEnsurePackage, replaceable in tests.
var localPackage = localEnsurePackage

// localEnsurePackage returns a package built from the installed files for a
// selection, generating it when the cached copy no longer matches the
// installation.
func localEnsurePackage(ctx context.Context, r PackageRequest, gameDir string) (string, error) {
	dir, err := helperDir()
	if err != nil {
		return "", err
	}
	m, err := localProvenance(r, gameDir, dir)
	if err != nil {
		return "", err
	}
	if path := localCached(r, m); path != "" {
		logSkin("using the package generated from the installed files " + r.SkinID)
		return path, nil
	}
	path, err := generatePackage(ctx, r, gameDir, dir, m)
	if err != nil {
		return "", err
	}
	logSkin("generated the package " + r.SkinID + " from the installed files")
	return path, nil
}

func runHelper(ctx context.Context, dir string, request helperRequest) (helperResult, error) {
	helperLifecycleMu.Lock()
	if err := ctx.Err(); err != nil {
		helperLifecycleMu.Unlock()
		return helperResult{}, err
	}
	helperProcesses.Add(1)
	helperLifecycleMu.Unlock()
	defer helperProcesses.Done()
	data, _ := json.Marshal(request)
	cmd := exec.CommandContext(ctx, filepath.Join(dir, "skin-generator.exe"))
	hideHelper(cmd)
	cmd.Stdin = bytes.NewReader(data)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, runErr := cmd.Output()
	var result helperResult
	if ctx.Err() != nil {
		return result, ctx.Err()
	}
	if err := json.Unmarshal(out, &result); err != nil {
		return result, fmt.Errorf("helper protocol: %v (%v)", err, runErr)
	}
	if runErr != nil || !result.OK {
		return result, fmt.Errorf("helper %s: %s (%v)", result.Kind, result.Error, runErr)
	}
	return result, nil
}

func generatePackage(ctx context.Context, r PackageRequest, gameDir, dir string, m *localMetadata) (string, error) {
	cache := skinDir(r.ChampionID, r.SkinID)
	if err := os.MkdirAll(cache, 0755); err != nil {
		return "", err
	}
	tmp, err := os.CreateTemp(cache, ".generate-*.fantome")
	if err != nil {
		return "", err
	}
	tmp.Close()
	defer os.Remove(tmp.Name())
	result, err := runHelper(ctx, dir, helperRequest{PackageRequest: r, GameDir: gameDir, Output: tmp.Name(), MeshCompatibility: m.MeshCompatibility, LTKCompatibility: m.LTKCompatibility})
	if err != nil {
		return "", err
	}
	if result.Generator != generatorVersion || result.MeshCompatibility != m.MeshCompatibility || result.LTKCompatibility != m.LTKCompatibility || !helperInputsMatch(result.Inputs, m.Inputs) {
		return "", fmt.Errorf("unexpected helper provenance")
	}
	for _, before := range m.Inputs {
		after, err := fingerprint(before.Path)
		if err != nil || after != before {
			return "", fmt.Errorf("installation changed during generation")
		}
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if err := validatePackage(tmp.Name()); err != nil {
		return "", err
	}
	content, err := os.ReadFile(tmp.Name())
	if err != nil {
		return "", err
	}
	m.Digest = fmt.Sprintf("%x", sha256.Sum256(content))
	m.Artifact = "local-" + m.Digest + ".fantome"
	path := filepath.Join(cache, m.Artifact)
	if existing, err := os.ReadFile(path); err != nil || !bytes.Equal(existing, content) {
		if err := os.Rename(tmp.Name(), path); err != nil {
			return "", err
		}
	}
	meta := cacheMetadata{Source: "local", Local: m, CheckedAt: time.Now().UTC()}
	encoded, _ := json.MarshalIndent(meta, "", "  ")
	if err := atomicWrite(skinCacheMetadataPath(r.ChampionID, r.SkinID), encoded); err != nil {
		return "", err
	}
	return path, nil
}

func helperInputsMatch(paths []string, inputs []inputFingerprint) bool {
	if len(paths) != len(inputs)-1 {
		return false
	}
	for i, path := range paths {
		if !strings.EqualFold(filepath.Clean(path), filepath.Clean(inputs[i+1].Path)) {
			return false
		}
	}
	return true
}
