package custommods

import (
	"archive/zip"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"github.com/hoangvu12/ame/internal/config"
)

// PickFileResult contains metadata extracted from a picked mod file.
type PickFileResult struct {
	FilePath            string `json:"filePath"`
	SuggestedName       string `json:"suggestedName"`
	SuggestedAuthor     string `json:"suggestedAuthor"`
	SuggestedChampionID int    `json:"suggestedChampionId"`
}

// modMeta represents META/info.json inside a mod archive.
type modMeta struct {
	Author   string `json:"Author"`
	Name     string `json:"Name"`
	Champion string `json:"Champion,omitempty"`
}

var hiddenAttr = &syscall.SysProcAttr{HideWindow: true}

// generateID creates a unique ID for a new custom mod.
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// PickModFile opens a native Windows file dialog for .fantome/.zip files
// and returns metadata extracted from the selected file.
func PickModFile() (*PickFileResult, error) {
	cmd := exec.Command("powershell", "-NoProfile", "-WindowStyle", "Hidden", "-Command",
		`Add-Type -AssemblyName System.Windows.Forms; `+
			`$form = New-Object System.Windows.Forms.Form -Property @{TopMost=$true;Width=0;Height=0}; `+
			`$f = New-Object System.Windows.Forms.OpenFileDialog; `+
			`$f.Filter = 'Mod Files (*.fantome;*.zip)|*.fantome;*.zip|All Files (*.*)|*.*'; `+
			`$f.Title = 'Select Custom Mod'; `+
			`if ($f.ShowDialog($form) -eq 'OK') { $f.FileName }`)
	cmd.SysProcAttr = hiddenAttr

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("file picker failed: %w", err)
	}

	filePath := strings.TrimSpace(string(out))
	if filePath == "" {
		return nil, nil // User cancelled
	}

	result := &PickFileResult{FilePath: filePath}

	// Try to extract metadata from archive
	meta, err := readModMeta(filePath)
	if err == nil && meta != nil {
		result.SuggestedName = meta.Name
		result.SuggestedAuthor = meta.Author
		if meta.Champion != "" {
			result.SuggestedChampionID = resolveChampionName(meta.Champion)
		}
	}

	// Auto-detect champion from filename if not found in meta
	if result.SuggestedChampionID == 0 {
		if name := detectChampionFromFilename(filePath); name != "" {
			result.SuggestedChampionID = resolveChampionName(name)
		}
	}

	// Fallback: use filename as name
	if result.SuggestedName == "" {
		base := filepath.Base(filePath)
		ext := filepath.Ext(base)
		result.SuggestedName = base[:len(base)-len(ext)]
	}

	return result, nil
}

// PickImageFile opens a native Windows file dialog for image files.
func PickImageFile() (string, error) {
	cmd := exec.Command("powershell", "-NoProfile", "-WindowStyle", "Hidden", "-Command",
		`Add-Type -AssemblyName System.Windows.Forms; `+
			`$form = New-Object System.Windows.Forms.Form -Property @{TopMost=$true;Width=0;Height=0}; `+
			`$f = New-Object System.Windows.Forms.OpenFileDialog; `+
			`$f.Filter = 'Image Files (*.png;*.jpg;*.jpeg;*.webp)|*.png;*.jpg;*.jpeg;*.webp|All Files (*.*)|*.*'; `+
			`$f.Title = 'Select Preview Image'; `+
			`if ($f.ShowDialog($form) -eq 'OK') { $f.FileName }`)
	cmd.SysProcAttr = hiddenAttr

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("image picker failed: %w", err)
	}

	return strings.TrimSpace(string(out)), nil
}

// readModMeta reads META/info.json from a mod archive.
func readModMeta(archivePath string) (*modMeta, error) {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	for _, f := range r.File {
		normalized := strings.ToLower(filepath.ToSlash(f.Name))
		if normalized == "meta/info.json" || strings.HasSuffix(normalized, "/meta/info.json") {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()

			var meta modMeta
			if err := json.NewDecoder(rc).Decode(&meta); err != nil {
				return nil, err
			}
			return &meta, nil
		}
	}

	return nil, nil
}

// ImportMod extracts a mod archive to custom-mods/{id}/ and saves metadata.
func ImportMod(filePath, name, author string, championID int, imageBase64 string) (*config.CustomMod, error) {
	id := generateID()
	modDir := filepath.Join(config.CustomModsDir, id)

	if err := extractMod(filePath, modDir); err != nil {
		os.RemoveAll(modDir)
		return nil, fmt.Errorf("failed to extract mod: %w", err)
	}

	hasImage := false
	if imageBase64 != "" {
		if err := saveImageBase64(modDir, imageBase64); err == nil {
			hasImage = true
		}
	}

	mod := config.CustomMod{
		ID:         id,
		Name:       name,
		Author:     author,
		ChampionID: championID,
		Enabled:    true,
		HasImage:   hasImage,
	}

	if err := config.AddCustomMod(mod); err != nil {
		os.RemoveAll(modDir)
		return nil, err
	}

	return &mod, nil
}

