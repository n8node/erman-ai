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
	journalSettings, err := s.journal.repo.GetSettings(ctx)
	if err != nil {
		return nil, err
	}
	creds := s.journal.llm.CredentialsFromStored(strategy.Config)
	provider := model.LLMProviderYandex
	apiKey := s.journal.llm.ResolveKey(provider, creds)
	if apiKey == "" {
		return nil, errors.New("yandex api key not configured")
	}
	ocrModel := NormalizeYandexOCRModel(journalSettings.Settings.OCRModel)
	var results []model.GeologicalJournalDocumentLLMResult
	for _, pageNumber := range req.PageNumbers {
		page, ok := selected[pageNumber]
		if !ok {
			continue
		}
		imagePath := documentPageOrientedAssetPath(page)
		if imagePath == "" {
			return nil, fmt.Errorf("oriented image is not available for page %d", pageNumber)
		}
		image, err := os.ReadFile(imagePath)
		if err != nil {
			return nil, fmt.Errorf("read oriented image for page %d: %w", pageNumber, err)
		}
		ocrText := strings.TrimSpace(page.OCRText)
		visionText, visionErr := s.journal.llm.RecognizeYandexVisionText(ctx, YandexVisionRecognizeParams{
			APIKey: apiKey, FolderID: creds.YandexFolderID, Image: image,
			MIME: documentPageImageMIME(imagePath), Model: ocrModel,
			Proxy: strategy.Config.ProxyForProvider(provider),
		})
		if visionErr == nil && strings.TrimSpace(visionText) != "" {
			ocrText = strings.TrimSpace(visionText)
			if err := s.journal.repo.UpdateDocumentPageOCRText(ctx, page.ID, ocrText); err != nil {
				return nil, err
			}
		} else if ocrText == "" {
			if visionErr != nil {
				return nil, fmt.Errorf("vision ocr page %d: %w", pageNumber, visionErr)
			}
			return nil, fmt.Errorf("vision ocr page %d returned empty text", pageNumber)
		} else if visionErr != nil {
			s.logger.Warn("vision OCR failed; using existing document OCR text", "page", pageNumber, "error", visionErr)
		}
		isTable := page.ContentType == "table" || page.ContentType == "mixed"
		var prompt string
		if isTable {
			prompt = fmt.Sprintf(`Распознай таблицу непосредственно по изображению страницы %d документа %q.
Предварительный анализ определил страницу как таблицу. Изображение является главным
источником истины, OCR-текст ниже используй только как дополнительную подсказку.
Определи реальные заголовки столбцов по изображению, не подставляй заранее заданную
схему геологического журнала. Сохрани порядок столбцов и строк. Не добавляй выводы,
объяснения, summary, facts, uncertainties, sources или рассуждения.

Ответь только валидным JSON-объектом строго такого вида. Никогда не возвращай OCR-текст,
список строк, markdown или обычный текст вместо этого JSON:
{"columns":["№","Наименование"],"rows":[{"№":"8","Наименование":"..."}]}

В columns перечисли только реально найденные заголовки. Ключи каждой строки должны
совпадать с columns. Для нечитаемой ячейки используй null или пустую строку. Не
оборачивай JSON в markdown.

Дополнительный OCR-текст:
%s`, pageNumber, doc.OriginalName, ocrText)
		} else {
			prompt = fmt.Sprintf(`Точно перепиши распознанный текст страницы %d документа %q.
Не анализируй текст, не делай выводов, не исправляй значения и не добавляй объяснений.

Ответь только валидным JSON-объектом строго такого вида:
{"transcription":"полный распознанный текст страницы"}

Не оборачивай JSON в markdown.

Текст страницы:
%s`, pageNumber, doc.OriginalName, ocrText)
		}
		visionModel := strings.TrimSpace(journalSettings.Settings.VisionModel)
		if visionModel == "" {
			return nil, errors.New("vision model is not configured in geological journal settings")
		}
		completion, err := s.journal.llm.CompleteWithImage(ctx, LLMImageCompletionRequest{
			Image:     image,
			ImageMIME: documentPageImageMIME(imagePath),
			LLMCompletionRequest: LLMCompletionRequest{
				Provider: provider, Model: visionModel,
				SystemPrompt: "Ты выполняешь точное распознавание данных с изображения документа. Для табличной страницы всегда возвращай только JSON с columns и rows. Никогда не возвращай плоский OCR-текст, список строк, markdown или рассуждения.",
				UserPrompt:   prompt, Temperature: 0.1, MaxTokens: 4096, APIKey: apiKey, FolderID: creds.YandexFolderID,
				Proxy: strategy.Config.ProxyForProvider(provider),
			},
		})
		if err != nil {
			return nil, err
		}
		_, resultJSON := parseDocumentLLMResponse(completion.Content)
		if tableResult, ok := parseDocumentLLMTableResult(completion.Content); ok {
			if err := s.journal.repo.UpdateDocumentPageTableResult(ctx, page.ID, tableResult); err != nil {
				return nil, err
			}
		}
		result, err := s.journal.repo.SaveDocumentLLMResult(ctx, documentID, page.ID, userID, req.Mode, resultJSON, completion.Model)
		if err != nil {
			return nil, err
		}
		results = append(results, *result)
	}
	return results, nil
}

