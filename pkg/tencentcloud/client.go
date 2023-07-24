package tencentcloud

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/hadesy/tencentcloudmonitor-datasource/pkg/model"
	"github.com/hadesy/tencentcloudmonitor-datasource/pkg/setting"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	monitor "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/monitor/v20180724"
	"time"
)

type Client struct {
	cvm     *cvm.Client
	monitor *monitor.Client
}

func NewTencentCloudClient(ctx context.Context, secretId string, secretKey string, region string) (*Client, error) {
	credential := common.NewCredential(secretId, secretKey)
	cvmClient, _ := cvm.NewClient(credential, region, profile.NewClientProfile())
	monitorClient, _ := monitor.NewClient(credential, region, profile.NewClientProfile())
	return &Client{cvmClient, monitorClient}, nil
}

func QueryCvmRegions(ctx context.Context, config *setting.DatasourceSecretSettings) ([]*cvm.RegionInfo, error) {
	tencentCloudClient, err := NewTencentCloudClient(ctx, config.SecretId, config.SecretKey, regions.Guangzhou)

	if err != nil {
		log.DefaultLogger.Error("Fail QueryCvmRegions", "error", err)
		return nil, err
	}

	request := cvm.NewDescribeRegionsRequest()
	response, err := tencentCloudClient.cvm.DescribeRegions(request)

	if err != nil {
		log.DefaultLogger.Error("Fail QueryCvmRegions", "error", err)
		return nil, err
	}

	return response.Response.RegionSet, nil

}

func QueryMonitorMetrics(ctx context.Context, config *setting.DatasourceSecretSettings, namespace string, region string) ([]*monitor.MetricSet, error) {
	tencentCloudClient, err := NewTencentCloudClient(ctx, config.SecretId, config.SecretKey, region)

	if err != nil {
		log.DefaultLogger.Error("Fail QueryMonitorMetrics", "error", err)
		return nil, err
	}

	request := monitor.NewDescribeBaseMetricsRequest()
	request.Namespace = &namespace
	response, err := tencentCloudClient.monitor.DescribeBaseMetrics(request)

	if err != nil {
		log.DefaultLogger.Error("Fail QueryMonitorMetrics", "error", err)
		return nil, err
	}

	return response.Response.MetricSet, nil

}

func QueryMonitorData(ctx context.Context, config *setting.DatasourceSecretSettings, query backend.DataQuery) ([]*data.Frame, error) {
	queryModel, err := GetQueryModel(query)
	tencentCloudClient, err := NewTencentCloudClient(ctx, config.SecretId, config.SecretKey, queryModel.Region)

	if err != nil {
		log.DefaultLogger.Error("Fail QueryMonitorMetrics", "error", err)
		return nil, err
	}

	if queryModel.Hide {
		return nil, nil
	}

	request := monitor.NewGetMonitorDataRequest()
	request.Namespace = common.StringPtr(queryModel.Service)
	request.MetricName = common.StringPtr(queryModel.Metric)
	request.Period = common.Uint64Ptr(queryModel.Period)
	request.StartTime = common.StringPtr(queryModel.StartTime)
	request.EndTime = common.StringPtr(queryModel.EndTime)

	var dimensions []*monitor.Dimension
	for _, dimension := range queryModel.Dimensions {
		dimensions = append(dimensions, &monitor.Dimension{
			Name:  common.StringPtr(dimension.Name),
			Value: common.StringPtr(dimension.Value),
		})
	}

	request.Instances = []*monitor.Instance{
		{
			Dimensions: dimensions,
		},
	}

	response, err := tencentCloudClient.monitor.GetMonitorData(request)

	if err != nil {
		log.DefaultLogger.Error("Fail QueryMonitorData", "error", err)
		return nil, err
	}

	return transformReportsResponseToDataFrames(response.Response), nil

}

func GetQueryModel(query backend.DataQuery) (*model.QueryModel, error) {
	model := &model.QueryModel{}
	err := json.Unmarshal(query.JSON, &model)
	if err != nil {
		return nil, fmt.Errorf("error reading query: %s", err.Error())
	}

	model.StartTime = query.TimeRange.From.Format("2006-01-02T15:04:05Z07:00")
	model.EndTime = query.TimeRange.To.Format("2006-01-02T15:04:05Z07:00")
	return model, nil
}

func transformReportsResponseToDataFrames(response *monitor.GetMonitorDataResponseParams) []*data.Frame {
	frame := data.NewFrame("")
	if len(response.DataPoints) == 0 {
		return []*data.Frame{frame}
	}
	labels := make(map[string]string)
	for _, v := range response.DataPoints[0].Dimensions {
		labels[*v.Name] = *v.Value
	}

	var timeValues []time.Time
	var pointValues []*float64
	for _, v := range response.DataPoints {
		for _, i := range v.Timestamps {
			timeValues = append(timeValues, time.Unix(int64(*i), 0))
		}
		pointValues = v.Values
	}

	frame.Fields = append(frame.Fields, data.NewField("Timestamps", labels, timeValues))
	frame.Fields = append(frame.Fields, data.NewField(*response.MetricName, labels, pointValues))

	return []*data.Frame{frame}
}
