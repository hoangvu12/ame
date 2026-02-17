package formicons

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed icons
var icons embed.FS

// ServeFormIcon handles GET /form-icon/{id}.png requests.
func ServeFormIcon(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/form-icon/")
	if name == "" {
		http.NotFound(w, r)
		return
	}

	sub, _ := fs.Sub(icons, "icons")
	data, err := fs.ReadFile(sub, name)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Write(data)
}
