package server

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	_ "net/http/pprof"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	filestorage "github.com/AbramovArseniy/YandexRuntimeMetrics/internal/server/fileStorage"
)

func Example() {
	fs := filestorage.NewFileStorage("/tmp/metrics-example.json", 5*time.Second, false)
	fs.SetFileStorage()
	s := NewServer("locashost:8080", false, fs, nil, "", "")
	handler := DecompressHandler(s.Router())
	handler = CompressHandler(handler)
	server := httptest.NewServer(handler)
	defer server.Close()
	requests := []struct {
		Name   string
		URL    string
		Method string
		Body   string
	}{
		{
			Name:   "Post gauge",
			URL:    server.URL + "/update/gauge/Alloc/200.10",
			Method: http.MethodPost,
		},
		{
			Name:   "Post counter",
			URL:    server.URL + "/update/counter/PollCount/5",
			Method: http.MethodPost,
		},
		{
			Name:   "Get gauge",
			URL:    server.URL + "/value/gauge/Alloc",
			Method: http.MethodGet,
		},
		{
			Name:   "Get counter",
			URL:    server.URL + "/value/counter/PollCount",
			Method: http.MethodGet,
		},
		{
			Name:   "Get all metrics",
			URL:    server.URL + "/",
			Method: http.MethodGet,
		},
		{
			Name:   "Post gauge",
			URL:    server.URL + "/update/",
			Method: http.MethodPost,
			Body: `{
				"id":"Alloc",
				"type":"gauge",
				"value":400
			}`,
		},
		{
			Name:   "Post counter",
			URL:    server.URL + "/update/counter/PollCount/5",
			Method: http.MethodPost,
			Body: `{
				"id":"PollCount",
				"type":"counter",
				"value":100
			}`,
		},
		{
			Name:   "Get counter",
			URL:    server.URL + "/value/",
			Method: http.MethodPost,
			Body: `{
				"id":"PollCount",
				"type":"counter"
			}`,
		},
		{
			Name:   "Get gauge",
			URL:    server.URL + "/value/",
			Method: http.MethodPost,
			Body: `{
				"id":"Alloc",
				"type":"gauge"
			}`,
		},
		{
			Name:   "Get all metrics",
			URL:    server.URL + "/",
			Method: http.MethodGet,
		},
	}
	for _, v := range requests {
		if v.Method == http.MethodPost {
			rdr := strings.NewReader(v.Body)
			resp, err := http.DefaultClient.Post(v.URL, "application/json", rdr)
			if err != nil {
				fmt.Println("error while getting response from server", err)
				return
			}
			body, _ := io.ReadAll(resp.Body)
			stringBody := string(body)
			fmt.Println(stringBody)
			resp.Body.Close()
		} else {
			resp, err := http.DefaultClient.Get(v.URL)
			if err != nil {
				fmt.Println("error while getting response from server", err)
				return
			}
			body, _ := io.ReadAll(resp.Body)
			stringBody := string(body)
			fmt.Println(stringBody)
			resp.Body.Close()
		}
	}
	// Output:
	// 	200.1
	// 5
	// PollCount: 5
	// Alloc: 200.100000
	//
	// {"id":"Alloc","type":"gauge","value":400}
	//
	// {"id":"PollCount","type":"counter","delta":10}
	// {"id":"Alloc","type":"gauge","value":400}
	// PollCount: 10
	// Alloc: 400.000000
}

// BenchmarkTextPlainMetricHandler benchmark for handlers with Content-Type 'text/plain'
func BenchmarkTextPlainMetricHandler(b *testing.B) {
	fs := filestorage.NewFileStorage("/tmp/metrics-test.json", 5*time.Second, false)
	fs.SetFileStorage()
	s := NewServer("locashost:8080", false, fs, nil, "", "")
	handler := DecompressHandler(s.Router())
	handler = CompressHandler(handler)
	server := httptest.NewServer(handler)
	defer server.Close()
	requests := []struct {
		Name   string
		URL    string
		Method string
	}{
		{
			Name:   "Post gauge",
			URL:    server.URL + "/update/gauge/Alloc/200.10",
			Method: http.MethodPost,
		},
		{
			Name:   "Post counter",
			URL:    server.URL + "/update/counter/PollCount/5",
			Method: http.MethodPost,
		},
		{
			Name:   "Get counter",
			URL:    server.URL + "/value/counter/PollCount",
			Method: http.MethodGet,
		},
		{
			Name:   "Get gauge",
			URL:    server.URL + "/value/gauge/Alloc",
			Method: http.MethodGet,
		},
	}
	for _, v := range requests {
		b.Run(v.Name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if v.Method == http.MethodPost {
					rdr := strings.NewReader("")
					resp, _ := http.DefaultClient.Post(v.URL, "text/plain", rdr)
					resp.Body.Close()
				} else {
					resp, _ := http.DefaultClient.Get(v.URL)
					resp.Body.Close()
				}
			}
		})
	}
}

