package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetMetricHandler(t *testing.T) {
	type want struct {
		code int
		body []string
	}
	tests := []struct {
		name   string
		URL    string
		method string
		want   want
	}{
		{
			name:   "200 Success update gauge number with dot",
			URL:    "/update/gauge/Alloc/200.12",
			method: http.MethodPost,
			want:   want{code: 200},
		},
		{
			name:   "200 Success update gauge number without dot",
			URL:    "/update/gauge/Alloc/200",
			method: http.MethodPost,
			want:   want{code: 200},
		},
		{
			name:   "200 Success update counter",
			URL:    "/update/counter/PollCount/5",
			method: http.MethodPost,
			want:   want{code: 200},
		},
		{
			name:   "200 Success Get counter",
			URL:    "/value/counter/PollCount",
			method: http.MethodGet,
			want: want{code: 200,
				body: []string{"5"},
			},
		},
		{
			name:   "200 Success get gauge",
			URL:    "/value/gauge/Alloc",
			method: http.MethodGet,
			want: want{
				code: 200,
				body: []string{"200"},
			},
		},
		{
			name:   "400 update gauge parse error",
			URL:    "/update/gauge/stringMetric/aaa",
			method: http.MethodPost,
			want:   want{code: 400},
		},
		{
			name:   "400 update counter parse error",
			URL:    "/update/counter/PollCounter/11.12",
			method: http.MethodPost,
			want:   want{code: 400},
		},
		{
			name:   "400 post: wrong metric type",
			URL:    "/update/wrongType/name/123",
			method: http.MethodPost,
			want:   want{code: 501},
		},
		{
			name:   "404 no such gauge",
			URL:    "/value/gauge/wrongGauge1223r412",
			method: http.MethodGet,
			want: want{code: 404,
				body: []string{"There is no metric you requested\n"}},
		},
		{
			name:   "404 no such counter",
			URL:    "/value/counter/adadadad",
			method: http.MethodGet,
			want: want{code: 404,
				body: []string{"There is no metric you requested\n"}},
		},
	}
	server := httptest.NewServer(Router())
	defer server.Close()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := RunRequest(t, server, tt.method, tt.URL, "")
			defer resp.Body.Close()
			assert.Equal(t, tt.want.code, resp.StatusCode)
			for _, s := range tt.want.body {
				assert.Equal(t, body, s)
			}
			assert.Equal(t, tt.want.code, resp.StatusCode)
		})
	}
}

func RunRequest(t *testing.T, ts *httptest.Server, method string, query string, body string) (*http.Response, string) {
	reader := strings.NewReader(body)
	req, err := http.NewRequest(method, ts.URL+query, reader)
	req.Header.Add("Content-Type", "text/plain")

	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}
