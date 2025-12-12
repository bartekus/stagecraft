// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.
*/

package inputs

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"
)

var reLowerHex64 = regexp.MustCompile(`^[0-9a-f]{64}$`)

// NormalizeString trims leading/trailing whitespace.
func NormalizeString(s string) string { return strings.TrimSpace(s) }

// NormalizeTags sorts tags lexicographically (in-place).
func NormalizeTags(tags []string) {
	sort.Strings(tags)
}

// NormalizeKV sorts by Key (in-place).
func NormalizeKV[T interface{ GetKey() string }](items []T) {
	sort.Slice(items, func(i, j int) bool { return items[i].GetKey() < items[j].GetKey() })
}

// PathNormalize validates + normalizes a path according to the spec.
// Rules:
// - must be relative (not absolute)
// - forward slashes
// - must not contain "." or ".." segments (but "." as standalone path is allowed)
// - must not be empty
func PathNormalize(p string) (string, error) {
	p = strings.ReplaceAll(p, `\`, `/`)
	p = strings.TrimSpace(p)
	if p == "" {
		return "", fmt.Errorf("path is empty")
	}
	// Reject absolute paths (including Windows drive-ish forms).
	if strings.HasPrefix(p, "/") || strings.HasPrefix(p, "~") || strings.Contains(p, ":/") || strings.Contains(p, `:\`) {
		return "", fmt.Errorf("path must be relative: %q", p)
	}

	// Allow "." as a standalone path (current directory)
	if p == "." {
		return ".", nil
	}

	// Check for . or .. segments BEFORE cleaning (path.Clean normalizes them away)
	parts := strings.Split(p, "/")
	for _, part := range parts {
		if part == "." || part == ".." {
			return "", fmt.Errorf("path must not contain '.' or '..' segments: %q", p)
		}
	}

	// Clean to normalize multiple slashes (apps//backend → apps/backend)
	clean := path.Clean(p)

	// path.Clean can produce "." if input was "." (already handled above) or empty
	if clean == "." {
		return ".", nil
	}
	if clean == "" {
		return "", fmt.Errorf("path invalid after clean: %q", p)
	}

	// Ensure no dot segments sneaked in after clean (defensive check)
	cleanParts := strings.Split(clean, "/")
	for _, part := range cleanParts {
		if part == "." || part == ".." {
			return "", fmt.Errorf("path must not contain '.' or '..' segments: %q", clean)
		}
	}

	return clean, nil
}

func ValidateSha256Hex64(hash string) error {
	if !reLowerHex64.MatchString(hash) {
		return fmt.Errorf("sha256 hash must be 64 lowercase hex chars: %q", hash)
	}
	return nil
}

// Sha256HexLower computes sha256 hex lowercase.
func Sha256HexLower(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// ---------- Shared leaf types ----------

type BuildArg struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (a BuildArg) GetKey() string { return a.Key }

type BuildLabel struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (l BuildLabel) GetKey() string { return l.Key }

type ComposeOverlay struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func (o ComposeOverlay) GetKey() string { return o.Name }

type ComposeVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (v ComposeVar) GetKey() string { return v.Key }

type HeaderKV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (h HeaderKV) GetKey() string { return h.Key }

type HealthEndpoint struct {
	Name           string     `json:"name"`
	URL            string     `json:"url"`
	ExpectedStatus int        `json:"expected_status"`
	Method         string     `json:"method"`
	Headers        []HeaderKV `json:"headers,omitempty"`
}

func (e HealthEndpoint) GetKey() string { return e.Name }
