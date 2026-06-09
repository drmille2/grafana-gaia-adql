package plugin

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// Ensure GaiaADQLDatasource implements the required interfaces.
var (
	_ backend.QueryDataHandler   = (*GaiaADQLDatasource)(nil)
	_ backend.CheckHealthHandler = (*GaiaADQLDatasource)(nil)
	_ instancemgmt.InstanceDisposer = (*GaiaADQLDatasource)(nil)
)

const (
	defaultBaseURL      = "https://gea.esac.esa.int/tap-server/tap"
	defaultMaxRows      = 1000
	httpTimeout         = 120 * time.Second
	dialTimeout         = 10 * time.Second
	tlsHandshakeTimeout = 10 * time.Second
)

// pluginSettings holds the JSON data stored in the datasource's jsonData field.
type pluginSettings struct {
	BaseURL string `json:"baseUrl"`
}

// queryModel represents a single query submitted from the frontend query editor.
type queryModel struct {
	AdqlQuery string `json:"adqlQuery"`
	MaxRows   int    `json:"maxRows"`
}

// GaiaADQLDatasource is the plugin backend struct. It handles query execution
// against the ESA Gaia archive TAP service.
type GaiaADQLDatasource struct {
	baseURL    string
	httpClient *http.Client
}

// NewGaiaADQLDatasource creates a new datasource instance for a given set
// of Grafana data source settings. It is called once per configured data source.
func NewGaiaADQLDatasource(ctx context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	var ps pluginSettings
	if err := json.Unmarshal(settings.JSONData, &ps); err != nil {
		return nil, fmt.Errorf("error unmarshalling datasource settings: %w", err)
	}

	baseURL := ps.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	// Trim trailing slash for consistent URL construction.
	baseURL = strings.TrimRight(baseURL, "/")

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   dialTimeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: tlsHandshakeTimeout,
	}

	return &GaiaADQLDatasource{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout:   httpTimeout,
			Transport: transport,
		},
	}, nil
}

// Dispose is called when the data source instance is removed.
func (d *GaiaADQLDatasource) Dispose() {
	// Clean up any resources here if needed.
}

// QueryData handles query requests from Grafana. Each request may contain
// multiple queries (one per panel query row). Results are returned as a
// map of data frames keyed by the query's RefID.
func (d *GaiaADQLDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	response := backend.NewQueryDataResponse()

	for _, q := range req.Queries {
		res := d.query(ctx, req.PluginContext, q)
		response.Responses[q.RefID] = res
	}

	return response, nil
}

// query executes a single ADQL query against the TAP service and returns
// the result as a backend.DataResponse containing a data frame.
func (d *GaiaADQLDatasource) query(ctx context.Context, _ backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	response := backend.DataResponse{}

	// Parse the query model from the JSON payload.
	var qm queryModel
	if err := json.Unmarshal(query.JSON, &qm); err != nil {
		response.Error = fmt.Errorf("error unmarshalling query: %w", err)
		return response
	}

	if strings.TrimSpace(qm.AdqlQuery) == "" {
		response.Error = fmt.Errorf("ADQL query is empty")
		return response
	}

	maxRows := qm.MaxRows
	if maxRows <= 0 {
		maxRows = defaultMaxRows
	}

	// Build the TAP sync query URL.
	tapURL := d.buildTAPURL(qm.AdqlQuery, maxRows)
	log.DefaultLogger.Debug("Executing ADQL query", "url", tapURL, "refId", query.RefID)

	// Execute the HTTP request to the TAP service.
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, tapURL, nil)
	if err != nil {
		response.Error = fmt.Errorf("error creating HTTP request: %w", err)
		return response
	}
	httpReq.Header.Set("Accept", "application/x-votable+xml, application/json")

	resp, err := d.httpClient.Do(httpReq)
	if err != nil {
		response.Error = fmt.Errorf("error executing TAP query: %w", err)
		return response
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		response.Error = fmt.Errorf("TAP service returned HTTP %d: %s", resp.StatusCode, string(body))
		return response
	}

	// Parse the CSV response into a Grafana data frame.
	frame, err := parseCSVResponse(resp.Body, query.RefID)
	if err != nil {
		response.Error = fmt.Errorf("error parsing CSV response: %w", err)
		return response
	}

	response.Frames = append(response.Frames, frame)
	return response
}

