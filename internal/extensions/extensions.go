package extensions

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/gorilla/websocket"
	"github.com/hoangvu12/ame/internal/config"
	"github.com/hoangvu12/ame/internal/display"
)

var extensionsDir = filepath.Join(config.AmeDir, "extensions")

// ExtensionFile represents a single extension JS file on disk.
type ExtensionFile struct {
	Filename string `json:"filename"`
	Source   string `json:"source"`
}

// GetExtensionsMessage is sent by frontend to list installed extensions.
type GetExtensionsMessage struct {
	Type string `json:"type"` // "getExtensions"
}

// AddExtensionMessage is sent by frontend to add an extension from URL.
type AddExtensionMessage struct {
	Type string `json:"type"` // "addExtension"
	URL  string `json:"url"`
}

// RemoveExtensionMessage is sent by frontend to remove an extension.
type RemoveExtensionMessage struct {
	Type     string `json:"type"` // "removeExtension"
	Filename string `json:"filename"`
}

// HandleGetExtensions reads all .js files from the extensions directory.
func HandleGetExtensions(conn *websocket.Conn) {
	os.MkdirAll(extensionsDir, os.ModePerm)

	entries, err := os.ReadDir(extensionsDir)
	if err != nil {
		display.Log(fmt.Sprintf("Extensions: failed to read dir: %v", err))
		sendExtensionsList(conn, []ExtensionFile{})
		return
	}

	var files []ExtensionFile
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".js") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(extensionsDir, e.Name()))
		if err != nil {
			continue
		}
		files = append(files, ExtensionFile{
			Filename: e.Name(),
			Source:   string(data),
		})
	}

	if files == nil {
		files = []ExtensionFile{}
	}
	sendExtensionsList(conn, files)
}

func sendExtensionsList(conn *websocket.Conn, files []ExtensionFile) {
	resp := map[string]interface{}{
		"type":       "extensions",
		"extensions": files,
	}
	data, _ := json.Marshal(resp)
	conn.WriteMessage(websocket.TextMessage, data)
}

// HandleAddExtension downloads a JS file from URL and saves it.
func HandleAddExtension(conn *websocket.Conn, raw json.RawMessage) {
	var msg AddExtensionMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		sendExtensionError(conn, "Invalid request")
		return
	}

	go func() {
		os.MkdirAll(extensionsDir, os.ModePerm)

		// Extract filename from URL
		filename := filenameFromURL(msg.URL)
		if filename == "" || !strings.HasSuffix(strings.ToLower(filename), ".js") {
			filename = "extension.js"
		}

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Get(msg.URL)
		if err != nil {
			display.Log(fmt.Sprintf("Extensions: download failed: %v", err))
			sendExtensionError(conn, fmt.Sprintf("Failed to download: %v", err))
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(io.LimitReader(resp.Body, 1*1024*1024)) // 1MB limit
		if err != nil {
			sendExtensionError(conn, "Failed to read response")
			return
		}

		destPath := filepath.Join(extensionsDir, filename)
		if err := os.WriteFile(destPath, body, 0644); err != nil {
			sendExtensionError(conn, "Failed to save file")
			return
		}

		display.Log(fmt.Sprintf("Extensions: added %s", filename))

		result := map[string]interface{}{
			"type": "extensionAdded",
			"extension": ExtensionFile{
				Filename: filename,
				Source:   string(body),
			},
		}
		data, _ := json.Marshal(result)
		conn.WriteMessage(websocket.TextMessage, data)
	}()
}

// HandleAddExtensionFromFile opens a file picker for .js files and copies it to extensions.
func HandleAddExtensionFromFile(conn *websocket.Conn) {
	go func() {
		filePath, err := pickJSFile()
		if err != nil {
			display.Log(fmt.Sprintf("Extensions: file picker failed: %v", err))
			sendExtensionError(conn, "File picker failed")
			return
		}
		if filePath == "" {
			// User cancelled
			sendExtensionError(conn, "")
			return
		}

		os.MkdirAll(extensionsDir, os.ModePerm)

		filename := filepath.Base(filePath)
		if !strings.HasSuffix(strings.ToLower(filename), ".js") {
			sendExtensionError(conn, "File must be a .js file")
			return
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			sendExtensionError(conn, "Failed to read file")
			return
		}
		if len(data) > 1*1024*1024 {
			sendExtensionError(conn, "File too large (max 1MB)")
			return
		}

		destPath := filepath.Join(extensionsDir, filename)
		if err := os.WriteFile(destPath, data, 0644); err != nil {
			sendExtensionError(conn, "Failed to save file")
			return
		}

		display.Log(fmt.Sprintf("Extensions: added %s from file", filename))

		result := map[string]interface{}{
			"type": "extensionAdded",
			"extension": ExtensionFile{
				Filename: filename,
				Source:   string(data),
			},
		}
		resp, _ := json.Marshal(result)
		conn.WriteMessage(websocket.TextMessage, resp)
	}()
}

