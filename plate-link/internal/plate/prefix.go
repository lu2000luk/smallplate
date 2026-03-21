package plate

import "strings"

func PrefixKey(plateID string, key string) string {
	return plateID + ":" + key
}

func PrefixPattern(plateID string, pattern string) string {
	if pattern == "" {
		return plateID + ":*"
	}
	return plateID + ":" + pattern
}

func UnprefixKey(plateID string, key string) string {
	prefix := plateID + ":"
	if strings.HasPrefix(key, prefix) {
		return strings.TrimPrefix(key, prefix)
	}
	return key
}