// extractMod extracts a .fantome/.zip archive to destDir.
func extractMod(archivePath, destDir string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer r.Close()

	os.MkdirAll(destDir, os.ModePerm)

	cleanDest := filepath.Clean(destDir)
	for _, f := range r.File {
		target := filepath.Join(destDir, filepath.FromSlash(f.Name))
		// Prevent zip-slip
		if !strings.HasPrefix(filepath.Clean(target), cleanDest+string(os.PathSeparator)) && filepath.Clean(target) != cleanDest {
			continue
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(target, os.ModePerm)
			continue
		}

		os.MkdirAll(filepath.Dir(target), os.ModePerm)

		outFile, err := os.Create(target)
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// DeleteMod removes a custom mod's files and config entry.
func DeleteMod(id string) error {
	modDir := filepath.Join(config.CustomModsDir, id)
	os.RemoveAll(modDir)
	return config.DeleteCustomMod(id)
}

// SaveImageFromPicker opens a file picker and saves the selected image to a mod's directory.
func SaveImageFromPicker(id string) (bool, error) {
	imgPath, err := PickImageFile()
	if err != nil {
		return false, err
	}
	if imgPath == "" {
		return false, nil // User cancelled
	}

	modDir := filepath.Join(config.CustomModsDir, id)
	if err := copyImageFile(imgPath, modDir); err != nil {
		return false, err
	}

	config.SetCustomModHasImage(id, true)
	return true, nil
}

// DeleteImage removes the preview image for a mod.
func DeleteImage(id string) error {
	modDir := filepath.Join(config.CustomModsDir, id)
	for _, ext := range []string{".png", ".jpg", ".jpeg", ".webp"} {
		os.Remove(filepath.Join(modDir, "preview"+ext))
	}
	return config.SetCustomModHasImage(id, false)
}

// GetImagePath returns the path to a mod's preview image, or empty if none.
func GetImagePath(id string) string {
	modDir := filepath.Join(config.CustomModsDir, id)
	for _, ext := range []string{".png", ".jpg", ".jpeg", ".webp"} {
		p := filepath.Join(modDir, "preview"+ext)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// ServeImage serves a mod's preview image over HTTP.
func ServeImage(w http.ResponseWriter, r *http.Request, id string) {
	imgPath := GetImagePath(id)
	if imgPath == "" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, imgPath)
}

// saveImageBase64 decodes a base64 image string and saves it as preview.
func saveImageBase64(modDir, b64 string) error {
	// Strip data URI prefix if present
	if idx := strings.Index(b64, ","); idx != -1 {
		b64 = b64[idx+1:]
	}

	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return err
	}

	// Detect format from magic bytes
	ext := ".png"
	if len(data) > 3 {
		if data[0] == 0xFF && data[1] == 0xD8 {
			ext = ".jpg"
		} else if len(data) > 12 && string(data[:4]) == "RIFF" && string(data[8:12]) == "WEBP" {
			ext = ".webp"
		}
	}

	return os.WriteFile(filepath.Join(modDir, "preview"+ext), data, 0644)
}

// copyImageFile copies an image file to a mod directory as preview.{ext}.
func copyImageFile(src, modDir string) error {
	// Remove existing preview images
	for _, ext := range []string{".png", ".jpg", ".jpeg", ".webp"} {
		os.Remove(filepath.Join(modDir, "preview"+ext))
	}

	ext := strings.ToLower(filepath.Ext(src))
	if ext == ".jpeg" {
		ext = ".jpg"
	}
	if ext == "" {
		ext = ".png"
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstPath := filepath.Join(modDir, "preview"+ext)
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// --- Junction helpers ---

// CreateJunction creates a Windows directory junction from linkDir to targetDir.
func CreateJunction(linkDir, targetDir string) error {
	cmd := exec.Command("cmd", "/c", "mklink", "/J", linkDir, targetDir)
	cmd.SysProcAttr = hiddenAttr
	return cmd.Run()
}

// isReparsePoint checks if a path is a junction or symlink (reparse point).
func isReparsePoint(path string) bool {
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return false
	}
	attrs, err := syscall.GetFileAttributes(p)
	if err != nil {
		return false
	}
	const FILE_ATTRIBUTE_REPARSE_POINT = 0x400
	return attrs&FILE_ATTRIBUTE_REPARSE_POINT != 0
}

// CleanModsDir safely removes the mods directory, handling junctions properly.
// Junctions are removed with os.Remove (just removes the reparse point),
// regular directories are removed with os.RemoveAll.
func CleanModsDir() {
	entries, err := os.ReadDir(config.ModsDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		entryPath := filepath.Join(config.ModsDir, entry.Name())
		if isReparsePoint(entryPath) {
			os.Remove(entryPath)
		} else {
			os.RemoveAll(entryPath)
		}
	}

	os.RemoveAll(config.ModsDir)
}

// SetupCustomModJunctions creates junctions in the mods directory for each
// enabled custom mod. Returns the list of mod directory names for --mods: arg.
func SetupCustomModJunctions() []string {
	enabled := config.GetEnabledCustomMods()
	if len(enabled) == 0 {
		return nil
	}

	var modNames []string
	for _, mod := range enabled {
		srcDir := filepath.Join(config.CustomModsDir, mod.ID)
		if _, err := os.Stat(srcDir); err != nil {
			continue
		}

		linkName := "custom_" + mod.ID
		linkDir := filepath.Join(config.ModsDir, linkName)

		if err := CreateJunction(linkDir, srcDir); err != nil {
			continue
		}

		modNames = append(modNames, linkName)
	}

	return modNames
}

// --- Champion auto-detection ---

var championNameRegex = regexp.MustCompile(`^([A-Za-z]+)[-_\s]`)

// detectChampionFromFilename tries to extract a champion name from a mod filename.
func detectChampionFromFilename(filePath string) string {
	base := filepath.Base(filePath)
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)]

	matches := championNameRegex.FindStringSubmatch(name)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

// championNameToID maps lowercase champion names to IDs.
// This is a subset of common champions for auto-detection.
// The frontend will have the full list and can correct the selection.
var championNameToID = map[string]int{
	"aatrox": 266, "ahri": 103, "akali": 84, "akshan": 166, "alistar": 12,
	"amumu": 32, "anivia": 34, "annie": 1, "aphelios": 523, "ashe": 22,
	"aurelionsol": 136, "azir": 268, "bard": 432, "belveth": 200, "blitzcrank": 53,
	"brand": 63, "braum": 201, "briar": 233, "caitlyn": 51, "camille": 164,
	"cassiopeia": 69, "chogath": 31, "corki": 42, "darius": 122, "diana": 131,
	"draven": 119, "drmundo": 36, "ekko": 245, "elise": 60, "evelynn": 28,
	"ezreal": 81, "fiddlesticks": 9, "fiora": 114, "fizz": 105, "galio": 3,
	"gangplank": 41, "garen": 86, "gnar": 150, "gragas": 79, "graves": 104,
	"gwen": 887, "hecarim": 120, "heimerdinger": 74, "hwei": 910, "illaoi": 420,
	"irelia": 39, "ivern": 427, "janna": 40, "jarvaniv": 59, "jax": 24,
	"jayce": 126, "jhin": 202, "jinx": 222, "kaisa": 145, "kalista": 429,
	"karma": 43, "karthus": 30, "kassadin": 38, "katarina": 55, "kayle": 10,
	"kayn": 141, "kennen": 85, "khazix": 121, "kindred": 203, "kled": 240,
	"kogmaw": 96, "ksante": 897, "leblanc": 7, "leesin": 64, "leona": 89,
	"lillia": 876, "lissandra": 127, "lucian": 236, "lulu": 117, "lux": 99,
	"malphite": 54, "malzahar": 90, "maokai": 57, "masteryi": 11, "milio": 902,
	"missfortune": 21, "mordekaiser": 82, "morgana": 25, "naafiri": 950,
	"nami": 267, "nasus": 75, "nautilus": 111, "neeko": 518, "nidalee": 76,
	"nilah": 895, "nocturne": 56, "nunu": 20, "olaf": 2, "orianna": 61,
	"ornn": 516, "pantheon": 80, "poppy": 78, "pyke": 555, "qiyana": 246,
	"quinn": 133, "rakan": 497, "rammus": 33, "reksai": 421, "rell": 526,
	"renata": 888, "renekton": 58, "rengar": 107, "riven": 92, "rumble": 68,
	"ryze": 13, "samira": 360, "sejuani": 113, "senna": 235, "seraphine": 147,
	"sett": 875, "shaco": 35, "shen": 98, "shyvana": 102, "singed": 27,
	"sion": 14, "sivir": 15, "skarner": 72, "smolder": 901, "sona": 37,
	"soraka": 16, "swain": 50, "sylas": 517, "syndra": 134, "tahmkench": 223,
	"taliyah": 163, "talon": 91, "taric": 44, "teemo": 17, "thresh": 412,
	"tristana": 18, "trundle": 48, "tryndamere": 23, "twistedfate": 4, "twitch": 29,
	"udyr": 77, "urgot": 6, "varus": 110, "vayne": 67, "veigar": 45,
	"velkoz": 161, "vex": 711, "vi": 254, "viego": 234, "viktor": 112,
	"vladimir": 8, "volibear": 106, "warwick": 19, "wukong": 62, "xayah": 498,
	"xerath": 101, "xinzhao": 5, "yasuo": 157, "yone": 777, "yorick": 83,
	"yuumi": 350, "zac": 154, "zed": 238, "zeri": 221, "ziggs": 115,
	"zilean": 26, "zoe": 142, "zyra": 143,
}

// resolveChampionName maps a champion name (case-insensitive) to its ID.
func resolveChampionName(name string) int {
	lower := strings.ToLower(strings.ReplaceAll(name, " ", ""))
	lower = strings.ReplaceAll(lower, "'", "")
	if id, ok := championNameToID[lower]; ok {
		return id
	}
	return 0
}
