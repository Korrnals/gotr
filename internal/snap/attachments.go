package snap

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Korrnals/gotr/internal/concurrent"
	"github.com/Korrnals/gotr/internal/models/data"
)

// AttachmentsAPI defines the API methods needed for attachment snapshot operations.
type AttachmentsAPI interface {
	GetAttachmentsForCase(ctx context.Context, caseID int64) (data.GetAttachmentsResponse, error)
	DownloadAttachment(ctx context.Context, attachmentID int64) (io.ReadCloser, error)
	AddAttachmentToCase(ctx context.Context, caseID int64, filePath string) (*data.AttachmentResponse, error)
}

// AttachmentManifest records saved attachment metadata within a snapshot.
type AttachmentManifest struct {
	Attachments []AttachmentEntry `json:"attachments"`
}

// AttachmentEntry describes a single saved attachment file.
type AttachmentEntry struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
	File        string `json:"file"`
	Compressed  bool   `json:"compressed"`
}

const attachmentsDir = "attachments"
const attachManifestFile = "attachments.json"

// PromptFunc asks the user whether to save a large attachment.
// Returns true to save, false to skip.
type PromptFunc func(att data.Attachment, maxMB int) bool

// SaveCaseAttachments downloads and stores attachment binaries for a case.
// Respects SnapConfig: save_binary mode, max file size, gzip compression.
// Use promptFn to interactively ask about large files (pass nil for non-interactive).
// Returns the number of attachments saved.
func SaveCaseAttachments(
	ctx context.Context,
	api AttachmentsAPI,
	store *Store,
	snapID string,
	caseID int64,
	cfg SnapConfig,
	promptFn ...PromptFunc,
) (int, error) {
	if cfg.AttachSaveBinary == "never" {
		return 0, nil
	}

	attachments, err := api.GetAttachmentsForCase(ctx, caseID)
	if err != nil {
		return 0, fmt.Errorf("list attachments for case %d: %w", caseID, err)
	}
	if len(attachments) == 0 {
		return 0, nil
	}

	toSave := selectAttachmentsForSave(attachments, cfg, resolvePromptFunc(promptFn))
	if len(toSave) == 0 {
		return 0, nil
	}

	dir := filepath.Join(store.SnapDir(snapID), attachmentsDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return 0, fmt.Errorf("create attachments dir: %w", err)
	}

	var mu sync.Mutex
	var manifest AttachmentManifest

	downloadOne := func(att data.Attachment, _ int) (*AttachmentEntry, error) {
		entry, err := downloadAndSave(ctx, api, dir, att, cfg.AttachCompress)
		if err != nil {
			return nil, err
		}
		mu.Lock()
		manifest.Attachments = append(manifest.Attachments, *entry)
		mu.Unlock()
		return entry, nil
	}

	saved, err := runAttachmentDownloads(ctx, toSave, downloadOne)
	if err != nil {
		return saved, err
	}
	if saved == 0 {
		return 0, nil
	}
	if err := saveAttachManifest(store, snapID, &manifest); err != nil {
		return saved, err
	}

	return saved, nil
}

// resolvePromptFunc returns the optional prompt callback when provided.
func resolvePromptFunc(promptFn []PromptFunc) PromptFunc {
	if len(promptFn) > 0 && promptFn[0] != nil {
		return promptFn[0]
	}
	return nil
}

// selectAttachmentsForSave applies snapshot policy and optional prompting.
func selectAttachmentsForSave(attachments []data.Attachment, cfg SnapConfig, prompt PromptFunc) []data.Attachment {
	maxBytes := int64(cfg.AttachMaxFileMB) * 1024 * 1024
	toSave := make([]data.Attachment, 0, len(attachments))
	for _, att := range attachments {
		if shouldSave(att, cfg, maxBytes) || shouldPromptSave(att, cfg, maxBytes, prompt) {
			toSave = append(toSave, att)
		}
	}
	return toSave
}

// shouldPromptSave checks whether an oversized attachment should be saved after prompt.
func shouldPromptSave(att data.Attachment, cfg SnapConfig, maxBytes int64, prompt PromptFunc) bool {
	if cfg.AttachSaveBinary != "auto" || att.Size <= maxBytes || !cfg.AttachPromptAboveThresh || prompt == nil {
		return false
	}
	return prompt(att, cfg.AttachMaxFileMB)
}

// runAttachmentDownloads executes download tasks sequentially or in parallel.
func runAttachmentDownloads(ctx context.Context, toSave []data.Attachment, downloadOne func(data.Attachment, int) (*AttachmentEntry, error)) (int, error) {
	saved := 0
	if len(toSave) >= concurrentThreshold {
		results, _ := concurrent.ParallelMap(ctx, toSave, defaultParallelism, downloadOne)
		for _, r := range results {
			if r.Error != nil {
				return saved, fmt.Errorf("save attachment: %w", r.Error)
			}
			saved++
		}
	} else {
		for i, att := range toSave {
			if _, err := downloadOne(att, i); err != nil {
				return saved, fmt.Errorf("save attachment %d (%s): %w", att.ID, att.Name, err)
			}
			saved++
		}
	}
	return saved, nil
}

