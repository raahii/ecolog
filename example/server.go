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
		Format: `{"time":"${time_rfc3339}","id":"${id}","remote_ip":"${remote_ip}",` +
			`"host":"${host}","method":"${method}","uri":"${uri}","user_agent":"${user_agent}"}`,
	}))

	e.GET("/", Hello)

	e.Logger.SetLevel(log.INFO)
	e.Logger.Fatal(e.Start(":1323"))

	// $ curl http://localhost:1323
	// {"time":"2022-12-18T20:00:22+09:00","id":"WNCQfBBFKh7dxl3t2mhN87INcXTc7uhg","remote_ip":"127.0.0.1","host":"localhost:1323","method":"GET","uri":"/","user_agent":"curl/7.79.1","message":"This is a log in Hello method."}
}