// HandleRemoveExtension deletes a JS file from the extensions directory.
func HandleRemoveExtension(conn *websocket.Conn, raw json.RawMessage) {
	var msg RemoveExtensionMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		sendExtensionError(conn, "Invalid request")
		return
	}

	// Sanitize: prevent path traversal
	filename := filepath.Base(msg.Filename)
	if filename == "." || filename == ".." {
		sendExtensionError(conn, "Invalid filename")
		return
	}

	path := filepath.Join(extensionsDir, filename)
	if err := os.Remove(path); err != nil {
		display.Log(fmt.Sprintf("Extensions: remove failed: %v", err))
		sendExtensionError(conn, "Failed to remove extension")
		return
	}

	display.Log(fmt.Sprintf("Extensions: removed %s", filename))

	result := map[string]interface{}{
		"type":     "extensionRemoved",
		"filename": filename,
	}
	data, _ := json.Marshal(result)
	conn.WriteMessage(websocket.TextMessage, data)
}

func sendExtensionError(conn *websocket.Conn, message string) {
	resp := map[string]interface{}{
		"type":    "status",
		"status":  "error",
		"message": message,
	}
	data, _ := json.Marshal(resp)
	conn.WriteMessage(websocket.TextMessage, data)
}

// --- Win32 file dialog for .js files ---

var (
	comdlg32             = syscall.NewLazyDLL("comdlg32.dll")
	procGetOpenFileNameW = comdlg32.NewProc("GetOpenFileNameW")
	user32               = syscall.NewLazyDLL("user32.dll")
	procGetForegroundWnd = user32.NewProc("GetForegroundWindow")
)

const (
	ofnFileMustExist = 0x00001000
	ofnPathMustExist = 0x00000800
	ofnNoChangeDir   = 0x00000008
)

type openFileNameW struct {
	lStructSize       uint32
	hwndOwner         uintptr
	hInstance         uintptr
	lpstrFilter       *uint16
	lpstrCustomFilter *uint16
	nMaxCustFilter    uint32
	nFilterIndex      uint32
	lpstrFile         *uint16
	nMaxFile          uint32
	lpstrFileTitle    *uint16
	nMaxFileTitle     uint32
	lpstrInitialDir   *uint16
	lpstrTitle        *uint16
	flags             uint32
	nFileOffset       uint16
	nFileExtension    uint16
	lpstrDefExt       *uint16
	lCustData         uintptr
	lpfnHook          uintptr
	lpTemplateName    *uint16
	pvReserved        uintptr
	dwReserved        uint32
	flagsEx           uint32
}

func pickJSFile() (string, error) {
	filterUTF16, _ := syscall.UTF16FromString("JavaScript Files (*.js)|*.js|All Files (*.*)|*.*|")
	for i := range filterUTF16 {
		if filterUTF16[i] == '|' {
			filterUTF16[i] = 0
		}
	}
	titleUTF16, _ := syscall.UTF16FromString("Select Extension File")

	var fileBuf [260]uint16
	hwnd, _, _ := procGetForegroundWnd.Call()

	ofn := openFileNameW{
		hwndOwner:    hwnd,
		lpstrFilter:  &filterUTF16[0],
		nFilterIndex: 1,
		lpstrFile:    &fileBuf[0],
		nMaxFile:     uint32(len(fileBuf)),
		lpstrTitle:   &titleUTF16[0],
		flags:        ofnFileMustExist | ofnPathMustExist | ofnNoChangeDir,
	}
	ofn.lStructSize = uint32(unsafe.Sizeof(ofn))

	ret, _, _ := procGetOpenFileNameW.Call(uintptr(unsafe.Pointer(&ofn)))
	if ret == 0 {
		return "", nil
	}
	return syscall.UTF16ToString(fileBuf[:]), nil
}

func filenameFromURL(rawURL string) string {
	parts := strings.Split(rawURL, "/")
	if len(parts) == 0 {
		return ""
	}
	name := parts[len(parts)-1]
	// Strip query string
	if idx := strings.IndexByte(name, '?'); idx >= 0 {
		name = name[:idx]
	}
	return name
}
