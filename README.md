# ame

A League of Legends skin changer for Windows. Lets you select and apply skins directly in the client during champion select. Heavily inspired by [Rose](https://github.com/Alban1911/Rose).

## How it works

1. Detects your League of Legends installation automatically
2. Sets up Pengu Loader and the bundled LTK runtime on first run
3. Injects a plugin into the League client that adds skin selection UI
4. Communicates over WebSocket between the plugin and the Go backend
5. Generates supported skins from installed game files and uses the signed LTK runtime to apply them

## Requirements

- Windows
- League of Legends installed
- Administrator privileges

## Installation

Download `ame.exe` from the [releases page](https://github.com/hoangvu12/ame/releases) and run it. First-run setup is automatic.

To uninstall, run `uninstall.bat` included in the release.

## Launch diagnostic logs

AME keeps launch, overlay, suspend/resume, and Unstuck diagnostics in
`%LOCALAPPDATA%\ame\logs\launch.jsonl`. Two rotated backups retain recent history
across application restarts and normal PC reboots, using approximately 1.5 MB total.
Use **Download logs** in AME settings to export retained backend history together
with the current plugin diagnostics. Routine skin selection and cache messages
are omitted. If a launch freezes, export after recovery and include the approximate
time of the failure. An abrupt power loss can still lose the most recent writes.

## Building from source

Requires Go 1.21+, Rust, and PowerShell on Windows.

```
./tools/ltk-runtime/build.ps1
go mod tidy
go build -trimpath -ldflags="-s -w" -o dist/ame.exe ./cmd/ame
```

The skin generator bundle is built and distributed separately: AME downloads
`skin-generator.zip` from the latest release on first use. For development
builds, point `AME_SKIN_GENERATOR_ZIP` at a local bundle zip instead.

Releases can also serve pre-generated skin packages from a remote catalog,
tried before on-device generation. Set `AME_SKIN_CDN_URL` and
`AME_SKIN_CDN_TOKEN` to point a build at one; without them ame generates
locally as before.

## Project structure

```
cmd/ame/        Main application entry point
internal/
  config/       Settings persistence
  game/         Game directory detection
  modtools/     Overlay runtime dispatch
  ltk/          Bundled signed host and overlay builder
  server/       WebSocket server
  setup/        First-run setup
  skin/         Skin download and extraction
  updater/      Auto-update from GitHub releases
src/            League client plugin (JavaScript)
```

## Features

- Auto-detects League of Legends game directory
- Skin selection UI injected into champion select
- Installed skin and chroma generation, including inherited animations and the 30 alternate forms listed in AME. Gear forms are fixed selections; original in-game progression/form-switching systems are not implemented.
- Auto-apply for ARAM and other game modes
- System tray with show/hide controls
- Automatic updates from GitHub releases

## Credits

- [Pengu Loader](https://github.com/PenguLoader/PenguLoader) — JavaScript plugin loader for the League of Legends Client
- [Rose](https://github.com/Alban1911/Rose) — Inspiration for this project

## Disclaimer

This is an unofficial fan-made tool. It is not endorsed by Riot Games. Use at your own risk.

## Runtime notes

The runtime is bundled with AME and applies generated packages through a
hash-pinned mod runtime; its build inputs live in `tools/ltk-runtime`. Audio
and visual output of generated skins varies per skin; locale WADs participate
in package cache invalidation. Unsupported local conversions return an error
instead of applying a stale package.

`go test ./...` covers the normal unit suite. Set `AME_TEST_GAME_DIR` to the installed `Game` directory for
offline generator/cache/combined-overlay tests. `AME_TEST_LTK=1` enables a signed
host configuration-only test; it does not scan for or attach to a game.
