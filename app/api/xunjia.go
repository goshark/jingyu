package api

import (
	"database/sql"
	"fmt"
	"github.com/gogf/gf/encoding/gjson"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/gtime"
	"github.com/gogf/gf/util/gconv"
	"jingyu/app/service/service"
	"jingyu/app/service/task"
	"jingyu/library/Uuid"
	"jingyu/library/resp"
)


/*
	获取询价前台数据
*/
func (c *Apis) GetKeyWordsData(r *ghttp.Request){
	res,err := c.db().Table("keywords").Fields("type").Group("type").Select()
	if err != nil && err != sql.ErrNoRows{
		r.Response.WriteJson(resp.Fail(err.Error()))
		r.Exit()
	}
	resmap := g.Map{}
	for _,v := range res.List(){
		res,_ := c.db().Table("keywords").Where("type=?",v["type"]).Select()
		resmap[gconv.String(v["type"])] = res.List()
	}
	r.Response.WriteJson(resp.Succ(resmap))
}


/*
	获取未接单询价订单列表(酒店用)
*/
func (c *Apis) QueryHotelEnquiryOrderCanReceiveList(r *ghttp.Request){
	params := r.GetRequestMap()
	user_uuid := params["user_uuid"]
	res := getHotelByUuid(user_uuid)
	if res == nil {
		r.Response.WriteJson(resp.NotBind("该用户未绑定企业信息"))
		r.Exit()
	}
	hotelmap := res
	hotel_uuid := res["uuid"]
	//未结单,1.先获取已接单的订单uuid,排除法
	dbres,err := c.db().Table("lv_order as o").LeftJoin("lv_order_receive as r","o.uuid = r.order_uuid").Where("r.hotel_uuid=? and o.status=?",hotel_uuid,0).FindAll()
	if err != nil && err != sql.ErrNoRows{
		r.Response.WriteJson(resp.Fail("查询失败"+err.Error()))
		r.Exit()
	}
	order_list := g.Slice{}
	for _,v := range dbres.List(){
		order_list = append(order_list, v["order_uuid"])
	}

	dbmodel := c.db().Table("lv_order as o")
	conditions := g.Map{
		"o.city":hotelmap["city"],
		"o.star":hotelmap["star"],
		"o.address":hotelmap["address"],
		"o.status":0,
	}
	if len(order_list) > 0{
		dbmodel.Where("o.uuid NOT IN(?)",order_list)
	}
	dbmodel.Where(conditions)
	dbres,err =dbmodel.FindAll()
	if err != nil && err != sql.ErrNoRows{
		r.Response.WriteJson(resp.Fail("查询失败"+err.Error()))
		r.Exit()
	}
	order_lists := dbres.List()
	for _,v := range order_lists{
		times,_ := c.db().Table("lv_ordertime").Where("order_uuid=?",v["uuid"]).FindAll()
		v["order_times"] = times.List()
	}

	r.Response.WriteJson(resp.Succ(order_lists))
}
/*
	获取询价订单列表(酒店用)
*/
func (c *Apis) QueryHotelEnquiryOrderList(r *ghttp.Request){
	params := r.GetRequestMap()
	user_uuid := params["user_uuid"]
	res := getHotelByUuid(user_uuid)
	if res == nil {
		r.Response.WriteJson(resp.NotBind("该用户未绑定企业信息"))
		r.Exit()
	}
	hotel_uuid := res["uuid"]
	resmap := g.Map{}

	//已接单
	dbres,err := c.db().Table("lv_order as o").LeftJoin("lv_order_receive as r","o.uuid = r.order_uuid").Where("r.hotel_uuid=? and o.status=?",hotel_uuid,0).Order("o.create_time desc").FindAll()

	order_list := dbres.List()
	for k,v := range order_list{
		if gconv.String(v["hotel_uuid"]) != ""{
			times,_ := c.db().Table("lv_ordertime").Where("order_uuid=?",v["uuid"]).FindAll()
			order_list[k]["order_times"] = times.List()
		}
	}
	resmap["received"] = order_list
	//其他状态单
	dbres,err = c.db().Table("lv_order").Where("hotel_uuid = ?",hotel_uuid).Order("create_time desc").FindAll()
	if err != nil && err != sql.ErrNoRows{
		r.Response.WriteJson(resp.Fail("查询失败"+err.Error()))
		r.Exit()
	}

	order_list1 := dbres.List()
	for k,v := range order_list1{
		if gconv.String(v["hotel_uuid"]) != ""{
			times,_ := c.db().Table("lv_ordertime").Where("order_uuid=?",v["uuid"]).FindAll()
			order_list1[k]["order_times"] = times.List()
		}
	}
	resmap["status"] = order_list1
	//已失效单
	dbres,err = c.db().Table("lv_order_receive as r").LeftJoin("lv_order as o","o.uuid = r.order_uuid").Where("r.hotel_uuid=?",hotel_uuid).And("o.hotel_uuid!=?",hotel_uuid).And("o.status!=",0).Order("o.create_time desc").Fields("o.*").FindAll()
	if err != nil && err != sql.ErrNoRows{
		r.Response.WriteJson(resp.Fail("查询失败"+err.Error()))
		r.Exit()
	}
	order_list2 := dbres.List()
	for k,v := range order_list2{
		if gconv.String(v["hotel_uuid"]) != ""{
			times,_ := c.db().Table("lv_ordertime").Where("order_uuid=?",v["uuid"]).FindAll()
			order_list2[k]["order_times"] = times.List()
		}
	}

	resmap["invalid"] = order_list2
	r.Response.WriteJson(resp.Succ(resmap))
}
/*
	获取询价订单列表
*/
func (c *Apis) QueryEnquiryOrderList(r *ghttp.Request){
	params := r.GetRequestMap()
	res,err := c.db().Table("lv_order o").LeftJoin("jiu_hotel hotel","hotel.uuid=o.hotel_uuid").Fields("o.*,hotel.name as hotel_name,hotel.real_address,hotel.address as hotel_address").Where("createby_uuid=?",params["user_uuid"]).Group("o.create_time desc").FindAll()
	if err != nil && err != sql.ErrNoRows{
		r.Response.WriteJson(resp.Fail("查询失败"+err.Error()))
		r.Exit()
	}
	order_list := res.List()
	for k,v := range order_list{

			times,_ := c.db().Table("lv_ordertime").Where("order_uuid=?",v["uuid"]).FindAll()
			order_list[k]["order_times"] = times.List()

	}
	r.Response.WriteJson(resp.Succ(order_list))
}

