package i18n

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hoangvu12/ame/internal/lcu"
)

const fallbackLocale = "en_US"

var currentLocale = fallbackLocale

var messages = map[string]map[string]string{
	"en_US": {
		"display.label.status":  "Status",
		"display.label.skin":    "Skin",
		"display.label.overlay": "Overlay",
		"display.label.party":   "Party",
		"display.label.log":     "Log",

		"display.value.waiting_client":       "Waiting for client",
		"display.value.connected":            "Connected",
		"display.value.none":                 "None",
		"display.value.overlay_active":       "Active",
		"display.value.overlay_inactive":     "Inactive",
		"display.value.party_off":            "Off",
		"display.value.party_in_room_waiting": "In room (waiting)",
		"display.value.party_in_room_teammates": "In room ({count} teammates)",
		"display.value.party_active_users":     "Active ({count} Ame users)",
	},
	"vi_VN": {
		"display.label.status":  "Trạng thái",
		"display.label.skin":    "Trang phục",
		"display.label.overlay": "Lớp phủ",
		"display.label.party":   "Nhóm",
		"display.label.log":     "Nhật ký",

		"display.value.waiting_client":       "Đang chờ client",
		"display.value.connected":            "Đã kết nối",
		"display.value.none":                 "Không có",
		"display.value.overlay_active":       "Hoạt động",
		"display.value.overlay_inactive":     "Tắt",
		"display.value.party_off":            "Tắt",
		"display.value.party_in_room_waiting": "Trong phòng (đang chờ)",
		"display.value.party_in_room_teammates": "Trong phòng ({count} đồng đội)",
		"display.value.party_active_users":     "Hoạt động ({count} người dùng Ame)",
	},
}

var localeRe = regexp.MustCompile(`^[a-z]{2}_[A-Z]{2}$`)

func normalizeLocale(loc string) string {
	loc = strings.TrimSpace(loc)
	loc = strings.ReplaceAll(loc, "-", "_")
	if loc == "" {
		return ""
	}
	parts := strings.Split(loc, "_")
	if len(parts) == 1 {
		return strings.ToLower(parts[0])
	}
	return strings.ToLower(parts[0]) + "_" + strings.ToUpper(parts[1])
}

func resolveLocale(preferred string) string {
	normalized := normalizeLocale(preferred)
	if normalized != "" {
		if _, ok := messages[normalized]; ok && localeRe.MatchString(normalized) {
			return normalized
		}
		if len(normalized) >= 2 {
			lang := normalized[:2]
			for loc := range messages {
				if strings.HasPrefix(strings.ToLower(loc), lang+"_") {
					return loc
				}
			}
		}
	}
	if _, ok := messages[fallbackLocale]; ok {
		return fallbackLocale
	}
	for loc := range messages {
		return loc
	}
	return fallbackLocale
}

// Init detects the current locale using the running League client.
// If detection fails, it falls back to English.
func Init() {
	locale, err := lcu.GetRegionLocale()
	if err != nil {
		currentLocale = resolveLocale("")
		return
	}
	currentLocale = resolveLocale(locale)
}

func T(key string, vars map[string]interface{}) string {
	if key == "" {
		return ""
	}
	msgs, ok := messages[currentLocale]
	if !ok {
		msgs = messages[fallbackLocale]
	}
	out, ok := msgs[key]
	if !ok {
		out = messages[fallbackLocale][key]
	}
	if out == "" {
		return key
	}
	if vars == nil || len(vars) == 0 {
		return out
	}
	for k, v := range vars {
		out = strings.ReplaceAll(out, "{"+k+"}", fmt.Sprint(v))
	}
	return out
}

func Locale() string {
	return currentLocale
}