func parseDocumentLLMTableResult(content string) (*model.GeologicalJournalOutput, bool) {
	content = strings.TrimSpace(trimGeologicalJournalFence(content))
	for _, candidate := range append([]string{content}, extractGeologicalJournalJSONObjects(content)...) {
		var envelope struct {
			Rows json.RawMessage `json:"rows"`
		}
		if err := json.Unmarshal([]byte(candidate), &envelope); err != nil || len(envelope.Rows) == 0 {
			continue
		}
		wrapped := append([]byte(`{"rows":`), envelope.Rows...)
		wrapped = append(wrapped, '}')
		result, err := ParseGeologicalJournalOutput(string(wrapped))
		if err == nil {
			return result, true
		}
	}
	return nil, false
}

func (s *GeologicalJournalDocumentService) SavePageTableResult(ctx context.Context, documentID, pageID, userID, role string, raw json.RawMessage) (*model.GeologicalJournalOutput, error) {
	if err := s.CheckAccess(ctx, documentID, userID, role); err != nil {
		return nil, err
	}
	pages, err := s.journal.repo.ListDocumentPages(ctx, documentID)
	if err != nil {
		return nil, err
	}
	var page *model.GeologicalJournalDocumentPage
	for index := range pages {
		if pages[index].ID == pageID {
			page = &pages[index]
			break
		}
	}
	if page == nil {
		return nil, fmt.Errorf("page not found")
	}
	var payload struct {
		Rows    json.RawMessage `json:"rows"`
		OCRText string          `json:"ocr_text"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil || len(payload.Rows) == 0 {
		return nil, fmt.Errorf("invalid result")
	}
	wrappedRows := append([]byte(`{"rows":`), payload.Rows...)
	wrappedRows = append(wrappedRows, '}')
	result, err := ParseGeologicalJournalOutput(string(wrappedRows))
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(payload.OCRText) != "" {
		if err := s.journal.repo.UpdateDocumentPageOCRText(ctx, pageID, payload.OCRText); err != nil {
			return nil, err
		}
	}
	ValidateGeologicalJournalOutput(result)
	if err := s.journal.repo.UpdateDocumentPageTableResult(ctx, pageID, result); err != nil {
		return nil, err
	}
	return result, nil
}

func documentPageOrientedAssetPath(page model.GeologicalJournalDocumentPage) string {
	if page.OrientedAssetPath != "" {
		return page.OrientedAssetPath
	}
	// Preprocessed images are generated from oriented.png and keep the same
	// readable orientation. Do not fall back to the original, potentially
	// rotated PDF render.
	return page.PreprocessedAssetPath
}

func documentPageImageMIME(path string) string {
	if strings.EqualFold(filepath.Ext(path), ".jpg") || strings.EqualFold(filepath.Ext(path), ".jpeg") {
		return "image/jpeg"
	}
	return "image/png"
}

func parseDocumentLLMResponse(content string) (string, json.RawMessage) {
	content = strings.TrimSpace(content)
	for _, candidate := range append([]string{trimGeologicalJournalFence(content)}, extractGeologicalJournalJSONObjects(content)...) {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		var payload struct {
			Transcription string `json:"transcription"`
		}
		if err := json.Unmarshal([]byte(candidate), &payload); err != nil {
			continue
		}
		if strings.TrimSpace(payload.Transcription) != "" {
			return strings.TrimSpace(payload.Transcription), json.RawMessage(candidate)
		}
	}
	if json.Valid([]byte(content)) {
		return "", json.RawMessage(content)
	}
	raw, _ := json.Marshal(map[string]string{"raw_response": content})
	return "", raw
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

func (s *GeologicalJournalDocumentService) StartAnalysis(ctx context.Context, id, userID, role string) (err error) {
	defer func() {
		if err != nil {
			_ = s.journal.repo.MarkDocumentError(context.Background(), id, err.Error())
		}
		if recovered := recover(); recovered != nil {
			_ = s.journal.repo.MarkDocumentError(context.Background(), id, fmt.Sprint(recovered))
		}
	}()
	if err := s.CheckAccess(ctx, id, userID, role); err != nil {
		return err
	}
	doc, err := s.journal.repo.GetDocument(ctx, id)
	if err != nil {
		return err
	}
	pdfPath := docAssetPath(doc, s.assetsDir)
	info, err := s.journal.preprocessor.PDFInfo(ctx, pdfPath)
	if err != nil {
		return err
	}
	if err := s.journal.repo.SetDocumentPageCount(ctx, id, info.PageCount); err != nil {
		return err
	}
	if err := s.journal.repo.MarkDocumentQueued(ctx, id); err != nil {
		return err
	}
	if err := s.journal.repo.CancelDocumentJobs(ctx, id); err != nil {
		return err
	}
	if err := s.journal.repo.MarkDocumentProcessing(ctx, id); err != nil {
		return err
	}
	previewRoot := filepath.Join(s.assetsDir, "documents", id)
	if err := os.MkdirAll(previewRoot, 0o777); err != nil {
		return err
	}
	for pageNumber := 1; pageNumber <= info.PageCount; pageNumber++ {
		page, err := s.journal.repo.CreateDocumentPage(ctx, id, pageNumber)
		if err != nil {
			return err
		}
		pageDir := filepath.Join(previewRoot, fmt.Sprintf("page-%04d", pageNumber))
		if err := os.MkdirAll(pageDir, 0o777); err != nil {
			return err
		}
		preview, err := s.journal.preprocessor.PreviewPDFPage(ctx, pdfPath, pageNumber, pageDir)
		if err != nil {
			return err
		}
		raw, _ := json.Marshal(preview.Analysis)
		if err := s.journal.repo.UpdateDocumentPageAnalysis(ctx, page.ID, preview.Status, preview.ContentType, preview.OrientationDegrees, preview.OrientationConfidence, preview.TableCount, preview.TextCharCount, preview.OCRText, raw, nil); err != nil {
			return err
		}
		if err := s.journal.repo.UpdateDocumentPageAssets(ctx, page.ID, preview.OriginalAssetPath, preview.OrientedAssetPath, ""); err != nil {
			return err
		}
	}
	if err := s.journal.repo.MarkDocumentPreviewReady(ctx, id); err != nil {
		return err
	}
	return nil
}

func (s *GeologicalJournalDocumentService) StartWorker(ctx context.Context, concurrency int) {
	// Deep page OCR is intentionally explicit now. The document button only
	// performs a cheap preview and never enqueues jobs for every page.
	if err := s.journal.repo.CancelAllDocumentJobs(ctx); err != nil {
		s.logger.Error("failed to cancel legacy geological journal jobs", "error", err)
	}
}

func (s *GeologicalJournalDocumentService) AnalyzeSelectedPageLocally(ctx context.Context, documentID, pageID, userID, role string) error {
	if err := s.CheckAccess(ctx, documentID, userID, role); err != nil {
		return err
	}
	doc, err := s.journal.repo.GetDocument(ctx, documentID)
	if err != nil {
		return err
	}
	pages, err := s.journal.repo.ListDocumentPages(ctx, documentID)
	if err != nil {
		return err
	}
	var page *model.GeologicalJournalDocumentPage
	for index := range pages {
		if pages[index].ID == pageID {
			page = &pages[index]
			break
		}
	}
	if page == nil {
		return fmt.Errorf("page not found")
	}
	pageDir := filepath.Join(s.assetsDir, "documents", documentID, fmt.Sprintf("page-%04d", page.PageNumber))
	if err := os.MkdirAll(pageDir, 0o777); err != nil {
		return err
	}
	pdf, err := os.ReadFile(docAssetPath(doc, s.assetsDir))
	if err != nil {
		return err
	}
	result, err := s.journal.preprocessor.AnalyzePDFPageBytes(ctx, pdf, page.PageNumber, pageDir)
	if err != nil {
		return err
	}
	raw, _ := json.Marshal(result.Analysis)
	if err := s.journal.repo.UpdateDocumentPageAnalysis(ctx, page.ID, result.Status, result.ContentType, result.OrientationDegrees, result.OrientationConfidence, result.TableCount, result.TextCharCount, result.OCRText, raw, nil); err != nil {
		return err
	}
	return s.journal.repo.UpdateDocumentPageAssets(ctx, page.ID, result.OriginalAssetPath, result.OrientedAssetPath, result.PreprocessedAssetPath)
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
			// Avoid hammering a restarting preprocessor after an OOM or connection
			// failure. The job remains retryable on the next explicit analysis run.
			time.Sleep(2 * time.Second)
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
