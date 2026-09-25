package skin

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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hoangvu12/ame/internal/config"
	"github.com/hoangvu12/ame/internal/game"
)

const generatorVersion = "ame-local-8"

// The generator bundle is a release asset, downloaded and installed on
// demand like the other setup components. A local path can override the
// download for development builds.
const generatorAssetURL = "https://github.com/hoangvu12/ame/releases/latest/download/skin-generator.zip"
const generatorAssetEnv = "AME_SKIN_GENERATOR_ZIP"
const generatorBundleLimit = 256 << 20

var generatorClient = &http.Client{Timeout: 5 * time.Minute}

// generatorInstall records which installed bundle the current build uses.
type generatorInstall struct {
	Dir     string `json:"dir"`
	SHA     string `json:"sha"`
	Version string `json:"version"`
}

type PackageRequest struct {
	ChampionID string `json:"championID"`
	SkinID     string `json:"skinID"`
	BaseSkinID string `json:"baseSkinID"`
}

func normalizeRequest(r PackageRequest) PackageRequest {
	// Reconnected plugin state can represent a non-chroma with its own ID as parent.
	if r.BaseSkinID == "0" || r.BaseSkinID == r.SkinID {
		r.BaseSkinID = ""
	}
	return r
}

type inputFingerprint struct {
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	Modified int64  `json:"modified"`
	Header   string `json:"headerSHA256"`
}

type localMetadata struct {
	Request           PackageRequest     `json:"request"`
	Generator         string             `json:"generator"`
	Bundle            string             `json:"bundle"`
	Schema            string             `json:"schema"`
	Build             string             `json:"gameBuild"`
	Inputs            []inputFingerprint `json:"inputs"`
	Artifact          string             `json:"artifact"`
	Digest            string             `json:"sha256"`
	MeshCompatibility bool               `json:"meshCompatibility,omitempty"`
	LTKCompatibility  bool               `json:"ltkCompatibility,omitempty"`
}

var helperMu sync.Mutex
var preparedHelper, preparedRoot string
var generationContext, stopGeneration = context.WithCancel(context.Background())
var helperLifecycleMu sync.Mutex
var helperProcesses sync.WaitGroup

// ShutdownLocalGenerator cancels queued work and waits for helper processes to
// exit. Caller cancellation is separate from application shutdown.
func ShutdownLocalGenerator() {
	helperLifecycleMu.Lock()
	stopGeneration()
	helperLifecycleMu.Unlock()
	helperProcesses.Wait()
}

// PrepareLocalGenerator installs the generator bundle ahead of selections.
// It starts no persistent process and is safe to run alongside server startup.
func PrepareLocalGenerator() {
	if _, err := helperDir(); err != nil {
		logSkin(err.Error() + "; skin generation will retry on demand")
	}
}

// Prepared bundles are immutable and versioned by their bytes. A failed unpack
// cannot be mistaken for an installed helper on the next launch.
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
		return install.Dir, nil
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

// readyMarker returns the content of a bundle's ready file, or "" when the
// bundle is not installed. The marker holds the bundle's content hash.
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

// Fingerprint the WAD directory/chunk hashes as well as file size and timestamp.
// This avoids rereading hundreds of MB of texture payloads on every lobby event.
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

func validRequest(r PackageRequest) error {
	c, e1 := strconv.Atoi(r.ChampionID)
	s, e2 := strconv.Atoi(r.SkinID)
	if e1 != nil || e2 != nil || c <= 0 || s/1000 != c || s%1000 == 0 || strconv.Itoa(c) != r.ChampionID || strconv.Itoa(s) != r.SkinID {
		return fmt.Errorf("invalid champion/full skin ID")
	}
	if r.BaseSkinID != "" && r.BaseSkinID != "0" {
		p, err := strconv.Atoi(r.BaseSkinID)
		if err != nil || p/1000 != c || p == s || strconv.Itoa(p) != r.BaseSkinID {
			return fmt.Errorf("invalid parent skin ID")
		}
	}
	return nil
}

func localProvenance(r PackageRequest, gameDir string) (*localMetadata, string, error) {
	dir, err := helperDir()
	if err != nil {
		return nil, "", err
	}
	data, err := os.ReadFile(filepath.Join(dir, "aliases.json"))
	if err != nil {
		return nil, "", err
	}
	aliases := map[string]string{}
	if err := json.Unmarshal(data, &aliases); err != nil {
		return nil, "", err
	}
	alias := aliases[r.ChampionID]
	if alias == "" || strings.ContainsAny(alias, `/\.:`) {
		return nil, "", fmt.Errorf("unsupported champion alias")
	}
	schema, err := os.ReadFile(filepath.Join(dir, "schema.json"))
	if err != nil {
		return nil, "", err
	}
	m := &localMetadata{Request: r, Generator: generatorVersion, Bundle: filepath.Base(dir), Schema: string(schema)}
	m.MeshCompatibility = true
	m.LTKCompatibility = true
	inputPaths := []string{filepath.Join(gameDir, "League of Legends.exe"), filepath.Join(gameDir, "DATA", "FINAL", "Champions", alias+".wad.client")}
	if m.LTKCompatibility {
		voices, err := filepath.Glob(filepath.Join(gameDir, "DATA", "FINAL", "Champions", alias+".*.wad.client"))
		if err != nil {
			return nil, "", err
		}
		inputPaths = append(inputPaths, voices...)
	}
	for _, path := range inputPaths {
		fp, err := fingerprint(path)
		if err != nil {
			return nil, "", fmt.Errorf("local input unavailable: %w", err)
		}
		m.Inputs = append(m.Inputs, fp)
	}
	m.Build = installedBuild(m.Inputs[0].Path)
	return m, dir, nil
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

func localCached(r PackageRequest, expected *localMetadata) string {
	m, err := readCacheMetadata(r.ChampionID, r.SkinID)
	if err != nil || m.Source != "local" || !sameInputs(m.Local, expected) {
		return ""
	}
	path := filepath.Join(skinDir(r.ChampionID, r.SkinID), m.Local.Artifact)
	if filepath.Base(m.Local.Artifact) != m.Local.Artifact {
		return ""
	}
	data, err := os.ReadFile(path)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(data)) != m.Local.Digest {
		return ""
	}
	return path
}

