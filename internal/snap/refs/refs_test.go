// Copyright (c) 2025 Igor "Breezefall" Vasilenko
// See LICENSE.md for details

package refs

import "testing"

func TestScanText_NumericID(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []Reference
	}{
		{
			name: "bare relative",
			in:   "see index.php?/attachments/get/123 here",
			want: []Reference{{AttachmentID: 123, URL: "index.php?/attachments/get/123", Field: "f"}},
		},
		{
			name: "root relative",
			in:   "url: /index.php?/attachments/get/42",
			want: []Reference{{AttachmentID: 42, URL: "/index.php?/attachments/get/42", Field: "f"}},
		},
		{
			name: "absolute https",
			in:   "https://tr.example.com/index.php?/attachments/get/7",
			want: []Reference{{AttachmentID: 7, URL: "https://tr.example.com/index.php?/attachments/get/7", Field: "f"}},
		},
		{
			name: "markdown image",
			in:   "![](/index.php?/attachments/get/55)",
			want: []Reference{{AttachmentID: 55, URL: "/index.php?/attachments/get/55", Field: "f"}},
		},
		{
			name: "markdown link",
			in:   "[file](https://h/index.php?/attachments/get/9)",
			want: []Reference{{AttachmentID: 9, URL: "https://h/index.php?/attachments/get/9", Field: "f"}},
		},
		{
			name: "multiple",
			in:   "a /index.php?/attachments/get/1 b /index.php?/attachments/get/2",
			want: []Reference{
				{AttachmentID: 1, URL: "/index.php?/attachments/get/1", Field: "f"},
				{AttachmentID: 2, URL: "/index.php?/attachments/get/2", Field: "f"},
			},
		},
		{name: "no refs", in: "nothing here", want: nil},
		{name: "empty", in: "", want: nil},
		{name: "false positive (different path)", in: "https://x/foo/bar/123", want: nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ScanText(tc.in, "f")
			if len(got) != len(tc.want) {
				t.Fatalf("len: got %d, want %d (got=%+v)", len(got), len(tc.want), got)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("[%d] got %+v, want %+v", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestScanText_MD5(t *testing.T) {
	in := "![](index.php?/attachments/get/abcdef0123456789abcdef0123456789)"
	got := ScanText(in, "field")
	if len(got) != 1 {
		t.Fatalf("len=%d", len(got))
	}
	if got[0].AttachmentMD5 != "abcdef0123456789abcdef0123456789" {
		t.Errorf("md5=%q", got[0].AttachmentMD5)
	}
	if got[0].AttachmentID != 0 {
		t.Errorf("AttachmentID should be 0 for md5 ref, got %d", got[0].AttachmentID)
	}
}

func TestScanText_MD5UpperNormalized(t *testing.T) {
	in := "url: index.php?/attachments/get/ABCDEF0123456789ABCDEF0123456789"
	got := ScanText(in, "f")
	if len(got) != 1 {
		t.Fatalf("len=%d", len(got))
	}
	if got[0].AttachmentMD5 != "abcdef0123456789abcdef0123456789" {
		t.Errorf("expected lowercased md5, got %q", got[0].AttachmentMD5)
	}
}

func TestScanText_NotMatchedPaths(t *testing.T) {
	for _, in := range []string{
		"https://example.com/index.php?/cases/view/123",
		"index.php?/attachments/list/123",
		"plain text 12345",
	} {
		if got := ScanText(in, "f"); len(got) != 0 {
			t.Errorf("expected no refs for %q, got %+v", in, got)
		}
	}
}
