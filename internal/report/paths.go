package report

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Category is a report classification bucket under ~/.gotr/reports/<Category>/.
type Category string

const (
	// CategoryMigrations — successful sync_* runs that produced a snapshot.
	CategoryMigrations Category = "migrations"
	// CategoryCoverage — compare/coverage audit reports (gotr_migration_*).
	CategoryCoverage Category = "coverage"
	// CategoryRollbacks — snap rollback reports.
	CategoryRollbacks Category = "rollbacks"
	// CategoryNoSnapshot — migrations that ran without creating a snapshot.
	CategoryNoSnapshot Category = "no-snapshot"
	// CategoryCleanupAttachments — `gotr attachments cleanup` deletion reports.
	CategoryCleanupAttachments Category = "cleanup-attachments"
	// CategoryTestrail — raw TestRail API dumps (gotr reports/plans/... list output).
	CategoryTestrail Category = "testrail"
	// CategoryUnclassified — filenames that do not match any known pattern.
	CategoryUnclassified Category = "_unclassified"

	// DefaultLabel is used when a report has no explicit label.
	DefaultLabel = "default"
)

// Classification is the result of inspecting a report filename.
// Path returns the directory (relative to the reports root) where the report
// should live: <Category>[/p<Project>]/<Label>/<YearMonth>/.
type Classification struct {
	Category  Category
	Label     string // "default" when not provided
	YearMonth string // e.g. "2026-04"; empty if indeterminable
	Project   int    // >0 for CategoryTestrail; otherwise 0
}

// RelDir returns the relative directory for a classified report, without the
// trailing filename. Caller joins the base reports dir.
func (c Classification) RelDir() string {
	switch c.Category {
	case CategoryTestrail:
		proj := fmt.Sprintf("p%d", c.Project)
		if c.Project <= 0 {
			proj = "p0"
		}
		if c.YearMonth == "" {
			return filepath.Join(string(c.Category), proj)
		}
		return filepath.Join(string(c.Category), proj, c.YearMonth)
	case CategoryUnclassified:
		return string(c.Category)
	default:
		label := c.Label
		if label == "" {
			label = DefaultLabel
		}
		if c.YearMonth == "" {
			return filepath.Join(string(c.Category), label)
		}
		return filepath.Join(string(c.Category), label, c.YearMonth)
	}
}

// tsRe matches the ISO-like timestamp used in report filenames, e.g.
// migration-20260424T020240Z-... or rollback-20260424T020240Z-...
var tsRe = regexp.MustCompile(`(\d{4})(\d{2})(\d{2})T\d{6}Z`)

// projectRe extracts a project id from "p<N>" anywhere in a filename.
var projectRe = regexp.MustCompile(`(?:^|[_-])p(\d+)(?:[_\-.]|$)`)

// ClassifyReport inspects a report filename (basename or relative path) and
// returns its classification. It never returns an error: unrecognized files
// fall into CategoryUnclassified.
//
// Recognized patterns (checked in order):
//
//	migration-<ts>-no_snapshot*        -> no-snapshot / default
//	migration-<ts>-sync_*              -> migrations / <label=default> (label comes from caller when available)
//	rollback-<ts>-*                    -> rollbacks / default
//	gotr_migration_<TAG>_*             -> coverage / <TAG> (TAG becomes the label)
//	testrail_*|*_p<N>_*                -> testrail / p<N>
//	everything else                    -> _unclassified
//
// When the caller knows the label (e.g. --snap-label for a freshly generated
// report), it should call ClassifyReportWithLabel to override the default.
func ClassifyReport(filename string) Classification {
	return ClassifyReportWithLabel(filename, "")
}

// ClassifyReportWithLabel is like ClassifyReport but uses explicitLabel when
// the filename itself doesn't carry a label. An empty explicitLabel falls back
// to DefaultLabel.
func ClassifyReportWithLabel(filename, explicitLabel string) Classification {
	name := filepath.Base(filename)
	low := strings.ToLower(name)

	ym := extractYearMonth(name)
	label := explicitLabel
	if label == "" {
		label = DefaultLabel
	}

	switch {
	case strings.HasPrefix(low, "migration-") && strings.Contains(low, "-no_snapshot"):
		// No snapshot → always bucketed under default (label is not meaningful).
		return Classification{Category: CategoryNoSnapshot, Label: DefaultLabel, YearMonth: ym}

	case strings.HasPrefix(low, "migration-"):
		return Classification{Category: CategoryMigrations, Label: label, YearMonth: ym}

	case strings.HasPrefix(low, "rollback-"):
		return Classification{Category: CategoryRollbacks, Label: label, YearMonth: ym}

	case strings.HasPrefix(low, "cleanup-attachments-"):
		return Classification{Category: CategoryCleanupAttachments, Label: label, YearMonth: ym}

	case strings.HasPrefix(low, "gotr_migration_"):
		// gotr_migration_<TAG>_p<A>_to_p<B>.pdf → TAG becomes the label when caller did not override.
		tag := extractCoverageTag(name)
		if explicitLabel == "" && tag != "" {
			label = tag
		}
		return Classification{Category: CategoryCoverage, Label: label, YearMonth: ym}

	case strings.HasPrefix(low, "testrail_") || projectRe.MatchString(low):
		proj := extractProjectID(low)
		return Classification{Category: CategoryTestrail, Project: proj, YearMonth: ym}
	}

	return Classification{Category: CategoryUnclassified, Label: "", YearMonth: ym}
}

