// SPDX-License-Identifier: AGPL-3.0-or-later
//
// Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.
//
// Copyright (C) 2025  Bartek Kus
//
// This program is free software licensed under the terms of the GNU AGPL v3 or later.
//
// See https://www.gnu.org/licenses/ for license details.

// Package main provides a tool to validate spec file references in Go source code.
package main

// Feature: GOV_V1_CORE
// Spec: spec/governance/GOV_V1_CORE.md

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type SpecReference struct {
	File string
	Line int
	Path string
}

type SpecError struct {
	File string
	Line int
	Path string
	Msg  string
}

func main() {
	if err := run("."); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func run(root string) error {
	files, err := walkGoFiles(root)
	if err != nil {
		return fmt.Errorf("walking go files: %w", err)
	}

	var specErrors []SpecError

	for _, f := range files {
		content, err := os.ReadFile(f) //nolint:gosec // G304: file path is from walkGoFiles, safe
		if err != nil {
			specErrors = append(specErrors, SpecError{
				File: f,
				Line: 0,
				Path: "",
				Msg:  fmt.Sprintf("reading file: %v", err),
			})
			continue
		}

		refs := extractSpecReferences(f, content)
		for _, r := range refs {
			if err := validateSpecPath(r.Path); err != nil {
				specErrors = append(specErrors, SpecError{
					File: r.File,
					Line: r.Line,
					Path: r.Path,
					Msg:  fmt.Sprintf("invalid spec path: %v", err),
				})
				continue
			}

			// Resolve spec path relative to root directory
			specPath := filepath.Join(root, r.Path)
			if _, err := os.Stat(specPath); err != nil {
				if os.IsNotExist(err) {
					specErrors = append(specErrors, SpecError{
						File: r.File,
						Line: r.Line,
						Path: r.Path,
						Msg:  "spec file does not exist",
					})
				} else {
					specErrors = append(specErrors, SpecError{
						File: r.File,
						Line: r.Line,
						Path: r.Path,
						Msg:  fmt.Sprintf("checking spec file: %v", err),
					})
				}
			}
		}
	}

	if len(specErrors) == 0 {
		return nil
	}

	for _, e := range specErrors {
		loc := e.File
		if e.Line > 0 {
			loc = fmt.Sprintf("%s:%d", e.File, e.Line)
		}
		if e.Path != "" {
			fmt.Fprintf(os.Stderr, "%s: Spec: %s: %s\n", loc, e.Path, e.Msg)
		} else {
			fmt.Fprintf(os.Stderr, "%s: %s\n", loc, e.Msg)
		}
	}

	return fmt.Errorf("spec reference validation failed with %d error(s)", len(specErrors))
}

func walkGoFiles(root string) ([]string, error) {
	var files []string

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			name := d.Name()
			switch name {
			case ".git", "vendor":
				return filepath.SkipDir
			}
			// Skip testdata and e2e tests
			if name == "testdata" || name == "e2e" {
				return filepath.SkipDir
			}
			return nil
		}

		if filepath.Ext(path) == ".go" {
			// Skip test files to avoid false positives from test data
			if strings.HasSuffix(path, "_test.go") {
				return nil
			}
			files = append(files, path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

func extractSpecReferences(filePath string, content []byte) []SpecReference {
	var refs []SpecReference

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if !strings.HasPrefix(trimmed, "//") {
			continue
		}

		// Strip leading comment marker and any spaces
		trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "//"))

		if trimmed == "" {
			continue
		}

		// Case-insensitive "Spec:" prefix
		lower := strings.ToLower(trimmed)
		const prefix = "spec:"
		if !strings.HasPrefix(lower, prefix) {
			continue
		}

		// Slice using original string to preserve case in path
		pathPart := strings.TrimSpace(trimmed[len(prefix):])
		if pathPart == "" {
			refs = append(refs, SpecReference{
				File: filePath,
				Line: lineNum,
				Path: "",
			})
			continue
		}

		refs = append(refs, SpecReference{
			File: filePath,
			Line: lineNum,
			Path: pathPart,
		})
	}

	return refs
}

func validateSpecPath(path string) error {
	if path == "" {
		return fmt.Errorf("empty path")
	}

	if !strings.HasPrefix(path, "spec/") {
		return fmt.Errorf("must start with 'spec/' (got %q)", path)
	}

	// Check for invalid characters before checking suffix
	// (since invalid chars might include newlines/tabs that affect suffix check)
	if strings.ContainsAny(path, " \t\r\n{}\"") {
		return fmt.Errorf("contains invalid characters (spaces, control chars, { } \")")
	}

	if !strings.HasSuffix(path, ".md") {
		return fmt.Errorf("must end with '.md' (got %q)", path)
	}

	// Normalized, simple rule set is enough for now.
	return nil
}
