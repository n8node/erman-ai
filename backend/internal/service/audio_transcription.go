package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/erman-ai/erman-ai/internal/config"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
)

const (
	AudioTranscriptionMaxUploadBytes int64 = 500 << 20 // 500 MiB server limit
	AudioTranscriptionAsyncMaxBytes  int64 = 50 << 20  // Yandex async body limit with margin
	AudioTranscriptionSyncMaxBytes     int64 = 900 << 10 // ~900 KiB sync margin
	AudioTranscriptionSyncMaxSec       float64 = 25
	AudioTranscriptionSegmentSec       float64 = 600 // 10 min async segments
)

var (
	ErrAudioTranscriptionForbidden         = errors.New("audio transcription access denied")
	ErrAudioTranscriptionInvalidAudio      = errors.New("invalid audio file")
	ErrAudioTranscriptionSettings          = errors.New("invalid audio transcription settings")
	ErrAudioTranscriptionAlreadyProcessing = errors.New("audio transcription already in progress")
)

func AudioTranscriptionHasAccess(role string, explicitlyEnabled bool) bool {
	return role == "superadmin" || explicitlyEnabled
}

type AudioTranscriptionService struct {
	cfg      *config.Config
	assetsDir string
	repo     *repository.AudioTranscriptionRepository
	runs     *repository.ToolRunRepository
	plans    *repository.PlanRepository
	billing  *BillingService
	llm      *LLMService
	strategy *StrategyLLMSettingsService
	usageLog *repository.UsageLogRepository
	logger   *slog.Logger
}

func NewAudioTranscriptionService(
	cfg *config.Config,
	repo *repository.AudioTranscriptionRepository,
	runs *repository.ToolRunRepository,
	plans *repository.PlanRepository,
	billing *BillingService,
	llm *LLMService,
	strategy *StrategyLLMSettingsService,
	usageLog *repository.UsageLogRepository,
	logger *slog.Logger,
) *AudioTranscriptionService {
	return &AudioTranscriptionService{
		cfg:       cfg,
		assetsDir: cfg.AudioTranscriptionAssetsDir,
		repo:      repo,
		runs:      runs,
		plans:     plans,
		billing:   billing,
		llm:       llm,
		strategy:  strategy,
		usageLog:  usageLog,
		logger:    logger,
	}
}

func (s *AudioTranscriptionService) EnsureAssetDirs() error {
	for _, sub := range []string{"uploads", "work", "transcripts"} {
		if err := os.MkdirAll(filepath.Join(s.assetsDir, sub), 0o750); err != nil {
			return err
		}
	}
	return nil
}

func (s *AudioTranscriptionService) CheckAccess(ctx context.Context, userID, role string) error {
	explicit, err := s.repo.HasExplicitAccess(ctx, userID)
	if err != nil {
		return err
	}
	if !AudioTranscriptionHasAccess(role, explicit) {
		return ErrAudioTranscriptionForbidden
	}
	return nil
}

func ValidateAudioTranscriptionFile(data []byte, contentType string) error {
	if len(data) == 0 || int64(len(data)) > AudioTranscriptionMaxUploadBytes {
		return ErrAudioTranscriptionInvalidAudio
	}
	ct := strings.ToLower(strings.TrimSpace(contentType))
	allowed := []string{
		"audio/mpeg", "audio/mp3", "audio/wav", "audio/x-wav", "audio/wave",
		"audio/ogg", "audio/opus", "audio/webm", "audio/mp4", "audio/x-m4a",
		"video/mp4", "video/webm",
	}
	for _, a := range allowed {
		if ct == a || strings.HasPrefix(ct, a+";") {
			return nil
		}
	}
	return ErrAudioTranscriptionInvalidAudio
}

func (s *AudioTranscriptionService) resolveSpeechKitParams(ctx context.Context, settings model.AudioTranscriptionSettings) (YandexSpeechKitParams, error) {
	strategyRec, err := s.strategy.GetStored(ctx)
	if err != nil {
		return YandexSpeechKitParams{}, err
	}
	creds := s.llm.CredentialsFromStored(strategyRec.Config)
	params := YandexSpeechKitParams{
		APIKey:       creds.YandexKey,
		FolderID:     creds.YandexFolderID,
		LanguageCode: settings.LanguageCode,
		Model:        settings.Model,
	}
	if strings.TrimSpace(params.APIKey) == "" || strings.TrimSpace(params.FolderID) == "" {
		return params, fmt.Errorf("%w: set Yandex API key and folder ID in Admin → AI Strategy LLM or server .env", ErrYandexSpeechKitNotConfigured)
	}
	return params, nil
}

