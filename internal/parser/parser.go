package parser

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/codewithboateng/jclift/internal/ir"
)

type Diagnostics struct {
	Warnings []string
}

func Parse(path string) (ir.Run, Diagnostics) {
	var run ir.Run
	run.IRVersion = ir.Version
	run.Source = filepath.Clean(path)
	diags := Diagnostics{}

	_ = filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		name := strings.ToLower(d.Name())
		if !strings.HasSuffix(name, ".jcl") && !strings.HasSuffix(name, ".txt") {
			return nil
		}
		job, perr := parseFile(p)
		if perr == nil && len(job.Steps) > 0 {
			run.Jobs = append(run.Jobs, job)
		}
		return nil
	})

	if len(run.Jobs) == 0 {
		diags.Warnings = append(diags.Warnings, "no JCL-like files found or no steps parsed")
	}
	return run, diags
}

func parseFile(p string) (ir.Job, error) {
	f, err := os.Open(p)
	if err != nil {
		return ir.Job{}, err
	}
	defer f.Close()

	job := ir.Job{Name: strings.TrimSuffix(filepath.Base(p), filepath.Ext(p))}
	var steps []ir.Step
	var cur *ir.Step
	var sysinCapturing bool
	var sysinBuf strings.Builder

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimRight(sc.Text(), "\r\n")

		// Capture inline SYSIN (between "DD *" and "/*")
		if sysinCapturing {
			if strings.TrimSpace(line) == "/*" {
				sysinCapturing = false
				if cur != nil {
					for i := len(cur.DD) - 1; i >= 0; i-- {
						if strings.EqualFold(cur.DD[i].DDName, "SYSIN") {
							cur.DD[i].Content = sysinBuf.String()
							break
						}
					}
				}
				continue
			}
			sysinBuf.WriteString(line)
			sysinBuf.WriteByte('\n')
			continue
		}

		trim := strings.TrimSpace(line)
		if !strings.HasPrefix(trim, "//") {
			continue
		}
		card := strings.TrimSpace(trim[2:]) // after //

		// New step: //<STEP> EXEC PGM=...
		if idx := strings.Index(card, "EXEC"); idx != -1 && strings.Contains(strings.ToUpper(card), "PGM=") {
			if cur != nil {
				steps = append(steps, *cur)
			}
			stepName := firstField(card)
			pgm := parsePGM(card)
			cond := parseCOND(card)

			cur = &ir.Step{
				Name:       stepName,
				Program:    strings.ToUpper(pgm),
				Ordinal:    len(steps) + 1,
				Conditions: cond,
			}
			continue
		}

		// DD statement: //<DDNAME> DD ...
		if idx := strings.Index(card, "DD "); idx != -1 {
			if cur == nil {
				cur = &ir.Step{Name: "STEP1", Program: "UNKNOWN", Ordinal: len(steps) + 1}
			}
			ddname := strings.ToUpper(strings.TrimSpace(card[:idx]))
			rest := strings.TrimSpace(card[idx+3:])
			upper := strings.ToUpper(rest)

			dd := ir.DD{DDName: ddname}

			// SYSIN DD *  → start capture
			if ddname == "SYSIN" && strings.HasPrefix(strings.TrimSpace(rest), "*") {
				sysinCapturing = true
				sysinBuf.Reset()
				cur.DD = append(cur.DD, dd)
				continue
			}
			// SYSIN DD DUMMY
			if ddname == "SYSIN" && strings.HasPrefix(upper, "DUMMY") {
				dd.Content = "DUMMY"
			}

			// DSN=
			if i := strings.Index(upper, "DSN="); i != -1 {
				val := rest[i+4:]
				end := indexAny(val, ", ")
				if end == -1 {
					end = len(val)
				}
				dd.Dataset = strings.TrimSpace(val[:end])
			}
			// DISP=
			if i := strings.Index(upper, "DISP="); i != -1 {
				val := rest[i+5:]
				if strings.HasPrefix(val, "(") {
					if j := strings.Index(val, ")"); j != -1 {
						dd.DISP = strings.TrimSpace(val[:j+1])
					}
				} else {
					end := indexAny(val, ", ")
					if end == -1 {
						end = len(val)
					}
					dd.DISP = strings.TrimSpace(val[:end])
				}
			}
			// SPACE= (raw capture for now)
			if i := strings.Index(upper, "SPACE="); i != -1 {
				val := strings.TrimSpace(rest[i+6:])
				end := indexAny(val, " ")
				if end == -1 {
					end = len(val)
				}
				dd.Space = strings.TrimRight(val[:end], ",")
			}

			cur.DD = append(cur.DD, dd)
			continue
		}
	}
	if cur != nil {
		steps = append(steps, *cur)
	}
	job.Steps = steps
	return job, sc.Err()
}

func firstField(s string) string {
	fs := strings.Fields(s)
	if len(fs) > 0 {
		return fs[0]
	}
	return "STEP"
}

func parsePGM(card string) string {
	u := strings.ToUpper(card)
	i := strings.Index(u, "PGM=")
	if i == -1 {
		return "UNKNOWN"
	}
	val := card[i+4:]
	end := indexAny(val, ", ")
	if end == -1 {
		end = len(val)
	}
	return strings.Trim(strings.TrimSpace(val[:end]), ",")
}

func parseCOND(card string) string {
	u := strings.ToUpper(card)
	i := strings.Index(u, "COND=")
	if i == -1 {
		return ""
	}
	rest := card[i+5:]
	// Capture until end or next blank after a keyword split; simple heuristic
	return strings.TrimSpace(rest)
}

func indexAny(s, any string) int {
	min := -1
	for _, ch := range any {
		if idx := strings.IndexRune(s, ch); idx != -1 {
			if min == -1 || idx < min {
				min = idx
			}
		}
	}
	return min
}
