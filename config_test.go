package goralens

import (
	"testing"

	mocksconfig "github.com/goravel/framework/mocks/config"
)

func TestSettingPrefersConfigThenEnv(t *testing.T) {
	cfg := mocksconfig.NewConfig(t)
	cfg.EXPECT().Get("goralens.token_storage_key").Return("from_config").Once()
	if got := setting(cfg, "goralens.token_storage_key", "GORALENS_TOKEN_STORAGE_KEY", ""); got != "from_config" {
		t.Fatalf("config value: %v", got)
	}

	cfg.EXPECT().Get("goralens.token_storage_key").Return(nil).Once()
	cfg.EXPECT().Env("GORALENS_TOKEN_STORAGE_KEY").Return("app_token").Once()
	if got := setting(cfg, "goralens.token_storage_key", "GORALENS_TOKEN_STORAGE_KEY", ""); got != "app_token" {
		t.Fatalf("env value: %v", got)
	}

	cfg.EXPECT().Get("goralens.enabled").Return(nil).Once()
	cfg.EXPECT().Env("GORALENS_ENABLED").Return(nil).Once()
	if got := setting(cfg, "goralens.enabled", "GORALENS_ENABLED", true); got != true {
		t.Fatalf("fallback: %v", got)
	}
}
