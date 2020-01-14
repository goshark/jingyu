package boot

import (
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/os/glog"
)



// 用于配置初始化.
func initConfig() {
	glog.Info("########service start...")
	c := g.Config()
	// 配置对象
	c.AddPath("config")
	// glog配置
	logPath := c.GetString("log-path")
	glog.SetPath(logPath)
	glog.SetStdoutPrint(true)

	//s_c  s_client客户端
	s_c := g.Server(c.GetString("client-name"))

	//静态目录设置
	if !gfile.Exists("./Uploads/"){
		gfile.Mkdir("./Uploads/")
	}

	s_c.SetServerRoot("./Uploads/")
	s_c.SetIndexFolder(true)
	s_c.AddStaticPath("/Uploads","./Uploads/")
	s_c.SetPort(c.GetInt(s_c.GetName()+".http-port"))
	s_c.SetNameToUriType(ghttp.URI_TYPE_FULLNAME)	//设置struct/method名称与uri的转换方式
	s_c.SetLogPath(logPath+"/server/"+s_c.GetName())
	s_c.SetErrorLogEnabled(true)
	s_c.SetAccessLogEnabled(true)

	//enableHttps(s_c)开启https
	glog.Info("########service finish.")

}

/*
	开启HTTPS
*/
func enableHttps(s ...*ghttp.Server){
	crtfilemap := map[string]string{
		"crtFile":"",
		"keyFile":"",
	}
	if len(s) > 0{
		for _,v := range s{
			v.EnableHTTPS(crtfilemap["crtFile"],crtfilemap["keyFile"])
			glog.Println("https服务已启动！")
		}
	}
}
/*
	启动初始化
*/
func init() {
	initConfig()
	initRouter()
	//开启扫描任务
	scanorder()
}