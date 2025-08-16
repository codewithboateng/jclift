package reporting

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/codewithboateng/jclift/internal/ir"
)

// WriteDiffJSON writes a placeholder diff between two runs.
// TODO: replace with real findings diff (new/removed/changed).
func WriteDiffJSON(baseID, headID, outDir string, base, head *ir.Run) (string, error) {
	payload := map[string]any{
		"base_id":         baseID,
		"head_id":         headID,
		"base_job_count":  len(base.Jobs),
		"head_job_count":  len(head.Jobs),
		"note":            "TODO: implement real findings diff",
	}
	path := filepath.Join(outDir, "diff_"+baseID+"__"+headID+".json")

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return "", err
	}
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(payload); err != nil {
		return "", err
	}
	return path, nil
}
