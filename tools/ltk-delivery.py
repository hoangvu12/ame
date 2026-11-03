#!/usr/bin/env python3
"""
ltk-delivery — download the current LTK patcher DLL from LeagueToolkit's delivery service.

Usage:
  python ltk-delivery.py fetch [--output <dir>]     Download patcher DLL
  python ltk-delivery.py refresh                     Extract fresh attestation from latest release

Protocol:
  1. GET  /v1/challenge?install=<uuid>     → challenge + difficulty
  2. POST /v1/installs                     → register with PoW solution + attestation → bearer token
  3. GET  /v1/offer?channel=stable&manager=<ver>  → minisign-signed manifest
  4. GET  /v1/dll/<sha256>                 → patcher DLL (SHA-256 verified)

The attestation is an Ed25519-signed token baked into every official ltk-manager.exe.
It sits in the binary's data section immediately after the minisign manifest public key.
Format: base64url({"manager":"X.Y.Z","commit":"<sha>"}).base64url(Ed25519_64byte_sig)

Rotation handling:
  - Key rotation: downloading the latest LTK Manager release yields a fresh attestation.
  - Version deprecation: the manager version in the attestation statement is used directly;
    no hardcoded version string survives a refresh.
  - 7z on PATH required for NSIS extraction during refresh.

Requires: Python 3.8+, requests.
"""

import argparse, base64, hashlib, json, os, re, shutil, subprocess, sys, tempfile, time, uuid
from pathlib import Path

import requests as req

BASE_URL = "https://patcher.leaguetoolkit.dev"
MANIFEST_KEY = "RWSk74swcHfkZfLblm3iSk9RLVHJ1xjtqn/zySIPaCBM1hecnlFy1Bxi"
RELEASES_URL = "https://api.github.com/repos/LeagueToolkit/ltk-manager/releases/latest"
MAX_DIFFICULTY = 24

# Ed25519 sig = 64 bytes = 86 base64url chars (no padding).
SIG_B64_LEN = 86

# Serves as initial state; overwritten on first refresh.
FALLBACK_ATTESTATION = (
    "eyJtYW5hZ2VyIjoiMS4zMi4yIiwiY29tbWl0Ijoi"
    "NjJhZmY0ODJjMThmZTI0ZGFjM2ZkYWU3MGEzMTY1OTUyMDRkOWVmNCJ9"
    ".55MKhasBjui5ix8dWaPZYi9HizK5Sap2wGN-btye1e6"
    "fjsegtM6zsb0Z7TMSXxsweOyiLpJ3ZRQJpJwX25GCBA"
)


# ---------------------------------------------------------------------------
#  Attestation
# ---------------------------------------------------------------------------

class Attestation:
    """Parsed attestation: the full token and the manager version it names."""

    def __init__(self, token: str):
        self.token = token
        self.manager = self._parse_manager(token)

    @staticmethod
    def _parse_manager(token: str) -> str:
        parts = token.split(".")
        if len(parts) != 2:
            return "1.32.2"
        try:
            stmt = json.loads(_b64url_decode(parts[0]))
            return stmt.get("manager", "1.32.2")
        except Exception:
            return "1.32.2"

    @property
    def user_agent(self) -> str:
        return f"ltk-manager/{self.manager}"


def _b64url_decode(s: str) -> bytes:
    return base64.urlsafe_b64decode(s + "=" * (-len(s) % 4))


def load_attestation(cache_dir: str | None = None) -> Attestation:
    if cache_dir:
        token_path = Path(cache_dir) / "attestation.txt"
        ver_path = Path(cache_dir) / "attestation_version.txt"
        if token_path.exists() and ver_path.exists():
            return Attestation(token_path.read_text().strip())
    return Attestation(FALLBACK_ATTESTATION)


def save_attestation(cache_dir: str, att: Attestation):
    p = Path(cache_dir)
    p.mkdir(parents=True, exist_ok=True)
    (p / "attestation.txt").write_text(att.token)
    (p / "attestation_version.txt").write_text(att.manager)


# ---------------------------------------------------------------------------
#  Proof-of-work
# ---------------------------------------------------------------------------

def _leading_zero_bits(digest: bytes) -> int:
    bits = 0
    for byte in digest:
        if byte != 0:
            return bits + (8 - byte.bit_length())
        bits += 8
    return bits


def solve_pow(challenge: str, difficulty: int) -> str:
    for n in range(2**32):
        nonce = f"{n:x}"
        h = hashlib.sha256((challenge + nonce).encode()).digest()
        if _leading_zero_bits(h) >= difficulty:
            return nonce
    raise RuntimeError("nonce space exhausted")


# ---------------------------------------------------------------------------
#  Delivery protocol
# ---------------------------------------------------------------------------

def _check(resp, label: str):
    if resp.status_code != 200:
        body = resp.text[:500]
        raise RuntimeError(f"{label}: HTTP {resp.status_code}: {body}")


