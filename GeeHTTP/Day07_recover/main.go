package main

import (
	"fmt"
	"gee"
	"log"
	"net/http"
	"time"
)

type student struct {
	Name string
	Age int8
}

func FormatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d-%02d-%02d", year, month, day)
}

func main() {
	r := gee.Default()
	r.GET("/", func(ctx *gee.Context) {
		ctx.String(http.StatusOK, "Hello hjh!\n")
	})

	// index out range of testing recovery
	r.GET("/panic", func(ctx *gee.Context) {
		names := []string{"hjh"}
		ctx.String(http.StatusOK, names[100])
	})

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
