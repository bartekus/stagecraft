// SPDX-License-Identifier: AGPL-3.0-or-later

// Feature: DEV_HOSTS
// Spec: spec/dev/hosts.md

package hosts

import "runtime"

// FilePath returns the platform-specific hosts file path.
//
// Returns:
//   - Linux/macOS: "/etc/hosts"
//   - Windows: "C:\\Windows\\System32\\drivers\\etc\\hosts"
func FilePath() string {
	switch runtime.GOOS {
	case "windows":
		return `C:\Windows\System32\drivers\etc\hosts`
	default: // linux, darwin, etc.
		return "/etc/hosts"
	}
}
