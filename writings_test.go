package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fixture returns a Config rooted in a fresh temporary directory.
func fixture(t *testing.T) Config {
	t.Helper()
	root := t.TempDir()
	cfg := Config{
		ContentRoot: filepath.Join(root, "content", "writings"),
		OutputRoot:  filepath.Join(root, "public"),
		SiteDir:     "", // no landing page in tests
	}
	if err := os.MkdirAll(cfg.ContentRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	return cfg
}

func writeArticle(t *testing.T, cfg Config, date, slug, body string) string {
	t.Helper()
	dir := filepath.Join(cfg.ContentRoot, date, slug)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func writeAttachment(t *testing.T, articleDir, name string, data []byte) {
	t.Helper()
	filesDir := filepath.Join(articleDir, "files")
	if err := os.MkdirAll(filesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(filesDir, name), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func mustBuild(t *testing.T, cfg Config) {
	t.Helper()
	if _, err := runBuild(cfg); err != nil {
		t.Fatalf("build failed unexpectedly: %v", err)
	}
}

func mustFailBuild(t *testing.T, cfg Config, wantSubstr string) {
	t.Helper()
	if _, err := runBuild(cfg); err == nil {
		t.Fatalf("build succeeded but should have failed (want error containing %q)", wantSubstr)
	} else if wantSubstr != "" && !strings.Contains(err.Error(), wantSubstr) {
		t.Fatalf("build failed with %q; want error containing %q", err.Error(), wantSubstr)
	}
}

func readOut(t *testing.T, cfg Config, parts ...string) string {
	t.Helper()
	p := filepath.Join(append([]string{cfg.OutputRoot}, parts...)...)
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("expected output file missing: %v", err)
	}
	return string(b)
}

func TestValidArticleGeneration(t *testing.T) {
	cfg := fixture(t)
	writeArticle(t, cfg, "2026-07-18", "why-i-like-plain-html",
		"# Why I like plain HTML\n\nThis is a short note with **bold** text.\n\n## A subheading\n\nMore text.\n")
	mustBuild(t, cfg)
	page := readOut(t, cfg, "writings", "2026-07-18", "why-i-like-plain-html", "index.html")
	for _, want := range []string{
		"<h1>Why I like plain HTML</h1>",
		"<p>2026-07-18</p>",
		`<p><a href="/writings">Writings</a></p>`,
		"<strong>bold</strong>",
		"<h2>A subheading</h2>",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("article page missing %q", want)
		}
	}
	if strings.Count(page, "<h1") != 1 {
		t.Errorf("title must not be rendered twice")
	}
	index := readOut(t, cfg, "writings", "index.html")
	if !strings.Contains(index,
		`<li>2026-07-18 - <a href="/writings/2026-07-18/why-i-like-plain-html">Why I like plain HTML</a></li>`) {
		t.Errorf("index entry malformed:\n%s", index)
	}
}

func TestEmptyIndexGeneration(t *testing.T) {
	cfg := fixture(t)
	mustBuild(t, cfg)
	index := readOut(t, cfg, "writings", "index.html")
	if !strings.Contains(index, "<h1>Writings</h1>\n<ul>\n</ul>") {
		t.Errorf("empty index must contain an empty <ul>:\n%s", index)
	}
}

func TestIndexOrderingDateDescending(t *testing.T) {
	cfg := fixture(t)
	writeArticle(t, cfg, "2026-01-01", "old", "# Old\n\nx\n")
	writeArticle(t, cfg, "2026-07-18", "new", "# New\n\nx\n")
	writeArticle(t, cfg, "2025-12-31", "oldest", "# Oldest\n\nx\n")
	mustBuild(t, cfg)
	index := readOut(t, cfg, "writings", "index.html")
	iNew := strings.Index(index, "2026-07-18")
	iOld := strings.Index(index, "2026-01-01")
	iOldest := strings.Index(index, "2025-12-31")
	if !(iNew < iOld && iOld < iOldest) {
		t.Errorf("index not in reverse chronological order:\n%s", index)
	}
}

func TestSameDateSlugAscending(t *testing.T) {
	cfg := fixture(t)
	writeArticle(t, cfg, "2026-07-18", "zebra", "# Zebra\n\nx\n")
	writeArticle(t, cfg, "2026-07-18", "alpha", "# Alpha\n\nx\n")
	mustBuild(t, cfg)
	index := readOut(t, cfg, "writings", "index.html")
	if !(strings.Index(index, "alpha") < strings.Index(index, "zebra")) {
		t.Errorf("same-date articles not sorted by slug ascending:\n%s", index)
	}
}

func TestInvalidCalendarDateRejected(t *testing.T) {
	cfg := fixture(t)
	writeArticle(t, cfg, "2026-02-30", "note", "# Note\n\nx\n")
	mustFailBuild(t, cfg, "2026-02-30")
}

func TestInvalidDateFormatRejected(t *testing.T) {
	cfg := fixture(t)
	writeArticle(t, cfg, "2026-7-18", "note", "# Note\n\nx\n")
	mustFailBuild(t, cfg, "2026-7-18")
}

func TestInvalidSlugRejected(t *testing.T) {
	for _, slug := range []string{"Plain-Html", "plain_html", "-plain", "plain-", "plain--html"} {
		cfg := fixture(t)
		writeArticle(t, cfg, "2026-07-18", slug, "# Note\n\nx\n")
		mustFailBuild(t, cfg, slug)
	}
}

func TestMissingTitleRejected(t *testing.T) {
	cfg := fixture(t)
	writeArticle(t, cfg, "2026-07-18", "note", "Just text, no heading.\n")
	mustFailBuild(t, cfg, "level-one")
}

func TestMultipleH1Rejected(t *testing.T) {
	cfg := fixture(t)
	writeArticle(t, cfg, "2026-07-18", "note", "# One\n\ntext\n\n# Two\n\nmore\n")
	mustFailBuild(t, cfg, "multiple level-one")
}

func TestRawHTMLDisabled(t *testing.T) {
	cfg := fixture(t)
	writeArticle(t, cfg, "2026-07-18", "note",
		"# Note\n\n<script>alert(1)</script>\n\n<div style=\"color:red\" onclick=\"x()\">hi</div>\n")
	mustBuild(t, cfg)
	page := readOut(t, cfg, "writings", "2026-07-18", "note", "index.html")
	for _, banned := range []string{"<script", "onclick", "style="} {
		if strings.Contains(page, banned) {
			t.Errorf("raw HTML leaked into output: found %q", banned)
		}
	}
}

func TestJavascriptAndDataURLsRejected(t *testing.T) {
	for _, link := range []string{"javascript:alert(1)", "data:text/html,x", "file:///etc/passwd"} {
		cfg := fixture(t)
		writeArticle(t, cfg, "2026-07-18", "note", "# Note\n\n[x]("+link+")\n")
		mustFailBuild(t, cfg, "rejected")
	}
}

func TestMissingAttachmentRejected(t *testing.T) {
	cfg := fixture(t)
	writeArticle(t, cfg, "2026-07-18", "note", "# Note\n\n![pic](files/photo.jpg)\n")
	mustFailBuild(t, cfg, "does not exist")
}

func TestPathTraversalRejected(t *testing.T) {
	for _, ref := range []string{"files/../secret.jpg", "../other/index.md", "files/%2e%2e/x.jpg"} {
		cfg := fixture(t)
		writeArticle(t, cfg, "2026-07-18", "note", "# Note\n\n[x]("+ref+")\n")
		mustFailBuild(t, cfg, "rejected")
	}
}

func TestUnsupportedAttachmentRejected(t *testing.T) {
	cfg := fixture(t)
	dir := writeArticle(t, cfg, "2026-07-18", "note", "# Note\n\nx\n")
	writeAttachment(t, dir, "page.svg", []byte("<svg/>"))
	mustFailBuild(t, cfg, "not supported")
}

func TestSymlinkAttachmentRejected(t *testing.T) {
	cfg := fixture(t)
	dir := writeArticle(t, cfg, "2026-07-18", "note", "# Note\n\nx\n")
	filesDir := filepath.Join(dir, "files")
	if err := os.MkdirAll(filesDir, 0o755); err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "real.jpg")
	if err := os.WriteFile(outside, []byte("jpegdata"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(filesDir, "photo.jpg")); err != nil {
		t.Skipf("symlinks not supported here: %v", err)
	}
	mustFailBuild(t, cfg, "symbolic link")
}

func TestAttachmentsCopiedByteForByte(t *testing.T) {
	cfg := fixture(t)
	dir := writeArticle(t, cfg, "2026-07-18", "note",
		"# Note\n\n![pic](files/photo.jpg)\n\n![screen](files/shot.png)\n\n[doc](files/paper.pdf)\n")
	jpg := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46}
	png := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00}
	pdf := []byte("%PDF-1.4\n%\xc3\xa4\xc3\xbc\n1 0 obj\nendobj\n")
	writeAttachment(t, dir, "photo.jpg", jpg)
	writeAttachment(t, dir, "shot.png", png)
	writeAttachment(t, dir, "paper.pdf", pdf)
	mustBuild(t, cfg)
	want := map[string]string{"photo.jpg": string(jpg), "shot.png": string(png), "paper.pdf": string(pdf)}
	for name, data := range want {
		got, err := os.ReadFile(filepath.Join(cfg.OutputRoot, "writings", "2026-07-18", "note", "files", name))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != data {
			t.Errorf("%s was not copied byte-for-byte", name)
		}
	}
}

