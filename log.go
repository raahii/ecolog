package ecolog

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/valyala/fasttemplate"
)

// Almost codes are copied from official logging middleware.
// See also: https://github.com/labstack/echo/blob/abecadc/middleware/logger.go

type AppLoggerConfig struct {
	// Tags to construct the logger format.
	// - time_rfc3339
	// - time_rfc3339_nano
	// - time_custom
	// - level
	// - prefix
	// - long_file
	// - short_file
	// - line
	// - id (Request ID)
	// - remote_ip
	// - host
	// - method
	// - uri
	// - path
	// - protocol
	// - route
	// - referer
	// - user_agent
	// - header:<NAME>
	// - query:<NAME>
	// - form:<NAME>
	//
	// Example "${remote_ip} ${status}"
	//
	// Optional. Default value DefaultLoggerConfig.Format.
	Format string `yaml:"format"`

	// Optional. Default value DefaultLoggerConfig.CustomTimeFormat.
	CustomTimeFormat string `yaml:"custom_time_format"`

	pool     *sync.Pool
	template *fasttemplate.Template
}

var (
	// DefaultLoggerConfig is the default Logger middleware config.
	DefaultLoggerConfig = AppLoggerConfig{
		Format: `{"time":"${time_rfc3339_nano}","id":"${id}","remote_ip":"${remote_ip}",` +
			`"host":"${host}","method":"${method}","uri":"${uri}","user_agent":"${user_agent}"}`,
		CustomTimeFormat: "2006-01-02 15:04:05.00000",
	}
)

// Logger returns a middleware that outputs application logs with request info.
func AppLogger() echo.MiddlewareFunc {
	return AppLoggerWithConfig(DefaultLoggerConfig)
}

// AppLoggerWithConfig returns a AppLogger middleware with config.
// See: `AppLogger()`.
func AppLoggerWithConfig(config AppLoggerConfig) echo.MiddlewareFunc {
	if config.Format == "" {
		config.Format = DefaultLoggerConfig.Format
	}

	config.template = fasttemplate.New(config.Format, "${", "}")
	config.pool = &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 256))
		},
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			req := c.Request()
			res := c.Response()

			buf := config.pool.Get().(*bytes.Buffer)
			buf.Reset()
			defer config.pool.Put(buf)

			if _, err = config.template.ExecuteFunc(buf, func(w io.Writer, tag string) (int, error) {
				switch tag {
				case "time_custom":
					return buf.WriteString(time.Now().Format(config.CustomTimeFormat))
				case "id":
					return buf.WriteString(res.Header().Get(echo.HeaderXRequestID))
				case "remote_ip":
					return buf.WriteString(c.RealIP())
				case "host":
					return buf.WriteString(req.Host)
				case "uri":
					return buf.WriteString(req.RequestURI)
				case "method":
					return buf.WriteString(req.Method)
				case "path":
					p := req.URL.Path
					if p == "" {
						p = "/"
					}
					return buf.WriteString(p)
				case "route":
					return buf.WriteString(c.Path())
				case "protocol":
					return buf.WriteString(req.Proto)
				case "referer":
					return buf.WriteString(req.Referer())
				case "user_agent":
					return buf.WriteString(req.UserAgent())
				default:
					switch {
					case strings.HasPrefix(tag, "context:"):
						return buf.WriteString(fmt.Sprintf("%s", c.Get(tag[8:])))
					case strings.HasPrefix(tag, "header:"):
						return buf.Write([]byte(c.Request().Header.Get(tag[7:])))
					case strings.HasPrefix(tag, "query:"):
						return buf.Write([]byte(c.QueryParam(tag[6:])))
					case strings.HasPrefix(tag, "form:"):
						return buf.Write([]byte(c.FormValue(tag[5:])))
					case strings.HasPrefix(tag, "cookie:"):
						cookie, err := c.Cookie(tag[7:])
						if err == nil {
							return buf.Write([]byte(cookie.Value))
						}
					}
					// Undo unsupported tags because they are handled by gommon.
					return buf.WriteString(fmt.Sprintf("${%s}", tag))
				}
			}); err != nil {
				return
			}

			c.Logger().SetHeader(string(buf.Bytes()))
			return next(c)
		}
	}
}
