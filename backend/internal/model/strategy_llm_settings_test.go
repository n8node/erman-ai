package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsValidYandexCloudFolderID(t *testing.T) {
	require.True(t, IsValidYandexCloudFolderID("b1gtnjtlhud3pof106in"))
	require.False(t, IsValidYandexCloudFolderID("erman.ai@yandex.ru"))
	require.False(t, IsValidYandexCloudFolderID(""))
	require.False(t, IsValidYandexCloudFolderID("b1"))
}
