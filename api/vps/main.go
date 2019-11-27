package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/John-Tonny/mnhost/api/vps/handler"
	"github.com/micro/go-micro/web"

	"github.com/John-Tonny/mnhost/common"
	"github.com/John-Tonny/mnhost/conf"
	vpsPb "github.com/John-Tonny/mnhost/interface/out/vps"

	"github.com/gin-gonic/gin"
)

const (
	service    = "apivps"
	vpsService = "vps"
)

var (
	serviceName    string
	vpsServiceName string
)

func init() {
	serviceName = config.GetServiceName(service)
	vpsServiceName = config.GetServiceName(vpsService)
}

func main() {
	srv := common.GetMicroWeb(service, web.Address("0.0.0.0:18889"))

	srvC := common.GetMicroClient(vpsService)
	// 创建 user-service 微服务的客户端
	vpsClient := vpsPb.NewVpsService(vpsServiceName, srvC.Client())

	vpsHandler := handler.GetVpsHandler(vpsClient)

	router := gin.Default()
	router.Use(Cors())

	v1 := router.Group("/api/v1")
	user := v1.Group("/vps")
	user.POST("/new", vpsHandler.NewNode)
	user.POST("/del", vpsHandler.DelNode)
	user.POST("/expand", vpsHandler.ExpandVolume)
	user.POST("/restart", vpsHandler.RestartNode)
	user.POST("/get", vpsHandler.GetAllVps)
	user.POST("/nodeofvps", vpsHandler.GetAllNodeFromVps)
	user.POST("/nodeofuser", vpsHandler.GetAllNodeFromUser)

	srv.Handle("/", router)
	if err := srv.Run(); err != nil {
		log.Fatal(err)
	}
}

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method               //请求方法
		origin := c.Request.Header.Get("Origin") //请求头部
		var headerKeys []string                  // 声明请求头keys
		for k, _ := range c.Request.Header {
			headerKeys = append(headerKeys, k)
		}
		headerStr := strings.Join(headerKeys, ", ")
		if headerStr != "" {
			headerStr = fmt.Sprintf("access-control-allow-origin, access-control-allow-headers, %s", headerStr)
		} else {
			headerStr = "access-control-allow-origin, access-control-allow-headers"
		}
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Origin", "*")                                       // 这是允许访问所有域
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE,UPDATE") //服务器支持的所有跨域请求的方法,为了避免浏览次请求的多次'预检'请求
			//  header的类型
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, X-CSRF-Token, Token,session,X_Requested_With,Accept, Origin, Host, Connection, Accept-Encoding, Accept-Language,DNT, X-CustomHeader, Keep-Alive, User-Agent, X-Requested-With, If-Modified-Since, Cache-Control, Content-Type, Pragma")
			//              允许跨域设置                                                                                                      可以返回其他子段
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers,Cache-Control,Content-Language,Content-Type,Expires,Last-Modified,Pragma,FooBar") // 跨域关键设置 让浏览器可以解析
			c.Header("Access-Control-Max-Age", "172800")                                                                                                                                                           // 缓存请求信息 单位为秒
			c.Header("Access-Control-Allow-Credentials", "false")                                                                                                                                                  //  跨域请求是否需要带cookie信息 默认设置为true
			c.Set("content-type", "application/json")                                                                                                                                                              // 设置返回格式是json
		}

		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.JSON(http.StatusOK, "Options Request!")
		}
		// 处理请求
		c.Next() //  处理请求
	}
}
