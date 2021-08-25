package Day01

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	r := gin.Default()
	r.GET("/", func(context *gin.Context) {
		context.String(200, "Hello!\n")
	})

	r.GET("/usr/:name", func(context *gin.Context) {
		name := context.Param("name")
		context.String(http.StatusOK, "Hello %s\n", name)
	})

	// 匹配users?name=xxx&role=xxx，role可选
	r.GET("/users", func(context *gin.Context) {
		name := context.Query("name")
		role := context.DefaultQuery("role", "teacher")
		context.String(http.StatusOK, "%s is a %s", name, role)
	})

	// curl http://localhost:9999/form  -X POST -d 'username=geektutu&password=1234'
	//{"password":"1234","username":"geektutu"}
	r.POST("form", func(context *gin.Context) {
		username := context.PostForm("username")
		password := context.PostForm("password")

		context.JSON(http.StatusOK, gin.H{
			"username": username,
			"password": password,
		})
	})

	// POST 混合 Query
	//$ curl "http://localhost:9999/posts?id=9876&page=7"  -X POST -d 'username=geektutu&password=1234'
	//{"id":"9876","page":"7","password":"1234","username":"geektutu"}
	r.POST("/posts", func(context *gin.Context) {
		id := context.Query("id")
		page := context.DefaultQuery("page", "0")
		username := context.PostForm("username")
		password := context.PostForm("password")

		context.JSON(http.StatusOK, gin.H{
			"id":       id,
			"page":     page,
			"username": username,
			"password": password,
		})
	})
	// Map 参数
	//$ curl -g "http://localhost:9999/postMap?ids[Jack]=001&ids[Tom]=002" -X POST -d 'names[a]=Sam&names[b]=David'
	//{"ids":{"Jack":"001","Tom":"002"},"names":{"a":"Sam","b":"David"}}
	r.POST("postMap", func(context *gin.Context) {
		ids := context.QueryMap("ids")
		names := context.PostFormMap("names")

		context.JSON(http.StatusOK, gin.H{
			"ids":   ids,
			"names": names,
		})
	})

	// 重定向
	// -i参数打印出服务器回应的 HTTP 标头。
	r.GET("/redirect", func(context *gin.Context) {
		context.Redirect(http.StatusMovedPermanently, "/index")
	})

	r.GET("goindex", func(context *gin.Context) {
		context.Request.URL.Path = "/"
		r.HandleContext(context)
	})

	// 分组路由
	defaultHandler := func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"path": c.FullPath(),
		})
	}

	// group v1
	v1 := r.Group("/v1")
	{
		v1.GET("/posts", defaultHandler)
		v1.GET("series", defaultHandler)
	}
	v2 := r.Group("/v2")
	{
		v2.GET("/posts", defaultHandler)
		v2.GET("series", defaultHandler)
	}

	// 长传文件
	r.POST("/upload1", func(context *gin.Context) {
		file, _ := context.FormFile("file")
		context.String(http.StatusOK, "%s upload!", file.Filename)
	})

	r.Run(":9999")
}
