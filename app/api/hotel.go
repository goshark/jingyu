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
	"jingyu/library/upload"
	"strings"
)

/*
	查询酒店列表
*/
func (c *Apis) QueryHotelList(r *ghttp.Request){
	res,err  := c.db().Table("jiu_hotel").Where("status=?",1).FindAll()
	if err != nil && err != sql.ErrNoRows{
		r.Response.WriteJson(resp.Fail("查询失败"+err.Error()))
		r.Exit()
	}
	r.Response.WriteJson(resp.Succ(res.List()))
}

/*
	查询酒店详情
*/
func (c *Apis) QueryHotelDetail(r *ghttp.Request){
	params := r.GetRequestMap()
	hotel_uuid := params["uuid"]
	res,err  := c.db().Table("jiu_hotel").Where("uuid = ? and status=?",hotel_uuid,1).One()
	if err != nil && err != sql.ErrNoRows{
		r.Response.WriteJson(resp.Fail("查询失败"+err.Error()))
		r.Exit()
	}
	hotel_map := res.Map()
	resmap := hotel_map
	houseres,_ := c.db().Table("jiu_house").Where("hotel_uuid=?",resmap["uuid"]).FindAll()
	resmap["house_list"] = houseres.List()
	r.Response.WriteJson(resp.Succ(resmap))
}

/*
	修改酒店基本信息
*/
func (c *Apis) UpdateHotelBasicInfo(r *ghttp.Request){
	params := r.GetRequestMap()
	user_uuid := params["user_uuid"]
	source := getHotelByUuid(user_uuid)
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
		"star":params["star"],
		"real_address":params["real_address"],
		"tel":params["tel"],
		"fac_id":params["fac_id"],//逗号隔开
		"pics":gjson.New(pics).MustToJsonString(),
	}
	tx,err := c.db().Begin()
	if err != nil {
		r.Response.WriteJson(resp.Fail("操作异常,请稍后再试！"+err.Error()))
		r.Exit()
	}
	_,err = tx.Table("jiu_hotel").Where("uuid=?",source["uuid"]).Data(data_map).Filter().Update()
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
func (c *Apis) UploadHotelBasicInfo(r *ghttp.Request){
	params := r.GetRequestMap()
	//查重
	conditions := g.Map{
		"name":params["name"],
		"city":params["city"],
		"address":params["address"],
		"star":params["star"],
	}
	hotels,err := c.db().Table("jiu_hotel").Where(conditions).FindAll()
	if err != nil && err != sql.ErrNoRows{
		r.Response.WriteJson(resp.Fail("查询失败"+err.Error()))
		r.Exit()
	}
	if hotels != nil && len(hotels.List()) > 0{
		r.Response.WriteJson(resp.Fail("该酒店信息已存在！"))
		r.Exit()
	}


	source := getHotelByUuid(params["user_uuid"])
	if source != nil {
		r.Response.WriteJson(resp.Fail("企业基本信息已上传,不能重复上传!"))
		r.Exit()
	}

	uuid := Uuid.Generate_uuid()
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
		"real_address":params["real_address"],
		"star":params["star"],
		"tel":params["tel"],
		"fac_id":params["fac_id"],//逗号隔开
		"create_time":gtime.Datetime(),
		"pics":gjson.New(pics).MustToJsonString(),
	}
	tx,err := c.db().Begin()
	if err != nil {
		r.Response.WriteJson(resp.Fail("操作异常,请稍后再试！"+err.Error()))
		r.Exit()
	}
	inres,err := tx.Table("jiu_hotel").Data(data_map).OmitEmpty().Insert()
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
	获取酒店信息
