package setting

import (
	"encoding/json"
	"fmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

type DatasourceSettings struct {
	Version string `json:"version"`
}

type DatasourceSecretSettings struct {
	SecretId  string `json:"secretId"`
	SecretKey string `json:"secretKey"`
}

func LoadSettings(ctx backend.PluginContext) (*DatasourceSecretSettings, error) {
	model := &DatasourceSecretSettings{}

	settings := ctx.DataSourceInstanceSettings
	err := json.Unmarshal(settings.JSONData, &model)
	if err != nil {
		return nil, fmt.Errorf("error reading settings: %s", err.Error())
	}
	model.SecretKey = settings.DecryptedSecureJSONData["secretKey"]

	return model, nil
}
