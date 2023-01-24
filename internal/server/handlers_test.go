package server

import (
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Decompress(rw io.ReadCloser, r *http.Request) ([]byte, error) {
	var reader io.Reader
	if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			log.Printf("gzip new reader error: %v", err)
			return nil, err
		}
		reader = gz
		defer gz.Close()
	} else {
		reader = r.Body
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		log.Printf("error ReadAll: %v", err)
		return nil, err
	}
	return body, nil
}

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
			URL:    "/update/gauge/Alloc/200.10",
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
			name:   "501 post: wrong metric type",
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
	s := NewServer()
	server := httptest.NewServer(DecompressHandler(s.Router()))
	defer server.Close()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := RunRequest(t, server, tt.method, tt.URL, "", "text/plain")
			defer resp.Body.Close()
			assert.Equal(t, tt.want.code, resp.StatusCode)
			for _, s := range tt.want.body {
				assert.Equal(t, body, s)
			}
			assert.Equal(t, tt.want.code, resp.StatusCode)
		})
	}
}

func TestJSONHandlers(t *testing.T) {
	type want struct {
		code int
		body []string
	}
	tests := []struct {
		name   string
		URL    string
		method string
		body   string
		want   want
	}{
		{
			name:   "200 Success JSON update gauge number with dot",
			URL:    "/update/",
			method: http.MethodPost,
			body:   `{"id":"Alloc","type":"gauge","value":200.10}`,
			want: want{code: 200,
				body: []string{`{"id":"Alloc","type":"gauge","value":200.1}`}},
		},
		{
			name:   "200 Success JSON update gauge number without dot",
			URL:    "/update/",
			method: http.MethodPost,
			body:   `{"id":"Alloc","type":"gauge","value":200}`,
			want: want{code: 200,
				body: []string{`{"id":"Alloc","type":"gauge","value":200}`}},
		},
		{
			name:   "200 Success JSON update counter",
			URL:    "/update/",
			method: http.MethodPost,
			body:   `{"id":"PollCount","type":"counter","delta":5}`,
			want: want{code: 200,
				body: []string{`{"id":"PollCount","type":"counter","delta":5}`}},
		},
		{
			name:   "200 Success JSON Get counter",
			URL:    "/value/",
			method: http.MethodPost,
			body:   `{"id":"PollCount","type":"counter"}`,
			want: want{code: 200,
				body: []string{`{"id":"PollCount","type":"counter","delta":5}`},
			},
		},
		{
			name:   "200 Success JSON get gauge",
			URL:    "/value/",
			method: http.MethodPost,
			body:   `{"id":"Alloc","type":"gauge"}`,
			want: want{
				code: 200,
				body: []string{`{"id":"Alloc","type":"gauge","value":200}`},
			},
		},
		{
			name:   "400 JSON update gauge parse error",
			URL:    "/update/",
			method: http.MethodPost,
			body:   `{"id":"stringMetric","type":"gauge","value":"aaa"}`,
			want:   want{code: 400},
		},
		{
			name:   "400 JSON update counter parse error",
			URL:    "/update/",
			method: http.MethodPost,
			body:   `{"id":"PollCounter","type":"counter","delta":11.12}`,
			want:   want{code: 400},
		},
		{
			name:   "501 JSON post: wrong metric type",
			URL:    "/update/",
			method: http.MethodPost,
			body:   `{"id":"name","type":"wrongType","value":123}`,
			want:   want{code: 501},
		},
		{
			name:   "404 JSON get no such gauge",
			URL:    "/value/",
			method: http.MethodPost,
			body:   `{"id":"wrongGauge2121331","type":"gauge"}`,
			want: want{code: 404,
				body: []string{"There is no metric you requested\n"}},
		},
		{
			name:   "404 JSON get no such counter",
			URL:    "/value/",
			method: http.MethodPost,
			body:   `{"id":"asdasda","type":"counter"}`,
			want: want{code: 404,
				body: []string{"There is no metric you requested\n"}},
		},
	}
	s := NewServer()
	server := httptest.NewServer(s.Router())
	defer server.Close()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log.Println(s.handler.storage)
			resp, body := RunRequest(t, server, tt.method, tt.URL, tt.body, "application/json")
			defer resp.Body.Close()
			assert.Equal(t, tt.want.code, resp.StatusCode)
			for _, s := range tt.want.body {
				assert.Equal(t, body, s)
			}
			assert.Equal(t, tt.want.code, resp.StatusCode)
		})
	}
}

func RunRequest(t *testing.T, ts *httptest.Server, method string, query string, body string, contentType string) (*http.Response, string) {
	reader := strings.NewReader(body)
	req, err := http.NewRequest(method, ts.URL+query, reader)
	req.Header.Add("Content-Type", contentType)

	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	RespBody, err := Decompress(resp.Body, req)
	require.NoError(t, err)
	return resp, string(RespBody)
}
