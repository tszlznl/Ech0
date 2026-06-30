// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025-2026 lin-snow

package uuid

import "testing"

func TestIsValid(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{name: "canonical v4", in: "550e8400-e29b-41d4-a716-446655440000", want: true},
		{name: "nil uuid", in: "00000000-0000-0000-0000-000000000000", want: true},
		{name: "uppercase", in: "550E8400-E29B-41D4-A716-446655440000", want: true},
		{name: "empty string", in: "", want: false},
		{name: "not a uuid", in: "hello-world", want: false},
		{name: "missing segment", in: "550e8400-e29b-41d4-a716", want: false},
		{name: "too long", in: "550e8400-e29b-41d4-a716-446655440000-extra", want: false},
		{name: "bad hex char", in: "zz0e8400-e29b-41d4-a716-446655440000", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValid(tt.in); got != tt.want {
				t.Fatalf("IsValid(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestNewV7RoundTrip(t *testing.T) {
	id, err := NewV7()
	if err != nil {
		t.Fatalf("NewV7() error = %v", err)
	}
	if !IsValid(id) {
		t.Fatalf("NewV7() produced invalid uuid: %q", id)
	}
}

func TestNewV7Unique(t *testing.T) {
	const n = 1000
	seen := make(map[string]struct{}, n)
	for i := 0; i < n; i++ {
		id, err := NewV7()
		if err != nil {
			t.Fatalf("NewV7() error = %v", err)
		}
		if !IsValid(id) {
			t.Fatalf("NewV7() produced invalid uuid: %q", id)
		}
		if _, dup := seen[id]; dup {
			t.Fatalf("NewV7() produced duplicate uuid: %q", id)
		}
		seen[id] = struct{}{}
	}
}

func TestMustNewV7(t *testing.T) {
	id := MustNewV7()
	if !IsValid(id) {
		t.Fatalf("MustNewV7() produced invalid uuid: %q", id)
	}
	// Second call must differ.
	if other := MustNewV7(); other == id {
		t.Fatalf("MustNewV7() returned identical ids: %q", id)
	}
}
