package repository

import (
	"context"
	"errors"
	"time"

	"github.com/erman-ai/erman-ai/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TelegramSupportThreadRepository struct {
	pool *pgxpool.Pool
}

func NewTelegramSupportThreadRepository(pool *pgxpool.Pool) *TelegramSupportThreadRepository {
	return &TelegramSupportThreadRepository{pool: pool}
}

func (r *TelegramSupportThreadRepository) GetByUserChatID(ctx context.Context, userChatID string) (*model.TelegramSupportThread, error) {
	const q = `
		SELECT user_chat_id, forum_chat_id, topic_id, display_name, created_at, updated_at
		FROM telegram_support_threads
		WHERE user_chat_id = $1
	`
	var t model.TelegramSupportThread
	err := r.pool.QueryRow(ctx, q, userChatID).Scan(
		&t.UserChatID, &t.ForumChatID, &t.TopicID, &t.DisplayName, &t.CreatedAt, &t.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TelegramSupportThreadRepository) GetByForumTopic(ctx context.Context, forumChatID string, topicID int) (*model.TelegramSupportThread, error) {
	const q = `
		SELECT user_chat_id, forum_chat_id, topic_id, display_name, created_at, updated_at
		FROM telegram_support_threads
		WHERE forum_chat_id = $1 AND topic_id = $2
	`
	var t model.TelegramSupportThread
	err := r.pool.QueryRow(ctx, q, forumChatID, topicID).Scan(
		&t.UserChatID, &t.ForumChatID, &t.TopicID, &t.DisplayName, &t.CreatedAt, &t.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *TelegramSupportThreadRepository) Upsert(ctx context.Context, t model.TelegramSupportThread) error {
	const q = `
		INSERT INTO telegram_support_threads (user_chat_id, forum_chat_id, topic_id, display_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT (user_chat_id) DO UPDATE SET
			forum_chat_id = EXCLUDED.forum_chat_id,
			topic_id = EXCLUDED.topic_id,
			display_name = EXCLUDED.display_name,
			updated_at = NOW()
	`
	_, err := r.pool.Exec(ctx, q, t.UserChatID, t.ForumChatID, t.TopicID, t.DisplayName)
	return err
}

func (r *TelegramSupportThreadRepository) Touch(ctx context.Context, userChatID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE telegram_support_threads SET updated_at = $2 WHERE user_chat_id = $1
	`, userChatID, time.Now())
	return err
}
