package boot

import (
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"jingyu/app/api"
)

const (
	c_api_path = "/api/client"
)




/*
绑定客户端路由
*/
func bindClientRouter() {
	urlPath := g.Config().GetString("client.url-path")
	s_c := g.Server(g.Config().GetString("client-name"))
	// 鉴权处理
	s_c.BindHookHandler(urlPath+c_api_path+"/*any", ghttp.HOOK_BEFORE_SERVE, BeforeClientServe)
	//商户端api路由注册
	client_api := new(api.Apis)
	s_c.BindObject(urlPath+c_api_path,client_api)
	s_c.SetLogStdout(true)

}






/*
	统一路由注册
*/
func initRouter() {
	bindClientRouter()
}




