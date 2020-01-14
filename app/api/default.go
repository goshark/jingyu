package api

import (
	"database/sql"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/gfile"
	"github.com/gogf/gf/os/glog"
	"github.com/gogf/gf/os/gtime"
	"github.com/gogf/gf/util/gconv"
	"jingyu/library/Queue"
	"jingyu/library/Uuid"
	"jingyu/library/resp"
	"jingyu/library/sms"
	"jingyu/library/voice"
	"strings"
)

/*
	测试推送
*/
func (c *Apis)TestPushList(r *ghttp.Request){
	params := r.GetRequestMap()

	server_type := gconv.Int(params["server_type"])

	if params["uuid"] == ""{
		r.Response.WriteJson("没有UUID")
		r.Exit()
	}
	if server_type == 0 {
		server_type = 0
	}
	server_type = 1

	uuids := strings.Split(gconv.String(params["uuid"]),",")
	msg := Queue.NewMsg()
	msg.To = uuids
	msg.SendType = Queue.TS
	msg.ToServer = 0
	msg.SetContent("{}","order","消息内容")
	Queue.GetInstance().Push(msg)
}


/*
	推送消息到前台
*/
func (c *Apis)SendPushMsg(r *ghttp.Request){
	params,_ := r.GetJson()
	user_uuid :=params.GetString("uuid")
	content := params.GetString("content")
	if  user_uuid == ""{
		r.Response.WriteJson("没有UUID")
		r.Exit()
	}
	uuids := strings.Split(gconv.String(user_uuid),",")
	msg := Queue.NewMsg()
	msg.To = uuids
	msg.SendType = Queue.TS
	msg.ToServer = 0
	msg.SetContent("","",gconv.String(content))
	Queue.GetInstance().Push(msg)
}
/*
	获取登录验证码
*/
func (c *Apis) SendSms(r *ghttp.Request){
	mobile := r.GetString("phone")
	if mobile == ""{
		r.Response.WriteJson(resp.Fail("请输入正确的手机号"))
		r.Exit()
	}
	code_type := r.GetInt("code_type")
	sms := sms.New()
	if code_type <= 0 {
		code_type = 1
	}
	err := sms.Smssend(code_type,mobile)
	if err != nil{
		r.Response.WriteJson(resp.Fail(err.Error()))
	}else{
		r.Response.WriteJson(resp.Succ("短信发送成功"))
	}
}
/*
	绑定用户的客户机信息
*/

func (c *Apis)BindCid(r *ghttp.Request){
	params := r.GetRequestMap()
	dbres,err := c.db().Table("device_bind").Where("user_uuid=?",params["user_uuid"]).One()
	if err != nil && err != sql.ErrNoRows{
		r.Response.WriteJson(resp.Error("绑定设备信息失败！"+err.Error()))
		r.Exit()
	}
	if tx,err := c.db().Begin();err == nil {
		if dbres == nil {
			_,err := tx.Table("device_bind").Data(g.Map{
				"cid":params["cid"],
				"user_uuid":params["user_uuid"],
				"alias":"",
				"server":r.Server.GetName(),
				"tags":"",
				"bind_time":gtime.Datetime(),
			}).Insert()
			if err != nil{
				tx.Rollback()
				r.Response.WriteJson(resp.Error("绑定设备信息失败！"+err.Error()))
				r.Exit()
			}
		}else{
			_,err := tx.Table("device_bind").Where("user_uuid=?",params["user_uuid"]).Data(g.Map{
				"cid":params["cid"],
				"alias":"",
				"tags":"",
			}).Update()
			if err != nil{
				tx.Rollback()
				r.Response.WriteJson(resp.Error("绑定设备信息失败！"+err.Error()))
				r.Exit()
			}

		}
		tx.Commit()
	}
	r.Response.WriteJson(resp.Succ("操作成功！"))
}
/*
	关于我们
*/
func (c *Apis) QueryAboutUs(r *ghttp.Request){
	res,err := c.db().Table("about").FindAll()
	if err != nil && err != sql.ErrNoRows{
		r.Response.WriteJson(resp.Fail("查询失败"+err.Error()))
		r.Exit()
	}
	r.Response.WriteJson(resp.Succ(res.List()))
}
/*
	问题反馈
*/
func (c *Apis) UploadErrorReport(r *ghttp.Request){
	params := r.GetRequestMap()
	content := params["content"]
	user_uuid := params["user_uuid"]
	if content == nil  || content == ""{
		r.Response.WriteJson(resp.Fail("反馈内容不能为空"))
		r.Exit()
	}
	userinfo,err  := c.db().Table("user").Where("uuid=?",user_uuid).One()
	if err != nil && err != sql.ErrNoRows{
		r.Response.WriteJson(resp.Fail("获取用户信息"+err.Error()))
		r.Exit()
	}
	usermap := userinfo.Map()
	if gconv.Int(usermap["type"]) == 0 {
		r.Response.WriteJson(resp.Fail("未知的用户类型"))
		r.Exit()
	}
	e_type := "旅行社"
	if gconv.Int(usermap["type"]) == 2{
		e_type = "酒店"
	}
	if gconv.Int(usermap["e_id"]) == 0{
		 r.Response.WriteJson(resp.NotBind("您还未绑定企业"))
		 r.Exit()
	}
 	data := g.Map{
		"create_time":gtime.Datetime(),
		"uuid":Uuid.Generate_uuid(),
		"createby_uuid":user_uuid,
		"content":content,
		"type":e_type,
		"e_id":usermap["e_id"],
	}
	_,err = c.db().Table("error_report").Data(data).Insert()
	if err != nil {
		r.Response.WriteJson(resp.Fail("反馈操作失败！请稍后再试！"))
		r.Exit()
	}
	r.Response.WriteJson(resp.Fail("操作成功！"))
}

