package snap

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/Korrnals/gotr/internal/models/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockAttachAPI implements AttachmentsAPI for testing.
type mockAttachAPI struct {
	attachments data.GetAttachmentsResponse
	downloads   map[int64][]byte
	uploaded    []uploadRecord
}

type uploadRecord struct {
	CaseID  int64
	Content []byte
}

func (m *mockAttachAPI) GetAttachmentsForCase(_ context.Context, caseID int64) (data.GetAttachmentsResponse, error) {
	return m.attachments, nil
}

func (m *mockAttachAPI) DownloadAttachment(_ context.Context, attachmentID int64) (io.ReadCloser, error) {
	d, ok := m.downloads[attachmentID]
	if !ok {
		return nil, io.EOF
	}
	return io.NopCloser(bytes.NewReader(d)), nil
}

func (m *mockAttachAPI) AddAttachmentToCase(_ context.Context, caseID int64, filePath string) (*data.AttachmentResponse, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	m.uploaded = append(m.uploaded, uploadRecord{CaseID: caseID, Content: content})
	return &data.AttachmentResponse{AttachmentID: 999}, nil
}

func setupAttachStore(t *testing.T) (*Store, string) {
	t.Helper()
	dir := t.TempDir()
	store, err := NewStoreAt(dir)
	require.NoError(t, err)
	return store, dir
}

func createSnapForAttach(t *testing.T, store *Store, snapID string) {
	t.Helper()
	meta := &Meta{
		ID:         snapID,
		Category:   "cases",
		Operation:  OpDelete,
		EntityType: "case",
		Status:     StatusAvailable,
		DataFile:   "data.json",
	}
	require.NoError(t, store.SaveMeta(meta))
}

