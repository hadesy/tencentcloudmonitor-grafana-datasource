package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/resource/httpadapter"
	"github.com/hadesy/tencentcloudmonitor-datasource/pkg/setting"
	"github.com/hadesy/tencentcloudmonitor-datasource/pkg/tencentcloud"
	"net/http"
)

// Make sure Datasource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. In this example datasource instance implements backend.QueryDataHandler,
// backend.CheckHealthHandler interfaces. Plugin should not implement all these
// interfaces- only those which are required for a particular task.
var (
	_ backend.QueryDataHandler      = (*TencentCloudDataSource)(nil)
	_ backend.CheckHealthHandler    = (*TencentCloudDataSource)(nil)
	_ instancemgmt.InstanceDisposer = (*TencentCloudDataSource)(nil)
)

// NewDatasource creates a new datasource instance.
func NewDatasource(_ backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	mux := http.NewServeMux()

	ds := &TencentCloudDataSource{
		resourceHandler: httpadapter.New(mux),
	}
	mux.HandleFunc("/cvm-regions", ds.handleQueryCvmRegions)
	mux.HandleFunc("/monitor/describeBaseMetrics", ds.handleQueryMetrics)
	return ds, nil
}

func (d *TencentCloudDataSource) CallResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	return d.resourceHandler.CallResource(ctx, req, sender)
}

func (d *TencentCloudDataSource) handleQueryCvmRegions(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		return
	}
	ctx := req.Context()
	config, err := setting.LoadSettings(httpadapter.PluginConfigFromContext(ctx))
	if err != nil {
		writeResult(rw, "?", nil, err)
		return
	}

	res, err := tencentcloud.QueryCvmRegions(ctx, config)
	writeResult(rw, "cvmRegions", res, err)
}

func (d *TencentCloudDataSource) handleQueryMetrics(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		return
	}
	ctx := req.Context()
	config, err := setting.LoadSettings(httpadapter.PluginConfigFromContext(ctx))
	if err != nil {
		writeResult(rw, "?", nil, err)
		return
	}

	query := req.URL.Query()
	var (
		namespace = query.Get("namespace")
		region    = query.Get("region")
	)

	res, err := tencentcloud.QueryMonitorMetrics(ctx, config, namespace, region)
	writeResult(rw, "metrics", res, err)
}

type TencentCloudDataSource struct {
	resourceHandler backend.CallResourceHandler
}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewSampleDatasource factory function.
func (d *TencentCloudDataSource) Dispose() {
	// Clean up datasource instance resources.
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *TencentCloudDataSource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := d.query(ctx, req.PluginContext, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

func (d *TencentCloudDataSource) query(ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	var response backend.DataResponse

	config, err := setting.LoadSettings(pCtx)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("LoadSettings error: %v", err.Error()))
	}

	res, err := tencentcloud.QueryMonitorData(ctx, config, query)

	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("Query error: %v", err.Error()))
	}

	response.Frames = res
	return response
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *TencentCloudDataSource) CheckHealth(_ context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	var status = backend.HealthStatusOk
	var message = "ok"

	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil
}

func writeResult(rw http.ResponseWriter, path string, val interface{}, err error) {
	response := make(map[string]interface{})
	code := http.StatusOK
	if err != nil {
		response["error"] = err.Error()
		code = http.StatusBadRequest
	} else {
		response[path] = val
	}

	body, err := json.Marshal(response)
	if err != nil {
		body = []byte(err.Error())
		code = http.StatusInternalServerError
	}
	_, err = rw.Write(body)
	if err != nil {
		code = http.StatusInternalServerError
	}
	rw.WriteHeader(code)
}
