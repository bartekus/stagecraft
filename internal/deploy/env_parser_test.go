// SPDX-License-Identifier: AGPL-3.0-or-later

/*
Stagecraft - Stagecraft is a Go-based CLI that orchestrates local-first development and scalable single-host to multi-host deployments for multi-service applications powered by Docker Compose.

Copyright (C) 2025  Bartek Kus

This program is free software licensed under the terms of the GNU AGPL v3 or later.

See https://www.gnu.org/licenses/ for license details.
*/

package deploy

import (
	"testing"
)

// Test cases copied from internal/core/env/env_test.go to match actual behavior

func TestParseEnvFileInto_QuotedValues(t *testing.T) {
	env := make(map[string]string)
	content := `KEY1="value with spaces"
KEY2='single quoted'
KEY3="value with \"quotes\""
KEY4="value\nwith\ttabs"
`
	parseEnvFileInto(env, []byte(content))

	if env["KEY1"] != "value with spaces" {
		t.Errorf("expected KEY1='value with spaces', got %q", env["KEY1"])
	}
	if env["KEY2"] != "single quoted" {
		t.Errorf("expected KEY2='single quoted', got %q", env["KEY2"])
	}
	if env["KEY3"] != "value with \"quotes\"" {
		t.Errorf("expected KEY3 to handle escaped quotes, got %q", env["KEY3"])
	}
	if env["KEY4"] != "value\nwith\ttabs" {
		t.Errorf("expected KEY4 to handle escape sequences, got %q", env["KEY4"])
	}
}

func TestParseEnvFileInto_Comments(t *testing.T) {
	env := make(map[string]string)
	content := `# Full line comment
KEY1=value1 # Inline comment
KEY2="value # not a comment"
KEY3=value3
`
	parseEnvFileInto(env, []byte(content))

	if _, ok := env["#"]; ok {
		t.Error("full line comments should not create variables")
	}
	if env["KEY1"] != "value1" {
		t.Errorf("expected KEY1='value1' (inline comment removed), got %q", env["KEY1"])
	}
	if env["KEY2"] != "value # not a comment" {
		t.Errorf("expected KEY2 to preserve # in quoted string, got %q", env["KEY2"])
	}
	if env["KEY3"] != "value3" {
		t.Errorf("expected KEY3='value3', got %q", env["KEY3"])
	}
}

func TestParseEnvFileInto_ExportKeyword(t *testing.T) {
	env := make(map[string]string)
	content := `export KEY1=value1
KEY2=value2
`
	parseEnvFileInto(env, []byte(content))

	if env["KEY1"] != "value1" {
		t.Errorf("expected KEY1='value1', got %q", env["KEY1"])
	}
	if env["KEY2"] != "value2" {
		t.Errorf("expected KEY2='value2', got %q", env["KEY2"])
	}
}

func TestParseEnvFileInto_Whitespace(t *testing.T) {
	env := make(map[string]string)
	content := `KEY1 = value1
KEY2=" value2 "
KEY3 = " value3 "
`
	parseEnvFileInto(env, []byte(content))

	if env["KEY1"] != "value1" {
		t.Errorf("expected KEY1='value1', got %q", env["KEY1"])
	}
	if env["KEY2"] != " value2 " {
		t.Errorf("expected KEY2=' value2 ', got %q", env["KEY2"])
	}
	if env["KEY3"] != " value3 " {
		t.Errorf("expected KEY3=' value3 ', got %q", env["KEY3"])
	}
}
