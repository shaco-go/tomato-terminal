package types

import (
	"encoding/json"
	"errors"
	"os"
)

var CharMap = cloneCharMap(defaultCharMap)

func LoadCharMap(filename string) error {
	raw, err := os.ReadFile(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	customMap := make(map[string]string)
	if err := json.Unmarshal(raw, &customMap); err != nil {
		return err
	}

	CharMap = customMap
	return nil
}

func cloneCharMap(src map[string]string) map[string]string {
	dst := make(map[string]string, len(src))
	for key, val := range src {
		dst[key] = val
	}
	return dst
}