// RestoreCaseAttachments uploads saved attachment binaries back to a case.
// Returns the number of attachments restored.
func RestoreCaseAttachments(
	ctx context.Context,
	api AttachmentsAPI,
	store *Store,
	snapID string,
	caseID int64,
) (int, error) {
	manifest, err := loadAttachManifest(store, snapID)
	if err != nil {
		return 0, err
	}
	if manifest == nil || len(manifest.Attachments) == 0 {
		return 0, nil
	}

	dir := filepath.Join(store.SnapDir(snapID), attachmentsDir)
	restored := 0

	for _, entry := range manifest.Attachments {
		srcPath := filepath.Join(dir, entry.File)

		uploadPath := srcPath
		tmpPath := ""
		if entry.Compressed {
			// Decompress to temp file for upload.
			tmp, err := decompressToTemp(srcPath, entry.Name)
			if err != nil {
				return restored, fmt.Errorf("decompress attachment %d: %w", entry.ID, err)
			}
			uploadPath = tmp
			tmpPath = tmp
		}

		_, err := api.AddAttachmentToCase(ctx, caseID, uploadPath)
		if err != nil {
			if tmpPath != "" {
				_ = os.Remove(tmpPath)
			}
			return restored, fmt.Errorf("upload attachment %d (%s): %w", entry.ID, entry.Name, err)
		}
		if tmpPath != "" {
			_ = os.Remove(tmpPath)
		}
		restored++
	}

	return restored, nil
}

// shouldSave determines whether an attachment should be saved based on config.
func shouldSave(att data.Attachment, cfg SnapConfig, maxBytes int64) bool {
	switch cfg.AttachSaveBinary {
	case "always":
		return true
	case "auto":
		return att.Size <= maxBytes
	default: // "never" handled upstream
		return false
	}
}

// downloadAndSave downloads a single attachment and writes it to disk.
func downloadAndSave(
	ctx context.Context,
	api AttachmentsAPI,
	dir string,
	att data.Attachment,
	compress bool,
) (*AttachmentEntry, error) {
	body, err := api.DownloadAttachment(ctx, att.ID)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	name := sanitizeAttachName(att.Name, att.ID)
	filename := name
	if compress {
		filename += ".gz"
	}

	outPath := filepath.Join(dir, filename)
	f, err := os.Create(outPath)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	var w io.WriteCloser = f
	if compress {
		gw := gzip.NewWriter(f)
		defer gw.Close()
		w = gw
	}

	if _, err := io.Copy(w, body); err != nil {
		return nil, fmt.Errorf("write: %w", err)
	}

	return &AttachmentEntry{
		ID:          att.ID,
		Name:        att.Name,
		Size:        att.Size,
		ContentType: att.ContentType,
		File:        filename,
		Compressed:  compress,
	}, nil
}

// decompressToTemp decompresses a gzip file into a named temp file.
func decompressToTemp(gzPath, originalName string) (string, error) {
	f, err := os.Open(gzPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return "", fmt.Errorf("gzip reader: %w", err)
	}
	defer gr.Close()

	tmp, err := os.CreateTemp("", "gotr-attach-*-"+originalName)
	if err != nil {
		return "", err
	}
	defer tmp.Close()

	if _, err := io.Copy(tmp, gr); err != nil {
		os.Remove(tmp.Name())
		return "", err
	}

	return tmp.Name(), nil
}

func saveAttachManifest(store *Store, snapID string, manifest *AttachmentManifest) error {
	dir := store.SnapDir(snapID)
	path := filepath.Join(dir, attachManifestFile)

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("save attachment manifest: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(manifest)
}

func loadAttachManifest(store *Store, snapID string) (*AttachmentManifest, error) {
	dir := store.SnapDir(snapID)
	path := filepath.Join(dir, attachManifestFile)

	raw, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("load attachment manifest: %w", err)
	}

	var manifest AttachmentManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return nil, fmt.Errorf("decode attachment manifest: %w", err)
	}
	return &manifest, nil
}

// sanitizeAttachName produces a safe filename from attachment metadata.
func sanitizeAttachName(name string, id int64) string {
	if name == "" {
		return fmt.Sprintf("%d", id)
	}
	safe := strings.Map(func(r rune) rune {
		if r == '/' || r == '\\' || r == ':' || r == '*' || r == '?' || r == '"' || r == '<' || r == '>' || r == '|' {
			return '_'
		}
		return r
	}, name)
	return fmt.Sprintf("%d_%s", id, safe)
}