type packageFlight struct {
	done chan struct{}
	path string
	err  error
}

var acquisition = struct {
	sync.Mutex
	flights map[string]*packageFlight
	slots   chan struct{}
}{flights: make(map[string]*packageFlight), slots: make(chan struct{}, 1)}

var acquirePackage = ensurePackage

// EnsurePackage is shared by own/teammate apply and prefetch. Caller cancellation
// stops waiting, not work another caller still needs. The worker has its own bound.
func EnsurePackage(ctx context.Context, r PackageRequest, gameDir string) (string, error) {
	r = normalizeRequest(r)
	if err := validRequest(r); err != nil {
		return "", err
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if err := generationContext.Err(); err != nil {
		return "", err
	}
	key := gameDir + "|" + r.ChampionID + "|" + r.SkinID + "|" + r.BaseSkinID
	acquisition.Lock()
	f := acquisition.flights[key]
	if f == nil {
		f = &packageFlight{done: make(chan struct{})}
		acquisition.flights[key] = f
		go func() {
			work, cancel := context.WithTimeout(generationContext, 2*time.Minute)
			defer cancel()
			select {
			case acquisition.slots <- struct{}{}:
				f.path, f.err = acquirePackage(work, r, gameDir)
				<-acquisition.slots
			case <-work.Done():
				f.err = work.Err()
			}
			acquisition.Lock()
			delete(acquisition.flights, key)
			close(f.done)
			acquisition.Unlock()
		}()
	}
	acquisition.Unlock()
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-f.done:
		return f.path, f.err
	}
}

func ensurePackage(ctx context.Context, r PackageRequest, gameDir string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	m, dir, localErr := localProvenance(r, gameDir)
	if localErr == nil {
		if path := localCached(r, m); path != "" {
			logSkin("using local package " + r.SkinID)
			return path, nil
		}
		var path string
		if path, localErr = generatePackage(ctx, r, gameDir, dir, m); localErr == nil {
			logSkin("generated local package " + r.SkinID)
			return path, nil
		}
	}
	return "", fmt.Errorf("skin generation unavailable for %s: %w", r.SkinID, localErr)
}

type helperRequest struct {
	PackageRequest
	GameDir           string `json:"gameDir"`
	Output            string `json:"output"`
	MeshCompatibility bool   `json:"meshCompatibility"`
	LTKCompatibility  bool   `json:"ltkCompatibility"`
}

type helperResult struct {
	OK                bool     `json:"ok"`
	Error             string   `json:"error"`
	Kind              string   `json:"kind"`
	Generator         string   `json:"generator"`
	Inputs            []string `json:"inputs"`
	MeshCompatibility bool     `json:"meshCompatibility"`
	LTKCompatibility  bool     `json:"ltkCompatibility"`
}

var invokeHelper = runHelper

func runHelper(ctx context.Context, dir string, request helperRequest) (helperResult, error) {
	helperLifecycleMu.Lock()
	if err := generationContext.Err(); err != nil {
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
	result, err := invokeHelper(ctx, dir, helperRequest{PackageRequest: r, GameDir: gameDir, Output: tmp.Name(), MeshCompatibility: m.MeshCompatibility, LTKCompatibility: m.LTKCompatibility})
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

func atomicWrite(path string, data []byte) error {
	f, err := os.CreateTemp(filepath.Dir(path), ".publish-")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}

// Download remains a compatibility entry point; all acquisitions share this worker.
func Download(championID, skinID, baseSkinID string) (string, error) {
	return EnsurePackage(context.Background(), PackageRequest{championID, skinID, baseSkinID}, game.FindGameDir())
}

func GetValidCachedPath(championID, skinID, baseSkinID string) string {
	r := PackageRequest{championID, skinID, baseSkinID}
	r = normalizeRequest(r)
	if validRequest(r) != nil {
		return ""
	}
	if m, _, err := localProvenance(r, game.FindGameDir()); err == nil {
		path := localCached(r, m)
		if path != "" {
			logSkin("using local package " + r.SkinID + " (validated installed inputs)")
		}
		return path
	}
	return ""
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
