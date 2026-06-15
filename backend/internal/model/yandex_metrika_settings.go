package model

import "time"

type YandexMetrikaSettings struct {
	Enabled     bool   `json:"enabled"`
	CounterCode string `json:"counter_code"`
}

type YandexMetrikaSettingsRecord struct {
	Config    YandexMetrikaSettings `json:"config"`
	UpdatedAt time.Time             `json:"updated_at"`
}

type YandexMetrikaAdminView struct {
	Settings  YandexMetrikaSettings `json:"settings"`
	UpdatedAt time.Time             `json:"updated_at"`
}

type YandexMetrikaPublicView struct {
	Enabled     bool   `json:"enabled"`
	CounterCode string `json:"counter_code"`
}

type YandexMetrikaAdminUpdateRequest struct {
	Settings YandexMetrikaSettings `json:"settings"`
}

func DefaultYandexMetrikaSettings() YandexMetrikaSettings {
	return YandexMetrikaSettings{
		Enabled:     false,
		CounterCode: "",
	}
}
