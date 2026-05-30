package repository

import (
	"context"
	"errors"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TelegramUserStateRepository struct {
	pool *pgxpool.Pool
}

func NewTelegramUserStateRepository(pool *pgxpool.Pool) *TelegramUserStateRepository {
	return &TelegramUserStateRepository{pool: pool}
}

func (r *TelegramUserStateRepository) GetMode(ctx context.Context, userChatID string) (string, error) {
	const q = `SELECT mode FROM telegram_bot_user_state WHERE user_chat_id = $1`
	var mode string
	err := r.pool.QueryRow(ctx, q, userChatID).Scan(&mode)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.TelegramUserModeIdle, nil
	}
	if err != nil {
		return "", err
	}
	return mode, nil
}

func (r *TelegramUserStateRepository) SetMode(ctx context.Context, userChatID, mode string) error {
	const q = `
		INSERT INTO telegram_bot_user_state (user_chat_id, mode, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (user_chat_id) DO UPDATE SET mode = EXCLUDED.mode, updated_at = NOW()
	`
	_, err := r.pool.Exec(ctx, q, userChatID, mode)
	return err
}
