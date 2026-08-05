//go:build unit

package versioninfo

import "testing"

func TestParseAcceptsAIWeLinkVersions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		upstream string
	}{
		{input: "0.1.170-1", upstream: "0.1.170"},
		{input: "0.1.170-2.4", upstream: "0.1.170"},
		{input: "12.34.567-8.9.10", upstream: "12.34.567"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse(%q) returned error: %v", tt.input, err)
			}
			if got.String() != tt.input {
				t.Fatalf("String() = %q, want %q", got.String(), tt.input)
			}
			if got.Upstream() != tt.upstream {
				t.Fatalf("Upstream() = %q, want %q", got.Upstream(), tt.upstream)
			}
		})
	}
}

func TestParseRejectsInvalidVersions(t *testing.T) {
	t.Parallel()

	invalid := []string{
		"",
		"v0.1.170-1",
		"0.1.170",
		"0.1-1",
		"0.1.170-0",
		"0.1.170-01",
		"0.1.170-1.0",
		"0.1.170-.1",
		"0.1.170-1.",
		"0.1.170-1-rc1",
		" 0.1.170-1",
	}

	for _, input := range invalid {
		input := input
		t.Run(input, func(t *testing.T) {
			t.Parallel()
			if _, err := Parse(input); err == nil {
				t.Fatalf("Parse(%q) succeeded, want error", input)
			}
		})
	}
}

func TestValidateRequiresMatchingUpstream(t *testing.T) {
	t.Parallel()

	if err := Validate("0.1.170-2.4", "0.1.170"); err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}
	if err := Validate("0.1.170-2.4", "0.1.169"); err == nil {
		t.Fatal("Validate succeeded with mismatched upstream")
	}
	if err := Validate("0.1.170-2.4", "v0.1.170"); err == nil {
		t.Fatal("Validate succeeded with invalid upstream")
	}
}

func TestCompareUsesBaselineAndEveryRevisionComponent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		left  string
		right string
		want  int
	}{
		{left: "0.1.170-2.4", right: "0.1.170-2", want: 1},
		{left: "0.1.170-2.4", right: "0.1.170-3", want: -1},
		{left: "0.1.171-1", right: "0.1.170-99", want: 1},
		{left: "0.1.170-2.4", right: "0.1.170-2.4", want: 0},
	}

	for _, tt := range tests {
		got, err := Compare(tt.left, tt.right)
		if err != nil {
			t.Fatalf("Compare(%q, %q) returned error: %v", tt.left, tt.right, err)
		}
		if got != tt.want {
			t.Fatalf("Compare(%q, %q) = %d, want %d", tt.left, tt.right, got, tt.want)
		}
	}
}

func TestCompareRejectsMalformedVersion(t *testing.T) {
	t.Parallel()

	if _, err := Compare("0.1.170", "0.1.170-1"); err == nil {
		t.Fatal("Compare accepted official-only version")
	}
}
