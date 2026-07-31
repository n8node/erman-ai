package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/prompts"
	"github.com/erman-ai/erman-ai/internal/repository"
	_ "golang.org/x/image/webp"
)

const GeologicalJournalMaxImageBytes int64 = 10 << 20

var (
	ErrGeologicalJournalForbidden    = errors.New("geological journal access denied")
	ErrGeologicalJournalInvalidImage = errors.New("invalid geological journal image")
	ErrGeologicalJournalSettings     = errors.New("invalid geological journal settings")
)

type ValidatedJournalImage struct {
	Bytes       []byte
	ContentType string
	Extension   string
	Width       int
	Height      int
}

func ValidateGeologicalJournalImage(data []byte) (*ValidatedJournalImage, error) {
	if len(data) == 0 || int64(len(data)) > GeologicalJournalMaxImageBytes {
		return nil, ErrGeologicalJournalInvalidImage
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || cfg.Width <= 0 || cfg.Height <= 0 || int64(cfg.Width)*int64(cfg.Height) > 100_000_000 {
		return nil, ErrGeologicalJournalInvalidImage
	}
	if _, _, err := image.Decode(bytes.NewReader(data)); err != nil {
		return nil, ErrGeologicalJournalInvalidImage
	}
	mimes := map[string]string{"jpeg": "image/jpeg", "png": "image/png", "gif": "image/gif", "webp": "image/webp"}
	exts := map[string]string{"jpeg": ".jpg", "png": ".png", "gif": ".gif", "webp": ".webp"}
	mime, ok := mimes[format]
	if !ok {
		return nil, ErrGeologicalJournalInvalidImage
	}
	return &ValidatedJournalImage{
		Bytes: data, ContentType: mime, Extension: exts[format], Width: cfg.Width, Height: cfg.Height,
	}, nil
}

func ParseGeologicalJournalOutput(content string) (*model.GeologicalJournalOutput, error) {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)
	var shape struct {
		Rows []map[string]json.RawMessage `json:"rows"`
	}
	if err := json.Unmarshal([]byte(content), &shape); err != nil {
		return nil, fmt.Errorf("invalid geological journal output: %w", err)
	}
	required := []string{
		"date", "drilling_diameter_mm", "depth_from_m", "depth_to_m", "drilling_run_m",
		"core_recovery_m", "core_recovery_pct", "rock_description", "sampling_interval",
		"sample_number", "notes", "uncertainties",
	}
	for i, row := range shape.Rows {
		for _, key := range required {
			if _, ok := row[key]; !ok {
				return nil, fmt.Errorf("invalid geological journal output: row %d missing %s", i, key)
			}
		}
	}
	var out model.GeologicalJournalOutput
	dec := json.NewDecoder(strings.NewReader(content))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&out); err != nil {
		return nil, fmt.Errorf("invalid geological journal output: %w", err)
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return nil, errors.New("invalid geological journal output: trailing data")
	}
	if out.Rows == nil {
		return nil, errors.New("invalid geological journal output: rows required")
	}
	for i := range out.Rows {
		if out.Rows[i].Uncertainties == nil {
			out.Rows[i].Uncertainties = []string{}
		}
	}
	return &out, nil
}

func GeologicalJournalHasAccess(role string, explicitlyEnabled bool) bool {
	return role == "superadmin" || explicitlyEnabled
}

type GeologicalJournalService struct {
	assetsDir string
	repo      *repository.GeologicalJournalRepository
	runs      *repository.ToolRunRepository
	plans     *repository.PlanRepository
	llm       *LLMService
	strategy  *StrategyLLMSettingsService
	usageLog  *repository.UsageLogRepository
	logger    *slog.Logger
}

func NewGeologicalJournalService(
	assetsDir string,
	repo *repository.GeologicalJournalRepository,
	runs *repository.ToolRunRepository,
	plans *repository.PlanRepository,
	llm *LLMService,
	strategy *StrategyLLMSettingsService,
	usageLog *repository.UsageLogRepository,
	logger *slog.Logger,
) *GeologicalJournalService {
	return &GeologicalJournalService{
		assetsDir: assetsDir, repo: repo, runs: runs, plans: plans, llm: llm,
		strategy: strategy, usageLog: usageLog, logger: logger,
	}
}

func (s *GeologicalJournalService) EnsureAssetDirs() error {
	for _, dir := range []string{filepath.Join(s.assetsDir, "pages"), filepath.Join(s.assetsDir, "examples")} {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return err
		}
	}
	return nil
}