func (s *AudioTranscriptionService) UploadAndTranscribe(ctx context.Context, userID, role, originalName, contentType string, data []byte) (*model.AudioTranscriptionFile, *model.ToolRun, error) {
	if err := s.CheckAccess(ctx, userID, role); err != nil {
		return nil, nil, err
	}
	if err := ValidateAudioTranscriptionFile(data, contentType); err != nil {
		return nil, nil, err
	}
	if err := s.billing.CheckToolLimit(ctx, userID, model.AudioTranscriptionToolSlug); err != nil {
		return nil, nil, err
	}
	if err := s.EnsureAssetDirs(); err != nil {
		return nil, nil, err
	}

	ext := audioExtensionFromContentType(contentType, originalName)
	path := filepath.Join(s.assetsDir, "uploads", randomAudioAssetName(ext))
	if err := os.WriteFile(path, data, 0o640); err != nil {
		return nil, nil, err
	}

	file, err := s.repo.CreateFile(ctx, userID, filepath.Base(originalName), path, contentType, int64(len(data)), nil)
	if err != nil {
		_ = os.Remove(path)
		return nil, nil, err
	}

	run, err := s.startTranscription(ctx, file.ID, userID, role)
	if err != nil {
		assetPath, transcriptPath, _ := s.repo.DeleteFile(ctx, file.ID, userID)
		_ = os.Remove(assetPath)
		if transcriptPath != nil {
			_ = os.Remove(*transcriptPath)
		}
		return nil, nil, err
	}
	return file, run, nil
}

func (s *AudioTranscriptionService) RetranscribeFile(ctx context.Context, fileID, userID, role string) (*model.ToolRun, error) {
	if err := s.CheckAccess(ctx, userID, role); err != nil {
		return nil, err
	}
	if err := s.billing.CheckToolLimit(ctx, userID, model.AudioTranscriptionToolSlug); err != nil {
		return nil, err
	}
	if _, err := s.repo.GetFile(ctx, fileID, userID); err != nil {
		return nil, err
	}
	runs, err := s.repo.ListFileRuns(ctx, fileID, userID)
	if err != nil {
		return nil, err
	}
	if len(runs) > 0 {
		latest := runs[0]
		if latest.Status == model.RunStatusPending || latest.Status == model.RunStatusProcessing {
			return nil, ErrAudioTranscriptionAlreadyProcessing
		}
	}
	if path, _, err := s.repo.TranscriptPath(ctx, fileID, userID); err == nil {
		_ = os.Remove(path)
	}
	_ = s.repo.ClearTranscript(ctx, fileID, userID)
	return s.startTranscription(ctx, fileID, userID, role)
}

func (s *AudioTranscriptionService) startTranscription(ctx context.Context, fileID, userID, role string) (*model.ToolRun, error) {
	if err := s.CheckAccess(ctx, userID, role); err != nil {
		return nil, err
	}
	up, err := s.plans.GetUserPlan(ctx, userID)
	if err != nil {
		return nil, err
	}
	input, _ := json.Marshal(model.AudioTranscriptionRunInput{FileID: fileID})
	run, err := s.runs.CreatePending(ctx, userID, model.AudioTranscriptionToolSlug, up.PlanSlug, input)
	if err != nil {
		return nil, err
	}
	if err := s.repo.AttachRun(ctx, fileID, run.ID); err != nil {
		_ = s.runs.UpdateRunError(ctx, run.ID, "failed to attach file run")
		return nil, err
	}
	go s.processRun(run.ID, fileID, userID)
	return run, nil
}

