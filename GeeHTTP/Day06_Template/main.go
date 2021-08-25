package main

import (
	"fmt"
	"gee"
	"html/template"
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
	r := gee.New()
	r.Use(gee.Logger())	// global middleware
	r.SetFuncMap(template.FuncMap{
		"FormatAsDate":	FormatAsDate,
	})
	r.LoadHTMLGlob("templates/*")
	r.Static("/assets", "./static")

	stu1 := &student{
		Name: "hjh",
		Age:  20,
	}
	stu2 := &student{
		Name: "hjh2",
		Age:  21,
	}
	r.GET("/", func(ctx *gee.Context) {
		ctx.HTML(http.StatusOK, "css.tmpl", nil)
	})
	r.GET("/stu", func(ctx *gee.Context) {
		ctx.HTML(http.StatusOK, "arr.tmpl", gee.H{
			"title":	"gee",
			"stuArr":	[2]*student{stu1, stu2},
		})
	})
	r.GET("/date", func(c *gee.Context) {
		c.HTML(http.StatusOK, "custom_func.tmpl", gee.H{
			"title": "gee",
			"now":   time.Date(2021, 5, 2, 0, 0, 0, 0, time.UTC),
		})
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
