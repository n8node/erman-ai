package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TelegramUrgentSendRepository struct {
	pool *pgxpool.Pool
}

func NewTelegramUrgentSendRepository(pool *pgxpool.Pool) *TelegramUrgentSendRepository {
	return &TelegramUrgentSendRepository{pool: pool}
}

func (r *TelegramUrgentSendRepository) CountSince(ctx context.Context, userChatID string, since time.Time) (int, error) {
	const q = `
		SELECT COUNT(*)::int FROM telegram_urgent_sends
		WHERE user_chat_id = $1 AND sent_at >= $2
	`
	var n int
	err := r.pool.QueryRow(ctx, q, userChatID, since).Scan(&n)
	return n, err
}

func (r *TelegramUrgentSendRepository) Insert(ctx context.Context, userChatID, preview string) error {
	const q = `
		INSERT INTO telegram_urgent_sends (user_chat_id, message_preview, sent_at)
		VALUES ($1, $2, NOW())
	`
	_, err := r.pool.Exec(ctx, q, userChatID, preview)
	return err
}

func startOfUTCDay(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func (r *TelegramUrgentSendRepository) CountTodayUTC(ctx context.Context, userChatID string) (int, error) {
	return r.CountSince(ctx, userChatID, startOfUTCDay(time.Now()))
}
