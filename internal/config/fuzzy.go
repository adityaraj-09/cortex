package config

import (
	"strings"
)

// LevenshteinDistance calculates the edit distance between two strings.
// This is the minimum number of single-character edits (insertions, deletions,
// or substitutions) required to change one string into the other.
func LevenshteinDistance(s1, s2 string) int {
	s1 = strings.ToLower(s1)
	s2 = strings.ToLower(s2)

	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	// Create matrix
	rows := len(s1) + 1
	cols := len(s2) + 1
	matrix := make([][]int, rows)
	for i := range matrix {
		matrix[i] = make([]int, cols)
	}

	// Initialize first row and column
	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	// Fill in the rest of the matrix
	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 1
			if s1[i-1] == s2[j-1] {
				cost = 0
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

// FindClosestMatch finds the closest matching string from a list of candidates.
// Returns the closest match and its distance. If no candidates are provided,
// returns empty string and -1.
// maxDistance limits how different a match can be (0 = exact match only).
// A typical threshold is len(input)/2 or 2-3 for short strings.
func FindClosestMatch(input string, candidates []string, maxDistance int) (string, int) {
	if len(candidates) == 0 {
		return "", -1
	}

	bestMatch := ""
	bestDistance := -1

	for _, candidate := range candidates {
		distance := LevenshteinDistance(input, candidate)

		// Only consider if within max distance threshold
		if maxDistance > 0 && distance > maxDistance {
			continue
		}

		if bestDistance == -1 || distance < bestDistance {
			bestDistance = distance
			bestMatch = candidate
		}
	}

	return bestMatch, bestDistance
}

// SuggestClosestMatch returns a suggestion string if a close match is found.
// Returns empty string if no good match is found.
func SuggestClosestMatch(input string, candidates []string) string {
	if len(candidates) == 0 {
		return ""
	}

	// Use a threshold based on the input length
	// For short strings (1-4 chars), allow up to 2 edits
	// For longer strings, allow up to len/2 edits
	maxDistance := len(input) / 2
	if maxDistance < 2 {
		maxDistance = 2
	}
	if maxDistance > 5 {
		maxDistance = 5 // Cap at 5 to avoid suggesting very different strings
	}

	match, distance := FindClosestMatch(input, candidates, maxDistance)

	// Only suggest if the match is reasonably close
	if match != "" && distance > 0 && distance <= maxDistance {
		return match
	}

	return ""
}

// min returns the minimum of three integers.
func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
