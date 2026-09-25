# Research: central skin hosting (pre-generated, private distribution)

> **Decision (final):** GitHub private repo + Cloudflare Worker proxy — not R2.
> The ame-skins private repo (`../ame-skins`) contains the CDN worker, batch
> build runner, and publish tooling. Packages are stored as an orphan-branch
> snapshot on a **separate GitHub account/org** (isolating ame's own release
> distribution from any takedown), force-pushed per patch; the worker proxies
> raw.githubusercontent.com with a PAT secret, bearer-token auth, and edge
> caching; pre-warming after publish keeps GitHub rate limits (5k/h/IP raw)
> from ever being hit in steady state. The data repo lives on the same account
> (`hoangvu12/ame-skins-data`) — accepted trade-off over account isolation. ame got a remote-first `EnsurePackage`
> with local-generation fallback (see `internal/skin/remote.go`). Release
> builds inject the CDN URL/token via ldflags; source builds stay local-only.
> R2 remains the documented fallback if the GitHub route ever fails
> (`publish.py` works with any git remote; the worker backend is swappable).

Date: October 2025 (researched against current League patch 16.19)

## TL;DR

Generate all skins once per patch on a machine with a League install (using the
existing `skin-generator.exe`), store the `.fantome` packages in **Cloudflare R2**
(≈3–4 GB total, ~$0.05/month, egress free), and serve them through a small
**Cloudflare Worker** with token auth — the same Cloudflare account already runs
`ame-rooms-ws`. ame's `EnsurePackage` gets a remote path tried before local
generation; local generation stays as fallback. The `cacheMetadata.Source`
field was designed for exactly this.

---

## 1. How "LeagueSkins" actually works (the model you like)

The ecosystem around `bettie9` (the repo your referenced layout comes from):

| Piece | What it is | Public? |
|---|---|---|
| `league-skin-fantome-builder` | Python CLI: `build.py --league <install> --out .\out` | Yes (MIT) |
| `bettie9/LeagueSkins` | 3.33 GB repo, ~11.7k `.fantome` packages + `index.json` catalog | Yes |
| `Sunshine` (app) | Tauri app that downloads packages per-skin from that repo on demand, applies via LTK Patcher | Yes |

Key mechanics, all directly relevant to ame:

- **The builder runs against a local League install.** Full roster (170+
  champions, skins + chromas + forms ≈ 9,700 packages) takes **~3–5 hours**.
  It vendors LtMAO (WAD/BIN reader) + `ritobin_cli`, and downloads CDragon
  hash tables (~50 MB) on first run.
- **Packages stay tiny (~300 KB average) because they don't ship assets.**
  Each `.fantome` contains a patched `skin0.bin` (+ animations bin) that
  redirects asset hashes to the skin's textures/models/VFX **already in the
  player's game files**. That's why the whole 9.7k-package library is only
  ~3.3 GB.
- **Rebuilt every patch**, incrementally: unchanged packages are reused, only
  catalog + changed packages are rewritten. The repo `BUILD-INFO.md` logs each
  rebuild with the game build it targets (e.g. `16.19.821.7343`).
- **Distribution is just GitHub itself**: `raw.githubusercontent.com` URLs per
  file + a catalog (`index.json`) mapping skin → filename + SHA-256.