/*
	获取系统配置
*/
func (c *Apis)getSysCfg(key string) string{
	res,_ := c.db().Table("sys_settings").Where("`key`=?",key).FindValue("`value`")
	if res != nil {
		return res.String()
	}
	return ""
}
/*
	生成uuid
*/
func (c *Apis) GenUuid(r *ghttp.Request){
	r.Response.WriteJson(resp.Succ(Uuid.Generate_uuid()))
}
/*
	语音合成
*/
func (c *Apis)Hecheng(r *ghttp.Request){
	params := r.GetRequestMap()
	if params["text"] == ""{
		params["text"] = "这是一条测试信息"
	}
	cr,err := voice.TextToSpeech(params["text"])
	if err != nil {
		r.Response.WriteJson(resp.Fail(err.Error()))
		r.Exit()
	}
	defer cr.Close()
	r.Response.Header().Set("Content-Type",cr.Header.Get("Content-Type"))
	r.Response.SetBuffer(cr.ReadAll())
}

/*
	多文件上传
*/
func UploadPics(r *ghttp.Request,parent_path string,pic_map g.Map,savename ...string){
	SERVER_NAME := r.Server.GetName()
	if SERVER_NAME == ""{
		SERVER_NAME = "Public"
	}
	UPLOAD_ROOT := g.Config().GetString("upload-path")

	if r.MultipartForm == nil || len(r.MultipartForm.File) == 0 {
		glog.Println("未检测到有图片")
		return
	}
	glog.Println(r.MultipartForm.File)

	//设置内存大小
	if err := r.ParseMultipartForm(32 << 20);err != nil {
		glog.Error(err)
		return
	}
	//获取上传的文件组
	filemap := r.MultipartForm.File

	glog.Println(filemap)

	for k,_ := range filemap{
		if gconv.String(pic_map[k]) != ""{
			RemovePic(gconv.String(pic_map[k]))
		}
		files := r.MultipartForm.File[k]
		length := len(files)
		for i := 0; i < length; i++ {
			//打开上传文件
			file, err := files[i].Open()
			defer file.Close()
			if err != nil {
				glog.Error(err)
			}

			buffer := make([]byte, files[i].Size)
			file.Read(buffer)

			//创建上传目录
			if !gfile.Exists(UPLOAD_ROOT){
				gfile.Mkdir(UPLOAD_ROOT)
			}
			pic_path := ""
			if len(savename) > 0 {
				pic_path =	"/"+SERVER_NAME+"/"+ parent_path +"/"+ savename[0]+gfile.Ext(gfile.Basename(files[i].Filename))
			}else{
				pic_path = "/"+SERVER_NAME+"/"+ parent_path +"/"+ gfile.Basename(files[i].Filename)
			}

			file_path := UPLOAD_ROOT+pic_path
			//if gfile.Exists(file_path){
			//	extfile,err := gfile.Open(file_path)
			//	defer extfile.Close()
			//	if err != nil {
			//		glog.Error(err)
			//}
			//	extfileinfo,err := extfile.Stat()
			// if err != nil {
			//	 glog.Error(err)
			// }
			//
			//	if extfileinfo.Size() != files[i].Size{
			//
			//		glog.Println("上传图片：",file_path)
			//	pic_paths = append(pic_paths, pic_path)
			//	gfile.PutBinContents(file_path, buffer)
			//}
			//}else{
			glog.Println("上传图片：",file_path)
			gfile.PutBytes(file_path,buffer)
			pic_map[k] = pic_path

			//}

		}
	}

	return


}