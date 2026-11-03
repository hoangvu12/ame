runtime.zip is the committed, prepared LTK runtime bundle: the unmodified
signed ltk_patcher_host.exe / ltk_patcher_dll.dll (hash- and
signature-pinned), our ame-ltk-overlay.exe helper, license/notice files and
the helper's source. A plain `go build` embeds it; no toolchain beyond Go
is required.

The bundled host and DLL are the offline fallback. On a fresh host start,
internal/ltk/delivery.go registers with LeagueToolkit's delivery service,
replaces ltk_patcher_dll.dll with the current signed build, and refreshes
ltk-manager.exe from the newest LTK Manager release (needs 7z on PATH).
Delivery failures are logged and the bundled files are used instead, so
Prepare() deliberately skips hash validation for those two files.

Rebuild the zip only when deliberately updating the fallback runtime
revision or the helper: run tools/ltk-runtime/build.ps1 (needs Rust and
Cargo.lock pins dependencies), then commit the regenerated zip together with
the updated hashes in that script.