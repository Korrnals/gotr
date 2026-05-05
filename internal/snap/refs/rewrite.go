// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package refs

import (
	"strconv"
	"strings"
)

// Rewrite substitutes every occurrence of `attachments/get/<oldID>`
// inside text with `attachments/get/<newID>` from idMap. Numeric IDs
// only — md5 references are left untouched and counted as "not
// rewritten" via the second return value (skipped count).
//
// Returns the new text, the number of successful substitutions, and
// the number of refs left untouched (md5 or unmapped numeric IDs).
func Rewrite(text string, idMap map[int64]int64) (out string, rewritten, skipped int) {
	if text == "" {
		return text, 0, 0
	}
	out = attachmentURLRe.ReplaceAllStringFunc(text, func(match string) string {
		sub := attachmentURLRe.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		token := sub[1]
		if isHexMD5(token) {
			skipped++
			return match
		}
		oldID, err := strconv.ParseInt(token, 10, 64)
		if err != nil {
			skipped++
			return match
		}
		newID, ok := idMap[oldID]
		if !ok {
			skipped++
			return match
		}
		// Replace just the trailing /<oldID> token to keep host/path
		// prefix intact. Substring guaranteed unique within match
		// because the regex anchors on /attachments/get/<token>.
		needle := "/" + token
		replacement := "/" + strconv.FormatInt(newID, 10)
		rewritten++
		return strings.Replace(match, needle, replacement, 1)
	})
	return out, rewritten, skipped
}
