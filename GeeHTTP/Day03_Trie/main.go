package main

import (
	"gee"
	"net/http"
)

func main() {
	r := gee.New()
	r.GET("/", func(ctx *gee.Context) {
		ctx.HTML(http.StatusOK, "<h1>Hello Hjh</h1>\n")
	})

	r.GET("/hello", func(ctx *gee.Context) {
		// expect /hello?name=hjh
		ctx.String(http.StatusOK, "Hello %s! you're at %s\n", ctx.Query("name"), ctx.Path)
	})

	r.GET("/hello/:name", func(c *gee.Context) {
		// expect /hello/hjh
		c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
	})

	r.GET("/assets/*filepath", func(c *gee.Context) {
		c.JSON(http.StatusOK, gee.H{"filepath": c.Param("filepath")})
	})

	r.Run(":9999")
}

