# chaporgin.com — writings

Source repository for the writings section of <https://chaporgin.com>, plus the
site's landing page. Everything is generated into ordinary static HTML files by
a small Go program (`main.go`, `writings.go`) using the maintained
[goldmark](https://github.com/yuin/goldmark) Markdown parser with raw HTML
disabled. There is no database, no CMS, no server-side rendering and no
JavaScript. The only styling is a single inline `<style>` rule that caps
embedded images at 1000px wide (`img{max-width:1000px;height:auto}`); there is
no external CSS and no inline `style=` attributes. Git-tracked source files were
deliberately selected as the durable source of truth for this small static
writings system: the Markdown files and attachments in this repository are
canonical, and all generated HTML can be reproduced from them at any time.

Every content paragraph is emitted with a unique, stable id (`p1`, `p2`, ...) in
document order, so any paragraph can be linked directly as
`/writings/DATE/slug/#p3`. There is no visible anchor marker; the ids are simply
present in the HTML.

Requirements: Go 1.21+ (`brew install go` or <https://go.dev/dl/>) and Bash.

## Source directory structure

```
content/
└── writings/
    └── YYYY-MM-DD/                  article publication date
        └── human-friendly-slug/     article slug
            ├── index.md             the article (Markdown, UTF-8)
            └── files/               optional attachments (.jpg, .jpeg, .png, .pdf)
site/
├── index.html                       landing page, copied verbatim to public/
├── favicon.ico                      site icon (32x32), copied verbatim to public/
└── favicon.svg                      site icon (vector), copied verbatim to public/
public/                              generated output (do not edit; not tracked)
bin/                                 the four commands documented below
```

## How date, slug, and title are determined

- **Date** comes from the `YYYY-MM-DD` directory name. It must be a real
  calendar date (`2026-02-30` fails the build).
- **Slug** comes from the article directory name. It must match
  `^[a-z0-9]+(?:-[a-z0-9]+)*$`.
- **Title** comes from the first non-empty line of `index.md`, which must be
  exactly one level-one Markdown heading (`# Article title`). All later
  headings must be level two through six. There is no front matter and no
  other metadata.

Every article found in the canonical directory structure is published; there
is no draft system.

## How to create a note

```bash
./bin/new-writing "Why I like plain HTML"
./bin/new-writing "Why I like plain HTML" --date 2026-07-18
./bin/new-writing "Why I like plain HTML" --slug custom-slug
```

This creates `content/writings/YYYY-MM-DD/slug/index.md` (pre-filled with the
title heading) and an empty `files/` directory. It refuses to overwrite an
existing article.

## How to edit a note

Edit `content/writings/YYYY-MM-DD/slug/index.md` in any text editor (UTF-8),
then rebuild.

## How to add an image or a PDF

Copy the file into the article's own `files/` directory, keeping its filename:

```bash
cp photo.jpg content/writings/2026-07-18/why-i-like-plain-html/files/
cp paper.pdf content/writings/2026-07-18/why-i-like-plain-html/files/
```

Only `.jpg`, `.jpeg`, `.png` and `.pdf` files are allowed. Attachments are stored
directly in the repository and copied byte-for-byte into the generated site —
never resized, recompressed or renamed.

## How to link an attachment from Markdown

```markdown
![A photograph](files/photo.jpg)

[Download the related PDF](files/document.pdf)
```

References must point into the article's own `files/` directory. The build
fails if a referenced attachment does not exist. `..`, absolute paths,
`file:`, `javascript:` and `data:` URLs are rejected. Ordinary external
`https:`/`http:` links are fine.

## How to build

```bash
./bin/build-writings
```

Validates every article first, builds into a temporary directory, and only
then replaces `public/writings/` (it refuses to touch a `public/writings`
directory it did not generate). Also copies `site/index.html` to
`public/index.html`. Output URLs:

- Index: `/writings/` → `public/writings/index.html`
- Article: `/writings/YYYY-MM-DD/slug/` → `public/writings/YYYY-MM-DD/slug/index.html`
- Attachment: `/writings/YYYY-MM-DD/slug/files/name.ext`

## How to test

```bash
./bin/test-writings
```

Runs the focused Go test suite (`writings_test.go`) covering generation,
ordering, validation failures, attachment safety, determinism and Unicode.

## How to preview

```bash
./bin/preview-writings          # http://127.0.0.1:8000/writings/
PORT=9999 ./bin/preview-writings
```

Builds first, then serves the complete generated site bound to `127.0.0.1`.

## How to publish

The site is the Netlify project **chaporgin** (<https://app.netlify.com/projects/chaporgin>),
deployed as a plain publish directory. After building:

```bash
./bin/build-writings
npx netlify-cli deploy --prod --dir public     # or drag public/ into the Netlify UI
```

Commit and push the source (`content/`, `site/`, code) to your Git remote so
the repository remains the backup.

## How to restore from a fresh Git clone

```bash
git clone <your-remote> writings && cd writings
./bin/build-writings    # first run fetches goldmark and writes go.sum
```

That reproduces the entire `public/` tree from the tracked sources; nothing
outside the repository is required except the pinned Go module (goldmark,
fetched by version from the Go module proxy, or vendored via `.gomodcache`).
