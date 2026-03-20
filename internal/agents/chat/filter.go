package chat

import "strings"

// IsEscapeSequence checks if a string is or contains a terminal escape sequence
// that should be filtered out (OSC responses, cursor position reports, mouse events, etc.)
// This is the canonical implementation - use this everywhere instead of duplicating logic.
func IsEscapeSequence(s string) bool {
	if len(s) == 0 {
		return false
	}

	// OSC responses start with ] (e.g., ]11;rgb:...)
	if s[0] == ']' {
		return true
	}

	// CSI sequences start with [ (e.g., [38;3R for cursor position)
	if s[0] == '[' {
		return true
	}

	// SGR mouse events start with < (e.g., <0;53;32M)
	if s[0] == '<' {
		return true
	}

	// Contains escape character
	if strings.ContainsRune(s, '\x1b') {
		return true
	}

	// RGB color responses (may come without leading ])
	if strings.Contains(s, "rgb:") || strings.Contains(s, "fafa") {
		return true
	}

	// Cursor position reports and mouse events: end with R, M, or m after digits/semicolons
	if len(s) > 1 {
		lastChar := s[len(s)-1]
		if lastChar == 'R' || lastChar == 'M' || lastChar == 'm' {
			allValid := true
			for i := 0; i < len(s)-1; i++ {
				c := s[i]
				if c != ';' && c != '<' && (c < '0' || c > '9') {
					allValid = false
					break
				}
			}
			if allValid {
				return true
			}
		}
	}

	// OSC responses starting with digits (missing leading ])
	// Pattern: "11;..." or "10;..." etc.
	if len(s) > 2 && s[0] >= '0' && s[0] <= '9' && strings.Contains(s, ";") {
		return true
	}

	// Control characters (except normal whitespace)
	for _, r := range s {
		if r < 32 && r != '\n' && r != '\r' && r != '\t' {
			return true
		}
	}

	return false
}

// CleanEscapeSequences removes escape sequence patterns from a string.
// Returns the cleaned string. Useful for cleaning textarea values before submission.
func CleanEscapeSequences(s string) string {
	if len(s) == 0 {
		return s
	}

	// Quick check - only clean if suspicious patterns exist
	if !strings.Contains(s, "rgb:") &&
		!strings.Contains(s, "fafa") &&
		!strings.ContainsRune(s, '\x1b') &&
		!strings.ContainsAny(s, "[]<") {
		return s
	}

	cleaned := s

	// Remove patterns iteratively
	for {
		changed := false

		// Remove "NN;rgb:XXXX/XXXX/XXXX" patterns
		if idx := strings.Index(cleaned, ";rgb:"); idx != -1 {
			start := idx
			for start > 0 && ((cleaned[start-1] >= '0' && cleaned[start-1] <= '9') || cleaned[start-1] == ';') {
				start--
			}
			end := idx + 5
			for end < len(cleaned) && (isHexCharFilter(cleaned[end]) || cleaned[end] == '/') {
				end++
			}
			cleaned = cleaned[:start] + cleaned[end:]
			changed = true
		}

		// Remove "fafa" patterns with surrounding hex/separators
		if idx := strings.Index(cleaned, "fafa"); idx != -1 {
			start := idx
			for start > 0 && (isHexCharFilter(cleaned[start-1]) || cleaned[start-1] == '/' || cleaned[start-1] == ';' || cleaned[start-1] == ':') {
				start--
			}
			for start > 0 && (cleaned[start-1] >= '0' && cleaned[start-1] <= '9') {
				start--
			}
			end := idx + 4
			for end < len(cleaned) && (isHexCharFilter(cleaned[end]) || cleaned[end] == '/' || cleaned[end] == ';' || cleaned[end] == ':') {
				end++
			}
			cleaned = cleaned[:start] + cleaned[end:]
			changed = true
		}

		// Remove cursor position reports: "NN;NNR"
		if idx := strings.Index(cleaned, "R"); idx > 0 {
			start := idx
			hasDigits := false
			hasSemi := false
			for start > 0 {
				c := cleaned[start-1]
				if c >= '0' && c <= '9' {
					hasDigits = true
					start--
				} else if c == ';' {
					hasSemi = true
					start--
				} else {
					break
				}
			}
			if hasDigits && hasSemi {
				cleaned = cleaned[:start] + cleaned[idx+1:]
				changed = true
			}
		}

		// Remove escape character sequences
		if idx := strings.IndexRune(cleaned, '\x1b'); idx != -1 {
			end := idx + 1
			for end < len(cleaned) && !((cleaned[end] >= 'A' && cleaned[end] <= 'Z') || (cleaned[end] >= 'a' && cleaned[end] <= 'z')) {
				end++
			}
			if end < len(cleaned) {
				end++ // Include terminating letter
			}
			cleaned = cleaned[:idx] + cleaned[end:]
			changed = true
		}

		if !changed {
			break
		}
	}

	// Remove bracket sequences
	for _, bracket := range []string{"[", "]", "<"} {
		for strings.Contains(cleaned, bracket) {
			idx := strings.Index(cleaned, bracket)
			end := idx + 1
			for end < len(cleaned) && cleaned[end] != ' ' && cleaned[end] != '\n' {
				if (cleaned[end] >= 'A' && cleaned[end] <= 'Z') || (cleaned[end] >= 'a' && cleaned[end] <= 'z') {
					end++
					break
				}
				end++
			}
			segment := cleaned[idx:end]
			if strings.ContainsAny(segment, "0123456789;") {
				cleaned = cleaned[:idx] + cleaned[end:]
			} else {
				break
			}
		}
	}

	return strings.TrimSpace(cleaned)
}

func isHexCharFilter(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}