// BenchmarkJSONMetricHandler benchmark for handlers with Content-Type 'application/json'
func BenchmarkJSONMetricHandler(b *testing.B) {
	fs := filestorage.NewFileStorage("/tmp/metrics-test.json", 5*time.Second, false)
	fs.SetFileStorage()
	s := NewServer("http://locashost:8080", false, fs, nil, "", "")
	handler := DecompressHandler(s.Router())
	handler = CompressHandler(handler)
	server := httptest.NewServer(handler)
	defer server.Close()
	requests := []struct {
		Name string
		URL  string
		Body string
	}{
		{
			Name: "Post gauge",
			URL:  server.URL + "/update/",
			Body: `{
				"id":"Alloc",
				"type":"gauge",
				"value":400
			}`,
		},
		{
			Name: "Post counter",
			URL:  server.URL + "/update/counter/PollCount/5",
			Body: `{
				"id":"Counter",
				"type":"counter",
				"value":100
			}`,
		},
		{
			Name: "Get counter",
			URL:  server.URL + "/value/counter/PollCount",
			Body: `{
				"id":"Counter",
				"type":"counter"
			}`,
		},
		{
			Name: "Get gauge",
			URL:  server.URL + "/value/gauge/Alloc",
			Body: `{
				"id":"Alloc",
				"type":"gauge",
			}`,
		},
	}
	for _, v := range requests {
		b.Run(v.Name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				rdr := strings.NewReader(v.Body)
				resp, _ := http.DefaultClient.Post(v.URL, "text/plain", rdr)
				resp.Body.Close()
			}
		})
	}
}

// TestPlainTextHandler tests handlers with Content-Type 'text/plain'
func TestPlainTextHandler(t *testing.T) {
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
			want:   want{code: 404},
		},
		{
			name:   "404 no such counter",
			URL:    "/value/counter/adadadad",
			method: http.MethodGet,
			want:   want{code: 404},
		},
	}
	fs := filestorage.NewFileStorage("/tmp/devops-metrics-db.json", 5*time.Second, false)
	s := NewServer("locashost:8080", false, fs, nil, "", "")
	handler := DecompressHandler(s.Router())
	handler = CompressHandler(handler)
	server := httptest.NewServer(handler)
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

// TestJSONHandlers tests handlers with Content-Type 'application/json'
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
			name:   "501 JSON post: no metric value in request",
			URL:    "/update/",
			method: http.MethodPost,
			body:   `{"id":"name","type":"gauge"}`,
			want:   want{code: 501},
		},
		{
			name:   "501 JSON post: wrong metric value type in request",
			URL:    "/update/",
			method: http.MethodPost,
			body:   `{"id":"name","type":"gauge","delta":200}`,
			want:   want{code: 501},
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
			want:   want{code: 404},
		},
		{
			name:   "404 JSON get no such counter",
			URL:    "/value/",
			method: http.MethodPost,
			body:   `{"id":"asdasda","type":"counter"}`,
			want:   want{code: 404},
		},
	}
	fs := filestorage.NewFileStorage("/tmp/devops-metrics-db.json", 5*time.Second, false)
	s := NewServer("locashost:8080", false, fs, nil, "", "")
	handler := DecompressHandler(s.Router())
	handler = CompressHandler(handler)
	server := httptest.NewServer(handler)
	defer server.Close()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

// RunRequest does request to a server
func RunRequest(t *testing.T, ts *httptest.Server, method string, query string, body string, contentType string) (*http.Response, string) {
	reader := strings.NewReader(body)
	req, err := http.NewRequest(method, ts.URL+query, reader)
	req.Header.Add("Content-Type", contentType)

	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	RespBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp, string(RespBody)
}
