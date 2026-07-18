// Command writings is the static generator for the writings section of
// chaporgin.com. See README.md.
//
// Subcommands:
//
//	new "Article title" [--date YYYY-MM-DD] [--slug slug]
//	build
//	preview [--port N]
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func usage() {
	fmt.Fprintln(os.Stderr, `usage:
  writings new "Article title" [--date YYYY-MM-DD] [--slug slug]
  writings build
  writings preview [--port N]`)
	os.Exit(2)
}

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	cfg := Config{
		ContentRoot: filepath.Join(cwd, "content", "writings"),
		OutputRoot:  filepath.Join(cwd, "public"),
		SiteDir:     filepath.Join(cwd, "site"),
	}
	switch os.Args[1] {
	case "new":
		err = cmdNew(cfg, os.Args[2:])
	case "build":
		err = cmdBuild(cfg)
	case "preview":
		err = cmdPreview(cfg, os.Args[2:])
	default:
		usage()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func cmdNew(cfg Config, args []string) error {
	if len(args) < 1 || len(args[0]) == 0 || args[0][0] == '-' {
		return fmt.Errorf(`usage: ./bin/new-writing "Article title" [--date YYYY-MM-DD] [--slug slug]`)
	}
	title := args[0]
	var date, slug string
	rest := args[1:]
	for i := 0; i < len(rest); i++ {
		switch rest[i] {
		case "--date":
			i++
			if i >= len(rest) {
				return fmt.Errorf("--date requires a value")
			}
			date = rest[i]
		case "--slug":
			i++
			if i >= len(rest) {
				return fmt.Errorf("--slug requires a value")
			}
			slug = rest[i]
		default:
			return fmt.Errorf("unknown argument %q", rest[i])
		}
	}
	indexPath, filesDir, err := runNew(cfg, title, date, slug)
	if err != nil {
		return err
	}
	repoRoot := filepath.Dir(filepath.Dir(cfg.ContentRoot))
	relIndex, _ := filepath.Rel(repoRoot, indexPath)
	relFiles, _ := filepath.Rel(repoRoot, filesDir)
	fmt.Println("Created:     " + relIndex)
	fmt.Println("Attachments: " + relFiles + string(filepath.Separator))
	fmt.Println("Build:       ./bin/build-writings")
	fmt.Println("Preview:     ./bin/preview-writings")
	return nil
}

func cmdBuild(cfg Config) error {
	articles, err := runBuild(cfg)
	if err != nil {
		return err
	}
	repoRoot := filepath.Dir(cfg.OutputRoot)
	rel, _ := filepath.Rel(repoRoot, filepath.Join(cfg.OutputRoot, "writings"))
	fmt.Printf("Built %d article(s) into %s%c\n", len(articles), rel, filepath.Separator)
	return nil
}
