package model

// LLMHTTPProxySettings configures outbound HTTP proxy for external LLM APIs (not Yandex).
type LLMHTTPProxySettings struct {
	Enabled           bool     `json:"enabled"`
	ProxyURLs         []string `json:"proxy_urls"`
	ProxyActiveURL    string   `json:"proxy_active_url"`
	ProxyAutoFailover bool     `json:"proxy_auto_failover"`
}

func DefaultLLMHTTPProxySettings() LLMHTTPProxySettings {
	return LLMHTTPProxySettings{
		ProxyAutoFailover: true,
		ProxyURLs:         []string{},
	}
}

func (s StrategyLLMSettings) ProxyForProvider(provider LLMProvider) *LLMHTTPProxySettings {
	switch provider {
	case LLMProviderDeepSeek:
		return s.deepSeekProxyPtr()
	case LLMProviderOpenRouter:
		return s.openRouterProxyPtr()
	default:
		return nil
	}
}

func (s StrategyLLMSettings) openRouterProxyPtr() *LLMHTTPProxySettings {
	cfg := s.OpenRouterProxy
	if !cfg.Enabled || len(cfg.ProxyURLs) == 0 {
		return nil
	}
	return &cfg
}

func (s StrategyLLMSettings) deepSeekProxyPtr() *LLMHTTPProxySettings {
	cfg := s.DeepSeekProxy
	if !cfg.Enabled || len(cfg.ProxyURLs) == 0 {
		return nil
	}
	return &cfg
}

func (s LegalScanLLMSettings) ProxyForProvider(provider LLMProvider) *LLMHTTPProxySettings {
	switch provider {
	case LLMProviderDeepSeek:
		cfg := s.DeepSeekProxy
		if !cfg.Enabled || len(cfg.ProxyURLs) == 0 {
			return nil
		}
		return &cfg
	case LLMProviderOpenRouter:
		cfg := s.OpenRouterProxy
		if !cfg.Enabled || len(cfg.ProxyURLs) == 0 {
			return nil
		}
		return &cfg
	default:
		return nil
	}
}
