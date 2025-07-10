package configuration

type ApplyProjectConfigurationRequest struct {
	Config                     map[string]any  `json:"config"`
	PreviousSave               string          `json:"previousSave"`
	Title                      string          `json:"title"`
	FastDataConfig             *map[string]any `json:"fastDataConfig,omitempty"`
	MicrofrontendPluginsConfig *map[string]any `json:"microfrontendPluginsConfig,omitempty"`
	ExtensionsConfig           *map[string]any `json:"extensionsConfig,omitempty"`
}

func CreateApplyConfigurationRequest(config map[string]any) ApplyProjectConfigurationRequest {
	applyConfig := ApplyProjectConfigurationRequest{}

	if fastDataConfig, ok := getConfig("fastDataConfig", config); ok && fastDataConfig != nil {
		applyConfig.FastDataConfig = &fastDataConfig
		delete(config, "fastDataConfig")
	}

	if extensionsConfig, ok := getConfig("extensionsConfig", config); ok && extensionsConfig != nil {
		applyConfig.ExtensionsConfig = &extensionsConfig
		delete(config, "extensionsConfig")
	}

	if microfrontendPluginsConfig, ok := getConfig("microfrontendPluginsConfig", config); ok && microfrontendPluginsConfig != nil {
		applyConfig.MicrofrontendPluginsConfig = &microfrontendPluginsConfig
		delete(config, "microfrontendPluginsConfig")
	}

	applyConfig.Config = config

	return applyConfig
}

func getConfig(configKey string, config map[string]any) (map[string]any, bool) {
	if configValue, ok := config[configKey]; ok {
		if configMap, ok := configValue.(map[string]any); ok {
			return configMap, true
		}
	}

	return nil, false
}
