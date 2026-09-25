runtime.zip is the committed, prepared LTK runtime bundle: the unmodified
signed ltk_patcher_host.exe / ltk_patcher_dll.dll (hash- and
signature-pinned), our ame-ltk-overlay.exe helper, license/notice files and
the helper's source. A plain `go build` embeds it; no toolchain beyond Go
is required.

Rebuild it only when deliberately updating the LTK revision or the helper:
run tools/ltk-runtime/build.ps1 (needs Rust and Cargo.lock pins
dependencies), then commit the regenerated zip together with the updated
hashes in that script.
