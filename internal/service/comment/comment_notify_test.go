package service

import "testing"

func TestUseCommentRecipient(t *testing.T) {
	tests := []struct {
		name string
		kind string
		want bool
	}{
		{
			name: "status uses commenter email",
			kind: "status",
			want: true,
		},
		{
			name: "hot uses commenter email",
			kind: "hot",
			want: true,
		},
		{
			name: "created keeps owner email",
			kind: "created",
			want: false,
		},
		{
			name: "other kinds keep owner email",
			kind: "test",
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := useCommentRecipient(tc.kind)
			if got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestParseValidEmail(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
		ok    bool
	}{
		{
			name:  "valid email",
			input: "author@example.com",
			want:  "author@example.com",
			ok:    true,
		},
		{
			name:  "trimmed valid email",
			input: "  author@example.com  ",
			want:  "author@example.com",
			ok:    true,
		},
		{
			name:  "empty email",
			input: "   ",
			want:  "",
			ok:    false,
		},
		{
			name:  "invalid format",
			input: "not-an-email",
			want:  "",
			ok:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := parseValidEmail(tc.input)
			if ok != tc.ok {
				t.Fatalf("expected ok=%v, got %v", tc.ok, ok)
			}
			if got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}
