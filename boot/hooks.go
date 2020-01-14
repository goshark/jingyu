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
	商户端鉴权处理
*/
//func BeforeBusinessServe(r *ghttp.Request) {
//	r.Response.CORSDefault()
//	//商户端action白名单
//	whitelist := WhiteList["business"]
//	uri := r.RequestURI
//	//过滤白名单
//	if gstr.Contains(uri,b_api_path){
//		method := gstr.SubStr(uri,len(b_api_path+"/"))
//		for _,v := range whitelist{
//			if gstr.Equal(v,method){
//				glog.Println("商户端免鉴权路由",uri)
//				return
//			}
//		}
//	}
//	glog.Println("请求地址",uri)
//	glog.Println("请求数据",r.GetRequestMap())
//
//	if r.GetRequestMap() == nil {
//		r.Response.WriteJson(resp.Fail("请传入有效参数"))
//		r.ExitAll()
//	}
//	params := r.GetRequestMap()
//	if params["access_token"] == ""{
//		r.Response.WriteJson(resp.Fail("未获取到Access_Token"))
//		r.ExitAll()
//	}
//
//	//解析access_token
//	cliams,err := jwtoken.ParseHStoken(params["access_token"])
//	if err != nil{
//		r.Response.WriteJson(resp.Fail(err.Error()))
//		r.ExitAll()
//	}
//	if cliams == nil {
//		r.Response.WriteJson(resp.Fail("无效的token,请重新登录！"))
//		r.ExitAll()
//	}else{
//
//		//唯一设备登录验证
//		if err := checkOneDevice(g.Map{
//			"access_token":params["access_token"],
//			"uuid":cliams["uuid"],
//		});err != nil{
//			r.Response.WriteJson(resp.Fail(err.Error()))
//			r.ExitAll()
//		}
//		glog.Println("当前操作用户uuid为",cliams["uuid"])
//		r.SetForm("user_uuid",cliams["uuid"].(string))
//
//		if cliams["hotel_uuid"] == nil || cliams["hotel_uuid"].(string) == "" {
//			r.Response.WriteJson(resp.Fail("你还没有绑定酒店信息,请联系酒店负责人或后台客服"))
//			r.ExitAll()
//		}
//
//		//查看该员工所在酒店是否通过审核;
//		dbres,err := g.DB().Table("b_license as l,b_hotel as h").Where("h.license_uuid = l.uuid").And("h.uuid = ?",cliams["hotel_uuid"]).One()
//		if err != nil && err != sql.ErrNoRows{
//			r.Response.WriteJson(resp.Fail("查询出错"+err.Error()))
//			r.ExitAll()
//		}
//		if dbres == nil {
//			r.Response.WriteJson(resp.Fail("暂未查询到该商户信息！"))
//			r.ExitAll()
//		}
//		jiudianmap := dbres.ToMap()
//		if jiudianmap["is_upload"].(int) == 0{
//			r.Response.WriteJson(resp.Resp{Code:-3,Msg:"该商户还未进行商户认证,请前往认证",Data:""})
//			r.ExitAll()
//		}
//		if jiudianmap["checked_result"].(int) != 1{
//			r.Response.WriteJson(resp.Fail("该商户认证还未通过审核!"))
//			r.ExitAll()
//		}
//		r.SetForm("hotel_uuid",cliams["hotel_uuid"].(string))
//		//r.SetParam("user_uuid",cliams["uuid"])
//	}
//
//}


/*
	后台PC鉴权
*/
//func BeforeAdminServe(r *ghttp.Request){
//	glog.Println("当前URI",r.RequestURI)
//	r.Response.CORS(ghttp.CORSOptions{
//		AllowOrigin:r.Request.Header.Get("origin"),
//		AllowCredentials:"true",
//		AllowMethods:     ghttp.HTTP_METHODS,
//		MaxAge:           3628800,
//	})
//
//	//管理端action白名单
//	whitelist := WhiteList["admin"]
//	uri := r.RequestURI
//	//过滤白名单
//	if gstr.Contains(uri,a_api_path){
//		method := gstr.SubStr(uri,len(a_api_path+"/"))
//		for _,v := range whitelist{
//			if gstr.Equal(v,method){
//				glog.Println("管理平台免鉴权路由",uri)
//				return
//			}
//		}
//	}
//	//glog.Println(r.Session.IsDirty(),r.Session.Size())
//	if r.Session.Size() == 0 || r.Session.IsDirty() {
//		r.Response.WriteJson(resp.Fail("登录超时,请重新登录！"))
//		r.Session.Clear()
//		r.ExitAll()
//	}
//}

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
