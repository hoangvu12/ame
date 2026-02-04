# Contributing Translations

Thanks for helping translate Ame.

## Quick Start
1. Copy `src/locales/en_US.json` to a new file named with the Riot locale code you want to add (example: `src/locales/vi_VN.json`).
2. Translate the values on the right-hand side. Keep the keys unchanged.
3. Add your new locale code to `src/locales/manifest.json`.

## Locale Codes
Use Riot-style locale codes (for example: `en_US`, `vi_VN`, `ja_JP`, `ko_KR`, `fr_FR`).

## Placeholders
Some strings include placeholders like `{label}`. Keep them exactly as-is.

Example:

- English: `"Ame: {label}"`
- Translated: `"Ame: {label}"`

## Tips
- Keep punctuation and capitalization consistent with English where appropriate.
- Shorter UI labels work best because some buttons are narrow.

## Testing Locally
If you use the local plugin copy script, make sure the locales folder is copied too.

Run:

```
copy-plugin.bat
```

Then restart the League client.

## Validation
To check that all locale files match the `en_US` keys, run:

```
node scripts/validate-locales.js
```
