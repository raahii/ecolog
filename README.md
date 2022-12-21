# ecolog 

[![Go package](https://github.com/raahii/ecolog/actions/workflows/test.yml/badge.svg)](https://github.com/raahii/ecolog/actions/workflows/test.yml)


Ecolog provides a middleware for [Go Echo framework](https://echo.labstack.com/) to output application logs with request context.

By using ecolog and Echo's standard gommon logging, You can add fields related to HTTP request such as method, URI, Request ID, etc.
```json
{
  "time": "2022-12-18T22:22:21+09:00",
  "level": "INFO",
  "id": "5aWJKbZ1hEDYyfwhidOnUcD7zRyYHaIa",
  "remote_ip": "127.0.0.1",
  "host": "localhost:1323",
  "method": "GET",
  "uri": "/",
  "user_agent": "curl/7.79.1",
  "message": "This is a log in Hello method."
}
```



## Installation

```shell
go get -u github.com/raahii/ecolog
```




## Example

1. Let's use ecolog middleware to override log format.

  ```go
  func main() {
    e := echo.New()
    
    // Use ecolog middleware.
    // See also ecolog.AppLoggerConfig doc.
    e.Use(ecolog.AppLoggerWithConfig(ecolog.AppLoggerConfig{
      Format: `{"time":"${time_rfc3339}","level": "${level}",id":"${id}","remote_ip":"${remote_ip}",` +
        `"host":"${host}","method":"${method}","uri":"${uri}","user_agent":"${user_agent}"}`,
    }))
    
    ...
  }
  ```



2. Define an endpoint, and output application log with `echo.Context.Logger()` in your handler.

  ```go
  func Hello(c echo.Context) error {
    c.Logger().Infof("This is a log in Hello method.")
    return c.JSON(http.StatusOK, "Hello, World")
  }

  func main() {
    ...
    e.GET("/", Hello)
    ...
  }
  ```



3. Then, we can observe the application log with the request context.

  ```shell
  ❯ go run example/server.go
  ⇨ http server started on [::]:1323
  {"time":"2022-12-18T22:22:21+09:00","level": "INFO","id":"5aWJKbZ1hEDYyfwhidOnUcD7zRyYHaIa","remote_ip":"127.0.0.1","host":"localhost:1323","method":"GET","uri":"/","user_agent":"curl/7.79.1","message":"This is a log in Hello method."}
  ```

See [example/server.go](https://github.com/raahii/ecolog/blob/main/example/server.go) for details.

