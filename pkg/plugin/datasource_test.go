package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sampleCSV returns a CSV document matching the columns used in the old VOTable fixture.
func sampleCSV() string {
	return "source_id,ra,dec,parallax,phot_g_mean_mag\n" +
		"4472832130942575872,280.46252,-60.08422,130.231,2.835\n" +
		"5853498713190525696,217.42896,-62.67949,768.067,-0.271\n" +
		"2947050466531873024,101.28715,-16.71312,374.490,-1.640\n"
}

func TestQueryData_BasicQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/tap/sync", r.URL.Path)
		assert.Equal(t, "doQuery", r.URL.Query().Get("REQUEST"))
		assert.Equal(t, "ADQL", r.URL.Query().Get("LANG"))
		assert.Equal(t, "csv", r.URL.Query().Get("FORMAT"))

		w.Header().Set("Content-Type", "text/csv")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, sampleCSV())
	}))
	defer server.Close()

	ds := &GaiaADQLDatasource{
		baseURL:    server.URL + "/tap",
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}

	queryJSON, err := json.Marshal(queryModel{
		AdqlQuery: "SELECT TOP 3 source_id, ra, dec, parallax, phot_g_mean_mag FROM gaiadr3.gaia_source WHERE parallax > 100",
		MaxRows:   3,
	})
	require.NoError(t, err)

	req := &backend.QueryDataRequest{
		Queries: []backend.DataQuery{
			{
				RefID: "A",
				JSON:  queryJSON,
			},
		},
	}

	resp, err := ds.QueryData(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	result, ok := resp.Responses["A"]
	require.True(t, ok)
	require.NoError(t, result.Error)
	require.Len(t, result.Frames, 1)

	frame := result.Frames[0]
	assert.Equal(t, "A", frame.Name)
	assert.Equal(t, 5, len(frame.Fields), "expected 5 columns: source_id, ra, dec, parallax, phot_g_mean_mag")
	assert.Equal(t, 3, frame.Rows(), "expected 3 rows")

	assert.Equal(t, "source_id", frame.Fields[0].Name)
	assert.Equal(t, "ra", frame.Fields[1].Name)
	assert.Equal(t, "dec", frame.Fields[2].Name)
	assert.Equal(t, "parallax", frame.Fields[3].Name)
	assert.Equal(t, "phot_g_mean_mag", frame.Fields[4].Name)
}

func TestQueryData_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, "service unavailable")
	}))
	defer server.Close()

	ds := &GaiaADQLDatasource{
		baseURL:    server.URL + "/tap",
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}

	queryJSON, _ := json.Marshal(queryModel{AdqlQuery: "SELECT TOP 1 source_id FROM gaiadr3.gaia_source", MaxRows: 1})
	req := &backend.QueryDataRequest{
		Queries: []backend.DataQuery{{RefID: "A", JSON: queryJSON}},
	}

	resp, err := ds.QueryData(context.Background(), req)
	require.NoError(t, err)

	result := resp.Responses["A"]
	assert.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "503")
}

func TestQueryData_EmptyQuery(t *testing.T) {
	ds := &GaiaADQLDatasource{
		baseURL:    "https://example.com/tap",
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}

	queryJSON, _ := json.Marshal(queryModel{AdqlQuery: "", MaxRows: 100})
	req := &backend.QueryDataRequest{
		Queries: []backend.DataQuery{{RefID: "A", JSON: queryJSON}},
	}

	resp, err := ds.QueryData(context.Background(), req)
	require.NoError(t, err)

	result := resp.Responses["A"]
	assert.Error(t, result.Error)
	assert.Contains(t, result.Error.Error(), "empty")
}

func TestQueryData_WhitespaceOnlyQuery(t *testing.T) {
	ds := &GaiaADQLDatasource{
		baseURL:    "https://example.com/tap",
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}

	queryJSON, _ := json.Marshal(queryModel{AdqlQuery: "   \t\n  ", MaxRows: 100})
	req := &backend.QueryDataRequest{
		Queries: []backend.DataQuery{{RefID: "A", JSON: queryJSON}},
	}

	resp, err := ds.QueryData(context.Background(), req)
	require.NoError(t, err)

	result := resp.Responses["A"]
	assert.Error(t, result.Error)
}

func TestBuildTAPURL(t *testing.T) {
	ds := &GaiaADQLDatasource{baseURL: "https://gea.esac.esa.int/tap-server/tap"}

	adql := "SELECT TOP 10 source_id FROM gaiadr3.gaia_source"
	tapURL := ds.buildTAPURL(adql, 1000)

	assert.Contains(t, tapURL, "https://gea.esac.esa.int/tap-server/tap/sync")
	assert.Contains(t, tapURL, "REQUEST=doQuery")
	assert.Contains(t, tapURL, "LANG=ADQL")
	assert.Contains(t, tapURL, "FORMAT=csv")
	assert.Contains(t, tapURL, "MAXREC=1000")
	assert.Contains(t, tapURL, "SELECT")
}

func TestCheckHealth_OK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/tap/availability", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `<availability><available>true</available></availability>`)
	}))
	defer server.Close()

	ds := &GaiaADQLDatasource{
		baseURL:    server.URL + "/tap",
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}

	result, err := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	require.NoError(t, err)
	assert.Equal(t, backend.HealthStatusOk, result.Status)
	assert.Contains(t, result.Message, "reachable")
}

func TestCheckHealth_ServiceError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	ds := &GaiaADQLDatasource{
		baseURL:    server.URL + "/tap",
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}

	result, err := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	require.NoError(t, err)
	assert.Equal(t, backend.HealthStatusError, result.Status)
}

func TestParseCSVResponse_BasicParsing(t *testing.T) {
	frame, err := parseCSVResponse(strings.NewReader(sampleCSV()), "test")
	require.NoError(t, err)
	require.NotNil(t, frame)

	assert.Equal(t, 5, len(frame.Fields))
	assert.Equal(t, 3, frame.Rows())

	assert.Equal(t, "source_id", frame.Fields[0].Name)
	assert.Equal(t, "ra", frame.Fields[1].Name)
	assert.Equal(t, "dec", frame.Fields[2].Name)
	assert.Equal(t, "parallax", frame.Fields[3].Name)
	assert.Equal(t, "phot_g_mean_mag", frame.Fields[4].Name)
}

func TestParseCSVResponse_EmptyResponse(t *testing.T) {
	csvData := "source_id\n"

	frame, err := parseCSVResponse(strings.NewReader(csvData), "empty")
	require.NoError(t, err)
	require.NotNil(t, frame)
	assert.Equal(t, 0, frame.Rows())
	assert.Equal(t, 1, len(frame.Fields))
}

func TestParseCSVResponse_NullValues(t *testing.T) {
	csvData := "ra,name\n45.12,Sirius\n,\n"

	frame, err := parseCSVResponse(strings.NewReader(csvData), "nulls")
	require.NoError(t, err)
	require.NotNil(t, frame)
	assert.Equal(t, 2, frame.Rows())

	// First row should have a value.
	raField := frame.Fields[0]
	raVal := raField.At(0).(*float64)
	assert.NotNil(t, raVal)
	assert.InDelta(t, 45.12, *raVal, 0.001)

	// Second row should be nil.
	assert.Nil(t, raField.At(1))
}
