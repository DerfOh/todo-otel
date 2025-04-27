package main

import "strings" // Import strings package for Contains

// contains checks if the query string is present in the text.
// Updated to use strings.Contains for simplicity and correctness.
func contains(text, query string) bool {
	// Basic validation
	if len(query) == 0 || len(text) == 0 {
		return false
	}
	// Case-insensitive comparison might be useful:
	// return strings.Contains(strings.ToLower(text), strings.ToLower(query))
	return strings.Contains(text, query)
}
