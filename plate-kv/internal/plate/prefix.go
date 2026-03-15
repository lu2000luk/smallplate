package plate

import "strings"

func PrefixKey(plateID string, key string) string {
	return plateID + ":" + key
}

func PrefixKeys(plateID string, keys []string) []string {
	prefixed := make([]string, 0, len(keys))
	for _, key := range keys {
		prefixed = append(prefixed, PrefixKey(plateID, key))
	}
	return prefixed
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

func UnprefixKeys(plateID string, keys []string) []string {
	trimmed := make([]string, 0, len(keys))
	for _, key := range keys {
		trimmed = append(trimmed, UnprefixKey(plateID, key))
	}
	return trimmed
}
