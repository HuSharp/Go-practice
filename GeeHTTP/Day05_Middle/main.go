package main

import (
	"gee"
	"log"
	"net/http"
	"time"
)

func main() {
	r := gee.New()
	r.Use(gee.Logger())	// global middleware
	r.GET("/index", func(c *gee.Context) {
		c.HTML(http.StatusOK, "<h1>Index Page</h1>")
	})

	v1 := r.Group("/v1")
	{
		v1.GET("/", func(ctx *gee.Context) {
			ctx.HTML(http.StatusOK, "<h1>Hello Hjh!</h1>")
		})
		v1.GET("/hello", func(ctx *gee.Context) {
			// expect /hello?name=hjh
			ctx.String(http.StatusOK, "Hello %s! you're at %s\n", ctx.Query("name"), ctx.Path)
		})
	}
	v2 := r.Group("/v2")
	v2.Use(onlyForV2())
	{
		v2.GET("/hello/:name", func(c *gee.Context) {
			// expect /hello/hjh
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
		})

		v2.GET("/assets/*filepath", func(c *gee.Context) {
			c.JSON(http.StatusOK, gee.H{"filepath": c.Param("filepath")})
		})
	}


	r.Run(":9999")
}

func onlyForV2() gee.HandlerFunc {
	return func(ctx *gee.Context) {
		// start time
		t := time.Now()
		// if a server error occured
		ctx.Fail(http.StatusInternalServerError, "Internal Error")
		// Calculate resolution time
		log.Printf("[%d] %s in %v for GroupV2", ctx.StatusCode, ctx.Req.RequestURI, time.Since(t))
	}
}
