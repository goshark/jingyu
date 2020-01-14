package api

import (
	"database/sql"
	"errors"
	"github.com/gogf/gf/crypto/gmd5"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/glog"
	"github.com/gogf/gf/os/gtime"
	"github.com/gogf/gf/util/gconv"
	"jingyu/library/Uuid"
	"jingyu/library/jwtoken"
	"jingyu/library/resp"
	"jingyu/library/sms"
)


/*
	修改个人基本信息
*/
func (c *Apis) UpdateUserBasicInto(r *ghttp.Request){
	params := r.GetRequestMap()
	user_uuid := params["user_uuid"]
	//parent_path := "users/"+gconv.String(user_uuid)+"/head_img"
	//多文件上传处理
	//pic_paths := upload.UploadMore(r,parent_path)

	//构建要插入的数据

	insdata := g.Map{

		"nickname":params["nickname"],
		"sex":gconv.Int(params["sex"]),
		"age":gconv.Int(params["age"]),
	}

	//if len(pic_paths) > 0 {
	//	insdata["head_img"] = strings.Join(pic_paths,";")
	//}
	_,err := c.db().Table("user").Where("uuid=?",user_uuid).Data(insdata).Update()
	if err != nil {
		r.Response.WriteJson(resp.Fail("操作失败！"+err.Error()))
		r.Exit()
	}
	r.Response.WriteJson(resp.Succ("操作成功！"))
}
/*
	商户端Api模板
	名称:注册新用户
	参数:...
*/
func (c *Apis) Register(r *ghttp.Request) {
	params := r.GetRequestMap()
	res,err := c.db().Table("user").Where("phone =?",params["phone"]).Count()
	if err != nil && err != sql.ErrNoRows{
		glog.Error(err)
		r.Response.WriteJson(resp.Fail("查询失败,请稍后再试！"))
		r.Exit()
	}

	if (res > 0){
		r.Response.WriteJson(resp.Fail("当前用户已经存在！"))
		r.Exit()
	}
	if params["reg_type"] == nil {
		r.Response.WriteJson(resp.Fail("请选择用户类型！"))
		r.Exit()
	}
	if gconv.Int(params["reg_type"]) != 1 && gconv.Int(params["reg_type"]) != 2 {
		r.Response.WriteJson(resp.Fail("请选择正确的用户类型！"))
		r.Exit()
	}
	if err := sms.CheckedSmsCode(sms.REG_TYPE,gconv.String(params["phone"]),gconv.String(params["code"]));err != nil {
		r.Response.WriteJson(resp.Fail("验证码错误！"))
		r.Exit()
	}
	_,err = c.db().Table("user").Data(g.Map{
		"phone":params["phone"],
		"type":params["reg_type"],
		"password":gmd5.MustEncrypt(params["password"]),
		"create_time":gtime.Now().Unix(),
		"uuid":Uuid.Generate_uuid(),
	}).Insert()
	if err != nil {
		r.Response.WriteJson(resp.Fail("注册失败,"+err.Error()))
		r.Exit()
	}
	r.Response.WriteJson(resp.Succ("注册成功！"))
}

/*
	商户端Api模板
	名称:在线修改密码
	参数:...
*/
func (c *Apis) ChangePassword(r *ghttp.Request) {
	params := r.GetRequestMap()
	uuid := params["user_uuid"]
	user,err := c.db().Table("user").Where("uuid = ?",uuid).One()
	if err != nil && err != sql.ErrNoRows{
		r.Response.WriteJson(resp.Fail("查询失败,"+err.Error()))
		r.Exit()
	}
	usermap := user.Map()
	if usermap["password"] != gmd5.MustEncrypt(params["oldpassword"]){
		r.Response.WriteJson(resp.Fail("旧密码错误"))
		r.Exit()
	}

	_,err = c.db().Table("user").Where("uuid = ?",uuid).Data(g.Map{
		"password":gmd5.MustEncrypt(params["password"]),
	}).Update()

	if err != nil {
		r.Response.WriteJson(resp.Fail("设置失败,"+err.Error()))
		r.Exit()
	}
	r.Response.WriteJson(resp.Succ("设置成功！"))
}

/*
	商户端Api模板
	名称:忘记密码修改密码
	参数:...
*/
func (c *Apis) ResetPassword(r *ghttp.Request) {
	params := r.GetRequestMap()
	res,err := c.db().Table("user").Where("phone = ?",params["phone"]).Count()
	if err != nil && err != sql.ErrNoRows{
		glog.Error(err)
		r.Response.WriteJson(resp.Fail("查询失败,请稍后再试！"))
		r.Exit()
	}
	if res == 0{
		r.Response.WriteJson(resp.Fail("该用户还未注册！"))
		r.Exit()
	}
	if err := sms.CheckedSmsCode(sms.SETPWD_TYPE,gconv.String(params["phone"]),gconv.String(params["code"]));err != nil {
		r.Response.WriteJson(resp.Fail("验证码错误！"))
		r.Exit()
	}

	_,err = c.db().Table("user").Where("phone = ?",params["phone"]).Data(g.Map{
		"password":gmd5.MustEncrypt(params["password"]),
	}).Update()

	if err != nil {
		r.Response.WriteJson(resp.Fail("设置失败,"+err.Error()))
		r.Exit()
	}
	r.Response.WriteJson(resp.Succ("设置成功！"))
}
/*
	创建token记录
*/
func (c *Apis) createToken(token,uuid,code string) error {
	if code == ""{
		return errors.New("code不能为空")
	}
	_,err := c.db().Table("tokens").Data(g.Map{
		"uuid":uuid,
		"access_token":token,
		"refresh_code":code,
		"refresh_time":gtime.Datetime(),
	}).Save()
	return err
}