func TestNewWritingCreatesSkeleton(t *testing.T) {
	cfg := fixture(t)
	indexPath, filesDir, err := runNew(cfg, "Why I like plain HTML", "2026-07-18", "")
	if err != nil {
		t.Fatal(err)
	}
	wantIndex := filepath.Join(cfg.ContentRoot, "2026-07-18", "why-i-like-plain-html", "index.md")
	if indexPath != wantIndex {
		t.Errorf("index path = %q, want %q", indexPath, wantIndex)
	}
	if fi, err := os.Stat(filesDir); err != nil || !fi.IsDir() {
		t.Errorf("files dir not created: %v", err)
	}
	b, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "# Why I like plain HTML\n\n" {
		t.Errorf("index.md content = %q", string(b))
	}
}

func TestNewWritingRefusesOverwrite(t *testing.T) {
	cfg := fixture(t)
	if _, _, err := runNew(cfg, "Note", "2026-07-18", "note"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := runNew(cfg, "Note", "2026-07-18", "note"); err == nil {
		t.Fatal("second new-writing with the same date and slug must fail")
	} else if !strings.Contains(err.Error(), "refusing to overwrite") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestNewWritingValidation(t *testing.T) {
	cfg := fixture(t)
	if _, _, err := runNew(cfg, "Note", "2026-02-30", ""); err == nil {
		t.Error("invalid calendar date accepted")
	}
	if _, _, err := runNew(cfg, "Note", "today", ""); err == nil {
		t.Error("non-date accepted")
	}
	if _, _, err := runNew(cfg, "Note", "", "Bad_Slug"); err == nil {
		t.Error("invalid slug accepted")
	}
	if _, _, err := runNew(cfg, "   ", "", ""); err == nil {
		t.Error("empty title accepted")
	}
}

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Why I like plain HTML": "why-i-like-plain-html",
		"Note 2":                "note-2",
		"  --Weird__title!!  ":  "weird-title",
		"Привет мир":            "note",
		"C'est déjà l'été":      "c-est-d-j-l-t",
	}
	for in, want := range cases {
		if got := slugify(in); got != want {
			t.Errorf("slugify(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCanonicalURLs(t *testing.T) {
	cfg := fixture(t)
	writeArticle(t, cfg, "2026-07-18", "note", "# Note\n\nx\n")
	mustBuild(t, cfg)
	page := readOut(t, cfg, "writings", "2026-07-18", "note", "index.html")
	if !strings.Contains(page, `<link rel="canonical" href="https://chaporgin.com/writings/2026-07-18/note">`) {
		t.Errorf("article canonical URL wrong:\n%s", page)
	}
	index := readOut(t, cfg, "writings", "index.html")
	if !strings.Contains(index, `<link rel="canonical" href="https://chaporgin.com/writings">`) {
		t.Errorf("index canonical URL wrong:\n%s", index)
	}
}

func TestNoHTMLSuffixInInternalLinks(t *testing.T) {
	cfg := fixture(t)
	writeArticle(t, cfg, "2026-07-18", "note", "# Note\n\n[back](/writings)\n")
	mustBuild(t, cfg)
	for _, p := range [][]string{
		{"writings", "index.html"},
		{"writings", "2026-07-18", "note", "index.html"},
	} {
		page := readOut(t, cfg, p...)
		if strings.Contains(page, ".html\"") || strings.Contains(page, ".html#") {
			t.Errorf("%v: internal links must not contain .html", p)
		}
	}
}

func TestNoCSSOrJavaScript(t *testing.T) {
	cfg := fixture(t)
	writeArticle(t, cfg, "2026-07-18", "note", "# Note\n\nSome *text* and `code`.\n")
	mustBuild(t, cfg)
	for _, p := range [][]string{
		{"writings", "index.html"},
		{"writings", "2026-07-18", "note", "index.html"},
	} {
		page := readOut(t, cfg, p...)
		for _, banned := range []string{"<script", "<style", "stylesheet", "style="} {
			if strings.Contains(page, banned) {
				t.Errorf("%v: generated page contains %q", p, banned)
			}
		}
	}
}

func TestNoLocalFilesystemPaths(t *testing.T) {
	cfg := fixture(t)
	writeArticle(t, cfg, "2026-07-18", "note", "# Note\n\nx\n")
	mustBuild(t, cfg)
	page := readOut(t, cfg, "writings", "2026-07-18", "note", "index.html")
	root := filepath.Dir(cfg.ContentRoot)
	if strings.Contains(page, root) {
		t.Errorf("generated page leaks local filesystem path %q", root)
	}
}

func TestDeterministicOutput(t *testing.T) {
	cfg := fixture(t)
	dir := writeArticle(t, cfg, "2026-07-18", "note",
		"# Note\n\nText with [a link](https://example.com) and an image.\n\n![p](files/photo.jpg)\n")
	writeAttachment(t, dir, "photo.jpg", []byte{0xFF, 0xD8, 0xFF})
	writeArticle(t, cfg, "2026-01-02", "other", "# Другой\n\nЮникод.\n")

	read := func() map[string]string {
		out := map[string]string{}
		err := filepath.Walk(filepath.Join(cfg.OutputRoot, "writings"),
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() {
					b, err := os.ReadFile(path)
					if err != nil {
						return err
					}
					rel, _ := filepath.Rel(cfg.OutputRoot, path)
					out[rel] = string(b)
				}
				return nil
			})
		if err != nil {
			t.Fatal(err)
		}
		return out
	}
	mustBuild(t, cfg)
	first := read()
	mustBuild(t, cfg)
	second := read()
	if len(first) != len(second) {
		t.Fatalf("file sets differ between builds: %d vs %d", len(first), len(second))
	}
	for name, content := range first {
		if second[name] != content {
			t.Errorf("%s differs between two identical builds", name)
		}
	}
}

func TestUnicodeTitleAndBody(t *testing.T) {
	cfg := fixture(t)
	writeArticle(t, cfg, "2026-07-18", "privet",
		"# Привет, мир — тест «кавычек» & <спецсимволов>\n\nТело статьи по-русски.\n")
	mustBuild(t, cfg)
	page := readOut(t, cfg, "writings", "2026-07-18", "privet", "index.html")
	if !strings.Contains(page, "Привет, мир — тест «кавычек» &amp; &lt;спецсимволов&gt;") {
		t.Errorf("unicode title not escaped/preserved correctly:\n%s", page)
	}
	if !strings.Contains(page, "Тело статьи по-русски.") {
		t.Errorf("unicode body missing")
	}
	index := readOut(t, cfg, "writings", "index.html")
	if !strings.Contains(index, "Привет, мир") {
		t.Errorf("unicode title missing from index")
	}
}

func TestOwnershipProtection(t *testing.T) {
	cfg := fixture(t)
	writeArticle(t, cfg, "2026-07-18", "note", "# Note\n\nx\n")
	// Pre-existing public/writings NOT generated by this tool.
	foreign := filepath.Join(cfg.OutputRoot, "writings")
	if err := os.MkdirAll(foreign, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(foreign, "precious.txt"), []byte("keep me"), 0o644); err != nil {
		t.Fatal(err)
	}
	mustFailBuild(t, cfg, "refusing to replace")
	if _, err := os.Stat(filepath.Join(foreign, "precious.txt")); err != nil {
		t.Fatalf("foreign file was deleted: %v", err)
	}
	// After removing the foreign dir, builds succeed and are repeatable.
	if err := os.RemoveAll(foreign); err != nil {
		t.Fatal(err)
	}
	mustBuild(t, cfg)
	mustBuild(t, cfg) // second build must replace its own output without error
}

func TestFaviconLinkPresent(t *testing.T) {
	cfg := fixture(t)
	writeArticle(t, cfg, "2026-07-18", "note", "# Note\n\nx\n")
	mustBuild(t, cfg)
	wants := []string{
		`<link rel="icon" href="/favicon.ico" sizes="any">`,
		`<link rel="icon" href="/favicon.svg" type="image/svg+xml">`,
	}
	for _, p := range [][]string{
		{"writings", "index.html"},
		{"writings", "2026-07-18", "note", "index.html"},
	} {
		page := readOut(t, cfg, p...)
		for _, want := range wants {
			if !strings.Contains(page, want) {
				t.Errorf("%v: missing favicon link %q", p, want)
			}
		}
	}
}

func TestRouteUsesDirectoryIndexLayout(t *testing.T) {
	cfg := fixture(t)
	writeArticle(t, cfg, "2026-07-18", "note", "# Note\n\nx\n")
	mustBuild(t, cfg)
	p := filepath.Join(cfg.OutputRoot, "writings", "2026-07-18", "note", "index.html")
	if fi, err := os.Stat(p); err != nil || fi.IsDir() {
		t.Fatalf("article must be a real static index.html file at %s: %v", p, err)
	}
}

func TestStrayFilesRejected(t *testing.T) {
	cfg := fixture(t)
	dir := writeArticle(t, cfg, "2026-07-18", "note", "# Note\n\nx\n")
	if err := os.WriteFile(filepath.Join(dir, "draft.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	mustFailBuild(t, cfg, "unexpected entry")
}
