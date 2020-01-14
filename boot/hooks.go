package boot

import (
	"database/sql"
	"errors"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/glog"
	"github.com/gogf/gf/os/gtime"
	"github.com/gogf/gf/text/gstr"
	"github.com/gogf/gf/util/gconv"
	"jingyu/library/jwtoken"
	"jingyu/library/resp"
)

var (
	//白名单,免鉴权接口
	WhiteList = map[string][]string{
		"client":{
			"Register",
			"Login",
			"SendSms",
			"RefreshToken",
			"Hecheng",//语音合成
			"ResetPassword",//忘记密码重置
			"TestPushList",//测试推送
			"SendPushMsg",
			"GenUuid",
			//"ShareRegister",
		},
	}
)



/*
	客户端鉴权处理
*/
func BeforeClientServe(r *ghttp.Request) {
	r.Response.CORSDefault()
	whitelist := WhiteList["client"]
	uri := r.RequestURI
	//过滤白名单
	if gstr.Contains(uri,c_api_path){
		method := gstr.SubStr(uri,len(c_api_path+"/"))
		for _,v := range whitelist{
			method = gstr.Split(method,"?")[0]
			if gstr.Equal(v,method){
				glog.Println("客户端免鉴权路由",uri)
				return
			}
		}
	}

	glog.Println("请求地址",uri)
	glog.Println("请求数据",r.GetRequestMap())

	if r.GetRequestMap() == nil {
		r.Response.WriteJson(resp.Error("请传入有效参数"))
		r.ExitAll()
	}
	params := r.GetRequestMap()
	if params["access_token"] == "" || params["access_token"] == nil{
		r.Response.WriteJson(resp.Unauthorized("未获取到Access_Token",nil))
		r.ExitAll()
	}

	//解析access_token
	cliams,err := jwtoken.ParseHStoken(gconv.String(params["access_token"]))
	if err != nil{
		r.Response.WriteJson(resp.Unauthorized(err.Error(),nil))
		r.ExitAll()
	}
	if cliams == nil {
		r.Response.WriteJson(resp.Unauthorized("无效的token",nil))
		r.ExitAll()
	}else{
		glog.Println("cliams",cliams)
		//{
		//"auth_id":9,
		//"exp":1577874872,
		//"iat":1577788472,
		//"phone":"13720734735",
		//"uuid":"b2195f8f-6407-4376-af31-911e2fa9b063"
		//}
		//唯一设备登录验证
		//if err := checkOneDevice(g.Map{
		//	"access_token":params["access_token"],
		//	"uuid":cliams["uuid"],
		//});err != nil{
		//	r.Response.WriteJson(resp.Unauthorized(err.Error(),nil))
		//	r.ExitAll()
		//}
		glog.Println("当前操作用户uuid为",cliams["uuid"])
		_,err := g.DB().Table("heartbeat").Save(g.Map{
			"user_uuid":cliams["uuid"],
			"update_time":gtime.Datetime(),
		})
		if err != nil {
			glog.Error("刷新操作时间失败！")
		}
		r.SetForm("user_uuid",cliams["uuid"].(string))
	}

}


/*
	验证是否保持是同一设备登录
	参数:map => uuid,access_token
*/
func checkOneDevice(check_access map[string]interface{}) error {
	if check_access["uuid"] == nil || check_access["access_token"] == nil {
		return errors.New("TOKEN缺少验证参数！")
	}
	ac_res,err := g.DB().Table("tokens").Where("uuid=?",check_access["uuid"]).And("access_token=?",check_access["access_token"]).One()
	if err != nil && err != sql.ErrNoRows{
		return err
	}
	if ac_res == nil {
		return errors.New("你已在其他设备登录,请重新登录")
	}
	return nil
}
