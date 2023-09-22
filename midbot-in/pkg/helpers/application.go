package helpers

func IsValidID(code string) bool {
	if len(code) != 25 {
		return false
	}

	for _, char := range code {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}