func (s *GeologicalJournalService) CheckAccess(ctx context.Context, userID, role string) error {
	explicit, err := s.repo.HasExplicitAccess(ctx, userID)
	if err != nil {
		return err
	}
	if !GeologicalJournalHasAccess(role, explicit) {
		return ErrGeologicalJournalForbidden
	}
	return nil
}

func (s *GeologicalJournalService) CreatePage(ctx context.Context, userID, role, originalName string, image *ValidatedJournalImage) (*model.GeologicalJournalPage, *model.ToolRun, error) {
	if err := s.CheckAccess(ctx, userID, role); err != nil {
		return nil, nil, err
	}
	if image == nil {
		return nil, nil, ErrGeologicalJournalInvalidImage
	}
	if err := s.EnsureAssetDirs(); err != nil {
		return nil, nil, err
	}
	path := filepath.Join(s.assetsDir, "pages", randomAssetName(image.Extension))
	if err := os.WriteFile(path, image.Bytes, 0o640); err != nil {
		return nil, nil, err
	}
	page, err := s.repo.CreatePage(ctx, userID, filepath.Base(originalName), path, image.ContentType, int64(len(image.Bytes)), image.Width, image.Height)
	if err != nil {
		_ = os.Remove(path)
		return nil, nil, err
	}
	run, err := s.StartAnalysis(ctx, page.ID, userID, role)
	if err != nil {
		_, _ = s.repo.DeletePage(ctx, page.ID, userID)
		_ = os.Remove(path)
		return nil, nil, err
	}
	return page, run, nil
}

func (s *GeologicalJournalService) StartAnalysis(ctx context.Context, pageID, userID, role string) (*model.ToolRun, error) {
	if err := s.CheckAccess(ctx, userID, role); err != nil {
		return nil, err
	}
	page, err := s.repo.GetPage(ctx, pageID, userID)
	if err != nil {
		return nil, err
	}
	up, err := s.plans.GetUserPlan(ctx, userID)
	if err != nil {
		return nil, err
	}
	input, _ := json.Marshal(map[string]string{"page_id": page.ID})
	run, err := s.runs.CreatePending(ctx, userID, model.GeologicalJournalToolSlug, up.PlanSlug, input)
	if err != nil {
		return nil, err
	}
	if err := s.repo.AttachRun(ctx, page.ID, run.ID); err != nil {
		_ = s.runs.UpdateRunError(ctx, run.ID, "failed to attach page")
		return nil, err
	}
	go s.processRun(run.ID, page.ID, userID)
	return run, nil
}

func (s *GeologicalJournalService) processRun(runID, pageID, userID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
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
	page, err := s.repo.GetPage(ctx, pageID, userID)
	if err != nil {
		fail(err)
		return
	}
	path, err := s.repo.PageAssetPath(ctx, pageID, userID)
	if err != nil {
		fail(err)
		return
	}
	imageData, err := os.ReadFile(path)
	if err != nil {
		fail(err)
		return
	}
	settingsRec, err := s.repo.GetSettings(ctx)
	if err != nil {
		fail(err)
		return
	}
	settings := settingsRec.Settings
	strategyRec, err := s.strategy.GetStored(ctx)
	if err != nil {
		fail(err)
		return
	}
	creds := s.llm.CredentialsFromStored(strategyRec.Config)
	provider := settings.Provider
	apiKey := s.llm.ResolveKey(provider, creds)
	if apiKey == "" {
		fail(errors.New("llm api key not configured"))
		return
	}
	prompt := settings.SystemPrompt
	if strings.TrimSpace(prompt) == "" {
		prompt = prompts.DefaultGeologicalJournalSystemPrompt
	}
	result, err := s.llm.CompleteWithImage(ctx, LLMImageCompletionRequest{
		LLMCompletionRequest: LLMCompletionRequest{
			Provider: provider, Model: settings.ActiveModel(), SystemPrompt: prompt,
			UserPrompt:  "Recognize this geological journal page and return the required strict JSON.",
			Temperature: settings.Temperature, MaxTokens: settings.MaxTokens, APIKey: apiKey,
			FolderID: creds.YandexFolderID, Proxy: strategyRec.Config.ProxyForProvider(provider),
		},
		Image: imageData, ImageMIME: page.ContentType,
	})
	if err != nil {
		fail(err)
		return
	}
	costUSD, costRUB := s.strategy.UsageCosts(strategyRec.Config, provider, result.Model, result.PromptTokens, result.CompletionTokens)
	usageCtx, cancelUsage := context.WithTimeout(context.Background(), 5*time.Second)
	if err := s.usageLog.Create(usageCtx, userID, runID, string(provider), result.Model, result.PromptTokens, result.CompletionTokens, costUSD, costRUB); err != nil {
		s.logger.Error("geological journal usage log failed", "run_id", runID, "error", err)
	}
	cancelUsage()

	output, err := ParseGeologicalJournalOutput(result.Content)
	if err != nil {
		fail(err)
		return
	}
	outJSON, _ := json.Marshal(output)
	if _, err := s.repo.SaveResult(ctx, pageID, userID, outJSON); err != nil {
		fail(err)
		return
	}
	if err := s.runs.UpdateRunDone(ctx, runID, outJSON, int64(result.TotalTokens), result.Model); err != nil {
		fail(err)
	}
}

