package router

import (
	"github.com/gogf/gf/frame/g"
	"jingyu/app/api/user"
)

func init() {
	// 用户模块 路由注册 - 使用执行对象注册方式
	g.Server().BindObject("/user", new(user.Controller))
}