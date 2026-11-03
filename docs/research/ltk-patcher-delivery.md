# Research: keeping the LTK patcher current (and why delivery is blocked)

Date: November 2026, tested against `ltk_patcher-26` (LTK Manager v1.32.2).

## Goal

`internal/ltk/assets/runtime.zip` pins a signed `ltk_patcher_host.exe` +
`ltk_patcher_dll.dll` pair. When Riot patches the game, the pinned DLL can stop
working and ame needs a new bundle + release. The experiment was to fetch the
current signed DLL from LeagueToolkit's delivery service at startup and to
refresh the host exe from the latest LTK Manager release, so the pinned bundle
becomes only an offline fallback.

## The delivery protocol (worked)

Reference implementation: `tools/ltk-delivery.py`
(`python ltk-delivery.py fetch --output .`). All of it works:

1. `GET /v1/challenge?install=<uuid>` → challenge + difficulty (≤ 24 bits)
2. `POST /v1/installs` with the PoW nonce and an Ed25519 **attestation** carved
   from any official `ltk-manager.exe` → bearer token
3. `GET /v1/offer?channel=stable&manager=<ver>` → base64 manifest + minisign
   envelope (DLL digest/size/version, protocol, manager ranges, EOL)
4. `GET /v1/dll/<sha256>` → patcher DLL, verified against the manifest digest

The attestation is `base64url({"manager","commit"}).base64url(sig)`, located in
the manager binary right after the minisign manifest public key
(`RWSk74swcHfkZfLblm3iSk9RLVHJ1xjtqn/zySIPaCBM1hecnlFy1Bxi`). Manager-version
rotation is handled by re-carving from the newest release.

A Go implementation of this reached the point of delivering
`ltk_patcher-26` (835,848 bytes, digest-verified) and exporting the host exe
from the release NSIS installer, launched as `ltk-manager.exe` with a
medium-integrity token so the host stays unelevated beneath an elevated ame.

## Why it cannot be used: the DLL blocks third-party hosts

Once the game started, the delivered DLL was injected and then refused to
overlay:

```
ERROR ltk_patcher_dll::verify: host verification failed, refusing overlay:
[VRF00000008] (3932,ltk_patcher_host.exe)
```

This is deliberate. LTK's maintainer answered the same report
(<https://github.com/LeagueToolkit/ltk-manager/issues/923>):

> This check is meant to block third-party apps from using the DLL.

The check arrived in Manager **1.30.0** (`ltk_patcher-25`) along with
`crates/ltk_patcher_dll/src/entry/imp_windows.rs`, whose string table pins the
host identity (`ltk-manager.exe`, `ltk_patcher_host.exe`, `\\?\GLOBALROOT…`,
`WinVerifyTrust`) and reports `VRF00000001`–`VRF00000009`. Renaming the host
file is not enough — the host reports its own image name, and the pinned bundle
in `runtime.zip` (764,160-byte DLL) simply predates the check: it contains no
`host verification failed` / `VRF0000000*` strings at all, which is why the
pinned pair keeps working.

## Decision

Do **not** ship the delivered DLL. The pinned `runtime.zip` pair is the runtime;
update it with `tools/ltk-runtime/build.ps1` when a game patch breaks it, and
release (see `CLAUDE.md` for the release flow). Bypassing `VRF00000008` would
mean tampering with an explicit anti-third-party control inside the game
process — risk to users (Vanguard) and to the project, so it is out of scope.

`tools/ltk-delivery.py` stays as the protocol reference for inspecting what
upstream ships; it is not part of the ame build.