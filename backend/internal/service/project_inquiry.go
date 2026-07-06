package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/erman-ai/erman-ai/internal/config"
	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/erman-ai/erman-ai/internal/repository"
)

var (
	ErrInquiryVerificationInvalid = errors.New("invalid or expired inquiry verification token")
	ErrInquiryInvalidTelegram     = errors.New("invalid telegram")
	ErrInquiryMissingFields       = errors.New("missing required fields")
	ErrProjectInquiryAlreadySent  = errors.New("project inquiry already submitted")
	telegramUsernamePattern       = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{4,31}$`)
	telegramPhonePattern          = regexp.MustCompile(`^\+?[0-9][0-9\s\-()]{6,18}$`)
	telegramHandleFallback        = regexp.MustCompile(`^[a-zA-Z0-9_]{3,32}$`)
)

const inquiryVerificationTTL = 24 * time.Hour

type ProjectInquiryInput struct {
	Name               string
	Email              string
	Telegram           string
	ProjectTitle       string
	ProjectDescription string
	CalculatorRunID    string
	CalculatorSnapshot json.RawMessage
	Locale             string
	Honeypot           string
}

type ProjectInquiryList struct {
	Items  []repository.AdminProjectInquiryRow `json:"items"`
	Total  int                                 `json:"total"`
	Limit  int                                 `json:"limit"`
	Offset int                                 `json:"offset"`
}

type ProjectInquiryService struct {
	inquiries *repository.ProjectInquiryRepository
	tokens    *repository.InquiryVerificationTokenRepository
	runs      *repository.ToolRunRepository
	users     *repository.UserRepository
	share     *ShareService
	mail      *MailService
	telegram  *TelegramService
	cfg       *config.Config
	keySalt   string
}

func NewProjectInquiryService(
	inquiries *repository.ProjectInquiryRepository,
	tokens *repository.InquiryVerificationTokenRepository,
	runs *repository.ToolRunRepository,
	users *repository.UserRepository,
	share *ShareService,
	mail *MailService,
	telegram *TelegramService,
	cfg *config.Config,
) *ProjectInquiryService {
	return &ProjectInquiryService{
		inquiries: inquiries,
		tokens:    tokens,
		runs:      runs,
		users:     users,
		share:     share,
		mail:      mail,
		telegram:  telegram,
		cfg:       cfg,
		keySalt:   cfg.APIKeySalt,
	}
}

func (s *ProjectInquiryService) CreatePublic(ctx context.Context, input ProjectInquiryInput, ipHash string) (*model.ProjectInquiry, error) {
	if err := validateProjectInquiryInput(input); err != nil {
		return nil, err
	}

	normalized := normalizeProjectInquiryInput(input)
	inq, err := s.inquiries.Create(
		ctx, nil, normalized.runID,
		normalized.name, normalized.email, normalized.telegram,
		normalized.projectTitle, normalized.projectDescription,
		string(model.ProjectInquiryStatusPendingEmail),
		normalized.locale, ipHash,
		normalized.calculatorSnapshot,
	)
	if err != nil {
		return nil, err
	}

	raw, hash, err := s.generateToken()
	if err != nil {
		return nil, err
	}
	if err := s.tokens.Upsert(ctx, inq.ID, hash, time.Now().Add(inquiryVerificationTTL)); err != nil {
		return nil, err
	}

	link := s.verificationURL(raw)
	subject, body := inquiryVerificationEmailContent(normalized.locale, link, normalized.name)
	if err := s.mail.Send(ctx, normalized.email, subject, body); err != nil {
		if s.cfg.Environment == "development" {
			slog.Info("inquiry verification link (smtp unavailable)", "email", normalized.email, "url", link, "err", err)
		} else {
			return nil, err
		}
	}

	return inq, nil
}

func (s *ProjectInquiryService) CreateAuthenticated(
	ctx context.Context,
	userID string,
	input ProjectInquiryInput,
) (*model.ProjectInquiry, error) {
	if err := validateProjectInquiryInput(input); err != nil {
		return nil, err
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !user.EmailVerified() {
		return nil, errors.New("email not verified")
	}

	normalized := normalizeProjectInquiryInput(input)
	normalized.email = strings.TrimSpace(user.Email)

	var runID *string
	if normalized.runID != nil {
		if err := s.validateRunOwnership(ctx, userID, *normalized.runID); err != nil {
			return nil, err
		}
		runID = normalized.runID
	}

	snapshot := normalized.calculatorSnapshot
	if runID != nil {
		snapshot = nil
	}

	uid := userID
	inq, err := s.inquiries.Create(
		ctx, &uid, runID,
		normalized.name, normalized.email, normalized.telegram,
		normalized.projectTitle, normalized.projectDescription,
		string(model.ProjectInquiryStatusNew),
		normalized.locale, "",
		snapshot,
	)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	inq.EmailVerifiedAt = &now

	s.attachCalculatorShare(ctx, inq.ID)
	s.notifyAdmin(ctx, inq.ID)
	s.sendReceivedConfirmation(ctx, normalized.locale, normalized.email, normalized.name, normalized.telegram)

	return inq, nil
}

func (s *ProjectInquiryService) Verify(ctx context.Context, rawToken string) (*model.ProjectInquiry, error) {
	hash := s.hashToken(rawToken)
	inquiryID, err := s.tokens.FindInquiryID(ctx, hash, time.Now())
	if err != nil {
		return nil, ErrInquiryVerificationInvalid
	}

	inq, err := s.inquiries.MarkVerified(ctx, inquiryID)
	if err != nil {
		return nil, ErrInquiryVerificationInvalid
	}

	_ = s.tokens.Delete(ctx, inquiryID)
	s.attachCalculatorShare(ctx, inq.ID)
	s.notifyAdmin(ctx, inq.ID)
	s.sendReceivedConfirmation(ctx, inq.Locale, inq.Email, inq.Name, inq.Telegram)

	return inq, nil
}

func (s *ProjectInquiryService) ListMine(ctx context.Context, userID string, limit, offset int) (*ProjectInquiryList, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	items, err := s.inquiries.ListByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []repository.AdminProjectInquiryRow{}
	}
	return &ProjectInquiryList{Items: items, Total: len(items), Limit: limit, Offset: offset}, nil
}

func (s *ProjectInquiryService) ListAdmin(ctx context.Context, limit, offset int) (*ProjectInquiryList, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	items, err := s.inquiries.ListAdmin(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []repository.AdminProjectInquiryRow{}
	}
	total, err := s.inquiries.CountAdmin(ctx)
	if err != nil {
		return nil, err
	}
	return &ProjectInquiryList{Items: items, Total: total, Limit: limit, Offset: offset}, nil
}

func (s *ProjectInquiryService) GetAdminDetail(ctx context.Context, id string) (*repository.AdminProjectInquiryDetail, error) {
	s.attachCalculatorShare(ctx, id)
	return s.inquiries.GetAdminDetail(ctx, id)
}

func (s *ProjectInquiryService) UpdateStatus(ctx context.Context, id, status string) (*model.ProjectInquiry, error) {
	switch model.ProjectInquiryStatus(status) {
	case model.ProjectInquiryStatusNew,
		model.ProjectInquiryStatusInProgress,
		model.ProjectInquiryStatusDone,
		model.ProjectInquiryStatusSpam:
	default:
		return nil, ErrInvalidInput
	}
	return s.inquiries.UpdateStatus(ctx, id, status)
}

func (s *ProjectInquiryService) notifyAdmin(ctx context.Context, inquiryID string) {
	detail, err := s.inquiries.GetAdminDetail(ctx, inquiryID)
	if err != nil {
		slog.Warn("project inquiry admin notify load failed", "id", inquiryID, "err", err)
		return
	}
	s.telegram.NotifyProjectInquiry(ctx, detail, s.cfg.PublicBaseURL())
}

func (s *ProjectInquiryService) attachCalculatorShare(ctx context.Context, inquiryID string) {
	if s.share == nil {
		return
	}

	row, err := s.inquiries.GetForShareAttach(ctx, inquiryID)
	if err != nil {
		slog.Warn("project inquiry share attach load failed", "id", inquiryID, "err", err)
		return
	}
	if row.ShareToken != nil && *row.ShareToken != "" {
		return
	}
	hasRun := row.CalculatorRunID != nil && *row.CalculatorRunID != ""
	hasSnapshot := len(row.CalculatorSnapshot) > 0
	if !hasRun && !hasSnapshot {
		return
	}

	ownerUserID := ""
	if row.UserID != nil && *row.UserID != "" {
		ownerUserID = *row.UserID
	} else {
		id, err := s.users.GetFirstSuperadminID(ctx)
		if err != nil {
			slog.Warn("project inquiry share attach owner lookup failed", "id", inquiryID, "err", err)
			return
		}
		ownerUserID = id
	}

	result, err := s.share.CreateInquiryShare(
		ctx,
		ownerUserID,
		row.CalculatorRunID,
		row.CalculatorSnapshot,
		s.cfg.PublicBaseURL(),
	)
	if err != nil {
		slog.Warn("project inquiry share attach failed", "id", inquiryID, "err", err)
		return
	}

	var runID *string
	if result.RunID != "" {
		runID = &result.RunID
	}
	if err := s.inquiries.SetShareAttachment(ctx, inquiryID, runID, result.Token); err != nil {
		slog.Warn("project inquiry share token save failed", "id", inquiryID, "err", err)
	}
}

func (s *ProjectInquiryService) sendReceivedConfirmation(ctx context.Context, locale, email, name, telegram string) {
	subject, body := inquiryReceivedEmailContent(locale, name, telegram)
	if err := s.mail.Send(ctx, email, subject, body); err != nil {
		slog.Warn("inquiry confirmation email failed", "email", email, "err", err)
	}
}

func (s *ProjectInquiryService) validateRunOwnership(ctx context.Context, userID, runID string) error {
	run, err := s.runs.GetByIDForUser(ctx, runID, userID)
	if err != nil {
		return err
	}
	if run.ToolSlug != "calculator" {
		return ErrInvalidInput
	}
	return nil
}

type normalizedInquiry struct {
	name, email, telegram, projectTitle, projectDescription, locale string
	runID                                                           *string
	calculatorSnapshot                                              json.RawMessage
}

func normalizeProjectInquiryInput(input ProjectInquiryInput) normalizedInquiry {
	locale := strings.TrimSpace(input.Locale)
	if locale == "" {
		locale = "ru"
	}
	telegram := strings.TrimSpace(input.Telegram)
	if !strings.HasPrefix(telegram, "@") {
		telegram = "@" + telegram
	}

	var runID *string
	if id := strings.TrimSpace(input.CalculatorRunID); id != "" {
		runID = &id
	}

	snapshot := normalizeCalculatorSnapshot(input.CalculatorSnapshot)

	return normalizedInquiry{
		name:               strings.TrimSpace(input.Name),
		email:              strings.ToLower(strings.TrimSpace(input.Email)),
		telegram:           telegram,
		projectTitle:       strings.TrimSpace(input.ProjectTitle),
		projectDescription: strings.TrimSpace(input.ProjectDescription),
		locale:             locale,
		runID:              runID,
		calculatorSnapshot: snapshot,
	}
}

func validateProjectInquiryInput(input ProjectInquiryInput) error {
	if len(input.CalculatorSnapshot) > 65536 {
		return ErrInvalidInput
	}
	n := normalizeProjectInquiryInput(input)
	if n.name == "" || n.email == "" || n.telegram == "" || n.projectDescription == "" {
		return ErrInquiryMissingFields
	}
	if !strings.Contains(n.email, "@") {
		return ErrInvalidInput
	}
	if !validateTelegramContact(n.telegram) {
		return ErrInquiryInvalidTelegram
	}
	if len(n.projectDescription) > 5000 {
		return ErrInvalidInput
	}
	return nil
}

func validateTelegramContact(telegram string) bool {
	telegram = strings.TrimSpace(telegram)
	if telegram == "" {
		return false
	}
	lower := strings.ToLower(telegram)
	if strings.Contains(lower, "t.me/") {
		return true
	}
	handle := strings.TrimPrefix(telegram, "@")
	if telegramUsernamePattern.MatchString(handle) {
		return true
	}
	if telegramPhonePattern.MatchString(telegram) {
		return true
	}
	return telegramHandleFallback.MatchString(handle)
}

func normalizeCalculatorSnapshot(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 || !json.Valid(raw) {
		return nil
	}
	var snap struct {
		Input  json.RawMessage `json:"input"`
		Output json.RawMessage `json:"output"`
	}
	if err := json.Unmarshal(raw, &snap); err != nil {
		return nil
	}
	if len(snap.Input) == 0 || len(snap.Output) == 0 || !json.Valid(snap.Input) || !json.Valid(snap.Output) {
		return nil
	}
	return raw
}

func (s *ProjectInquiryService) verificationURL(rawToken string) string {
	return fmt.Sprintf("%s/dashboard/discuss/confirmed?token=%s", strings.TrimRight(s.cfg.PublicBaseURL(), "/"), rawToken)
}

func (s *ProjectInquiryService) generateToken() (raw, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	raw = hex.EncodeToString(b)
	return raw, s.hashToken(raw), nil
}

func (s *ProjectInquiryService) hashToken(raw string) string {
	sum := sha256.Sum256([]byte(s.keySalt + ":inquiry:" + raw))
	return hex.EncodeToString(sum[:])
}

func inquiryVerificationEmailContent(locale, link, name string) (subject, html string) {
	if locale == "en" {
		return "Confirm your project discussion request — Erman AI",
			fmt.Sprintf(`<p>Hello, %s!</p>
<p>Please confirm your email to submit the project discussion request.</p>
<p><a href="%s">Confirm request</a></p>
<p>The link expires in 24 hours.</p>`, name, link)
	}
	return "Подтвердите заявку «Обсудить проект» — Erman AI",
		fmt.Sprintf(`<p>Здравствуйте, %s!</p>
<p>Подтвердите email, чтобы мы получили вашу заявку на обсуждение проекта.</p>
<p><a href="%s">Подтвердить заявку</a></p>
<p>Ссылка действует 24 часа.</p>`, name, link)
}

func inquiryReceivedEmailContent(locale, name, telegram string) (subject, html string) {
	if locale == "en" {
		return "We received your project discussion request — Erman AI",
			fmt.Sprintf(`<p>Hello, %s!</p>
<p>We received your request and will contact you on Telegram %s based on your calculator data and project description.</p>
<p>Erman AI team</p>`, name, telegram)
	}
	return "Заявка «Обсудить проект» получена — Erman AI",
		fmt.Sprintf(`<p>Здравствуйте, %s!</p>
<p>Мы получили вашу заявку и свяжемся с вами в Telegram %s — обсудим автоматизацию на базе ваших данных и расчёта.</p>
<p>Команда Erman AI</p>`, name, telegram)
}

func HashIP(ip, salt string) string {
	sum := sha256.Sum256([]byte(salt + ":ip:" + strings.TrimSpace(ip)))
	return hex.EncodeToString(sum[:16])
}
