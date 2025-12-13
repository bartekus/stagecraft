// SPDX-License-Identifier: AGPL-3.0-or-later

/*

Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.

*/

// Tiny repo indexer (Go). Writes data/index.json. Usage: go run index.go [path]
package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileEntry struct {
	Path  string `json:"path"`
	Size  int64  `json:"size"`
	Lines int    `json:"lines"`
	Ext   string `json:"ext"`
}
type ModuleHit struct {
	Kind   string `json:"kind"`
	File   string `json:"file"`
	Name   string `json:"name,omitempty"`
	Module string `json:"module,omitempty"`
}
type Index struct {
	Root        string           `json:"root"`
	IndexedAt   string           `json:"indexedAt"`
	Files       []FileEntry      `json:"files"`
	Languages   map[string]int64 `json:"languages"`
	TopDirs     map[string]int64 `json:"topDirs"`
	ModuleFiles []ModuleHit      `json:"moduleFiles"`
	Digest      string           `json:"digest"`
}

var defaultIgnores = map[string]bool{
	".git": true, "node_modules": true, "dist": true, "build": true, "out": true,
	"target": true, "vendor": true, ".cache": true, ".tmp": true, "coverage": true,
}

func readIgnores(root string) map[string]bool {
	ign := map[string]bool{}
	for k := range defaultIgnores {
		ign[k] = true
	}
	b, err := os.ReadFile(filepath.Join(root, "tools", "context-compiler", "xray", "ignore.rules"))
	if err != nil {
		return ign
	}
	for _, ln := range strings.Split(string(b), "\n") {
		ln = strings.TrimSpace(ln)
		if ln != "" {
			ign[ln] = true
		}
	}
	return ign
}

func countLines(p string) int {
	f, err := os.Open(p)
	if err != nil {
		return 0
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	i := 0
	for sc.Scan() {
		i++
	}
	return i
}

func main() {
	root, _ := filepath.Abs(".")
	if len(os.Args) > 1 {
		root, _ = filepath.Abs(os.Args[1])
	}
	ign := readIgnores(root)

	files := []FileEntry{}
	filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(root, p)
		parts := strings.Split(rel, string(os.PathSeparator))
		for _, part := range parts {
			if ign[part] {
				return filepath.SkipDir
			}
		}
		if d.Type().IsRegular() {
			st, err := os.Stat(p)
			if err != nil {
				return nil
			}
			ext := strings.ToLower(filepath.Ext(p))
			files = append(files, FileEntry{
				Path: filepath.ToSlash(rel), Size: st.Size(), Lines: countLines(p), Ext: ext,
			})
		}
		return nil
	})

	langs := map[string]int64{}
	top := map[string]int64{}
	for _, f := range files {
		dir := "."
		if i := strings.IndexByte(f.Path, '/'); i >= 0 {
			dir = f.Path[:i]
		}
		top[dir] += f.Size
		langs[f.Ext] += f.Size
	}

	mods := []ModuleHit{}
	if b, err := os.ReadFile(filepath.Join(root, "package.json")); err == nil {
		name := ""
		if i := strings.Index(string(b), `"name"`); i >= 0 {
			name = "package"
		}
		mods = append(mods, ModuleHit{Kind: "npm", File: "package.json", Name: name})
	}
	if b, err := os.ReadFile(filepath.Join(root, "go.mod")); err == nil {
		lines := strings.Split(string(b), "\n")
		for _, ln := range lines {
			if strings.HasPrefix(ln, "module ") {
				mods = append(mods, ModuleHit{Kind: "go", File: "go.mod", Module: strings.TrimSpace(strings.TrimPrefix(ln, "module"))})
				break
			}
		}
	}
	if _, err := os.Stat(filepath.Join(root, ".git")); err == nil {
		mods = append(mods, ModuleHit{Kind: "git", File: ".git"})
	}

	idx := Index{
		Root:      filepath.Base(root),
		IndexedAt: time.Now().UTC().Format(time.RFC3339),
		Files:     files, Languages: langs, TopDirs: top, ModuleFiles: mods,
	}
	raw, _ := json.Marshal(idx)
	sum := fmt.Sprintf("%x", sha256.Sum256(raw))[:16]
	idx.Digest = sum

	_ = os.MkdirAll(filepath.Join(root, "data"), 0o755)
	out := filepath.Join(root, "data", "index.json")
	blob, _ := json.MarshalIndent(idx, "", "  ")
	_ = os.WriteFile(out, blob, 0o644)
	fmt.Printf("Wrote %s (%d files)\n", out, len(files))
}
