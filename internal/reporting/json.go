package reporting

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/codewithboateng/jclift/internal/ir"
)

func WriteJSON(runID, outDir string, run *ir.Run) (string, error) {
	path := filepath.Join(outDir, runID+".json")
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(run); err != nil {
		return "", err
	}
	return path, nil
}