func (s *GeologicalJournalService) ListPages(ctx context.Context, userID, role string) ([]model.GeologicalJournalPage, error) {
	if err := s.CheckAccess(ctx, userID, role); err != nil {
		return nil, err
	}
	return s.repo.ListPages(ctx, userID)
}

func (s *GeologicalJournalService) GetPage(ctx context.Context, pageID, userID, role string) (*model.GeologicalJournalPageDetail, error) {
	if err := s.CheckAccess(ctx, userID, role); err != nil {
		return nil, err
	}
	page, err := s.repo.GetPage(ctx, pageID, userID)
	if err != nil {
		return nil, err
	}
	runs, err := s.repo.ListPageRuns(ctx, pageID, userID)
	if err != nil {
		return nil, err
	}
	versions, err := s.repo.ListVersions(ctx, pageID, userID)
	if err != nil {
		return nil, err
	}
	return &model.GeologicalJournalPageDetail{GeologicalJournalPage: *page, Runs: runs, Versions: versions}, nil
}

func (s *GeologicalJournalService) PageImage(ctx context.Context, pageID, userID, role string) (string, string, error) {
	if err := s.CheckAccess(ctx, userID, role); err != nil {
		return "", "", err
	}
	page, err := s.repo.GetPage(ctx, pageID, userID)
	if err != nil {
		return "", "", err
	}
	path, err := s.repo.PageAssetPath(ctx, pageID, userID)
	return path, page.ContentType, err
}

func (s *GeologicalJournalService) SaveCorrectedResult(ctx context.Context, pageID, userID, role string, result json.RawMessage) (*model.GeologicalJournalResultVersion, error) {
	if err := s.CheckAccess(ctx, userID, role); err != nil {
		return nil, err
	}
	parsed, err := ParseGeologicalJournalOutput(string(result))
	if err != nil {
		return nil, ErrInvalidInput
	}
	normalized, err := json.Marshal(parsed)
	if err != nil {
		return nil, err
	}
	return s.repo.SaveResult(ctx, pageID, userID, normalized)
}

func (s *GeologicalJournalService) DeletePage(ctx context.Context, pageID, userID, role string) error {
	if err := s.CheckAccess(ctx, userID, role); err != nil {
		return err
	}
	runs, err := s.repo.ListPageRuns(ctx, pageID, userID)
	if err != nil {
		return err
	}
	for _, run := range runs {
		if err := s.runs.DeleteForUser(ctx, run.ID, userID); err != nil {
			return err
		}
	}
	path, err := s.repo.DeletePage(ctx, pageID, userID)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		s.logger.Warn("geological journal asset delete failed", "path", path, "error", err)
	}
	return nil
}

func (s *GeologicalJournalService) GetSettings(ctx context.Context) (*model.GeologicalJournalSettingsRecord, error) {
	rec, err := s.repo.GetSettings(ctx)
	if err != nil {
		return nil, err
	}
	if rec.Settings.SystemPrompt == "" {
		rec.Settings.SystemPrompt = prompts.DefaultGeologicalJournalSystemPrompt
	}
	adminView, err := s.strategy.GetAdminView(ctx)
	if err != nil {
		return nil, err
	}
	rec.Providers = geologicalJournalProviderStatuses(adminView.Providers)
	return rec, nil
}

