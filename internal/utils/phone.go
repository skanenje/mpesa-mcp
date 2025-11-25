package utils

import "strings"

// FormatPhoneNumber ensures phone number is in format 254XXXXXXXXX
func FormatPhoneNumber(phone string) string {
	// Remove any spaces or special characters
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "+", "")

	// If starts with 0, replace with 254
	if strings.HasPrefix(phone, "0") {
		return "254" + phone[1:]
	}

	// If doesn't start with 254, add it
	if !strings.HasPrefix(phone, "254") {
		return "254" + phone
	}

	return phone
}
