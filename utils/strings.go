package utils

import "strings"

// ExtractVersion extracts version information from a string
// Examples: "Stockfish 16" -> "16", "Fruit 2.1" -> "2.1", "Toga II 3.0" -> "3.0", "Crafty-23.4" -> "23.4"
func ExtractVersion(name string) string {
	// First try splitting by hyphen (e.g., "Crafty-23.4")
	if strings.Contains(name, "-") {
		parts := strings.Split(name, "-")
		for i := len(parts) - 1; i >= 0; i-- {
			part := strings.TrimSpace(parts[i])
			if len(part) > 0 && strings.ContainsAny(part, "0123456789") {
				// Remove common prefixes like "v"
				part = strings.TrimPrefix(part, "v")
				part = strings.TrimPrefix(part, "V")
				return part
			}
		}
	}

	// Then try splitting by spaces
	parts := strings.Fields(name)
	for i := len(parts) - 1; i >= 0; i-- {
		part := parts[i]
		// Check if this looks like a version (contains digits)
		if len(part) > 0 && (strings.ContainsAny(part, "0123456789")) {
			// Remove common prefixes like "v"
			part = strings.TrimPrefix(part, "v")
			part = strings.TrimPrefix(part, "V")
			return part
		}
	}
	return ""
}

// TitleCase capitalizes the first letter of each word
// Words can be separated by hyphens or spaces
func TitleCase(s string) string {
	// Handle hyphen-separated words
	words := strings.Split(s, "-")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}
