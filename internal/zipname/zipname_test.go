package zipname

import "testing"

func TestUnsafe(t *testing.T) {
	cases := []struct {
		name string
		want bool
	}{
		{"WAD/characters/ahri/skin0.bin", false},
		{"META/info.json", false},
		{"ltk_patcher_host.exe", false},
		{"/etc/passwd", true},         // absolute path
		{`C:\Windows\System32`, true}, // drive colon
		{`C:/Windows/System32`, true}, // drive colon, forward slash
		{"WAD/data:a", true},          // ADS colon
		{"WAD/../escape.bin", true},   // parent segment
		{"..", true},                  // bare parent
		{"..\\escape", true},          // parent, backslash separator
		{"WAD/../../escape", true},    // repeated parents
		{"weird..name.bin", true},     // strict: any ".." occurrence
		{"ltrs/./here.bin", false},    // "." segment is not unsafe
	}
	for _, c := range cases {
		if got := Unsafe(c.name); got != c.want {
			t.Errorf("Unsafe(%q) = %v, want %v", c.name, got, c.want)
		}
	}
}
