package addon

import (
	"encoding/json"
	"os"
)

func LoadDataset(path string) ([]DatasetEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var entries []DatasetEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	return entries, nil
}