func extractYearMonth(name string) string {
	m := tsRe.FindStringSubmatch(name)
	if len(m) < 3 {
		return ""
	}
	return m[1] + "-" + m[2]
}

func extractProjectID(name string) int {
	m := projectRe.FindStringSubmatch(name)
	if len(m) < 2 {
		return 0
	}
	n := 0
	for _, r := range m[1] {
		n = n*10 + int(r-'0')
	}
	return n
}

// extractCoverageTag pulls the TAG out of "gotr_migration_<TAG>_p<A>_to_p<B>...".
// Returns "" if the shape does not match.
func extractCoverageTag(name string) string {
	const prefix = "gotr_migration_"
	if !strings.HasPrefix(strings.ToLower(name), prefix) {
		return ""
	}
	rest := name[len(prefix):]
	// Stop at the first "_p<digit>" marker which starts the project segment.
	idx := strings.Index(strings.ToLower(rest), "_p")
	if idx <= 0 {
		return ""
	}
	tag := rest[:idx]
	// A TAG should contain at least one non-numeric character to be considered
	// a real label (otherwise it is just a project id).
	hasLetter := false
	for _, r := range tag {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '-' {
			hasLetter = true
			break
		}
	}
	if !hasLetter {
		return ""
	}
	return tag
}

// ReportEntry is a single report file discovered under the reports root.
type ReportEntry struct {
	AbsPath  string    // absolute path on disk
	Rel      string    // path relative to the reports root (forward slashes)
	Basename string    // filename only
	ModTime  time.Time // filesystem mtime (used by "latest" lookups)
	Size     int64
}

// RecursiveListReports walks baseDir and returns every regular file whose
// extension looks like a report (.md, .pdf, .json, .txt). Hidden entries and
// the "_unclassified" bucket are included; INDEX.md is excluded from the
// returned slice because it is a derived artifact.
func RecursiveListReports(baseDir string) ([]ReportEntry, error) {
	var out []ReportEntry
	err := filepath.WalkDir(baseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		name := d.Name()
		if name == "INDEX.md" {
			return nil
		}
		if !reportLikeExt(name) {
			return nil
		}
		info, ierr := d.Info()
		if ierr != nil {
			return nil //nolint:nilerr // skip unreadable file rather than abort the walk
		}
		rel, rerr := filepath.Rel(baseDir, path)
		if rerr != nil {
			rel = path
		}
		out = append(out, ReportEntry{
			AbsPath:  path,
			Rel:      filepath.ToSlash(rel),
			Basename: name,
			ModTime:  info.ModTime(),
			Size:     info.Size(),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ModTime.After(out[j].ModTime) })
	return out, nil
}

func reportLikeExt(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".md", ".pdf", ".json", ".txt", ".csv":
		return true
	}
	return false
}

// ResolveReportPath finds a single report file in baseDir that matches the
// given input. Lookup rules (first match wins):
//
//  1. input == "latest"          -> newest report by mtime
//  2. input is an absolute path  -> returned as-is if it exists
//  3. input has path separators  -> joined with baseDir, used as-is
//  4. input matches a basename exactly in any subdirectory
//  5. input matches a basename without its extension (adds .md/.pdf/.json)
//
// Returns os.ErrNotExist when nothing matches.
func ResolveReportPath(baseDir, input string) (string, error) {
	if input == "" {
		return "", errors.New("empty report name")
	}
	if input == "latest" {
		entries, err := RecursiveListReports(baseDir)
		if err != nil {
			return "", err
		}
		if len(entries) == 0 {
			return "", fs.ErrNotExist
		}
		return entries[0].AbsPath, nil
	}
	if filepath.IsAbs(input) {
		if _, err := os.Stat(input); err != nil {
			return "", err
		}
		return input, nil
	}
	if strings.ContainsAny(input, "/\\") {
		cand := filepath.Join(baseDir, input)
		if _, err := os.Stat(cand); err == nil {
			return cand, nil
		}
	}

	entries, err := RecursiveListReports(baseDir)
	if err != nil {
		return "", err
	}
	// exact basename match
	for _, e := range entries {
		if e.Basename == input {
			return e.AbsPath, nil
		}
	}
	// basename without extension
	for _, e := range entries {
		noExt := strings.TrimSuffix(e.Basename, filepath.Ext(e.Basename))
		if noExt == input {
			return e.AbsPath, nil
		}
	}
	return "", fs.ErrNotExist
}

// IsFlatLayout returns true (and the number of reports found in the root) when
// baseDir contains at least one report-like file directly in its root and
// therefore predates the categorized hierarchy. Missing baseDir is not an
// error; it reports (false, 0, nil).
func IsFlatLayout(baseDir string) (flat bool, count int, err error) {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return false, 0, nil
		}
		return false, 0, err
	}
	n := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if e.Name() == "INDEX.md" {
			continue
		}
		if reportLikeExt(e.Name()) {
			n++
		}
	}
	return n > 0, n, nil
}
