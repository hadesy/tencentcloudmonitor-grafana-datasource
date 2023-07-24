package plugin

import (
	"context"
	"github.com/hadesy/tencentcloudmonitor-datasource/pkg/setting"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

type TencentCloudService interface {
	QueryCvmRegions(ctx context.Context, config *setting.DatasourceSecretSettings) ([]*cvm.RegionInfo, error)
}
