runtime.zip is the committed, prepared LTK runtime bundle: the unmodified
signed ltk_patcher_host.exe / ltk_patcher_dll.dll (hash- and
signature-pinned), our ame-ltk-overlay.exe helper, license/notice files and
the helper's source. A plain `go build` embeds it; no toolchain beyond Go
is required.

The bundled host and DLL are the offline fallback. On a fresh manager start,
internal/ltk/delivery.go registers with LeagueToolkit's delivery service and
replaces both files with the current signed pair (DLL from the service, host
and ltk-manager.exe from the newest LTK Manager release installer; needs 7z
on PATH). Delivery failures are logged and the bundled files are used
instead, so Prepare() deliberately skips hash validation for those two files
and never overwrites a delivered pair. ltk-manager.exe itself is never
bundled: ame launches the copy DeliveryFetch extracted from the installer
and drives it over the WebView2 DevTools protocol (internal/ltk/cdp.go).

The manager runs hidden with --shell-relaunched (so it stays ame's child and
keeps ame's environment) and unelevated (see LTK ADR-0067 and ADR-0074): the
DLL verifies its host by image name and signature and refuses an elevated
one, so ame hands exec.Cmd a medium-integrity token.

Rebuild the zip only when deliberately updating the fallback runtime
revision or the helper: run tools/ltk-runtime/build.ps1 (needs Rust and
Cargo.lock pins dependencies), then commit the regenerated zip together with
the updated hashes in that script.
