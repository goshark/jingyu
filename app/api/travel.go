package api

import (
	"database/sql"
	"github.com/gogf/gf/encoding/gjson"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/glog"
	"github.com/gogf/gf/os/gtime"
	"github.com/gogf/gf/util/gconv"
	"jingyu/library/Uuid"
	"jingyu/library/resp"
)



/*
	修改旅行社基本信息
*/
func (c *Apis) UpdateTravelBasicInfo(r *ghttp.Request){
	params := r.GetRequestMap()
	user_uuid := params["user_uuid"]
	source := getTravelByUuid(user_uuid)
	if source == nil {
		r.Response.WriteJson(resp.Fail("上传酒店照片失败！未获取到企业信息"))
		r.Exit()
	}

	pics := gjson.New(source["pics"]).ToMap()
	if pics == nil {
		pics = g.Map{}
	}
	parent_path := "hotels/"+gconv.String(source["uuid"])+"/pics"
	UploadPics(r,parent_path,pics)
	data_map := g.Map{

		"name":params["name"],
		"city":params["city"],
		"address":params["address"],
		"tel":params["tel"],
		"pics":gjson.New(pics).MustToJsonString(),
	}
	tx,err := c.db().Begin()
	if err != nil {
		r.Response.WriteJson(resp.Fail("操作异常,请稍后再试！"+err.Error()))
		r.Exit()
	}
	_,err = tx.Table("lv_travel").Where("uuid=?",source["uuid"]).Data(data_map).Filter().Update()
	if err != nil {
		tx.Rollback()
		r.Response.WriteJson(resp.Fail("上传基本信息失败！"))
		r.Exit()
	}

	tx.Commit()
	r.Response.WriteJson(resp.Succ("上传基本信息成功！"))
}
/*
	上传旅行社基本信息
*/
func (c *Apis) UploadTravelBasicInfo(r *ghttp.Request){
	params := r.GetRequestMap()
	//查重
	conditions := g.Map{
		"name":params["name"],
		"city":params["city"],
		"address":params["address"],
		"star":params["star"],
	}
	travels,err := c.db().Table("jiu_hotel").Where(conditions).FindAll()
	if err != nil && err != sql.ErrNoRows{
		r.Response.WriteJson(resp.Fail("查询失败"+err.Error()))
		r.Exit()
	}
	if travels != nil && len(travels.List()) > 0{
		r.Response.WriteJson(resp.Fail("该旅行社信息已存在！"))
		r.Exit()
	}

	uuid := Uuid.Generate_uuid()
	source := getTravelByUuid(params["user_uuid"])
	if source != nil {
		r.Response.WriteJson(resp.Fail("企业基本信息已上传,不能重复上传!"))
		r.Exit()
	}
	pics := gjson.New(source["pics"]).ToMap()
	if pics == nil {
		pics = g.Map{}
	}
	parent_path := "hotels/"+uuid+"/pics"
	UploadPics(r,parent_path,pics)
	data_map := g.Map{
		"uuid":uuid,
		"name":params["name"],
		"city":params["city"],
		"address":params["address"],
		"tel":params["tel"],
		"create_time":gtime.Datetime(),
		"pics":gjson.New(pics).MustToJsonString(),
	}
	tx,err := c.db().Begin()
	if err != nil {
		r.Response.WriteJson(resp.Fail("操作异常,请稍后再试！"+err.Error()))
		r.Exit()
	}
	inres,err := tx.Table("lv_travel").Data(data_map).Insert()
	if err != nil {
		tx.Rollback()
		r.Response.WriteJson(resp.Fail("上传基本信息失败！"))
		r.Exit()
	}
	last_id,err := inres.LastInsertId()
	if err != nil {
		tx.Rollback()
		r.Response.WriteJson(resp.Fail("上传基本信息失败！未获取到企业id"+err.Error()))
		r.Exit()
	}
	e_id := gconv.Int(last_id)
	_,err = tx.Table("user").Where("uuid = ?",params["user_uuid"]).Data(g.Map{
		"e_id":e_id,
	}).Update()
	if err != nil {
		tx.Rollback()
		r.Response.WriteJson(resp.Fail("上传基本信息失败！"))
		r.Exit()
	}
	tx.Commit()
	r.Response.WriteJson(resp.Succ("上传基本信息成功！"))
}

/*
	获取旅行社信息
*/
func getTravelByUuid(user_uuid interface{}) g.Map{
	res,err := g.DB().Table("lv_travel as travel").LeftJoin("user as u","travel.id = u.e_id").Where("u.uuid=?",user_uuid).Fields("travel.*").One()
	if err != nil {
		glog.Error(err)
		return nil
	}
	resmap := res.Map()
	if resmap["uuid"] == nil {
		return nil
	}
	return resmap
}