def delivery_register(session, att: Attestation) -> str:
    install_id = str(uuid.uuid4())
    # Challenge
    r = session.get(f"{BASE_URL}/v1/challenge", params={"install": install_id}, timeout=30)
    _check(r, "challenge")
    ch = r.json()
    if ch["difficulty"] > MAX_DIFFICULTY:
        raise RuntimeError(f"difficulty {ch['difficulty']} > {MAX_DIFFICULTY}")
    nonce = solve_pow(ch["challenge"], ch["difficulty"])
    # Register
    r = session.post(
        f"{BASE_URL}/v1/installs",
        json={
            "install": install_id,
            "manager": att.manager,
            "attestation": att.token,
            "challenge": ch["challenge"],
            "nonce": nonce,
        },
        timeout=30,
    )
    _check(r, "register")
    return r.json()["token"]


def delivery_offer(session, token: str, manager: str) -> dict:
    r = session.get(
        f"{BASE_URL}/v1/offer",
        params={"channel": "stable", "manager": manager},
        headers={"Authorization": f"Bearer {token}"},
        timeout=30,
    )
    if r.status_code == 409:
        raise RuntimeError(f"manager {manager} not served by any release range")
    _check(r, "offer")
    envelope = r.json()["manifest"]
    manifest_bytes = base64.b64decode(envelope["manifest"])
    manifest = json.loads(manifest_bytes)
    # Verify trusted comment
    expected = f"trusted comment: {manifest['channel']} {manifest['sequence']} {manifest['release']['version']}"
    if expected not in envelope["signature"]:
        raise RuntimeError(f"trusted comment mismatch: {expected} not in signature")
    return manifest


def delivery_download(session, token: str, release: dict, dst: str) -> str:
    r = session.get(
        f"{BASE_URL}/v1/dll/{release['digest']}",
        headers={"Authorization": f"Bearer {token}"},
        timeout=120,
        stream=True,
    )
    _check(r, "download")
    h = hashlib.sha256()
    written = 0
    with open(dst, "wb") as f:
        for chunk in r.iter_content(8192):
            f.write(chunk)
            h.update(chunk)
            written += len(chunk)
    if written != release["size"]:
        raise RuntimeError(f"size mismatch: {written} != {release['size']}")
    if h.hexdigest() != release["digest"]:
        raise RuntimeError(f"digest mismatch: {h.hexdigest()} != {release['digest']}")
    return dst


# ---------------------------------------------------------------------------
#  Attestation refresh (carve from latest LTK Manager release)
# ---------------------------------------------------------------------------

def refresh_attestation(cache_dir: str | None = None) -> tuple[Attestation, str]:
    """Download latest LTK Manager, carve attestation, return (att, tag)."""
    # 1. Latest release info
    r = req.get(
        RELEASES_URL,
        headers={
            "Accept": "application/vnd.github+json",
            "X-GitHub-Api-Version": "2022-11-28",
            "User-Agent": "ltk-manager/1.32.2",
        },
        timeout=30,
    )
    r.raise_for_status()
    release_data = r.json()
    tag = release_data["tag_name"]

    setup_url = None
    for asset in release_data["assets"]:
        if asset["name"].endswith("-setup.exe") and not asset["name"].endswith(".sig"):
            setup_url = asset["browser_download_url"]
            break
    if not setup_url:
        raise RuntimeError(f"no setup.exe in {tag}")

    # 2. Download + extract
    tmp = tempfile.mkdtemp(prefix="ltk-refresh-")
    try:
        setup_path = os.path.join(tmp, "setup.exe")
        r = req.get(setup_url, timeout=120)
        r.raise_for_status()
        with open(setup_path, "wb") as f:
            f.write(r.content)

        extract_dir = os.path.join(tmp, "extracted")
        os.makedirs(extract_dir, exist_ok=True)
        subprocess.run(
            ["7z", "x", f"-o{extract_dir}", "-y", setup_path],
            capture_output=True,
            check=True,
        )

        exe_path = None
        for root, _dirs, files in os.walk(extract_dir):
            for f in files:
                if f.lower() == "ltk-manager.exe":
                    exe_path = os.path.join(root, f)
                    break
            if exe_path:
                break
        if not exe_path:
            raise RuntimeError("ltk-manager.exe not found in extracted installer")

        att = carve_attestation(exe_path)
        save_attestation(cache_dir, att)
        return att, tag
    finally:
        shutil.rmtree(tmp, ignore_errors=True)


def carve_attestation(exe_path: str) -> Attestation:
    """Carve the attestation token from a ltk-manager.exe binary."""
    data = Path(exe_path).read_bytes()

    pk_idx = data.find(MANIFEST_KEY.encode())
    if pk_idx < 0:
        raise RuntimeError("manifest public key not found in binary")

    search_start = pk_idx + len(MANIFEST_KEY)
    att_rel = data[search_start:].find(b"eyJ")
    if att_rel < 0:
        raise RuntimeError("attestation header not found in binary")
    att_start = search_start + att_rel

    # eyJ<stmt>.{86 Ed25519 sig}
    m = re.match(
        rb"^(eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]{" + str(SIG_B64_LEN).encode() + rb"})",
        data[att_start:],
    )
    if not m:
        raise RuntimeError("attestation format not recognized")
    return Attestation(m.group(1).decode())


