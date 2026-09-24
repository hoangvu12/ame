// Package zipname holds AME's single rule for validating archive entry
// names before they are used as filesystem paths.
//
// Entry names are untrusted input: a crafted archive can name a file
// outside its destination directory (zip slip). Every unpack path — the
// bundled runtime, generated skin packages, and extracted mods — shares
// this rule so the checks cannot drift apart again.
package zipname

import "strings"

// Unsafe reports whether an archive entry name may escape its destination:
// absolute paths, Windows drive or alternate-data-stream colons, and
// parent-directory references. The ".." check is deliberately strict — any
// occurrence is rejected, not only exact path segments — because AME
// archives contain only plain relative names.
func Unsafe(name string) bool {
	name = strings.ReplaceAll(name, `\`, "/")
	return strings.HasPrefix(name, "/") ||
		strings.Contains(name, ":") ||
		strings.Contains(name, "..")
}
