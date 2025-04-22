// Package log is based on uber zap
package log

import (
	"encoding/json"
)

const (
	notAvaliable = "N/A"
)

// Mask content
func Mask(content string) string {
	bcontent := []byte(content)

	if !json.Valid(bcontent) {
		return notAvaliable
	}

	contentMap := map[string]interface{}{}

	if err := json.Unmarshal(bcontent, &contentMap); err != nil {
		return notAvaliable
	}

	newContentMap := maskMap(contentMap)

	newContent, err := json.Marshal(newContentMap)
	if err != nil {
		return notAvaliable
	}

	return string(newContent)
}

func maskMap(a map[string]interface{}) map[string]interface{} {
	newMap := map[string]interface{}{}

	for k, v := range a {
		switch t := v.(type) {
		case map[string]interface{}:
			newMap[k] = maskMap(t)
		default:
			value := a[k]

			if isSensitive(k) {
				value = "****"
			}

			newMap[k] = value
		}
	}

	return newMap
}

func isSensitive(key string) bool {
	switch key {
	case
		"access_token",
		"refresh_token":
		return true
	}

	return false
}
