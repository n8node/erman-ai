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

	"github.com/erman-ai/erman-ai/internal/model"
)

const GeologicalJournalMaxPDFBytes int64 = 250 << 20

var ErrGeologicalJournalInvalidPDF = errors.New("invalid geological journal pdf")

type GeologicalJournalDocumentService struct {
	assetsDir string
	journal   *GeologicalJournalService
	logger    *slog.Logger
}

func NewGeologicalJournalDocumentService(assetsDir string, journal *GeologicalJournalService, logger *slog.Logger) *GeologicalJournalDocumentService {
	return &GeologicalJournalDocumentService{assetsDir: assetsDir, journal: journal, logger: logger}
}

func (s *GeologicalJournalDocumentService) CheckAccess(ctx context.Context, documentID, userID, role string) error {
	return s.journal.CheckDocumentAccess(ctx, documentID, userID, role)
}

func (s *GeologicalJournalDocumentService) Create(ctx context.Context, userID, role, name string, body io.Reader, size int64) (*model.GeologicalJournalDocument, error) {
	if err := s.journal.CheckAccess(ctx, userID, role); err != nil {
		return nil, err
	}
	if size <= 0 || size > GeologicalJournalMaxPDFBytes {
		return nil, ErrGeologicalJournalInvalidPDF
	}
	if err := s.journal.EnsureAssetDirs(); err != nil {
		return nil, err
	}
	dir := filepath.Join(s.assetsDir, "documents")
	// The PDF is read by the non-root journal-preprocessor container through the
	// shared volume. The directory/file must be traversable/readable there.
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	fileName := randomAssetName(".pdf")
	path := filepath.Join(dir, fileName)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o644)
	if err != nil {
		return nil, err
	}
	_, copyErr := io.CopyN(file, body, size)
	closeErr := file.Close()
	if copyErr != nil && !errors.Is(copyErr, io.EOF) {
		_ = os.Remove(path)
		return nil, copyErr
	}
	if closeErr != nil {
		_ = os.Remove(path)
		return nil, closeErr
	}
	doc, err := s.journal.repo.CreateDocument(ctx, userID, filepath.Base(name), path, "application/pdf", size)
	if err != nil {
		_ = os.Remove(path)
		return nil, err
	}
	return doc, nil
}

func (s *GeologicalJournalDocumentService) List(ctx context.Context, userID, role string) ([]model.GeologicalJournalDocument, error) {
	if err := s.journal.CheckAccess(ctx, userID, role); err != nil {
		return nil, err
	}
	return s.journal.repo.ListDocuments(ctx, userID, true)
}

func (s *GeologicalJournalDocumentService) Get(ctx context.Context, id, userID, role string) (*model.GeologicalJournalDocumentDetail, error) {
	if err := s.CheckAccess(ctx, id, userID, role); err != nil {
		return nil, err
	}
	doc, err := s.journal.repo.GetDocument(ctx, id)
	if err != nil {
		return nil, err
	}
	pages, err := s.journal.repo.ListDocumentPages(ctx, id)
	if err != nil {
		return nil, err
	}
	return &model.GeologicalJournalDocumentDetail{Document: *doc, Pages: pages}, nil
}

func (s *GeologicalJournalDocumentService) SetShared(ctx context.Context, id, userID, role string, shared bool) error {
	doc, err := s.journal.repo.GetDocument(ctx, id)
	if err != nil {
		return err
	}
	if role != "superadmin" && doc.OwnerID != userID {
		return ErrGeologicalJournalForbidden
	}
	return s.journal.repo.SetDocumentShared(ctx, id, shared)
}

func (s *GeologicalJournalDocumentService) Delete(ctx context.Context, id, userID, role string) error {
	if err := s.CheckAccess(ctx, id, userID, role); err != nil {
		return err
	}
	path, err := s.journal.repo.DeleteDocument(ctx, id, userID, role == "superadmin")
	if err != nil {
		return err
	}
	if path != "" {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			s.logger.Warn("geological journal document asset delete failed", "path", path, "error", err)
		}
	}
	documentDir := filepath.Join(s.assetsDir, "documents", id)
	if err := os.RemoveAll(documentDir); err != nil {
		s.logger.Warn("geological journal document artifacts delete failed", "path", documentDir, "error", err)
	}
	return nil
}

func (s *GeologicalJournalDocumentService) PageAsset(ctx context.Context, documentID, pageID, kind string) (string, error) {
	return s.journal.repo.GetDocumentPageAssetForDocument(ctx, documentID, pageID, kind)
}

