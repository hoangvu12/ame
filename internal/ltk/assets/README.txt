runtime.zip is the committed, prepared LTK runtime bundle: the unmodified
signed ltk_patcher_host.exe / ltk_patcher_dll.dll (hash- and
signature-pinned), our ame-ltk-overlay.exe helper, license/notice files and
the helper's source. A plain `go build` embeds it; no toolchain beyond Go
is required.

The bundled host and DLL pair is the offline fallback of the old
architecture; the manager runtime ame actually launches never comes from
this zip. On a fresh manager start, internal/ltk/delivery.go downloads the
newest LTK Manager release from GitHub and extracts the signed
ltk-manager.exe plus its bundled ltk_patcher_host.exe into the shared
%LOCALAPPDATA%\ame\manager cache with a pure-Go NSIS reader (internal/
ltk/nsis.go) — no 7-Zip on PATH is required, and nothing is re-downloaded
while the release tag matches. The patcher DLL itself is delivered by the
manager, from LeagueToolkit's delivery service, when its patcher starts.

Prepare() deliberately skips hash validation for the bundled host and DLL
and never overwrites them: a runtime directory written by an older ame may
hold the pair its delivery service installed.

The manager runs hidden with --shell-relaunched (so it stays ame's child and
keeps ame's environment) and unelevated (see LTK ADR-0067 and ADR-0074): the
DLL verifies its host by image name and signature and refuses an elevated
one, so ame hands exec.Cmd a medium-integrity token.

Rebuild the zip only when deliberately updating the fallback runtime
revision or the helper: run tools/ltk-runtime/build.ps1 (needs Rust and
Cargo.lock pins dependencies), then commit the regenerated zip together with
the updated hashes in that script.
