-- +goose Up
CREATE TABLE consultation_services (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(120) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    duration_minutes INTEGER NOT NULL CHECK (duration_minutes > 0),
    price_rub INTEGER NOT NULL CHECK (price_rub >= 0),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    min_notice_minutes INTEGER NOT NULL DEFAULT 180 CHECK (min_notice_minutes >= 0),
    max_advance_days INTEGER NOT NULL DEFAULT 30 CHECK (max_advance_days > 0),
    buffer_before_minutes INTEGER NOT NULL DEFAULT 0 CHECK (buffer_before_minutes >= 0),
    buffer_after_minutes INTEGER NOT NULL DEFAULT 0 CHECK (buffer_after_minutes >= 0),
    meeting_url TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 100,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE consultation_availability_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_id UUID NOT NULL REFERENCES consultation_services(id) ON DELETE CASCADE,
    weekday SMALLINT NOT NULL CHECK (weekday BETWEEN 0 AND 6),
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    slot_step_minutes INTEGER NOT NULL DEFAULT 15 CHECK (slot_step_minutes > 0),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (start_time < end_time)
);

CREATE INDEX idx_consultation_availability_service ON consultation_availability_rules(service_id, weekday);

CREATE TABLE consultation_blackouts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_id UUID REFERENCES consultation_services(id) ON DELETE CASCADE,
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    reason TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (starts_at < ends_at)
);

CREATE INDEX idx_consultation_blackouts_range ON consultation_blackouts(starts_at, ends_at);

CREATE TABLE consultation_bookings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_id UUID NOT NULL REFERENCES consultation_services(id),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    customer_name VARCHAR(255) NOT NULL,
    customer_email VARCHAR(255) NOT NULL,
    customer_phone VARCHAR(100) NOT NULL DEFAULT '',
    customer_telegram VARCHAR(100) NOT NULL DEFAULT '',
    customer_note TEXT NOT NULL DEFAULT '',
    starts_at TIMESTAMPTZ NOT NULL,
    ends_at TIMESTAMPTZ NOT NULL,
    timezone VARCHAR(100) NOT NULL DEFAULT 'Europe/Moscow',
    status VARCHAR(50) NOT NULL DEFAULT 'pending_payment',
    amount_rub INTEGER NOT NULL CHECK (amount_rub >= 0),
    provider VARCHAR(50),
    external_id VARCHAR(255),
    inv_id BIGINT,
    meeting_url TEXT NOT NULL DEFAULT '',
    expires_at TIMESTAMPTZ NOT NULL,
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (starts_at < ends_at)
);

CREATE INDEX idx_consultation_bookings_service_time ON consultation_bookings(service_id, starts_at, ends_at);
CREATE INDEX idx_consultation_bookings_status ON consultation_bookings(status);
CREATE INDEX idx_consultation_bookings_user ON consultation_bookings(user_id);
CREATE UNIQUE INDEX idx_consultation_bookings_provider_external ON consultation_bookings(provider, external_id)
    WHERE external_id IS NOT NULL;
CREATE UNIQUE INDEX idx_consultation_bookings_inv ON consultation_bookings(inv_id)
    WHERE inv_id IS NOT NULL;
CREATE INDEX idx_consultation_bookings_slot ON consultation_bookings(service_id, starts_at);

INSERT INTO consultation_services (
    slug,
    name,
    description,
    duration_minutes,
    price_rub,
    min_notice_minutes,
    max_advance_days,
    sort_order
) VALUES (
    'strategic-session',
    'Стратегическая консультация',
    'Разберём текущие бизнес-процессы, найдём быстрые AI/automation wins и посчитаем экономический эффект.',
    90,
    100,
    180,
    30,
    10
);

INSERT INTO consultation_availability_rules (service_id, weekday, start_time, end_time, slot_step_minutes)
SELECT id, weekday, '10:00'::time, '18:00'::time, 15
FROM consultation_services
CROSS JOIN generate_series(1, 5) AS weekday
WHERE slug = 'strategic-session';

-- +goose Down
DROP TABLE IF EXISTS consultation_bookings;
DROP TABLE IF EXISTS consultation_blackouts;
DROP TABLE IF EXISTS consultation_availability_rules;
DROP TABLE IF EXISTS consultation_services;