/*
	清除token记录
*/
func (c *Apis) removeToken(uuid string) error {
	_,err := c.db().Table("tokens").Where("uuid=?",uuid).Delete()
	return err
}


/*
	用户刷新token
*/
func (c *Apis) RefreshToken(r *ghttp.Request){
	token := r.GetString("access_token")
	code := r.GetString("refresh_code")
	res,err := c.db().Table("tokens").Where("access_token=?",token).And("refresh_code=?",code).One()
	if err != nil && err != sql.ErrNoRows{
		r.Response.WriteJson(resp.BadToken("登录超时,请重新登录！"))
		r.Exit()
	}
	if res == nil {
		r.Response.WriteJson(resp.BadToken("登录超时,请重新登录！"))
		r.Exit()
	}else{
		res,err = c.db().Table("user").Where("uuid=?",res.Map()["uuid"]).One()
		if err != nil && err != sql.ErrNoRows{
			r.Response.WriteJson(resp.BadToken("登录超时,请重新登录！"))
			r.Exit()
		}
		if res == nil {
			r.Response.WriteJson(resp.BadToken("登录超时,请重新登录！"))
			r.Exit()
		}
		resmap := res.Map()
		access_token := jwtoken.GetHStoken(
			map[string]interface{}{
				"phone":resmap["phone"],
				"auth_id":resmap["id"],
				"uuid":resmap["uuid"],
			})
		refresh_code := sms.CreateCaptcha()
		_,err := c.db().Table("tokens").Where("uuid=?",resmap["uuid"]).Data(g.Map{
			"access_token":access_token,
			"refresh_time":gtime.Datetime(),
			"refresh_code":refresh_code,
		}).Update()
		if err != nil {
			glog.Error(err)
			r.Response.WriteJson(resp.BadToken("登录超时,请重新登录！"))
		}else{
			r.Response.WriteJson(resp.Succ(g.Map{
				"access_token":access_token,
				"refresh_code":refresh_code,
			}))
		}

	}

}

/*
	客户端Api模板
	名称:客户端用户登录
	参数:手机号或者用户名
*/
func (c *Apis) Login(r *ghttp.Request) {
	params := r.GetRequestMap()
	glog.Println(params)
	if params == nil {
		r.Response.WriteJson(resp.Fail("参数无效"))
		r.Exit()
	}

	useres,err := c.db().Table("user").Where("phone=?",params["phone"]).One()
	if err != nil && err != sql.ErrNoRows{
		r.Response.WriteJson(resp.Fail("查询出错,请稍后再试！"))
		r.Exit()
	}
	uuid := ""//用户uuid
	if useres == nil{
		//查无此人,新用户
		r.Response.WriteJson(resp.Fail("该手机号暂未注册！"))
		r.Exit()
	}
	user_info := useres.Map()
	uuid = gconv.String(user_info["uuid"])
	if user_info["password"] != gmd5.MustEncrypt(params["password"]){
		r.Response.WriteJson(resp.Fail("登录密码错误！"))
		r.Exit()
	}
	//生成token
	access_token := jwtoken.GetHStoken(g.Map{
		"phone":params["phone"],
		"auth_id":user_info["id"],
		"uuid":user_info["uuid"],
	})
	if access_token == ""{
		r.Response.WriteJson(resp.Fail("token生成失败！"))
		r.Exit()
	}
	code := sms.CreateCaptcha()
	if err := c.createToken(access_token,uuid,code);err != nil {
		r.Response.WriteJson(resp.Fail(err.Error()))
		r.Exit()
	}
	//绑定用户的客户机信息
	r.Response.WriteJson(resp.Succ(g.Map{
		"access_token":access_token,
		"uuid":uuid,
		"refresh_code":code,
	}))

}


/*
	获取用户信息
*/
func (c *Apis)GetUserInfo(r *ghttp.Request){
	params := r.GetRequestMap()
	user,err := c.db().Table("user as u").Where("uuid = ?",params["user_uuid"]).One()
	if err != nil && err != sql.ErrNoRows{
		r.Response.WriteJson(resp.Fail("查询失败"+err.Error()))
		r.Exit()
	}
	usermap := user.Map()
	glog.Println(usermap)
	if gconv.Int(usermap["type"]) == 0 || gconv.Int(usermap["type"]) > 2{
		r.Response.WriteJson(resp.Fail("未知的用户类型！"))
		r.Exit()
	}
	res := g.Map{}
	res["userinfo"] =usermap
	table := "lv_travel"
	if gconv.Int(usermap["type"]) == 1{
		table = "lv_travel"
	}else{
		table = "jiu_hotel"
	}
	res["company"] = nil
	if gconv.Int(usermap["e_id"]) > 0 {
		e_info,err := c.db().Table(table).Where("id=?",usermap["e_id"]).One()
		if err != nil && err != sql.ErrNoRows{
			r.Response.WriteJson(resp.Fail("查询失败"+err.Error()))
			r.Exit()
		}
		if e_info != nil {
			res["company"] = e_info.Map()
		}

	}

	r.Response.WriteJson(resp.Succ(res))
}

