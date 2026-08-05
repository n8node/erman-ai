package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/erman-ai/erman-ai/internal/config"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
)

const VideoTranscriptionMaxUploadBytes int64 = 500 << 20 // 500 MiB

var (
	ErrVideoTranscriptionForbidden         = errors.New("video transcription access denied")
	ErrVideoTranscriptionInvalidVideo      = errors.New("invalid video file")
	ErrVideoTranscriptionSettings          = errors.New("invalid video transcription settings")
	ErrVideoTranscriptionAlreadyProcessing = errors.New("video transcription already in progress")
	ErrVideoTranscriptionNoAudio           = errors.New("video has no audio track")
)

func VideoTranscriptionHasAccess(role string, explicitlyEnabled bool) bool {
	return role == "superadmin" || explicitlyEnabled
}

type VideoTranscriptionService struct {
	cfg       *config.Config
	assetsDir string
	repo      *repository.VideoTranscriptionRepository
	runs      *repository.ToolRunRepository
	plans     *repository.PlanRepository
	billing   *BillingService
	llm       *LLMService
	strategy  *StrategyLLMSettingsService
	usageLog  *repository.UsageLogRepository
	logger    *slog.Logger
}

func NewVideoTranscriptionService(
	cfg *config.Config,
	repo *repository.VideoTranscriptionRepository,
	runs *repository.ToolRunRepository,
	plans *repository.PlanRepository,
	billing *BillingService,
	llm *LLMService,
	strategy *StrategyLLMSettingsService,
	usageLog *repository.UsageLogRepository,
	logger *slog.Logger,
) *VideoTranscriptionService {
	return &VideoTranscriptionService{
		cfg:       cfg,
		assetsDir: cfg.VideoTranscriptionAssetsDir,
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

func (s *VideoTranscriptionService) EnsureAssetDirs() error {
	for _, sub := range []string{"uploads", "work", "transcripts"} {
		if err := os.MkdirAll(filepath.Join(s.assetsDir, sub), 0o750); err != nil {
			return err
		}
	}
	return nil
}

func (s *VideoTranscriptionService) CheckAccess(ctx context.Context, userID, role string) error {
	explicit, err := s.repo.HasExplicitAccess(ctx, userID)
	if err != nil {
		return err
	}
	if !VideoTranscriptionHasAccess(role, explicit) {
		return ErrVideoTranscriptionForbidden
	}
	return nil
}

func ValidateVideoTranscriptionContentType(contentType string) error {
	ct := strings.ToLower(strings.TrimSpace(contentType))
	allowed := []string{
		"video/mp4", "video/webm", "video/quicktime", "video/x-msvideo",
		"video/x-matroska", "video/mpeg", "video/ogg",
	}
	for _, a := range allowed {
		if ct == a || strings.HasPrefix(ct, a+";") {
			return nil
		}
	}
	return ErrVideoTranscriptionInvalidVideo
}

func ValidateVideoTranscriptionFile(data []byte, contentType string) error {
	if len(data) == 0 || int64(len(data)) > VideoTranscriptionMaxUploadBytes {
		return ErrVideoTranscriptionInvalidVideo
	}
	return ValidateVideoTranscriptionContentType(contentType)
}

func (s *VideoTranscriptionService) resolveSpeechKitParams(ctx context.Context, settings model.VideoTranscriptionSettings) (YandexSpeechKitParams, error) {
	strategyRec, err := s.strategy.GetStored(ctx)
	if err != nil {
		return YandexSpeechKitParams{}, err
	}
	creds := s.llm.CredentialsFromStored(strategyRec.Config)
	settings = model.ApplyVideoTranscriptionDefaults(settings)
	params := YandexSpeechKitParams{
		APIKey:                   creds.YandexKey,
		FolderID:                 creds.YandexFolderID,
		LanguageCode:             settings.LanguageCode,
		Model:                    settings.Model,
		TextNormalizationEnabled: settings.TextNormalizationEnabled,
		LiteratureText:           settings.LiteratureText,
		ProfanityFilter:          settings.ProfanityFilter,
	}
	if strings.TrimSpace(params.APIKey) == "" || strings.TrimSpace(params.FolderID) == "" {
		return params, fmt.Errorf("%w: set Yandex API key and folder ID in Admin → AI Strategy LLM or server .env", ErrYandexSpeechKitNotConfigured)
	}
	return params, nil
}

func (s *VideoTranscriptionService) UploadAndTranscribe(ctx context.Context, userID, role, originalName, contentType string, src io.Reader) (*model.VideoTranscriptionFile, *model.ToolRun, error) {
	if err := s.CheckAccess(ctx, userID, role); err != nil {
		return nil, nil, err
	}
	if err := ValidateVideoTranscriptionContentType(contentType); err != nil {
		return nil, nil, err
	}
	if err := s.billing.CheckToolLimit(ctx, userID, model.VideoTranscriptionToolSlug); err != nil {
		return nil, nil, err
	}
	if err := s.EnsureAssetDirs(); err != nil {
		return nil, nil, err
	}

	ext := videoExtensionFromContentType(contentType, originalName)
	path := filepath.Join(s.assetsDir, "uploads", randomMediaAssetName(ext))
	out, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o640)
	if err != nil {
		return nil, nil, err
	}
	limited := io.LimitReader(src, VideoTranscriptionMaxUploadBytes+1)
	written, copyErr := io.Copy(out, limited)
	closeErr := out.Close()
	if copyErr != nil {
		_ = os.Remove(path)
		return nil, nil, copyErr
	}
	if closeErr != nil {
		_ = os.Remove(path)
		return nil, nil, closeErr
	}
	if written == 0 || written > VideoTranscriptionMaxUploadBytes {
		_ = os.Remove(path)
		return nil, nil, ErrVideoTranscriptionInvalidVideo
	}

	file, err := s.repo.CreateFile(ctx, userID, filepath.Base(originalName), path, contentType, written, nil)
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

func (s *VideoTranscriptionService) RetranscribeFile(ctx context.Context, fileID, userID, role string) (*model.ToolRun, error) {
	if err := s.CheckAccess(ctx, userID, role); err != nil {
		return nil, err
	}
	if err := s.billing.CheckToolLimit(ctx, userID, model.VideoTranscriptionToolSlug); err != nil {
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
			return nil, ErrVideoTranscriptionAlreadyProcessing
		}
	}
	if path, _, err := s.repo.TranscriptPath(ctx, fileID, userID); err == nil {
		_ = os.Remove(path)
	}
	_ = s.repo.ClearTranscript(ctx, fileID, userID)
	return s.startTranscription(ctx, fileID, userID, role)
}

func (s *VideoTranscriptionService) startTranscription(ctx context.Context, fileID, userID, role string) (*model.ToolRun, error) {
	if err := s.CheckAccess(ctx, userID, role); err != nil {
		return nil, err
	}
	up, err := s.plans.GetUserPlan(ctx, userID)
	if err != nil {
		return nil, err
	}
	input, _ := json.Marshal(model.VideoTranscriptionRunInput{FileID: fileID})
	run, err := s.runs.CreatePending(ctx, userID, model.VideoTranscriptionToolSlug, up.PlanSlug, input)
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

func (s *VideoTranscriptionService) processRun(runID, fileID, userID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Minute)
	defer cancel()

	fail := func(err error) {
		s.logger.Error("video transcription run failed", "run_id", runID, "file_id", fileID, "error", err)
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
	settings := model.ApplyVideoTranscriptionDefaults(settingsRec.Settings)
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
	if err := ensureReadableVideoAsset(assetPath); err != nil {
		fail(err)
		return
	}
	if err := s.EnsureAssetDirs(); err != nil {
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

	extractedPath := filepath.Join(workDir, "extracted.wav")
	if err := ffmpegExtractAudioFromVideo(ctx, assetPath, extractedPath); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "no audio track") {
			fail(ErrVideoTranscriptionNoAudio)
			return
		}
		fail(fmt.Errorf("audio extract: %w", err))
		return
	}

	params, err := s.resolveSpeechKitParams(ctx, settings)
	if err != nil {
		fail(err)
		return
	}

	result, err := transcribeSpeechKitAsset(ctx, s.llm, extractedPath, workDir, speechKitTranscriptionSettings{
		Model:                    settings.Model,
		LanguageCode:             settings.LanguageCode,
		TextNormalizationEnabled: settings.TextNormalizationEnabled,
		LiteratureText:           settings.LiteratureText,
		ProfanityFilter:          settings.ProfanityFilter,
	}, params)
	if err != nil {
		fail(err)
		return
	}
	fullText := result.Text
	duration := result.DurationSec

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
	output, _ := json.Marshal(model.VideoTranscriptionOutput{
		Text:        fullText,
		Language:    settings.LanguageCode,
		Model:       settings.Model,
		DurationSec: durVal,
		ChunkCount:  result.ChunkCount,
		Preview:     preview,
	})

	artifactURL := fmt.Sprintf("/api/v1/tools/video-transcription/files/%s/download", fileID)
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

func (s *VideoTranscriptionService) ListFiles(ctx context.Context, userID, role string) ([]model.VideoTranscriptionFile, error) {
	if err := s.CheckAccess(ctx, userID, role); err != nil {
		return nil, err
	}
	return s.repo.ListFiles(ctx, userID)
}

func (s *VideoTranscriptionService) GetFile(ctx context.Context, fileID, userID, role string) (*model.VideoTranscriptionFileDetail, error) {
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
	return &model.VideoTranscriptionFileDetail{VideoTranscriptionFile: *file, Runs: runs}, nil
}

func (s *VideoTranscriptionService) TranscriptDownload(ctx context.Context, fileID, userID, role string) (path, downloadName string, err error) {
	if err := s.CheckAccess(ctx, userID, role); err != nil {
		return "", "", err
	}
	return s.repo.TranscriptPath(ctx, fileID, userID)
}

func (s *VideoTranscriptionService) DeleteFile(ctx context.Context, fileID, userID, role string) error {
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

func (s *VideoTranscriptionService) ListAccessUsers(ctx context.Context) ([]model.VideoTranscriptionAccessUser, error) {
	return s.repo.ListAccessUsers(ctx)
}

func (s *VideoTranscriptionService) SetAccess(ctx context.Context, userID string, enabled bool) error {
	return s.repo.SetAccess(ctx, userID, enabled)
}

func (s *VideoTranscriptionService) GetSettings(ctx context.Context) (*model.VideoTranscriptionSettingsRecord, error) {
	rec, err := s.repo.GetSettings(ctx)
	if err != nil {
		return nil, err
	}
	rec.Settings = model.ApplyVideoTranscriptionDefaults(rec.Settings)
	return rec, nil
}

func (s *VideoTranscriptionService) UpdateSettings(ctx context.Context, settings model.VideoTranscriptionSettings) (*model.VideoTranscriptionSettingsRecord, error) {
	settings = model.ApplyVideoTranscriptionDefaults(settings)
	settings.Model = normalizeSpeechKitModel(settings.Model)
	settings.LanguageCode = normalizeSpeechKitLanguage(settings.LanguageCode)
	if settings.LiteratureText && !settings.TextNormalizationEnabled {
		settings.TextNormalizationEnabled = true
	}
	if settings.PriceRUBPerMinute < 0 {
		return nil, ErrVideoTranscriptionSettings
	}
	rec, err := s.repo.UpdateSettings(ctx, settings)
	if err != nil {
		return nil, err
	}
	rec.Settings = model.ApplyVideoTranscriptionDefaults(rec.Settings)
	return rec, nil
}

func videoExtensionFromContentType(contentType, originalName string) string {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "video/mp4":
		return ".mp4"
	case "video/webm":
		return ".webm"
	case "video/quicktime":
		return ".mov"
	case "video/x-msvideo":
		return ".avi"
	case "video/x-matroska":
		return ".mkv"
	case "video/mpeg":
		return ".mpeg"
	case "video/ogg":
		return ".ogv"
	default:
		ext := strings.ToLower(filepath.Ext(originalName))
		if ext != "" {
			return ext
		}
		return ".bin"
	}
}
