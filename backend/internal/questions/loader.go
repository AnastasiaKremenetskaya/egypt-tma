package questions

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Bank struct {
	Maat []MaatQuestion
	Seth []SethQuestion
}

func Load(dataDir string) (*Bank, error) {
	maat, err := loadMaat(filepath.Join(dataDir, "questions_maat.json"))
	if err != nil {
		return nil, fmt.Errorf("maat: %w", err)
	}
	seth, err := loadSeth(filepath.Join(dataDir, "questions_seth.json"))
	if err != nil {
		return nil, fmt.Errorf("seth: %w", err)
	}
	return &Bank{Maat: maat, Seth: seth}, nil
}

func loadMaat(path string) ([]MaatQuestion, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var qs []MaatQuestion
	if err := json.Unmarshal(data, &qs); err != nil {
		return nil, err
	}
	return qs, nil
}

func loadSeth(path string) ([]SethQuestion, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var qs []SethQuestion
	if err := json.Unmarshal(data, &qs); err != nil {
		return nil, err
	}
	return qs, nil
}
