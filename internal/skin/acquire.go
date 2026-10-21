package skin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/hoangvu12/ame/internal/config"
	"github.com/hoangvu12/ame/internal/game"
)

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

var acquireContext, stopAcquire = context.WithCancel(context.Background())

// Shutdown cancels queued and in-flight package downloads.
func Shutdown() { stopAcquire() }

// RemoveLocalGenerator deletes the skin generator bundles that versions with
// on-device generation installed. Packages now come only from the CDN.
func RemoveLocalGenerator() {
	dir := filepath.Join(config.AmeDir, "generators")
	if _, err := os.Stat(dir); err != nil {
		return
	}
	if err := os.RemoveAll(dir); err != nil {
		logSkin("could not remove old skin generator: " + err.Error())
	}
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
	if err := acquireContext.Err(); err != nil {
		return "", err
	}
	key := gameDir + "|" + r.ChampionID + "|" + r.SkinID + "|" + r.BaseSkinID
	acquisition.Lock()
	f := acquisition.flights[key]
	if f == nil {
		f = &packageFlight{done: make(chan struct{})}
		acquisition.flights[key] = f
		go func() {
			work, cancel := context.WithTimeout(acquireContext, 2*time.Minute)
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
	path, err := remoteEnsurePackage(ctx, r, gameDir)
	if err != nil {
		return "", fmt.Errorf("skin package unavailable for %s: %w", r.SkinID, err)
	}
	logSkin("using remote package " + r.SkinID)
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
	r := normalizeRequest(PackageRequest{championID, skinID, baseSkinID})
	if validRequest(r) != nil {
		return ""
	}
	gameDir := game.FindGameDir()
	if gameDir == "" {
		return ""
	}
	path := remoteValidCachedPath(r, gameBuild(gameDir))
	if path != "" {
		logSkin("using remote package " + r.SkinID + " (validated installed build)")
	}
	return path
}
