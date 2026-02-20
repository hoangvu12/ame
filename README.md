# ame

A League of Legends skin changer for Windows. Lets you select and apply skins directly in the client during champion select. Heavily inspired by [Rose](https://github.com/Alban1911/Rose). Skins are sourced from [LeagueSkins](https://github.com/Alban1911/LeagueSkins/tree/main/skins).

## How it works

1. Detects your League of Legends installation automatically
2. Downloads and sets up Pengu Loader and mod-tools on first run
3. Injects a plugin into the League client that adds skin selection UI
4. Communicates over WebSocket between the plugin and the Go backend
5. Uses mod-tools to apply the selected skin before the game starts

## Requirements

- Windows
- League of Legends installed
- Administrator privileges (required for mod-tools)

## Installation

Download `ame.exe` from the [releases page](https://github.com/hoangvu12/ame/releases) and run it. First-run setup is automatic.

To uninstall, run `uninstall.bat` included in the release.

## Building from source

Requires Go 1.21+ and 7z.

```
go mod tidy
go build -trimpath -ldflags="-s -w" -o dist/ame.exe ./cmd/ame
```

## Project structure

```
cmd/ame/        Main application entry point
internal/
  config/       Settings persistence
  game/         Game directory detection
  modtools/     mod-tools integration
  server/       WebSocket server
  setup/        First-run setup
  skin/         Skin download and extraction
  updater/      Auto-update from GitHub releases
src/            League client plugin (JavaScript)
```

## Features

- Auto-detects League of Legends game directory
- Skin selection UI injected into champion select
- Chroma variant support
- Auto-apply for ARAM and other game modes
- System tray with show/hide controls
- Automatic updates from GitHub releases

## Credits

- [Pengu Loader](https://github.com/PenguLoader/PenguLoader) — JavaScript plugin loader for the League of Legends Client
- [Rose](https://github.com/Alban1911/Rose) — Inspiration for this project
- [LeagueSkins](https://github.com/Alban1911/LeagueSkins) — Skin data source

## Disclaimer

This is an unofficial fan-made tool. It is not endorsed by Riot Games. Use at your own risk.