func (s *GeologicalJournalDocumentService) SummarizeSelectedPages(ctx context.Context, documentID, userID, role string, req model.GeologicalJournalDocumentLLMRequest) ([]model.GeologicalJournalDocumentLLMResult, error) {
	if err := s.CheckAccess(ctx, documentID, userID, role); err != nil {
		return nil, err
	}
	if len(req.PageNumbers) == 0 {
		return nil, fmt.Errorf("at least one page must be selected")
	}
	doc, err := s.journal.repo.GetDocument(ctx, documentID)
	if err != nil {
		return nil, err
	}
	pages, err := s.journal.repo.ListDocumentPages(ctx, documentID)
	if err != nil {
		return nil, err
	}
	selected := make(map[int]model.GeologicalJournalDocumentPage)
	for _, page := range pages {
		selected[page.PageNumber] = page
	}
	strategy, err := s.journal.strategy.GetStored(ctx)
	if err != nil {
		return nil, err
	}
	creds := s.journal.llm.CredentialsFromStored(strategy.Config)
	provider := model.LLMProviderYandex
	apiKey := s.journal.llm.ResolveKey(provider, creds)
	if apiKey == "" {
		return nil, errors.New("yandex api key not configured")
	}
	var results []model.GeologicalJournalDocumentLLMResult
	for _, pageNumber := range req.PageNumbers {
		page, ok := selected[pageNumber]
		if !ok {
			continue
		}
		prompt := fmt.Sprintf("Проанализируй страницу %d документа %q. Определи ключевые факты, числа, геологические объекты и сомнительные места. Ответь на русском JSON-объектом с ключами summary, facts, uncertainties, sources. OCR:\n%s", pageNumber, doc.OriginalName, page.OCRText)
		completion, err := s.journal.llm.Complete(ctx, LLMCompletionRequest{
			Provider: provider, Model: strategy.Config.YandexModel, SystemPrompt: "Ты аккуратный аналитик геологических документов. Не выдумывай факты. Всегда указывай номер страницы и uncertainty.",
			UserPrompt: prompt, Temperature: 0.1, MaxTokens: 4096, APIKey: apiKey, FolderID: creds.YandexFolderID,
			Proxy: strategy.Config.ProxyForProvider(provider),
		})
		if err != nil {
			return nil, err
		}
		result, err := s.journal.repo.SaveDocumentLLMResult(ctx, documentID, page.ID, userID, req.Mode, json.RawMessage(completion.Content), completion.Model)
		if err != nil {
			return nil, err
		}
		results = append(results, *result)
	}
	return results, nil
}

func (s *GeologicalJournalDocumentService) Chat(ctx context.Context, documentID, userID, role string, req model.GeologicalJournalDocumentChatRequest) (*model.GeologicalJournalDocumentChatMessage, error) {
	if err := s.CheckAccess(ctx, documentID, userID, role); err != nil {
		return nil, err
	}
	if len(req.PageNumbers) == 0 || len(req.Message) < 2 {
		return nil, fmt.Errorf("select pages and enter a question")
	}
	doc, err := s.journal.repo.GetDocument(ctx, documentID)
	if err != nil {
		return nil, err
	}
	pages, err := s.journal.repo.ListDocumentPages(ctx, documentID)
	if err != nil {
		return nil, err
	}
	selected := make(map[int]model.GeologicalJournalDocumentPage)
	for _, page := range pages {
		selected[page.PageNumber] = page
	}
	var contextText strings.Builder
	for _, number := range req.PageNumbers {
		if page, ok := selected[number]; ok {
			contextText.WriteString(fmt.Sprintf("\n[Страница %d]\n%s\n", number, page.OCRText))
		}
	}
	if req.IncludeNeighbors {
		for _, number := range req.PageNumbers {
			if page, ok := selected[number-1]; ok {
				contextText.WriteString(fmt.Sprintf("\n[Соседняя страница %d]\n%s\n", number-1, page.OCRText))
			}
			if page, ok := selected[number+1]; ok {
				contextText.WriteString(fmt.Sprintf("\n[Соседняя страница %d]\n%s\n", number+1, page.OCRText))
			}
		}
	}
	strategy, err := s.journal.strategy.GetStored(ctx)
	if err != nil {
		return nil, err
	}
	creds := s.journal.llm.CredentialsFromStored(strategy.Config)
	provider := model.LLMProviderYandex
	apiKey := s.journal.llm.ResolveKey(provider, creds)
	if apiKey == "" {
		return nil, errors.New("yandex api key not configured")
	}
	completion, err := s.journal.llm.Complete(ctx, LLMCompletionRequest{Provider: provider, Model: strategy.Config.YandexModel, SystemPrompt: "Отвечай только по переданным страницам. Не выдумывай. В ответе укажи номера страниц и уровень уверенности.", UserPrompt: "Документ: " + doc.OriginalName + "\nВопрос: " + req.Message + "\nКонтекст:" + contextText.String(), Temperature: 0.1, MaxTokens: 4096, APIKey: apiKey, FolderID: creds.YandexFolderID, Proxy: strategy.Config.ProxyForProvider(provider)})
	if err != nil {
		return nil, err
	}
	if req.SessionID == "" {
		req.SessionID, err = s.journal.repo.CreateChatSession(ctx, documentID, userID, req.PageNumbers, req.IncludeNeighbors)
		if err != nil {
			return nil, err
		}
	}
	return s.journal.repo.AddChatMessage(ctx, req.SessionID, "assistant", completion.Content, json.RawMessage(fmt.Sprintf(`[{"pages":%v}]`, req.PageNumbers)), "medium")
}

