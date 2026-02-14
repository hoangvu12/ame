# Commits

- Always use a single-line commit message (no body, no co-authored-by, no credits)
- Keep messages short and descriptive (e.g. `fix: room party teammate skins not showing`)

# Releasing

When creating a new release (pushing a version tag):

1. Overwrite `changes.md` with the list of changes for this version
2. Use simple bullet points, one change per line
3. Write changes in plain language (no technical jargon)
4. Commit `changes.md` before creating the tag
5. Follow **semver** for the version number:
   - **patch** (x.y.Z): bug fixes only
   - **minor** (x.Y.0): new features (may include fixes)
   - **major** (X.0.0): breaking changes

Example `changes.md`:

```
- Added room party for sharing skins with teammates
- Fixed skin not applying on certain champions
- Improved startup speed
```

The release workflow automatically appends installation instructions below the changelog.
