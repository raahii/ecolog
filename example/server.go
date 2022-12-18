package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/raahii/ecolog"
)

func Hello(c echo.Context) error {
	c.Logger().Infof("This is a log in Hello method.")
	return c.JSON(http.StatusOK, "Hello, World")
}

func main() {
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.RequestID())

	// Use ecolog for contextual logging.
	// See also document of ecolog.AppLoggerConfig.
	e.Use(ecolog.AppLoggerWithConfig(ecolog.AppLoggerConfig{
		Format: `{"time":"${time_rfc3339}","level": "${level}","id":"${id}","remote_ip":"${remote_ip}",` +
			`"host":"${host}","method":"${method}","uri":"${uri}","user_agent":"${user_agent}"}`,
	}))

	e.GET("/", Hello)

	e.Logger.SetLevel(log.INFO)
	e.Logger.Fatal(e.Start(":1323"))
}