func (s *GeologicalJournalDocumentService) StartAnalysis(ctx context.Context, id, userID, role string) error {
	if err := s.CheckAccess(ctx, id, userID, role); err != nil {
		return err
	}
	doc, err := s.journal.repo.GetDocument(ctx, id)
	if err != nil {
		return err
	}
	pdf, err := os.ReadFile(docAssetPath(doc, s.assetsDir))
	if err != nil {
		return err
	}
	info, err := s.journal.preprocessor.PDFInfoBytes(ctx, pdf)
	if err != nil {
		return err
	}
	if err := s.journal.repo.SetDocumentPageCount(ctx, id, info.PageCount); err != nil {
		return err
	}
	if err := s.journal.repo.MarkDocumentQueued(ctx, id); err != nil {
		return err
	}
	if err := s.journal.repo.ResetDocumentJobs(ctx, id); err != nil {
		return err
	}
	for pageNumber := 1; pageNumber <= info.PageCount; pageNumber++ {
		page, err := s.journal.repo.CreateDocumentPage(ctx, id, pageNumber)
		if err != nil {
			return err
		}
		if err := s.journal.repo.CreateDocumentJob(ctx, id, page.ID); err != nil {
			return err
		}
	}
	return nil
}

func (s *GeologicalJournalDocumentService) StartWorker(ctx context.Context, concurrency int) {
	if concurrency < 1 {
		concurrency = 1
	}
	for i := 0; i < concurrency; i++ {
		go s.workerLoop(ctx)
	}
}

func (s *GeologicalJournalDocumentService) workerLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		job, err := s.journal.repo.ClaimDocumentJob(ctx, 300)
		if err != nil {
			s.logger.Error("journal document job claim failed", "error", err)
			time.Sleep(time.Second)
			continue
		}
		if job == nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}
		if err := s.processDocumentPage(ctx, job); err != nil {
			message := err.Error()
			if len(message) > 500 {
				message = message[:500]
			}
			_ = s.journal.repo.FailDocumentJob(context.Background(), job.ID, job.PageID, message)
			s.logger.Error("journal document page failed", "job_id", job.ID, "error", err)
			continue
		}
		_ = s.journal.repo.CompleteDocumentJob(context.Background(), job.ID, job.PageID, "done", "done")
	}
}

func (s *GeologicalJournalDocumentService) processDocumentPage(ctx context.Context, job *model.GeologicalJournalDocumentJob) error {
	documentDir := filepath.Join(s.assetsDir, "documents", job.DocumentID)
	pageDir := filepath.Join(documentDir, fmt.Sprintf("page-%04d", job.PageNumber))
	// Existing documents may have been created by an older image with mode 0750.
	// The non-root preprocessor must be able to traverse the document directory.
	if err := os.Chmod(documentDir, 0o777); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	// The preprocessor writes page artifacts into the shared volume as uid 10001.
	if err := os.MkdirAll(pageDir, 0o777); err != nil {
		return err
	}
	if err := os.Chmod(pageDir, 0o777); err != nil {
		return err
	}
	pdf, err := os.ReadFile(job.DocumentPath)
	if err != nil {
		return err
	}
	result, err := s.journal.preprocessor.AnalyzePDFPageBytes(ctx, pdf, job.PageNumber, pageDir)
	if err != nil {
		return err
	}
	raw, _ := json.Marshal(result.Analysis)
	if err := s.journal.repo.UpdateDocumentPageAnalysis(ctx, job.PageID, result.Status, result.ContentType,
		result.OrientationDegrees, result.OrientationConfidence, result.TableCount, result.TextCharCount,
		result.OCRText, raw, nil); err != nil {
		return err
	}
	if err := s.journal.repo.UpdateDocumentPageAssets(ctx, job.PageID, result.OriginalAssetPath, result.OrientedAssetPath, result.PreprocessedAssetPath); err != nil {
		return err
	}
	return s.journal.repo.RefreshDocumentStatus(ctx, job.DocumentID)
}

func docAssetPath(doc *model.GeologicalJournalDocument, assetsDir string) string {
	if doc == nil {
		return ""
	}
	if doc.AssetPath != "" {
		return doc.AssetPath
	}
	return filepath.Join(assetsDir, "documents", doc.OriginalName)
}
