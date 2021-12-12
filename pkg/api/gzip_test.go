package api_test

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kyverno/policy-reporter/pkg/api"
	v1 "github.com/kyverno/policy-reporter/pkg/api/v1"
	"github.com/kyverno/policy-reporter/pkg/target"
)

func Test_GzipCompression(t *testing.T) {
	t.Run("GzipRespose", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/targets", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("Accept-Encoding", "gzip")

		rr := httptest.NewRecorder()
		handler := api.Gzip(v1.TargetsHandler(make([]target.Client, 0)))

		handler.ServeHTTP(rr, req)

		reader, _ := gzip.NewReader(rr.Body)
		defer reader.Close()

		buf := new(bytes.Buffer)
		buf.ReadFrom(reader)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := "[]"
		if !strings.Contains(buf.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", buf.String(), expected)
		}
	})
	t.Run("Uncompressed Respose", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/targets", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := api.Gzip(v1.TargetsHandler(make([]target.Client, 0)))

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := "[]"
		if !strings.Contains(rr.Body.String(), expected) {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})

	t.Run("Uncompressed Respose", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/targets", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("Accept-Encoding", "gzip")

		rr := httptest.NewRecorder()
		handler := api.Gzip(func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(204)
		})

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusNoContent {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
	})
}