func (s *AudioTranscriptionService) processRun(runID, fileID, userID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Minute)
	defer cancel()

	fail := func(err error) {
		msg := err.Error()
		if len(msg) > 500 {
			msg = msg[:500]
		}
		_ = s.runs.UpdateRunError(context.Background(), runID, msg)
	}

	if err := s.runs.UpdateStatus(ctx, runID, model.RunStatusProcessing); err != nil {
		return
	}

	settingsRec, err := s.repo.GetSettings(ctx)
	if err != nil {
		fail(err)
		return
	}
	settings := settingsRec.Settings
	if strings.TrimSpace(settings.Model) == "" {
		settings.Model = "general"
	}
	if strings.TrimSpace(settings.LanguageCode) == "" {
		settings.LanguageCode = "ru-RU"
	}

	assetPath, err := s.repo.FileAssetPath(ctx, fileID, userID)
	if err != nil {
		fail(err)
		return
	}

	workDir := filepath.Join(s.assetsDir, "work", fileID)
	_ = os.RemoveAll(workDir)
	if err := os.MkdirAll(workDir, 0o750); err != nil {
		fail(err)
		return
	}
	defer os.RemoveAll(workDir)

	normalized := filepath.Join(workDir, "normalized.ogg")
	if err := ffmpegNormalizeAudio(ctx, assetPath, normalized); err != nil {
		fail(fmt.Errorf("audio normalize: %w", err))
		return
	}

	duration, err := ffprobeDuration(ctx, normalized)
	if err != nil {
		s.logger.Warn("audio transcription duration probe failed", "file_id", fileID, "error", err)
	}

	normalizedData, err := os.ReadFile(normalized)
	if err != nil {
		fail(err)
		return
	}

	params, err := s.resolveSpeechKitParams(ctx, settings)
	if err != nil {
		fail(err)
		return
	}
	var chunks [][]byte
	if int64(len(normalizedData)) <= AudioTranscriptionAsyncMaxBytes {
		chunks = [][]byte{normalizedData}
	} else {
		segmentDir := filepath.Join(workDir, "segments")
		if err := ffmpegSegmentAudio(ctx, normalized, segmentDir, AudioTranscriptionSegmentSec); err != nil {
			fail(fmt.Errorf("audio split: %w", err))
			return
		}
		entries, err := os.ReadDir(segmentDir)
		if err != nil {
			fail(err)
			return
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".ogg") {
				continue
			}
			part, err := os.ReadFile(filepath.Join(segmentDir, entry.Name()))
			if err != nil {
				fail(err)
				return
			}
			if int64(len(part)) > AudioTranscriptionAsyncMaxBytes {
				subParts, err := s.splitWithSyncChunks(ctx, part, workDir)
				if err != nil {
					fail(err)
					return
				}
				chunks = append(chunks, subParts...)
				continue
			}
			chunks = append(chunks, part)
		}
		if len(chunks) == 0 {
			fail(errors.New("no audio segments produced"))
			return
		}
	}

	var transcriptParts []string
	for i, chunk := range chunks {
		text, err := s.transcribeChunk(ctx, params, chunk, "audio/ogg")
		if err != nil {
			fail(fmt.Errorf("chunk %d: %w", i+1, err))
			return
		}
		transcriptParts = append(transcriptParts, strings.TrimSpace(text))
	}

	fullText := strings.TrimSpace(strings.Join(transcriptParts, "\n\n"))
	if fullText == "" {
		fail(ErrYandexSpeechKitEmptyResult)
		return
	}

	transcriptPath := filepath.Join(s.assetsDir, "transcripts", fileID+".txt")
	if err := os.WriteFile(transcriptPath, []byte(fullText), 0o640); err != nil {
		fail(err)
		return
	}
	if err := s.repo.SetTranscriptPath(ctx, fileID, userID, transcriptPath, duration); err != nil {
		fail(err)
		return
	}

	preview := fullText
	if len(preview) > 500 {
		preview = preview[:500] + "…"
	}
	durVal := 0.0
	if duration != nil {
		durVal = *duration
	}
	output, _ := json.Marshal(model.AudioTranscriptionOutput{
		Text:        fullText,
		Language:    settings.LanguageCode,
		Model:       settings.Model,
		DurationSec: durVal,
		ChunkCount:  len(chunks),
		Preview:     preview,
	})

	artifactURL := fmt.Sprintf("/api/v1/tools/audio-transcription/files/%s/download", fileID)
	modelUsed := settings.Model
	if err := s.runs.UpdateRunDoneWithArtifact(ctx, runID, output, 0, modelUsed, artifactURL); err != nil {
		fail(err)
		return
	}

	costRUB := 0.0
	if settings.PriceRUBPerMinute > 0 && durVal > 0 {
		costRUB = (durVal / 60) * settings.PriceRUBPerMinute
	}
	_ = s.usageLog.Create(ctx, userID, runID, "yandex-speechkit", settings.Model, 0, 0, 0, costRUB)
}

func (s *AudioTranscriptionService) transcribeChunk(ctx context.Context, params YandexSpeechKitParams, chunk []byte, contentType string) (string, error) {
	if int64(len(chunk)) <= AudioTranscriptionSyncMaxBytes {
		text, syncErr := s.llm.TranscribeYandexSpeechKitSync(ctx, params, chunk, "oggopus")
		if syncErr == nil {
			return text, nil
		}
	}
	return s.llm.TranscribeYandexSpeechKitAsync(ctx, params, chunk, contentType)
}

func (s *AudioTranscriptionService) splitWithSyncChunks(ctx context.Context, part []byte, workDir string) ([][]byte, error) {
	partPath := filepath.Join(workDir, "oversized.ogg")
	if err := os.WriteFile(partPath, part, 0o640); err != nil {
		return nil, err
	}
	syncDir := filepath.Join(workDir, "sync_segments")
	if err := ffmpegSegmentAudio(ctx, partPath, syncDir, AudioTranscriptionSyncMaxSec); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(syncDir)
	if err != nil {
		return nil, err
	}
	var out [][]byte
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".ogg") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(syncDir, entry.Name()))
		if err != nil {
			return nil, err
		}
		out = append(out, data)
	}
	return out, nil
}

