package ecolog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	NOT_EMPTY = "<not_empty>"
)

func TestLoggerTemplate(t *testing.T) {
	e := echo.New()
	e.Logger.SetLevel(log.INFO)

	// Target endpoint.
	e.POST("/foobar", func(c echo.Context) error {
		c.Logger().Infof("hello")
		return c.String(http.StatusOK, "OK")
	})

	// echo.Context.Logger() uses Echo.Logger as default, so capture output with Echo.Logger.SetOutput
	// - https://github.com/labstack/echo/blob/895121d/context.go#L627
	// - https://github.com/labstack/echo/blob/895121d/echo.go#L350-L360
	buf := new(bytes.Buffer)
	e.Logger.SetOutput(buf)

	e.Use(middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Generator: func() string {
			return "Test Request ID"
		},
		RequestIDHandler: func(c echo.Context, rid string) {
			c.Set("custom", rid)
		},
	}))

	allTags := []string{}

	// Tags suported by gommon
	// https://github.com/labstack/gommon/blob/6267eb7/log/log.go#L372-L388
	gommonTags := map[string]string{
		"time_rfc3339":      NOT_EMPTY,
		"time_rfc3339_nano": NOT_EMPTY,
		"level":             "INFO",
		"prefix":            "echo",
		"long_file":         NOT_EMPTY,
		"short_file":        NOT_EMPTY,
		"line":              NOT_EMPTY,
	}
	for tag := range gommonTags {
		allTags = append(allTags, toTemplate(tag))
	}

	ourTags := map[string]string{
		"id":             "Test Request ID",
		"remote_ip":      NOT_EMPTY,
		"host":           "localhost:80",
		"method":         "POST",
		"uri":            "/foobar?custom=Test+Custom+Query",
		"path":           "/foobar",
		"protocol":       "HTTP/1.1",
		"route":          "/foobar",
		"referer":        "Test Referer",
		"user_agent":     "Test UA",
		"header:Custom":  "Test Custom Header",
		"query:custom":   "Test Custom Query",
		"form:custom":    "Test Custom Form",
		"context:custom": "Test Request ID",
	}

	for tag := range ourTags {
		allTags = append(allTags, toTemplate(tag))
	}

	// Configure our middleware
	e.Use(AppLoggerWithConfig(AppLoggerConfig{
		Format: "{" + strings.Join(allTags, ", ") + "}",
	}))

	req := httptest.NewRequest(http.MethodPost, "/foobar?custom=Test+Custom+Query", nil)
	req.Host = "localhost:80"
	req.Header.Add("User-Agent", "Test UA")
	req.Header.Add("Referer", "Test Referer")
	req.Header.Add("Custom", "Test Custom Header")
	req.Form = url.Values{"custom": []string{"Test Custom Form"}}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	var logMap map[string]string
	logLine := buf.String()
	json.Unmarshal([]byte(logLine), &logMap)
	require.Len(
		t, logMap, len(gommonTags)+len(ourTags)+1,
		"Log line doesn't contain expected number of keys. logLine=%s", logLine,
	)

	assertLog(t, logMap, "message", "hello")
	for k, v := range gommonTags {
		assertLog(t, logMap, k, v)
	}
	for k, v := range ourTags {
		assertLog(t, logMap, k, v)
	}
}

func assertLog(t *testing.T, logs map[string]string, key, expected string) {
	assert.Contains(t, logs, key)

	actual := logs[key]
	if expected == NOT_EMPTY {
		assert.NotEmpty(t, actual, "%s is '%s'", key, actual)
	} else {
		assert.Equal(t, expected, actual, "expected %s is '%s' but '%s'", key, expected, actual)
	}
}

func toTemplate(key string) string {
	return fmt.Sprintf(`"%s": "${%s}"`, key, key)
}