- **Client sync (Sunshine/Rose pattern)**: check catalog version/commit SHA →
  download only what the user selects. (Rose instead downloads the whole repo
  ZIP at startup — worse; don't copy that part.)

Rose (what ame is inspired by) does the same but dumber: full-repo ZIP
download + incremental updates via the GitHub compare API when <200 files
changed.

**What this means for ame:** you already have the equivalent of the builder
(`skin-generator.exe` bundle, currently downloaded from your GitHub releases
and run per-user). The missing piece is only (a) a batch runner over all
(champion, skin, chroma, form) IDs, (b) storage, (c) a catalog, and (d) a
client fetch path.

## 2. Why private matters: the actual threat model

This concern is well-founded, with receipts:

- **darkseal-org** (ran `lol-skins`, `ritoskin`, `Exalted` — the biggest public
  skin library): on **14 November 2025** received formal legal notice from
  **Tencent** (as League of Legends rights holder) through counsel. They
  complied, took everything down, and closed the org. Their parting note:
  "We knew the risk when we started. We just thought we'd have longer."
- **LeagueSandbox/GameServer**: Riot C&D (Aug 2022) → archived.
- `bettie9/LeagueSkins` is public *today*, and its packages contain
  patched/derived game data — same DMCA exposure. It just hasn't been noticed
  or prioritized yet.

Be honest with yourself about what "private" buys: the endpoint is embedded in
the ame binary, so a motivated Riot/Tencent employee can download ame and
extract it in an afternoon. What privacy actually removes is the **zero-effort
attack surface**: search engines, GitHub search, Discord links, DMCA-scraping
bots. The darkseal takedown letter targeted public, indexed, advertised
repositories. A non-indexed domain with auth + unguessable keys + no public
mention makes you invisible to that low-effort layer — that's all.

Also: central hosting means you are **redistributing** derived Riot assets at
scale, which is strictly more exposed than ame's current model (each user
generates from their own install). Mitigations below, but this is the real
trade-off.

## 3. Private storage options compared

| Option | Egress cost | Storage cost | Privacy mechanics | Verdict |
|---|---|---|---|---|
| **Cloudflare R2 + Worker** | **$0 (forever)** | $0.015/GB-mo (10 GB free) | Bucket is private by default; only the Worker (R2 binding) can read; clients hit Worker URL with token | **Recommended** |
| Public R2 (r2.dev / custom domain) | $0 | same | No auth possible (no per-object ACLs in R2); WAF HMAC needs Pro plan | Rejected — "public" is what you're avoiding |
| R2 presigned URLs | $0 | same | Work only on `<account>.r2.cloudflarestorage.com` (leaks account+bucket in URL); ≤7-day expiry; bearer token | OK alternative; uglier URLs, key rotation awkward |
| Private GitHub repo + PAT in binary | free | free | Fine-grained PAT scoped to repo contents; **shared token = 5k req/hr** for all users combined; token extractable; account ban risk for API-as-CDN | Rejected |
| GitHub Releases on private repo | free | free | Asset download still needs auth (same PAT problem) | Rejected |
| Backblaze B2 | free up to 3× storage/day, then $0.01/GB | $6/TB-mo | app keys + bucket private | Fine, but egress metering and no worker integration |
| S3 | $0.09/GB | $0.023/GB | full signing support | Rejected — egress dominates |
| Self-hosted (Hetzner/OVH + nginx) | cheap/unmetered | included | hidden domain + tokens; also responds to legal notices eventually | Viable; you babysit TLS, DDoS, updates, disk |
| jsDelivr / public CDNs | free | free | none | Rejected |

### Why R2 + Worker specifically

- **You already run Cloudflare infra** (`worker/`, wrangler 4.x, Durable
  Objects for room party). One more worker, same account, same deploy flow.
- **Zero egress** is the killer feature: users downloading only what they
  select ≈ 60 MB/month per active user; 1,000 active users ≈ 60 GB egress =
  $0. On S3 that's ~$5.4/mo and it scales linearly forever.
- **R2 is private by default.** No public URL exists until you explicitly add
  one. All access goes through the Worker binding.
- **The Worker hides everything**: clients only ever see
  `https://<unlisted-domain>/...`; bucket name, account ID, key layout never
  appear. Objects can use non-guessable keys (e.g. `sha256(...)` of content),
  so even a leaked base URL allows zero enumeration.
- **Edge caching for free**: `Cache-Control` + Cloudflare cache on the Worker
  route → hot packages served from edge, R2 read (Class B op) not even hit.
- Streaming from R2 in a Worker is a one-liner
  (`new Response(object.body)`), supports range requests for resumable
  downloads.

## 4. Recommended architecture for ame