*/
func getHotelByUuid(user_uuid interface{}) g.Map{
	res,err := g.DB().Table("jiu_hotel as hotel").LeftJoin("user as u","hotel.id = u.e_id").Where("u.uuid=?",user_uuid).Fields("hotel.*").One()
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



/*
	获取房型列表
*/
func (c *Apis) GetHotelHouseList(r *ghttp.Request){
	params := r.GetRequestMap()
	user_uuid := params["user_uuid"]
	travel := getHotelByUuid(user_uuid)
	if travel == nil {
		r.Response.WriteJson(resp.Fail("获取房型列表失败！未获取到企业信息"))
		r.Exit()
	}
	house,err := c.db().Table("jiu_house").Where("hotel_uuid=?",travel["uuid"]).FindAll()
	if err != nil && err != sql.ErrNoRows{
		r.Response.WriteJson(resp.Fail("查询异常！"+err.Error()))
		r.Exit()
	}
	r.Response.WriteJson(resp.Succ(house.List()))
}

/*
	酒店房型添加
*/
func (c *Apis) UploadHotelHouse(r *ghttp.Request){
	params := r.GetRequestMap()
	user_uuid := params["user_uuid"]
	travel := getHotelByUuid(user_uuid)
	if travel == nil {
		r.Response.WriteJson(resp.Fail("酒店房型添加失败！未获取到企业信息"))
		r.Exit()
	}
	house_name := params["name"]
	if house_name == nil {
		house_name = "house"
	}
	parent_path := "hotels/"+gconv.String(travel["uuid"])+"/house_"+gconv.String(house_name)
	//多文件上传处理
	pic_paths := upload.UploadMore(r,parent_path)

	//构建要插入的数据
	insdata := g.Map{
		"pic":"",
	}
	if len(pic_paths) > 0 {
		insdata["pic"] = strings.Join(pic_paths,";")
	}
	_,err := c.db().Table("jiu_house").Data(g.Map{
		"uuid":Uuid.Generate_uuid(),
		"name":params["name"],
		"hotel_uuid":travel["uuid"],
		"pic":insdata["pic"],
		"price":params["price"],
		"createby_uuid":params["user_uuid"],
		"create_time":gtime.Datetime(),
	}).OmitEmpty().Filter().Insert()

	if err != nil {
		r.Response.WriteJson(resp.Fail("酒店房型添加失败！"+err.Error()))
		r.Exit()
	}
	r.Response.WriteJson(resp.Succ("操作成功！"))
}
/*
	Save抢单设置
*/
func (c *Apis) SaveAutoReceive(r *ghttp.Request){
	params := r.GetRequestMap()
	user_uuid := params["user_uuid"]
	source := getHotelByUuid(user_uuid)
	if source == nil {
		r.Response.WriteJson(resp.Fail("抢单设置保存失败！未获取到企业信息"))
		r.Exit()
	}
	_,err := c.db().Table("jiu_hotel").Where("uuid=?", source["uuid"]).Data(g.Map{
		"auto_receive":gconv.Int(params["auto"]),
		"receive_minprice":gconv.Int(params["min"]),
		"receive_maxprice":gconv.Int(params["max"]),
	}).Update()
	if err != nil {
		r.Response.WriteJson(resp.Fail("保存失败！"+err.Error()))
		r.Exit()
	}
	r.Response.WriteJson(resp.Succ("操作成功！"))
}

/*
	酒店房型修改
*/
func (c *Apis) UpdateHotelHouse(r *ghttp.Request){
	params := r.GetRequestMap()
	user_uuid := params["user_uuid"]
	source := getHotelByUuid(user_uuid)
	if source == nil {
		r.Response.WriteJson(resp.Fail("酒店房型修改失败！未获取到企业信息"))
		r.Exit()
	}
	house_name := params["name"]
	if house_name == nil {
		house_name = "house"
	}
	parent_path := "hotels/"+gconv.String(source["uuid"])+"/house_"+gconv.String(house_name)
	//多文件上传处理
	pic_paths := upload.UploadMore(r,parent_path)

	//构建要插入的数据
	insdata := g.Map{}

	if len(pic_paths) > 0 {
		insdata["pic"] = strings.Join(pic_paths,";")
	}
	_,err := c.db().Table("jiu_house").Where("uuid=?",params["uuid"]).Data(g.Map{
		"name":params["name"],
		"hotel_uuid":source["uuid"],
		"pic":insdata["pic"],
		"price":params["price"],
		"createby_uuid":params["user_uuid"],
		"create_time":gtime.Datetime(),
	}).Filter().OmitEmpty().Update()

	if err != nil {
		r.Response.WriteJson(resp.Fail("酒店房型修改失败！"+err.Error()))
		r.Exit()
	}
	r.Response.WriteJson(resp.Succ("操作成功！"))
}