func (s *AudioTranscriptionService) ListFiles(ctx context.Context, userID, role string) ([]model.AudioTranscriptionFile, error) {
	if err := s.CheckAccess(ctx, userID, role); err != nil {
		return nil, err
	}
	return s.repo.ListFiles(ctx, userID)
}

func (s *AudioTranscriptionService) GetFile(ctx context.Context, fileID, userID, role string) (*model.AudioTranscriptionFileDetail, error) {
	if err := s.CheckAccess(ctx, userID, role); err != nil {
		return nil, err
	}
	file, err := s.repo.GetFile(ctx, fileID, userID)
	if err != nil {
		return nil, err
	}
	runs, err := s.repo.ListFileRuns(ctx, fileID, userID)
	if err != nil {
		return nil, err
	}
	return &model.AudioTranscriptionFileDetail{AudioTranscriptionFile: *file, Runs: runs}, nil
}

func (s *AudioTranscriptionService) TranscriptDownload(ctx context.Context, fileID, userID, role string) (path, downloadName string, err error) {
	if err := s.CheckAccess(ctx, userID, role); err != nil {
		return "", "", err
	}
	return s.repo.TranscriptPath(ctx, fileID, userID)
}

func (s *AudioTranscriptionService) DeleteFile(ctx context.Context, fileID, userID, role string) error {
	if err := s.CheckAccess(ctx, userID, role); err != nil {
		return err
	}
	assetPath, transcriptPath, err := s.repo.DeleteFile(ctx, fileID, userID)
	if err != nil {
		return err
	}
	_ = os.Remove(assetPath)
	if transcriptPath != nil {
		_ = os.Remove(*transcriptPath)
	}
	_ = os.RemoveAll(filepath.Join(s.assetsDir, "work", fileID))
	return nil
}

func (s *AudioTranscriptionService) ListAccessUsers(ctx context.Context) ([]model.AudioTranscriptionAccessUser, error) {
	return s.repo.ListAccessUsers(ctx)
}

func (s *AudioTranscriptionService) SetAccess(ctx context.Context, userID string, enabled bool) error {
	return s.repo.SetAccess(ctx, userID, enabled)
}

func (s *AudioTranscriptionService) GetSettings(ctx context.Context) (*model.AudioTranscriptionSettingsRecord, error) {
	return s.repo.GetSettings(ctx)
}

func (s *AudioTranscriptionService) UpdateSettings(ctx context.Context, settings model.AudioTranscriptionSettings) (*model.AudioTranscriptionSettingsRecord, error) {
	settings.Model = normalizeSpeechKitModel(settings.Model)
	settings.LanguageCode = normalizeSpeechKitLanguage(settings.LanguageCode)
	if settings.PriceRUBPerMinute < 0 {
		return nil, ErrAudioTranscriptionSettings
	}
	return s.repo.UpdateSettings(ctx, settings)
}

func randomAudioAssetName(ext string) string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	if ext == "" {
		ext = ".bin"
	}
	return hex.EncodeToString(buf) + ext
}

func audioExtensionFromContentType(contentType, originalName string) string {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "audio/mpeg", "audio/mp3":
		return ".mp3"
	case "audio/wav", "audio/x-wav", "audio/wave":
		return ".wav"
	case "audio/ogg", "audio/opus":
		return ".ogg"
	case "audio/webm", "video/webm":
		return ".webm"
	case "audio/mp4", "audio/x-m4a", "video/mp4":
		return ".m4a"
	default:
		ext := strings.ToLower(filepath.Ext(originalName))
		if ext != "" {
			return ext
		}
		return ".bin"
	}
}

func ffmpegNormalizeAudio(ctx context.Context, input, output string) error {
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-y", "-i", input,
		"-ac", "1", "-ar", "16000",
		"-c:a", "libopus", "-b:a", "32k",
		output,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func ffmpegSegmentAudio(ctx context.Context, input, outputPattern string, segmentSec float64) error {
	pattern := filepath.Join(outputPattern, "chunk_%03d.ogg")
	if err := os.MkdirAll(outputPattern, 0o750); err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-y", "-i", input,
		"-f", "segment",
		"-segment_time", fmt.Sprintf("%.0f", segmentSec),
		"-c", "copy",
		pattern,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func ffprobeDuration(ctx context.Context, input string) (*float64, error) {
	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		input,
	)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	val, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil || val <= 0 {
		return nil, fmt.Errorf("invalid duration")
	}
	return &val, nil
}