```
                ┌────────────────────────────────────────────┐
                │  Build machine (your PC or a small runner) │
                │  LoL installed + skin-generator.exe        │
                │  per patch: iterate all (champ, skin) IDs  │
                │  → out/<champID>/<skinID>/<skinID>.fantome │
                │  → index.json {id, sha256, size, build}    │
                └───────────────┬────────────────────────────┘
                                │ wrangler r2 object put / aws cli
                                ▼
                ┌────────────────────────────────────────────┐
                │  Cloudflare R2 bucket "ame-skins" (private)│
                │  ~3–4 GB, ~10k objects, keys = content hash│
                └───────────────┬────────────────────────────┘
                                │ R2 binding (only reader)
                                ▼
                ┌────────────────────────────────────────────┐
                │  Worker "ame-cdn" (new, next to rooms-ws)  │
                │  GET /v{v}/catalog    → index.json          │
                │  GET /v{v}/{skinId}   → package             │
                │  auth: HMAC token, per-app-release secret   │
                │  rate-limit per token, no directory listing │
                └───────────────┬────────────────────────────┘
                                │ HTTPS (unlisted domain)
                                ▼
                ┌────────────────────────────────────────────┐
                │  ame Go backend                              │
                │  EnsurePackage(): remote try → cache →       │
                │  fallback local generation (unchanged path)  │
                └────────────────────────────────────────────┘
```

### Build side

1. Write a batch runner that enumerates all champion/skin/chroma/form IDs
   (source of truth: the same LCU/Data Dragon data the plugin already uses to
   render the picker) and shells the existing helper once per package —
   mirroring bettie9's `--only`/`--limit` flags for testing.