# ---------------------------------------------------------------------------
#  Host exe extraction
# ---------------------------------------------------------------------------

def download_host_exe(dst_dir: str) -> str:
    """Extract ltk_patcher_host.exe from the latest LTK Manager installer."""
    host_path = os.path.join(dst_dir, "ltk_patcher_host.exe")
    if os.path.exists(host_path):
        return host_path

    r = req.get(RELEASES_URL, headers={
        "Accept": "application/vnd.github+json",
        "X-GitHub-Api-Version": "2022-11-28",
        "User-Agent": "ltk-manager/1.32.2",
    }, timeout=30)
    r.raise_for_status()
    release_data = r.json()

    setup_url = None
    for asset in release_data["assets"]:
        if asset["name"].endswith("-setup.exe") and not asset["name"].endswith(".sig"):
            setup_url = asset["browser_download_url"]
            break
    if not setup_url:
        raise RuntimeError("no setup.exe")

    tmp = tempfile.mkdtemp(prefix="ltk-host-")
    try:
        setup_path = os.path.join(tmp, "setup.exe")
        r = req.get(setup_url, timeout=120)
        r.raise_for_status()
        with open(setup_path, "wb") as f:
            f.write(r.content)

        extract_dir = os.path.join(tmp, "extracted")
        os.makedirs(extract_dir, exist_ok=True)
        subprocess.run(["7z", "x", f"-o{extract_dir}", "-y", setup_path],
                       capture_output=True, check=True)

        for root, _dirs, files in os.walk(extract_dir):
            for f in files:
                if f.lower() == "ltk_patcher_host.exe":
                    src = os.path.join(root, f)
                    os.makedirs(dst_dir, exist_ok=True)
                    shutil.copy2(src, host_path)
                    return host_path
        raise RuntimeError("ltk_patcher_host.exe not found")
    finally:
        shutil.rmtree(tmp, ignore_errors=True)


# ---------------------------------------------------------------------------
#  CLI
# ---------------------------------------------------------------------------

def cmd_fetch(args):
    output = args.output or os.getcwd()
    os.makedirs(output, exist_ok=True)

    att = load_attestation(output)
    s = req.Session()
    s.headers["User-Agent"] = att.user_agent

    print(f"[1/4] Registering (manager {att.manager})...")
    try:
        token = delivery_register(s, att)
    except RuntimeError as e:
        if "HTTP 401" in str(e) or "HTTP 403" in str(e):
            print("  Attestation rejected. Refreshing from latest release...")
            att, tag = refresh_attestation(output)
            print(f"  Fresh attestation from {tag} (manager {att.manager})")
            s.headers["User-Agent"] = att.user_agent
            token = delivery_register(s, att)
        else:
            raise

    print(f"[2/4] Fetching manifest...")
    manifest = delivery_offer(s, token, att.manager)
    release = manifest["release"]
    print(f"  Version:  {release['version']}")
    print(f"  Digest:   {release['digest']}")
    print(f"  Size:     {release['size']:,} bytes")
    print(f"  EOL:      {time.strftime('%Y-%m-%d', time.gmtime(release['eol']))}")

    dll_path = os.path.join(output, "ltk_patcher_dll.dll")
    if os.path.exists(dll_path):
        existing = hashlib.sha256(Path(dll_path).read_bytes()).hexdigest()
        if existing == release["digest"]:
            print(f"[3/4] DLL current at {dll_path}")
            print(f"[4/4] Done. {release['version']}")
            return

    print(f"[3/4] Downloading DLL ({release['size'] / 1024:.0f} KB)...")
    delivery_download(s, token, release, dll_path)
    print(f"  Saved to {dll_path}")
    print(f"[4/4] Done. {release['version']}")


def cmd_refresh(args):
    output = args.output or os.getcwd()
    att, tag = refresh_attestation(output)
    print(f"Attestation from {tag}:")
    print(att.token)
    parts = att.token.split(".")
    stmt = json.loads(_b64url_decode(parts[0]))
    sig = _b64url_decode(parts[1])
    print(f"  Manager:   {stmt['manager']}")
    print(f"  Commit:    {stmt['commit']}")
    print(f"  Sig:       {len(sig)} bytes Ed25519")


def cmd_host(args):
    output = args.output or os.getcwd()
    path = download_host_exe(output)
    print(f"Host exe: {path}")


if __name__ == "__main__":
    ap = argparse.ArgumentParser(description="LTK patcher DLL delivery client")
    sp = ap.add_subparsers(dest="command")

    f = sp.add_parser("fetch", help="Download patcher DLL")
    f.add_argument("--output", "-o", help="Output directory")

    r = sp.add_parser("refresh", help="Extract fresh attestation from latest release")
    r.add_argument("--output", "-o", help="Cache directory for attestation")

    h = sp.add_parser("host", help="Extract ltk_patcher_host.exe from latest release")
    h.add_argument("--output", "-o", help="Output directory")

    args = ap.parse_args()
    if args.command == "fetch":
        cmd_fetch(args)
    elif args.command == "refresh":
        cmd_refresh(args)
    elif args.command == "host":
        cmd_host(args)
    else:
        ap.print_help()