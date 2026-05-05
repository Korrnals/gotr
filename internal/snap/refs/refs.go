// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

// Package refs scans markdown text fields of TestRail entities for
// attachment references (index.php?/attachments/get/<id-or-md5>) and
// produces a structured index that drives reference rewriting on
// snapshot restore.
package refs

import (
	"regexp"
	"strconv"
	"strings"
)

// Reference is a single attachment URL found inside an entity's
// markdown field. Either AttachmentID (numeric) or AttachmentMD5 (hex)
// is set, never both.
type Reference struct {
	// AttachmentID is the numeric attachment ID. Set when the URL was
	// /attachments/get/<digits>.
	AttachmentID int64 `json:"attachment_id,omitempty"`
	// AttachmentMD5 is the legacy 32-hex MD5 attachment locator. Set
	// when the URL was /attachments/get/<32-hex>.
	AttachmentMD5 string `json:"attachment_md5,omitempty"`
	// URL is the exact matched URL token (relative or absolute) as it
	// appears in the source text. Used as the substitution key at
	// rewrite time.
	URL string `json:"url"`
	// Field is the dotted path of the field that owned this URL. For
	// `custom_steps_separated[]` items the index is included, e.g.
	// `custom_steps_separated[2].content`.
	Field string `json:"field"`
}

// EntityRefs aggregates all references discovered inside a single
// entity. EntityType is one of "case", "result", "run", "plan",
// "milestone".
type EntityRefs struct {
	EntityType string      `json:"entity_type"`
	EntityID   int64       `json:"entity_id"`
	Refs       []Reference `json:"refs"`
}

// attachmentURLRe matches every flavour of TestRail attachment URL we
// expect inside markdown text: absolute (https://host/...), root-relative
// (/index.php?...) and bare (index.php?...). The capture group returns
// either a numeric ID or a 32-character hex MD5. A single regex avoids
// the overhead of pulling in goldmark just for URL extraction.
var attachmentURLRe = regexp.MustCompile(
	`(?i)(?:https?://[^\s)\]]+?)?(?:/)?index\.php\?/attachments/get/([0-9]+|[a-f0-9]{32})`,
)

// ScanText returns every attachment reference found inside text in
// document order. fieldPath is propagated as Reference.Field so that
// the caller can later rewrite the exact field that contained the
// match.
func ScanText(text, fieldPath string) []Reference {
	if text == "" {
		return nil
	}
	matches := attachmentURLRe.FindAllStringSubmatchIndex(text, -1)
	if len(matches) == 0 {
		return nil
	}
	out := make([]Reference, 0, len(matches))
	for _, m := range matches {
		// m[0]:m[1] is the full match, m[2]:m[3] is the first capture.
		fullURL := text[m[0]:m[1]]
		token := text[m[2]:m[3]]
		ref := Reference{URL: fullURL, Field: fieldPath}
		if isHexMD5(token) {
			ref.AttachmentMD5 = strings.ToLower(token)
		} else if id, err := strconv.ParseInt(token, 10, 64); err == nil {
			ref.AttachmentID = id
		} else {
			// Defensive: regex shouldn't allow this, skip rather than
			// emit a bogus reference.
			continue
		}
		out = append(out, ref)
	}
	return out
}

// isHexMD5 reports whether token is a 32-character lowercase hex
// string. The regex already enforces 32-char and [a-f0-9], but we
// re-check defensively to keep ScanText side-effect free if the regex
// is ever tightened or relaxed.
func isHexMD5(token string) bool {
	if len(token) != 32 {
		return false
	}
	for i := 0; i < len(token); i++ {
		c := token[i]
		switch {
		case c >= '0' && c <= '9':
		case c >= 'a' && c <= 'f':
		case c >= 'A' && c <= 'F':
		default:
			return false
		}
	}
	return true
}