func TestSaveCaseAttachments_NeverMode(t *testing.T) {
	store, _ := setupAttachStore(t)
	snapID := "cases/20260413T120000_delete_1"
	createSnapForAttach(t, store, snapID)

	api := &mockAttachAPI{}
	cfg := DefaultConfig()
	cfg.AttachSaveBinary = "never"

	count, err := SaveCaseAttachments(context.Background(), api, store, snapID, 1, cfg)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestSaveCaseAttachments_NoAttachments(t *testing.T) {
	store, _ := setupAttachStore(t)
	snapID := "cases/20260413T120000_delete_1"
	createSnapForAttach(t, store, snapID)

	api := &mockAttachAPI{attachments: nil}
	cfg := DefaultConfig()

	count, err := SaveCaseAttachments(context.Background(), api, store, snapID, 1, cfg)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestSaveCaseAttachments_AutoMode_SmallFile(t *testing.T) {
	store, _ := setupAttachStore(t)
	snapID := "cases/20260413T120000_delete_1"
	createSnapForAttach(t, store, snapID)

	content := []byte("hello attachment binary data")
	api := &mockAttachAPI{
		attachments: data.GetAttachmentsResponse{
			{ID: 10, Name: "test.png", Size: int64(len(content)), ContentType: "image/png"},
		},
		downloads: map[int64][]byte{10: content},
	}
	cfg := DefaultConfig() // auto, 10MB, compress=true

	count, err := SaveCaseAttachments(context.Background(), api, store, snapID, 1, cfg)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Verify manifest
	manifest, err := loadAttachManifest(store, snapID)
	require.NoError(t, err)
	require.Len(t, manifest.Attachments, 1)
	assert.Equal(t, int64(10), manifest.Attachments[0].ID)
	assert.Equal(t, "test.png", manifest.Attachments[0].Name)
	assert.True(t, manifest.Attachments[0].Compressed)
	assert.Equal(t, "10_test.png.gz", manifest.Attachments[0].File)

	// Verify gzip file exists and is valid
	gzPath := filepath.Join(store.SnapDir(snapID), attachmentsDir, "10_test.png.gz")
	f, err := os.Open(gzPath)
	require.NoError(t, err)
	defer f.Close()

	gr, err := gzip.NewReader(f)
	require.NoError(t, err)
	defer gr.Close()

	got, err := io.ReadAll(gr)
	require.NoError(t, err)
	assert.Equal(t, content, got)
}

func TestSaveCaseAttachments_AutoMode_LargeFile_Skipped(t *testing.T) {
	store, _ := setupAttachStore(t)
	snapID := "cases/20260413T120000_delete_1"
	createSnapForAttach(t, store, snapID)

	api := &mockAttachAPI{
		attachments: data.GetAttachmentsResponse{
			{ID: 20, Name: "huge.bin", Size: 50 * 1024 * 1024}, // 50MB > 10MB threshold
		},
		downloads: map[int64][]byte{20: []byte("data")},
	}
	cfg := DefaultConfig() // auto, max 10MB

	count, err := SaveCaseAttachments(context.Background(), api, store, snapID, 1, cfg)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestSaveCaseAttachments_AlwaysMode_LargeFile_Saved(t *testing.T) {
	store, _ := setupAttachStore(t)
	snapID := "cases/20260413T120000_delete_1"
	createSnapForAttach(t, store, snapID)

	content := []byte("large file content")
	api := &mockAttachAPI{
		attachments: data.GetAttachmentsResponse{
			{ID: 30, Name: "huge.bin", Size: 50 * 1024 * 1024, ContentType: "application/octet-stream"},
		},
		downloads: map[int64][]byte{30: content},
	}
	cfg := DefaultConfig()
	cfg.AttachSaveBinary = "always"

	count, err := SaveCaseAttachments(context.Background(), api, store, snapID, 1, cfg)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestSaveCaseAttachments_NoCompress(t *testing.T) {
	store, _ := setupAttachStore(t)
	snapID := "cases/20260413T120000_delete_1"
	createSnapForAttach(t, store, snapID)

	content := []byte("raw content")
	api := &mockAttachAPI{
		attachments: data.GetAttachmentsResponse{
			{ID: 40, Name: "doc.txt", Size: int64(len(content))},
		},
		downloads: map[int64][]byte{40: content},
	}
	cfg := DefaultConfig()
	cfg.AttachCompress = false

	count, err := SaveCaseAttachments(context.Background(), api, store, snapID, 1, cfg)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Verify raw file (not gzip)
	rawPath := filepath.Join(store.SnapDir(snapID), attachmentsDir, "40_doc.txt")
	got, err := os.ReadFile(rawPath)
	require.NoError(t, err)
	assert.Equal(t, content, got)

	manifest, err := loadAttachManifest(store, snapID)
	require.NoError(t, err)
	assert.False(t, manifest.Attachments[0].Compressed)
}

func TestRestoreCaseAttachments_Success(t *testing.T) {
	store, _ := setupAttachStore(t)
	snapID := "cases/20260413T120000_delete_1"
	createSnapForAttach(t, store, snapID)

	// Create a compressed attachment file
	dir := filepath.Join(store.SnapDir(snapID), attachmentsDir)
	require.NoError(t, os.MkdirAll(dir, 0o755))

	content := []byte("original attachment data")
	gzPath := filepath.Join(dir, "10_test.png.gz")
	f, err := os.Create(gzPath)
	require.NoError(t, err)
	gw := gzip.NewWriter(f)
	_, err = gw.Write(content)
	require.NoError(t, err)
	require.NoError(t, gw.Close())
	require.NoError(t, f.Close())

	// Create manifest
	manifest := &AttachmentManifest{
		Attachments: []AttachmentEntry{
			{ID: 10, Name: "test.png", Size: int64(len(content)), File: "10_test.png.gz", Compressed: true},
		},
	}
	require.NoError(t, saveAttachManifest(store, snapID, manifest))

	api := &mockAttachAPI{}
	count, err := RestoreCaseAttachments(context.Background(), api, store, snapID, 42)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
	require.Len(t, api.uploaded, 1)
	assert.Equal(t, int64(42), api.uploaded[0].CaseID)

	// Verify the uploaded file was decompressed correctly
	assert.Equal(t, content, api.uploaded[0].Content)
}

func TestRestoreCaseAttachments_NoManifest(t *testing.T) {
	store, _ := setupAttachStore(t)
	snapID := "cases/20260413T120000_delete_1"
	createSnapForAttach(t, store, snapID)

	api := &mockAttachAPI{}
	count, err := RestoreCaseAttachments(context.Background(), api, store, snapID, 1)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestSaveCaseAttachments_PromptAccepts(t *testing.T) {
	store, _ := setupAttachStore(t)
	snapID := "cases/20260413T120000_delete_1"
	createSnapForAttach(t, store, snapID)

	bigContent := bytes.Repeat([]byte("x"), 100)
	api := &mockAttachAPI{
		attachments: data.GetAttachmentsResponse{
			{ID: 1, Name: "huge.bin", Size: 20 * 1024 * 1024, ContentType: "application/octet-stream"},
		},
		downloads: map[int64][]byte{1: bigContent},
	}

	cfg := SnapConfig{
		AttachSaveBinary:        "auto",
		AttachMaxFileMB:         10,
		AttachCompress:          false,
		AttachPromptAboveThresh: true,
	}

	// Prompt says yes → should save.
	prompt := func(_ data.Attachment, _ int) bool { return true }
	count, err := SaveCaseAttachments(context.Background(), api, store, snapID, 42, cfg, prompt)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestSaveCaseAttachments_PromptDeclines(t *testing.T) {
	store, _ := setupAttachStore(t)
	snapID := "cases/20260413T120000_delete_1"
	createSnapForAttach(t, store, snapID)

	api := &mockAttachAPI{
		attachments: data.GetAttachmentsResponse{
			{ID: 1, Name: "huge.bin", Size: 20 * 1024 * 1024, ContentType: "application/octet-stream"},
		},
		downloads: map[int64][]byte{1: []byte("data")},
	}

	cfg := SnapConfig{
		AttachSaveBinary:        "auto",
		AttachMaxFileMB:         10,
		AttachCompress:          false,
		AttachPromptAboveThresh: true,
	}

	// Prompt says no → should skip.
	prompt := func(_ data.Attachment, _ int) bool { return false }
	count, err := SaveCaseAttachments(context.Background(), api, store, snapID, 42, cfg, prompt)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestSaveCaseAttachments_NoPromptNoThresh(t *testing.T) {
	store, _ := setupAttachStore(t)
	snapID := "cases/20260413T120000_delete_1"
	createSnapForAttach(t, store, snapID)

	api := &mockAttachAPI{
		attachments: data.GetAttachmentsResponse{
			{ID: 1, Name: "huge.bin", Size: 20 * 1024 * 1024, ContentType: "application/octet-stream"},
		},
		downloads: map[int64][]byte{1: []byte("data")},
	}

	cfg := SnapConfig{
		AttachSaveBinary:        "auto",
		AttachMaxFileMB:         10,
		AttachCompress:          false,
		AttachPromptAboveThresh: false, // disabled
	}

	// No prompt, threshold disabled → large file skipped.
	count, err := SaveCaseAttachments(context.Background(), api, store, snapID, 42, cfg)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestShouldSave(t *testing.T) {
	maxBytes := int64(10 * 1024 * 1024)
	tests := []struct {
		name   string
		mode   string
		size   int64
		expect bool
	}{
		{"auto_small", "auto", 1024, true},
		{"auto_exact_threshold", "auto", maxBytes, true},
		{"auto_over_threshold", "auto", maxBytes + 1, false},
		{"always_large", "always", maxBytes + 1, true},
		{"never", "never", 100, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			att := data.Attachment{Size: tc.size}
			cfg := SnapConfig{AttachSaveBinary: tc.mode}
			assert.Equal(t, tc.expect, shouldSave(att, cfg, maxBytes))
		})
	}
}

func TestSanitizeAttachName(t *testing.T) {
	tests := []struct {
		name   string
		id     int64
		expect string
	}{
		{"simple.txt", 1, "1_simple.txt"},
		{"path/file.txt", 2, "2_path_file.txt"},
		{"", 3, "3"},
		{"bad:chars*here?.txt", 4, "4_bad_chars_here_.txt"},
	}
	for _, tc := range tests {
		t.Run(tc.expect, func(t *testing.T) {
			assert.Equal(t, tc.expect, sanitizeAttachName(tc.name, tc.id))
		})
	}
}

func TestSaveAndLoadAttachManifest(t *testing.T) {
	store, _ := setupAttachStore(t)
	snapID := "cases/20260413T120000_delete_1"
	createSnapForAttach(t, store, snapID)

	manifest := &AttachmentManifest{
		Attachments: []AttachmentEntry{
			{ID: 1, Name: "a.txt", Size: 100, File: "1_a.txt.gz", Compressed: true},
			{ID: 2, Name: "b.png", Size: 200, File: "2_b.png", Compressed: false},
		},
	}

	err := saveAttachManifest(store, snapID, manifest)
	require.NoError(t, err)

	loaded, err := loadAttachManifest(store, snapID)
	require.NoError(t, err)
	require.Len(t, loaded.Attachments, 2)
	assert.Equal(t, "a.txt", loaded.Attachments[0].Name)
	assert.Equal(t, "b.png", loaded.Attachments[1].Name)

	// Verify file is valid JSON
	path := filepath.Join(store.SnapDir(snapID), attachManifestFile)
	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.True(t, json.Valid(raw))
}