// buildTAPURL constructs the URL for a TAP synchronous query request.
// The TAP sync endpoint accepts ADQL queries via the QUERY parameter.
func (d *GaiaADQLDatasource) buildTAPURL(adqlQuery string, maxRows int) string {
	params := url.Values{}
	params.Set("REQUEST", "doQuery")
	params.Set("LANG", "ADQL")
	params.Set("FORMAT", "csv")
	params.Set("QUERY", adqlQuery)
	if maxRows > 0 {
		params.Set("MAXREC", strconv.Itoa(maxRows))
	}

	return fmt.Sprintf("%s/sync?%s", d.baseURL, params.Encode())
}

// CheckHealth verifies the TAP service is reachable by executing a minimal
// metadata query to the TAP capabilities endpoint.
func (d *GaiaADQLDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	// Probe the TAP availability endpoint.
	availabilityURL := fmt.Sprintf("%s/availability", d.baseURL)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, availabilityURL, nil)
	if err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("Failed to create health check request: %s", err.Error()),
		}, nil
	}

	resp, err := d.httpClient.Do(httpReq)
	if err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("Cannot reach TAP service at %s: %s", d.baseURL, err.Error()),
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("TAP service returned server error: HTTP %d", resp.StatusCode),
		}, nil
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: fmt.Sprintf("TAP service is reachable at %s", d.baseURL),
	}, nil
}

// ---------------------------------------------------------------------------
// CSV parsing
// ---------------------------------------------------------------------------

// parseCSVResponse reads a CSV response from the TAP service and converts it
// to a Grafana data frame. Column types are inferred by attempting to parse
// all values in each column as int64, then float64, falling back to string.
func parseCSVResponse(r io.Reader, refID string) (*data.Frame, error) {
	cr := csv.NewReader(r)

	headers, err := cr.Read()
	if err != nil {
		return nil, fmt.Errorf("reading CSV header: %w", err)
	}

	records, err := cr.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("reading CSV data: %w", err)
	}

	frame := data.NewFrame(refID)
	if len(headers) == 0 {
		return frame, nil
	}

	numCols := len(headers)

	// Determine per-column type by scanning all values.
	// Start optimistic (int64) and downgrade as needed.
	const (
		typeInt64   = 0
		typeFloat64 = 1
		typeString  = 2
	)
	colTypes := make([]int, numCols)
	for _, row := range records {
		for j := 0; j < numCols && j < len(row); j++ {
			cell := row[j]
			if cell == "" {
				continue
			}
			switch colTypes[j] {
			case typeInt64:
				if _, err := strconv.ParseInt(cell, 10, 64); err != nil {
					if _, err := strconv.ParseFloat(cell, 64); err != nil {
						colTypes[j] = typeString
					} else {
						colTypes[j] = typeFloat64
					}
				}
			case typeFloat64:
				if _, err := strconv.ParseFloat(cell, 64); err != nil {
					colTypes[j] = typeString
				}
			}
		}
	}

	// Allocate typed column slices.
	columns := make([]interface{}, numCols)
	for j, ct := range colTypes {
		switch ct {
		case typeInt64:
			columns[j] = make([]*int64, 0, len(records))
		case typeFloat64:
			columns[j] = make([]*float64, 0, len(records))
		default:
			columns[j] = make([]*string, 0, len(records))
		}
	}

	// Populate rows.
	for _, row := range records {
		for j := 0; j < numCols && j < len(row); j++ {
			cell := row[j]
			switch colTypes[j] {
			case typeInt64:
				s := columns[j].([]*int64)
				if cell == "" {
					columns[j] = append(s, nil)
				} else {
					v, _ := strconv.ParseInt(cell, 10, 64)
					columns[j] = append(s, &v)
				}
			case typeFloat64:
				s := columns[j].([]*float64)
				if cell == "" {
					columns[j] = append(s, nil)
				} else {
					v, _ := strconv.ParseFloat(cell, 64)
					columns[j] = append(s, &v)
				}
			default:
				s := columns[j].([]*string)
				if cell == "" {
					columns[j] = append(s, nil)
				} else {
					cp := cell
					columns[j] = append(s, &cp)
				}
			}
		}
	}

	// Add typed fields to the frame.
	for j, name := range headers {
		var field *data.Field
		switch v := columns[j].(type) {
		case []*int64:
			field = data.NewField(name, nil, v)
		case []*float64:
			field = data.NewField(name, nil, v)
		case []*string:
			field = data.NewField(name, nil, v)
		}
		if field != nil {
			frame.Fields = append(frame.Fields, field)
		}
	}

	return frame, nil
}
