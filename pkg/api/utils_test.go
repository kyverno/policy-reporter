package api_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/kyverno/policy-reporter/pkg/api"
	db "github.com/kyverno/policy-reporter/pkg/database"
)

func TestSendResponseSuccess(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	server := api.NewServer(engine, api.WithRecovery())

	engine.GET("/send", func(ctx *gin.Context) {
		api.SendResponse(ctx, "data", "", nil)
	})

	req, _ := http.NewRequest("GET", "/send", nil)
	w := httptest.NewRecorder()

	server.Serve(w, req)

	assert := assert.New(t)
	assert.Equal(http.StatusOK, w.Code)
	assert.Equal(`"data"`, string(w.Body.Bytes()))
}

func TestSendResponseError(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	server := api.NewServer(engine, api.WithRecovery())

	engine.GET("/send", func(ctx *gin.Context) {
		api.SendResponse(ctx, nil, "errorMsg", errors.New("error"))
	})

	req, _ := http.NewRequest("GET", "/send", nil)
	w := httptest.NewRecorder()

	server.Serve(w, req)

	assert := assert.New(t)
	assert.Equal(http.StatusInternalServerError, w.Code)
	assert.Equal("", string(w.Body.Bytes()))
}

func TestBuildFilter(t *testing.T) {
	filter := api.BuildFilter(&gin.Context{
		Request: &http.Request{
			URL: &url.URL{
				RawQuery: "labels=env:test&labels=app:nginx&labels=invalid&exclude=kyverno:Pod&exclude=kyverno:Job&exclude=kyverno&status=pass&namespaced=true",
			},
		},
	})

	assert := assert.New(t)
	assert.Equal(db.Filter{
		ReportLabel: map[string]string{
			"env": "test",
			"app": "nginx",
		},
		Exclude: map[string][]string{
			"kyverno": {"Pod", "Job"},
		},
		Status:     []string{"pass"},
		Namespaced: true,
	}, filter)
}

func TestPaginationDefaults(t *testing.T) {
	pagination := api.BuildPagination(&gin.Context{
		Request: &http.Request{
			URL: &url.URL{
				RawQuery: "",
			},
		},
	}, []string{"namespace", "source"})

	assert := assert.New(t)
	assert.Equal(db.Pagination{
		Page:      0,
		Offset:    0,
		SortBy:    []string{"namespace", "source"},
		Direction: "ASC",
	}, pagination)
}

func TestPaginationFromURL(t *testing.T) {
	pagination := api.BuildPagination(&gin.Context{
		Request: &http.Request{
			URL: &url.URL{
				RawQuery: "page=5&offset=10&direction=desc&sortBy=namespace&sortBy=kind",
			},
		},
	}, []string{"namespace", "source"})

	assert := assert.New(t)
	assert.Equal(db.Pagination{
		Page:      5,
		Offset:    10,
		SortBy:    []string{"namespace", "kind"},
		Direction: "DESC",
	}, pagination)
}