/*
	查看询价详情
*/
func (c *Apis) QueryEnquiryOrderDetail(r *ghttp.Request){
	params := r.GetRequestMap()
	order_uuid  := params["uuid"]
	res,err := c.db().Table("lv_order").Where("uuid=?",order_uuid).One()
	if err != nil && err != sql.ErrNoRows{
		r.Response.WriteJson(resp.Fail("查询失败"+err.Error()))
		r.Exit()
	}
	ordermap := res.Map()
	times,_ := c.db().Table("lv_ordertime").Where("order_uuid=?",order_uuid).FindAll()
	ordermap["order_times"] = times.List()

	if gconv.Int(ordermap["status"]) == 0{
		receivers,err := c.db().Table("lv_order_receive as orv").LeftJoin("jiu_hotel as h","orv.hotel_uuid = h.uuid").Where("orv.order_uuid=?",order_uuid).FindAll()
		if err != nil && err != sql.ErrNoRows{
			r.Response.WriteJson(resp.Fail("查询失败"+err.Error()))
			r.Exit()
		}
		ordermap["receivers"] = receivers.List()
	}
	r.Response.WriteJson(resp.Succ(ordermap))
}


/*
	酒店接受询价订单
*/
func (c *Apis) ReceiveEnquiryOrder(r *ghttp.Request){
	params := r.GetRequestMap()
	user_uuid := params["user_uuid"]
	hotel_info := getHotelByUuid(user_uuid)
	order_uuid := params["uuid"]
	if hotel_info == nil {
		r.Response.WriteJson(resp.Fail("未获取到企业信息！"))
		r.Exit()
	}
	if gconv.Int(hotel_info["status"]) != 1{
		r.Response.WriteJson(resp.Fail("企业信息未审核或审核未通过！"))
		r.Exit()
	}

	dbres,err := c.db().Table("jiu_hotel_receive").Where(g.Map{
		"hotel_uuid":hotel_info["uuid"],
		"order_uuid":order_uuid,
	}).One()
	if err != nil && err != sql.ErrNoRows{
		r.Response.WriteJson(resp.Fail("获取询价信息失败"+err.Error()))
		r.Exit()
	}
	if dbres != nil {
		r.Response.WriteJson(resp.Fail("接单失败！不能重复接单！"))
		r.Exit()
	}
	dbres,err = c.db().Table("lv_order").Where("uuid=?",order_uuid).One()
	if err != nil && err != sql.ErrNoRows{
		r.Response.WriteJson(resp.Fail("获取询价订单信息失败"+err.Error()))
		r.Exit()
	}
	ordermap := dbres.Map()
	if gconv.Int(ordermap["status"]) != 0{
		r.Response.WriteJson(resp.Fail("该询价订单当前状态不能接单！,状态码："+gconv.String(ordermap["status"])))
		r.Exit()
	}
	if gconv.String(ordermap["hotel_uuid"]) != ""{
		r.Response.WriteJson(resp.Fail("操作失败！该询价订单已确认酒店！"))
		r.Exit()
	}


	data := g.Map{
		"create_time":gtime.Datetime(),
		"createby_uuid":user_uuid,
		"hotel_uuid":hotel_info["uuid"],
		"order_uuid":params["uuid"],
		"e_id":hotel_info["id"],
	}
	_,err = c.db().Table("lv_order_receive").Save(data)
	if err != nil {
		r.Response.WriteJson(resp.Fail("操作失败！"+err.Error()))
		r.Exit()
	}
	go service.TsOrderMsgToTravel(order_uuid,hotel_info["uuid"])
	r.Response.WriteJson(resp.Succ("操作成功！"))
}


