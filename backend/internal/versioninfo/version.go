package versioninfo

import (
	"fmt"
	"strconv"
	"strings"
)

// Version is a validated AIWeLink release version.
type Version struct {
	raw      string
	upstream string
	parts    []int
}

// Parse validates and parses an AIWeLink version without a leading v.
func Parse(value string) (Version, error) {
	if value == "" || strings.TrimSpace(value) != value {
		return Version{}, fmt.Errorf("invalid AIWeLink version %q", value)
	}

	segments := strings.Split(value, "-")
	if len(segments) != 2 {
		return Version{}, fmt.Errorf("AIWeLink version %q must contain one revision separator", value)
	}

	upstreamParts := strings.Split(segments[0], ".")
	if len(upstreamParts) != 3 {
		return Version{}, fmt.Errorf("AIWeLink version %q must contain a three-part upstream version", value)
	}

	revisionParts := strings.Split(segments[1], ".")
	if len(revisionParts) == 0 {
		return Version{}, fmt.Errorf("AIWeLink version %q must contain a revision", value)
	}

	parts := make([]int, 0, len(upstreamParts)+len(revisionParts))
	for _, part := range upstreamParts {
		parsed, err := parseCanonicalInteger(part, false)
		if err != nil {
			return Version{}, fmt.Errorf("invalid upstream component in %q: %w", value, err)
		}
		parts = append(parts, parsed)
	}
	for _, part := range revisionParts {
		parsed, err := parseCanonicalInteger(part, true)
		if err != nil {
			return Version{}, fmt.Errorf("invalid AIWeLink revision component in %q: %w", value, err)
		}
		parts = append(parts, parsed)
	}

	return Version{raw: value, upstream: segments[0], parts: parts}, nil
}

func parseCanonicalInteger(value string, positive bool) (int, error) {
	if value == "" {
		return 0, fmt.Errorf("component is empty")
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 || strconv.Itoa(parsed) != value {
		return 0, fmt.Errorf("component %q is not a canonical non-negative integer", value)
	}
	if positive && parsed == 0 {
		return 0, fmt.Errorf("component %q must be positive", value)
	}
	return parsed, nil
}

func (v Version) String() string {
	return v.raw
}

func (v Version) Upstream() string {
	return v.upstream
}

// Validate confirms that full is valid and based on the exact upstream value.
func Validate(full, upstream string) error {
	parsed, err := Parse(full)
	if err != nil {
		return err
	}
	if _, err := parseUpstream(upstream); err != nil {
		return err
	}
	if parsed.Upstream() != upstream {
		return fmt.Errorf("AIWeLink version %q is based on %q, not %q", full, parsed.Upstream(), upstream)
	}
	return nil
}

func parseUpstream(value string) ([3]int, error) {
	raw := value
	parts := strings.Split(value, ".")
	if len(parts) != 3 {
		return [3]int{}, fmt.Errorf("upstream version %q must contain three components", value)
	}
	var parsed [3]int
	for i, part := range parts {
		parsedPart, err := parseCanonicalInteger(part, false)
		if err != nil {
			return [3]int{}, fmt.Errorf("invalid upstream version %q: %w", raw, err)
		}
		parsed[i] = parsedPart
	}
	return parsed, nil
}

// Compare returns -1, 0, or 1 when left is older, equal to, or newer than right.
func Compare(left, right string) (int, error) {
	leftVersion, err := Parse(left)
	if err != nil {
		return 0, err
	}
	rightVersion, err := Parse(right)
	if err != nil {
		return 0, err
	}

	length := len(leftVersion.parts)
	if len(rightVersion.parts) > length {
		length = len(rightVersion.parts)
	}
	for i := 0; i < length; i++ {
		leftPart := componentAt(leftVersion.parts, i)
		rightPart := componentAt(rightVersion.parts, i)
		if leftPart < rightPart {
			return -1, nil
		}
		if leftPart > rightPart {
			return 1, nil
		}
	}
	return 0, nil
}

func componentAt(parts []int, index int) int {
	if index >= len(parts) {
		return 0
	}
	return parts[index]
}
