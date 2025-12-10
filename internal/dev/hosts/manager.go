// SPDX-License-Identifier: AGPL-3.0-or-later

// Feature: DEV_HOSTS
// Spec: spec/dev/hosts.md

package hosts

import (
	"context"
	"fmt"
	"os"
	"time"
)

// Manager manages hosts file entries for dev domains.
type Manager interface {
	// AddEntries adds dev domain entries to the hosts file.
	// Entries are marked as Stagecraft-managed and are idempotent.
	AddEntries(ctx context.Context, domains []string) error

	// RemoveEntries removes Stagecraft-managed entries for the given domains from the hosts file.
	// Only entries marked as Stagecraft-managed are removed.
	RemoveEntries(ctx context.Context, domains []string) error

	// Cleanup removes all Stagecraft-managed entries from the hosts file.
	Cleanup(ctx context.Context) error
}

// Options captures hosts file behavior.
type Options struct {
	HostsFilePath string // Optional override, default is platform-specific
	Verbose       bool
}

// manager implements Manager interface.
type manager struct {
	hostsPath string
	verbose   bool
}

// NewManager creates a new hosts file manager for the current platform.
func NewManager() Manager {
	return &manager{
		hostsPath: FilePath(),
		verbose:   false,
	}
}

// NewManagerWithOptions creates a manager with explicit options (for testing).
func NewManagerWithOptions(opts Options) Manager {
	hostsPath := opts.HostsFilePath
	if hostsPath == "" {
		hostsPath = FilePath()
	}

	return &manager{
		hostsPath: hostsPath,
		verbose:   opts.Verbose,
	}
}

// AddEntries adds dev domain entries to the hosts file.
func (m *manager) AddEntries(ctx context.Context, domains []string) error {
	if len(domains) == 0 {
		return nil
	}

	// Parse hosts file (with retry for file locking)
	file, err := m.parseWithRetry(ctx)
	if err != nil {
		return fmt.Errorf("parse hosts file: %w", err)
	}

	// Add managed entry (idempotent)
	file.AddManagedEntry(domains)

	// Write file (with retry for file locking)
	if err := m.writeWithRetry(ctx, file); err != nil {
		return fmt.Errorf("write hosts file: %w", err)
	}

	return nil
}

// RemoveEntries removes Stagecraft-managed entries for the given domains.
func (m *manager) RemoveEntries(ctx context.Context, domains []string) error {
	if len(domains) == 0 {
		return nil
	}

	// Parse hosts file
	file, err := m.parseWithRetry(ctx)
	if err != nil {
		return fmt.Errorf("parse hosts file: %w", err)
	}

	// Remove entries matching the domains
	var filtered []Entry
	for _, entry := range file.Entries {
		if entry.Managed {
			// Check if this entry matches any of the requested domains
			match := false
			for _, d := range domains {
				for _, entryDomain := range entry.Domains {
					if entryDomain == d {
						match = true
						break
					}
				}
				if match {
					break
				}
			}
			if !match {
				// Keep this managed entry (different domains)
				filtered = append(filtered, entry)
			}
			// Otherwise, skip it (remove)
		} else {
			// Keep non-managed entries
			filtered = append(filtered, entry)
		}
	}
	file.Entries = filtered

	// Write file
	if err := m.writeWithRetry(ctx, file); err != nil {
		return fmt.Errorf("write hosts file: %w", err)
	}

	return nil
}

// Cleanup removes all Stagecraft-managed entries from the hosts file.
func (m *manager) Cleanup(ctx context.Context) error {
	// Parse hosts file
	file, err := m.parseWithRetry(ctx)
	if err != nil {
		return fmt.Errorf("parse hosts file: %w", err)
	}

	// Remove all managed entries
	file.RemoveManagedEntries()

	// Write file
	if err := m.writeWithRetry(ctx, file); err != nil {
		return fmt.Errorf("write hosts file: %w", err)
	}

	return nil
}

// parseWithRetry parses the hosts file with retry logic for file locking.
func (m *manager) parseWithRetry(ctx context.Context) (*File, error) {
	var lastErr error
	for i := 0; i < 3; i++ {
		if i > 0 {
			// Exponential backoff: 100ms, 200ms, 400ms
			//nolint:gosec // G115: i is bounded (0-2), no overflow possible
			backoff := time.Duration(100*(1<<uint(i-1))) * time.Millisecond
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		file, err := ParseFile(m.hostsPath)
		if err == nil {
			return file, nil
		}

		// Check if it's a permission error
		if os.IsPermission(err) {
			return nil, fmt.Errorf("cannot modify hosts file: permission denied. Run with sudo or administrator privileges, or use --no-hosts to skip hosts file modification")
		}

		lastErr = err
	}

	return nil, fmt.Errorf("hosts file is locked by another process. Please close other applications that may be using the hosts file: %w", lastErr)
}

// writeWithRetry writes the hosts file with retry logic for file locking.
func (m *manager) writeWithRetry(ctx context.Context, file *File) error {
	var lastErr error
	for i := 0; i < 3; i++ {
		if i > 0 {
			// Exponential backoff: 100ms, 200ms, 400ms
			//nolint:gosec // G115: i is bounded (0-2), no overflow possible
			backoff := time.Duration(100*(1<<uint(i-1))) * time.Millisecond
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		err := WriteFile(m.hostsPath, file)
		if err == nil {
			return nil
		}

		// Check if it's a permission error
		if os.IsPermission(err) {
			return fmt.Errorf("cannot modify hosts file: permission denied. Run with sudo or administrator privileges, or use --no-hosts to skip hosts file modification")
		}

		lastErr = err
	}

	return fmt.Errorf("hosts file is locked by another process. Please close other applications that may be using the hosts file: %w", lastErr)
}