/*
	处理询价订单,旅行社确认订单
*/
func (c *Apis) ConfirmEnquiryOrder(r *ghttp.Request){
	params := r.GetRequestMap()
	//user_uuid := params["user_uuid"]
	hotel_uuid := params["hotel_uuid"]
	order_uuid := params["order_uuid"]
	order,err := c.db().Table("lv_order").Where("uuid=?",order_uuid).One()
	if err != nil && err != sql.ErrNoRows{
		r.Response.WriteJson(resp.Fail("获取询价订单信息失败"+err.Error()))
		r.Exit()
	}
	ordermap := order.Map()
	if gconv.Int(ordermap["status"]) != 0{
		r.Response.WriteJson(resp.Fail("当前订单状态不能接单！状态码："+gconv.String(ordermap["status"])))
		r.Exit()
	}
	_,err = c.db().Table("lv_order").Where("uuid=?",order_uuid).Data(g.Map{
		"status":1,
		"hotel_uuid":hotel_uuid,
	}).Update()
	if err != nil {
		r.Response.WriteJson(resp.Fail("操作失败！"+err.Error()))
		r.Exit()
	}

	//订单已确认,到入住时间时,自动修改订单状态为进行中
	go task.AutoChangeOrderStatus(order_uuid)
	go service.TsConfirmOrderMsgToHotel(order_uuid)
	r.Response.WriteJson(resp.Succ("操作成功！"))
}

/*
	发布询价
*/
func (c *Apis) CreateEnquiryOrder(r *ghttp.Request){
	params := r.GetRequestMap()
	order_times := gjson.New(params["order_times"]).ToArray()
	order_uuid := Uuid.Generate_uuid()
	//获取该企业是否审核通过
	travel := getTravelByUuid(params["user_uuid"])
	if travel == nil {
		r.Response.WriteJson(resp.NotBind("未绑定企业信息！"))
		r.Exit()
	}
	if gconv.Int(travel["status"]) != 1{
		r.Response.WriteJson(resp.Fail("企业信息未审核或审核未通过！"))
		r.Exit()
	}
	order_map := g.Map{
		"uuid":order_uuid,
		"team_num":params["team_num"],
		"city":params["city"],
		"address":params["address"],
		"star":params["star"],
		"room_type":params["room_type"],
		"room_num":params["room_num"],
		"room_price":params["room_price"],
		"remarks":params["remarks"],
		"create_time":gtime.Datetime(),
		"createby_uuid":params["user_uuid"],
	}
	room_num := 5
	if res:=c.getSysCfg("room_num"); res != ""{
		room_num = gconv.Int(res)
	}
	room_price := 70
	if res:=c.getSysCfg("room_num"); res != ""{
		room_num = gconv.Int(res)
	}
	if gconv.Int(params["room_num"]) < room_num{
		r.Response.WriteJson(resp.Fail(fmt.Sprintf("房间数量限制最少为%d\n",room_num)))
		r.Exit()
	}
	if gconv.Int(params["room_price"]) < room_price{
		r.Response.WriteJson(resp.Fail(fmt.Sprintf("所询价格限制最少为%d\n",room_price)))
		r.Exit()
	}
	if tx,err := c.db().Begin();err == nil {
		_,err = tx.Table("lv_order").Save(order_map)
		if err != nil {
			r.Response.WriteJson(resp.Fail("操作异常"+err.Error()))
			r.Exit()
		}
		if len(order_times) <= 0{
			tx.Rollback()
			r.Response.WriteJson(resp.Fail("至少选择一个入住时间！"))
			r.Exit()
		}
		order_timelist := g.List{}
		format_time := "06:00"
		if res:=c.getSysCfg("time"); res != ""{
			format_time = res
		}

		for _,v := range order_times{
			order_time_map := gconv.Map(v)
			order_time_map["order_uuid"] = order_uuid
			order_time_map["start_time"] =gtime.NewFromStr(gconv.String(order_time_map["start_time"])).Format("Y-m-d "+format_time)
			order_time_map["end_time"] = gtime.NewFromStr(gconv.String(order_time_map["end_time"])).Format("Y-m-d "+format_time)

			order_timelist = append(order_timelist, order_time_map)
		}

		_,err = tx.Table("lv_ordertime").Insert(order_timelist)
		if err != nil {
			tx.Rollback()
			r.Response.WriteJson(resp.Fail("操作异常"+err.Error()))
			r.Exit()
		}
		tx.Commit()

	}else{
		r.Response.WriteJson(resp.Fail("操作异常"+err.Error()))
		r.Exit()
	}
	//系统15分后自动取消订单
	go task.SysCancelOrder(order_uuid)
	go service.TsOrderMsgToHotel(order_uuid)
	go task.AutoReceiveOrder(order_uuid)
	r.Response.WriteJson(resp.Succ("操作成功！"))

}

/*
	修改询价
*/
func (c *Apis) UpdateEnquiryOrder(r *ghttp.Request){

}


