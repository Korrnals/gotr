// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package refs

import "testing"

func TestRewrite_NumericReplaces(t *testing.T) {
	in := "before /index.php?/attachments/get/1 mid /index.php?/attachments/get/2 end"
	idMap := map[int64]int64{1: 100, 2: 200}
	out, rewritten, skipped := Rewrite(in, idMap)
	if rewritten != 2 || skipped != 0 {
		t.Errorf("rewritten=%d skipped=%d", rewritten, skipped)
	}
	want := "before /index.php?/attachments/get/100 mid /index.php?/attachments/get/200 end"
	if out != want {
		t.Errorf("\ngot:  %q\nwant: %q", out, want)
	}
}

func TestRewrite_UnmappedSkipped(t *testing.T) {
	in := "/index.php?/attachments/get/1 and /index.php?/attachments/get/9"
	idMap := map[int64]int64{1: 100}
	_, rewritten, skipped := Rewrite(in, idMap)
	if rewritten != 1 || skipped != 1 {
		t.Errorf("rewritten=%d skipped=%d", rewritten, skipped)
	}
}

func TestRewrite_MD5SkippedNotRewritten(t *testing.T) {
	in := "/index.php?/attachments/get/abcdef0123456789abcdef0123456789"
	out, rewritten, skipped := Rewrite(in, map[int64]int64{1: 100})
	if rewritten != 0 || skipped != 1 {
		t.Errorf("rewritten=%d skipped=%d", rewritten, skipped)
	}
	if out != in {
		t.Errorf("md5 ref must be left intact, got %q", out)
	}
}

func TestRewrite_PreservesHostAndQuery(t *testing.T) {
	in := "see [link](https://tr.example.com/index.php?/attachments/get/5)"
	out, rewritten, _ := Rewrite(in, map[int64]int64{5: 99})
	want := "see [link](https://tr.example.com/index.php?/attachments/get/99)"
	if rewritten != 1 || out != want {
		t.Errorf("\ngot:  %q\nwant: %q", out, want)
	}
}

func TestRewrite_EmptyAndNoMatch(t *testing.T) {
	if out, r, s := Rewrite("", nil); out != "" || r != 0 || s != 0 {
		t.Errorf("empty case failed")
	}
	if out, r, s := Rewrite("plain text", map[int64]int64{1: 2}); out != "plain text" || r != 0 || s != 0 {
		t.Errorf("no-match case failed")
	}
}