func (s *GeologicalJournalService) UpdateSettings(ctx context.Context, settings model.GeologicalJournalSettings) (*model.GeologicalJournalSettingsRecord, error) {
	if err := validateGeologicalJournalSettings(settings); err != nil {
		return nil, err
	}
	if _, err := s.repo.UpdateSettings(ctx, settings); err != nil {
		return nil, err
	}
	return s.GetSettings(ctx)
}

func (s *GeologicalJournalService) RefreshModels(ctx context.Context, provider model.LLMProvider) (*model.StrategyLLMTestConnectionResult, error) {
	if provider != model.LLMProviderOpenRouter && provider != model.LLMProviderYandex {
		return nil, fmt.Errorf("%w: provider must be openrouter or yandex", ErrGeologicalJournalSettings)
	}
	return s.strategy.TestConnection(ctx, provider)
}

func validateGeologicalJournalSettings(s model.GeologicalJournalSettings) error {
	if s.Provider != model.LLMProviderOpenRouter && s.Provider != model.LLMProviderYandex {
		return fmt.Errorf("%w: provider must be openrouter or yandex", ErrGeologicalJournalSettings)
	}
	if strings.TrimSpace(s.OpenRouterModel) == "" || strings.TrimSpace(s.YandexModel) == "" {
		return fmt.Errorf("%w: model names required", ErrGeologicalJournalSettings)
	}
	if s.Temperature < 0 || s.Temperature > 2 || s.MaxTokens < 256 || s.MaxTokens > 32000 {
		return ErrGeologicalJournalSettings
	}
	return nil
}

func (s *GeologicalJournalService) ListAccessUsers(ctx context.Context) ([]model.GeologicalJournalAccessUser, error) {
	return s.repo.ListAccessUsers(ctx)
}

func (s *GeologicalJournalService) SetAccess(ctx context.Context, userID string, enabled bool) error {
	return s.repo.SetAccess(ctx, userID, enabled)
}

func (s *GeologicalJournalService) ListExamples(ctx context.Context, userID, role string, admin bool) ([]model.GeologicalJournalExample, error) {
	if !admin {
		if err := s.CheckAccess(ctx, userID, role); err != nil {
			return nil, err
		}
	}
	return s.repo.ListExamples(ctx, !admin)
}

func (s *GeologicalJournalService) ExampleImage(ctx context.Context, id, userID, role string) (string, string, error) {
	if err := s.CheckAccess(ctx, userID, role); err != nil {
		return "", "", err
	}
	return s.repo.ExampleAssetPath(ctx, id, role != "superadmin")
}

func (s *GeologicalJournalService) CreateExample(ctx context.Context, meta model.GeologicalJournalExampleMetadata, image *ValidatedJournalImage) (*model.GeologicalJournalExample, error) {
	if image == nil || strings.TrimSpace(meta.Title) == "" {
		return nil, ErrInvalidInput
	}
	if err := s.EnsureAssetDirs(); err != nil {
		return nil, err
	}
	path := filepath.Join(s.assetsDir, "examples", randomAssetName(image.Extension))
	if err := os.WriteFile(path, image.Bytes, 0o640); err != nil {
		return nil, err
	}
	item, err := s.repo.CreateExample(ctx, strings.TrimSpace(meta.Title), strings.TrimSpace(meta.Description), path, image.ContentType, int64(len(image.Bytes)), meta.SortOrder, meta.IsPublished)
	if err != nil {
		_ = os.Remove(path)
	}
	return item, err
}

func (s *GeologicalJournalService) UpdateExample(ctx context.Context, id string, meta model.GeologicalJournalExampleMetadata) (*model.GeologicalJournalExample, error) {
	if strings.TrimSpace(meta.Title) == "" {
		return nil, ErrInvalidInput
	}
	return s.repo.UpdateExample(ctx, id, meta)
}

func (s *GeologicalJournalService) DeleteExample(ctx context.Context, id string) error {
	path, err := s.repo.DeleteExample(ctx, id)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		s.logger.Warn("geological journal example delete failed", "path", path, "error", err)
	}
	return nil
}

func randomAssetName(ext string) string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	}
	return hex.EncodeToString(b[:]) + ext
}

func geologicalJournalProviderStatuses(providers []model.LLMProviderStatus) []model.LLMProviderStatus {
	filtered := make([]model.LLMProviderStatus, 0, 2)
	for _, provider := range providers {
		if provider.ID == model.LLMProviderOpenRouter || provider.ID == model.LLMProviderYandex {
			filtered = append(filtered, provider)
		}
	}
	return filtered
}