2. Deduplicate: skip packages whose (game build, skin ID, content hash) match
   the previous run (bettie9 does this: "repeating the build reused all
   packages without rewriting them").
3. Emit `index.json`:
   `{"build": "16.19.821.7343", "generatedAt": ..., "packages": {"<skinId>": {"sha256": ..., "size": ...}}}`.
4. Cadence: rerun after each patch (LoL patches every 2 weeks). The catalog
   records the game build; the client compares against the installed build
   (you already read it via `installedBuild`/`league.exe` fingerprint) and
   falls back to local generation on mismatch.
5. Where to run it: your own PC on patch day is fine (3–5 h). A cheap
   always-on Windows VPS makes it hands-off and lets you rebuild mid-patch
   for hotfixes. GitHub Actions is awkward: Windows runners have ~14 GB disk
   and the game's Champions WADs alone eat most of it; you'd also need to
   fetch WADs from Riot's CDN via release manifests (Flint does this) —
   doable but extra work for no real gain.

### Serving side (the private part)

- New worker `ame-cdn`, R2 binding, no public bucket URL, no custom domain on
  the bucket itself.
- Auth: client sends `Authorization: Bearer <token>` (or HMAC query param).
  The token = HMAC(release-secret, app-version) baked into each ame release;
  **rotate the secret every release** so a leaked old token dies at the next
  update. Extraction from the binary is always possible — rotation bounds the
  blast radius and breaks scrapers quietly.
- Object keys = content hash (or `sha256` prefix) rather than readable
  `<champ>/<skin>.fantome`, so even with the token you can't enumerate or
  harvest the library without the catalog, which the Worker hands out one
  skin at a time.
- Rate-limit per token + per IP in the Worker; return 404 (not 403) on bad
  auth so probing looks like noise.
- Set `Cache-Control: public, max-age` on package responses — safe to cache
  because URLs are content-addressed and immutable; hot skins come off the
  edge and never touch R2 reads.
- Keep the domain out of: the public repo, Discord, search engines, the
  README, and binary strings (assemble it at runtime from parts).
- Zero logs of user identity; log only counts.

### Client side (integration points already in the code)

- `internal/skin/local.go`:
  - `ensurePackage` — insert remote attempt between cache check and
    `generatePackage`. On success: validate with the existing
    `validatePackage()`, verify SHA-256 against the catalog, write package +
    `cacheMetadata` with `Source: "remote"` (the field already exists and
    `ArtifactKey` already handles non-local sources by falling back to plain
    skinID keys).
  - `Download()` is already a "compatibility entry point" that all
    acquisitions share — one seam to change.
- Keep **local generation as the fallback** (don't delete it): it covers
  users on a stale patch before you finish the rebuild, covers a build
  machine failure, and keeps the app working if the CDN ever gets taken down.
  This also de-risks the whole legal exposure: central hosting is an
  accelerator, not a dependency.
- Prefetch on champion hover already exists (per `EnsurePackage` comment) —
  with remote packages it becomes a ~300 KB download instead of a generation
  job, which is the actual UX win: selection feels instant.

### Cost estimate

- Storage: 3–4 GB ≈ $0.05/mo (free tier covers 10 GB).
- Class B (GETs): 10M/mo free. 1k active users × 200 packages/mo = 200k reads
  ≈ 2% of the free tier, and edge caching eats into even that.
- Worker requests: 100k/day free (then $5/mo plan). Fine well past 10k users.
- **Realistic total: $0 → $5/month.**

## 5. Risks and mitigations

| Risk | Mitigation |
|---|---|
| DMCA/legal notice (the darkseal scenario) | No public repo, no ads, unlisted domain, token auth, per-release rotation, 404-on-unauth, non-enumerable keys. Local generation fallback means a takedown = slower UX, not a dead app. Keep no identity logs. |
| Token extracted from binary | Rotate per release; rate-limit per token; keys are content-addressed so a token alone yields nothing without per-skin requests; optionally bind token to a challenge from the LCU presence (harder, later). |
| Package broken by a patch | Catalog carries the game build; client compares to installed build and falls back to local generation; rebuild within hours of patch (same as today's per-user regeneration, but once). |
| CDN domain blocked / worker suspended | Two names ready (swap domain), or fall back to local generation entirely. Keep the catalog+packages layout portable (plain objects + JSON) so moving to B2/Hetzner is a re-upload, not a redesign. |
| Cost blow-up | R2 has no egress fees; hard ceiling is storage (~$0.05/mo) + $5 worker plan. Set a Cloudflare billing alert. |
| User on old game patch | Build check per package; stale client ⇒ local generation (current behavior). |

## 6. Open decisions

1. **Chroma/form coverage**: the plugin already models `baseSkinID` — confirm
   the batch runner enumerates chromas and multi-form skins the way the UI
   presents them (bettie9's catalog has 9.7k entries including these).
2. **Whole-library vs popular-only central hosting**: hosting everything is
   ~3.5 GB and simplest; hosting top-N champions only shrinks legal surface
   but adds cache-miss logic. Recommend starting with everything.
3. **Domain strategy**: free `*.workers.dev` subdomain (fine, not indexed) vs
   a throwaway custom domain (~$10/yr, less obviously Cloudflare-shared).
4. **Team room party teammates**: `ExtractPackage`/teammate flow already
   consumes published packages by hash — remote source plugs in with zero
   changes there.

## Sources

- bettie9/LeagueSkins repo + README, BUILD-INFO.md (sizes, rebuild cadence,
  catalog format): https://github.com/bettie9/LeagueSkins
- bettie9/league-skin-fantome-builder (build pipeline, local install input,
  3–5 h full roster, ~300 KB packages): https://github.com/bettie9/league-skin-fantome-builder
- bettie9/Sunshine (per-skin on-demand download model, LTK Patcher):
  https://github.com/bettie9/Sunshine
- Alban1911/Rose (repo-as-CDN client sync, whole-ZIP vs incremental):
  https://github.com/Alban1911/Rose
- darkseal-org archive (Tencent legal notice 2025-11-14, org shut down):
  https://github.com/darkseal-org
- LeagueSandbox/GameServer (Riot C&D precedent):
  https://github.com/LeagueSandbox/GameServer
- Cloudflare R2 pricing (zero egress, $0.015/GB-mo, 10 GB free tier):
  https://developers.cloudflare.com/r2/pricing/
- R2 privacy model + presigned URL limitations (no object ACLs, presign only
  on S3 domain, custom domain = public):
  https://developers.cloudflare.com/r2/api/s3/presigned-urls/
  and https://community.cloudflare.com/t/how-to-private-all-files-directly-and-leave-only-signed-url/742654
- Worker + R2 private serving (bindings, bearer auth, streaming):
  https://developers.cloudflare.com/workers/tutorials/upload-assets-with-r2/
