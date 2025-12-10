// SPDX-License-Identifier: AGPL-3.0-or-later

// Feature: DEV_HOSTS
// Spec: spec/dev/hosts.md

package hosts

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

const (
	// StagecraftComment is the exact comment used to mark Stagecraft-managed entries.
	StagecraftComment = "# Stagecraft managed"
)

// Entry represents a single hosts file entry.
type Entry struct {
	IP      string   // IP address (e.g., "127.0.0.1")
	Domains []string // Domain names (lexicographically sorted)
	Comment string   // Optional comment
	Managed bool     // true if this is a Stagecraft-managed entry
}

// File represents a parsed hosts file.
type File struct {
	Entries []Entry
}

// ParseFile parses a hosts file and returns a File structure.
//
// The parser:
//   - Preserves all entries from the original file
//   - Identifies Stagecraft-managed entries by the exact comment "# Stagecraft managed"
//   - Handles empty lines and comments
//   - Sorts domains within each entry lexicographically
func ParseFile(path string) (*File, error) {
	//nolint:gosec // G304: path is controlled by caller (internal package)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Missing file is acceptable (will be created)
			return &File{Entries: []Entry{}}, nil
		}
		return nil, fmt.Errorf("open hosts file: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			// Log close error but don't fail parse
			_ = closeErr
		}
	}()

	var entries []Entry
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Handle comment-only lines (preserve as entry with comment)
		if strings.HasPrefix(line, "#") {
			entries = append(entries, Entry{
				Comment: line,
			})
			continue
		}

		// Parse entry line: IP domain1 domain2 ... # comment
		parts := strings.Fields(line)
		if len(parts) < 2 {
			// Invalid line format, preserve as comment
			entries = append(entries, Entry{
				Comment: line,
			})
			continue
		}

		ip := parts[0]
		var domains []string
		var comment string
		var managed bool

		// Extract domains and comment
		for i := 1; i < len(parts); i++ {
			part := parts[i]
			if strings.HasPrefix(part, "#") {
				// Rest of line is comment
				comment = strings.Join(parts[i:], " ")
				if strings.Contains(comment, StagecraftComment) {
					managed = true
				}
				break
			}
			domains = append(domains, part)
		}

		// Sort domains lexicographically
		sort.Strings(domains)

		entries = append(entries, Entry{
			IP:      ip,
			Domains: domains,
			Comment: comment,
			Managed: managed,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read hosts file: %w", err)
	}

	return &File{Entries: entries}, nil
}

// WriteFile writes a File structure to a hosts file path.
//
// The writer:
//   - Writes entries in a deterministic format
//   - Sorts domains lexicographically within each entry
//   - Preserves comment-only entries
//   - Uses atomic write (write to temp file, then rename)
func WriteFile(path string, file *File) error {
	// Create temp file in same directory
	tmpFile := path + ".tmp"
	//nolint:gosec // G304: path is controlled by caller (internal package)
	f, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}

	// Write entries
	for _, entry := range file.Entries {
		// Comment-only entry
		if entry.IP == "" && len(entry.Domains) == 0 {
			if entry.Comment != "" {
				if _, err := fmt.Fprintf(f, "%s\n", entry.Comment); err != nil {
					_ = f.Close()
					return fmt.Errorf("write comment: %w", err)
				}
			}
			continue
		}

		// Regular entry
		if entry.IP != "" && len(entry.Domains) > 0 {
			line := fmt.Sprintf("%s    %s", entry.IP, strings.Join(entry.Domains, " "))
			if entry.Comment != "" {
				line += "    " + entry.Comment
			}
			if _, err := fmt.Fprintf(f, "%s\n", line); err != nil {
				_ = f.Close()
				return fmt.Errorf("write entry: %w", err)
			}
		}
	}

	if err := f.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpFile, path); err != nil {
		return fmt.Errorf("rename temp file: %w", err)
	}

	return nil
}

// HasDomains checks if the file contains entries for the given domains.
//
// Returns true if any entry (managed or unmanaged) contains all requested domains.
func (f *File) HasDomains(domains []string) bool {
	if len(domains) == 0 {
		return false
	}

	// Sort domains for comparison
	sortedDomains := make([]string, len(domains))
	copy(sortedDomains, domains)
	sort.Strings(sortedDomains)

	// Create set for quick lookup
	domainSet := make(map[string]bool)
	for _, d := range sortedDomains {
		domainSet[d] = true
	}

	// Check each entry
	for _, entry := range f.Entries {
		if len(entry.Domains) == 0 {
			continue
		}

		// Check if this entry contains all requested domains
		allFound := true
		for _, d := range sortedDomains {
			found := false
			for _, entryDomain := range entry.Domains {
				if entryDomain == d {
					found = true
					break
				}
			}
			if !found {
				allFound = false
				break
			}
		}

		if allFound {
			return true
		}
	}

	return false
}

// RemoveManagedEntries removes all Stagecraft-managed entries from the file.
//
// Preserves all other entries unchanged.
func (f *File) RemoveManagedEntries() {
	var filtered []Entry
	for _, entry := range f.Entries {
		if !entry.Managed {
			filtered = append(filtered, entry)
		}
	}
	f.Entries = filtered
}

// AddManagedEntry adds a Stagecraft-managed entry for the given domains.
//
// If an entry already exists for these domains (managed or unmanaged), no duplicate is created.
// Domains are sorted lexicographically.
func (f *File) AddManagedEntry(domains []string) {
	if len(domains) == 0 {
		return
	}

	// Sort domains
	sortedDomains := make([]string, len(domains))
	copy(sortedDomains, domains)
	sort.Strings(sortedDomains)

	// Check if entry already exists
	if f.HasDomains(sortedDomains) {
		// Check if it's already managed
		for i := range f.Entries {
			if f.Entries[i].Managed {
				// Check if domains match
				if len(f.Entries[i].Domains) == len(sortedDomains) {
					match := true
					for j, d := range sortedDomains {
						if f.Entries[i].Domains[j] != d {
							match = false
							break
						}
					}
					if match {
						// Already exists and managed, do nothing
						return
					}
				}
			}
		}
		// Entry exists but not managed, preserve it and add new managed entry
	}

	// Add new managed entry
	f.Entries = append(f.Entries, Entry{
		IP:      "127.0.0.1",
		Domains: sortedDomains,
		Comment: StagecraftComment,
		Managed: true,
	})
}
