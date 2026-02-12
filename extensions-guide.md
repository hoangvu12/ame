# Writing a Browse Extension

An extension is a single `.js` file that exports a `createSource(helpers)` function. Place it in `%LOCALAPPDATA%\ame\extensions\` or add it via the Extensions manager using a raw GitHub URL.

## Skeleton

```js
function createSource({ fetch, parseHtml }) {
  const BASE = 'https://example.com';

  return {
    // --- Required metadata ---
    id: 'example',            // unique identifier
    name: 'Example Site',     // display name
    baseUrl: BASE,
    lang: 'en',
    version: 1,

    // --- Required methods ---

    async getPopular(page) {
      const res = await fetch(`${BASE}/popular?page=${page}`);
      const doc = parseHtml(await res.text());

      const items = [...doc.querySelectorAll('.skin-card')].map(card => ({
        id: card.querySelector('a').href.split('/').pop(),
        title: card.querySelector('.title').textContent.trim(),
        author: card.querySelector('.author')?.textContent.trim() || '',
        thumbnailUrl: card.querySelector('img')?.src || '',
        championId: 0,  // 0 = unknown/all
      }));

      return {
        items,
        hasNextPage: !!doc.querySelector('.next-page'),
      };
    },

    async search(query, page, filters) {
      const sort = filters.sort || 'popular';
      const res = await fetch(`${BASE}/search?q=${encodeURIComponent(query)}&sort=${sort}&page=${page}`);
      const doc = parseHtml(await res.text());
      // ... same parsing pattern as getPopular
      return { items: [], hasNextPage: false };
    },

    async getDetails(skinMod) {
      const res = await fetch(`${BASE}/skin/${skinMod.id}`);
      const doc = parseHtml(await res.text());

      return {
        id: skinMod.id,
        title: doc.querySelector('h1').textContent.trim(),
        author: doc.querySelector('.author')?.textContent.trim() || '',
        description: doc.querySelector('.desc')?.textContent || '',
        championId: 0,
        images: [...doc.querySelectorAll('.screenshot img')].map(i => i.src),
        downloads: [{
          url: doc.querySelector('.download-btn').href,
          label: 'Download',
          size: null,
        }],
      };
    },

    // --- Optional methods ---

    getFilters() {
      return [
        {
          type: 'sort', name: 'Sort by', key: 'sort', default: 'popular',
          options: [
            { label: 'Popular', value: 'popular' },
            { label: 'Recent', value: 'recent' },
          ],
        },
      ];
    },

    // async getLatest(page) { ... }  // defaults to getPopular if omitted
  };
}
```

## Helpers

The runtime passes two helpers to `createSource`:

| Helper | Description |
|--------|-------------|
| `fetch(url, opts?)` | Proxied HTTP fetch through the Go backend. Returns a Response-like object with `.text()` and `.json()`. All requests go through a per-domain rate limiter (2 req/s, burst 5). |
| `parseHtml(html)` | Parses an HTML string into a `Document` (uses `DOMParser`). Use standard DOM methods like `querySelectorAll` to extract data. |

## Data Types

**SkinMod** — search/browse results:

| Field | Type | Description |
|-------|------|-------------|
| `id` | `string` | Unique ID on the source (URL slug, etc.) |
| `title` | `string` | Skin name |
| `author` | `string` | Creator name |
| `thumbnailUrl` | `string` | External image URL (proxied automatically) |
| `championId` | `number` | Champion ID, or `0` for unknown/all |

**SkinModPage** — paginated result:

| Field | Type | Description |
|-------|------|-------------|
| `items` | `SkinMod[]` | Results for this page |
| `hasNextPage` | `boolean` | Whether more pages exist |

**SkinDetail** — detail view:

| Field | Type | Description |
|-------|------|-------------|
| `id` | `string` | Same ID as the SkinMod |
| `title` | `string` | Full title |
| `author` | `string` | Creator name |
| `description` | `string` | Mod description |
| `championId` | `number` | Champion ID, or `0` |
| `images` | `string[]` | Screenshot/preview URLs |
| `downloads` | `DownloadInfo[]` | Downloadable file variants |

**DownloadInfo**:

| Field | Type | Description |
|-------|------|-------------|
| `url` | `string` | Direct download URL for `.fantome`/`.zip` |
| `label` | `string` | Button label, e.g. `"Download"`, `"HD"`, `"Lite"` |
| `size` | `number \| null` | File size in bytes (optional) |

**Filter** — declared by `getFilters()`, rendered by the UI:

```
{ type: 'select',   name, key, default, options: [{label, value}] }
{ type: 'sort',     name, key, default, options: [{label, value}] }
{ type: 'text',     name, key, default }
{ type: 'checkbox', name, key, default }
```

## Tips

- **`page` is 1-indexed.** The first call is `getPopular(1)`, then `getPopular(2)` when the user clicks "Load More".
- **`filters`** is a plain object keyed by your filter `key` values. On the first call it contains the defaults from `getFilters()`.
- **`getDetails`** receives the same `SkinMod` object the user clicked, so you can use `skinMod.id` to build the detail URL.
- **`downloads[].url`** must point to a direct `.fantome` or `.zip` file. The backend downloads it and imports it as a custom mod.
- **Thumbnails and images** are proxied through the backend automatically — just return the original external URLs.
- **No npm/bundler needed.** Extensions are plain JS evaluated with `new Function()`. Don't use `import`/`export`/`require`. Just define `createSource` at the top level.
- Keep the file small — there's a 1MB download limit when adding via URL.
