package api

import (
	"database/sql"
	"fmt"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/util/gconv"
	"jingyu/app/service/service"
	"jingyu/library/resp"
)

/*
	询价订单房间数量修改
	params:room_num,uuid,comment
*/
func (c *Apis) UpdateEnquiryOrderHotelNum(r *ghttp.Request){
	params := r.GetRequestMap()
	res,err := c.db().Table("lv_order").Where("uuid=?",params["uuid"]).One()
	if err != nil && err != sql.ErrNoRows{
		r.Response.WriteJson(resp.Fail("服务器未响应，请稍后再试！"))
		r.Exit()
	}
	ordermap := res.Map()
	order_status := gconv.Int(ordermap["status"])
	room_num := gconv.Int(ordermap["room_num"])
	//大于3,订单已完成或已失效,无法取消;
	if order_status >= 3{
		r.Response.WriteJson(resp.Fail("操作失败,当前订单状态不能取消！"))
		r.Exit()
	}
	tx,err := c.db().Begin()
	if err != nil {
		r.Response.WriteJson(resp.Fail("服务器未响应,请稍后再试！"))
		r.Exit()
	}
	_,err = tx.Table("lv_order").Where("uuid=?",params["uuid"]).Data(g.Map{
		"room_num":params["room_num"],
	}).Update()
	if err != nil {
		tx.Rollback()
		r.Response.WriteJson(resp.Fail("操作失败"+err.Error()))
		r.Exit()
	}
	//订单状态未确定，或者改的房数为减,不需要审核；
	if order_status == 0 || room_num >= gconv.Int(params["room_num"]){
		tx.Commit()
		r.Response.WriteJson(resp.Succ("操作成功！"))
		r.Exit()
	}
	if gconv.String(ordermap["hotel_uuid"]) == ""{
		tx.Rollback()
		r.Response.WriteJson(resp.Fail("操作异常,订单所绑定酒店信息异常！"))
		r.Exit()
	}
	_,err = tx.Table("lv_order_modify").Data(g.Map{

		"order_uuid":params["uuid"],
		"hotel_uuid":ordermap["hotel_uuid"],
		"comment":params["comment"],
		"field":"room_num",
		"value":ordermap["room_num"],
		"oper_status":0,//0:待确认;1:已确认;2:已拒绝;
	}).OmitEmpty().Save()
	if err != nil {
		tx.Rollback()
		r.Response.WriteJson(resp.Fail("操作失败"+err.Error()))
		r.Exit()
	}
	tx.Commit()
	go service.TsModifyOrderMsgToHotel(ordermap["uuid"],fmt.Sprint("有订单信息被修改！"))
	r.Response.WriteJson(resp.Succ("操作成功！"))
}



/*
	取消询价订单
	params :uuid,comment
*/
func (c *Apis) CancelEnquiryOrder(r *ghttp.Request){
	params := r.GetRequestMap()
	res,err := c.db().Table("lv_order").Where("uuid=?",params["uuid"]).One()
	if err != nil && err != sql.ErrNoRows{
		r.Response.WriteJson(resp.Fail("服务器未响应，请稍后再试！"))
		r.Exit()
	}
	ordermap := res.Map()
	order_status := gconv.Int(ordermap["status"])
	//大于3,订单已完成或已失效,无法取消;
	if order_status >= 3{
		r.Response.WriteJson(resp.Fail("操作失败,当前订单状态不能取消！"))
		r.Exit()
	}
	tx,err := c.db().Begin()
	if err != nil {
		r.Response.WriteJson(resp.Fail("服务器未响应,请稍后再试！"))
		r.Exit()
	}
	_,err = tx.Table("lv_order").Where("uuid=?",params["uuid"]).Data(g.Map{
		"status":5,
	}).Update()
	if err != nil {
		tx.Rollback()
		r.Response.WriteJson(resp.Fail("操作失败"+err.Error()))
		r.Exit()
	}

	if order_status == 0 {
		tx.Commit()
		r.Response.WriteJson(resp.Succ("操作成功,订单已取消"))
		r.Exit()
	}
	if gconv.String(ordermap["hotel_uuid"]) == ""{
		tx.Rollback()
		r.Response.WriteJson(resp.Fail("操作异常,订单所绑定酒店信息异常！"))
		r.Exit()
	}
	_,err = tx.Table("lv_order_modify").Data(g.Map{
		"order_uuid":params["uuid"],
		"hotel_uuid":ordermap["hotel_uuid"],
		"comment":params["comment"],
		"field":"status",
		"value":ordermap["status"],
		"oper_status":0,//0:待确认;1:已确认;2:已拒绝;
	}).Save()
	if err != nil {
		tx.Rollback()
		r.Response.WriteJson(resp.Fail("操作失败"+err.Error()))
		r.Exit()
	}
	tx.Commit()
	go service.TsModifyOrderMsgToHotel(ordermap["uuid"],fmt.Sprint("有订单被取消！"))
	r.Response.WriteJson(resp.Succ("操作成功！"))
}

/*
	查被修改的订单详情
*/
func (c *Apis) QueryModifiedEnquiryOrderDetail(r *ghttp.Request){
	params := r.GetRequestMap()
	res,err := c.db().Table("lv_order o").LeftJoin("lv_order_modify om","om.order_uuid = o.uuid").Where("o.uuid=?",params["uuid"]).One()
	if err != nil && err != sql.ErrNoRows{
		r.Response.WriteJson(resp.Fail("服务器未响应，请稍后再试！"))
		r.Exit()
	}
	//入住时间
	ordermap := res.Map()
	times,_ := c.db().Table("lv_ordertime").Where("order_uuid=?",ordermap["order_uuid"]).FindAll()
	ordermap["order_times"] = times.List()
	r.Response.WriteJson(resp.Succ(ordermap))
}

/*
	酒店确认修改信息
	params:uuid=>order_uuid;
	status => 0:待酒店确认，1:已确认，2:已拒绝
*/
func (c *Apis) CheckModifiedEnquiryOrder(r *ghttp.Request){
	params := r.GetRequestMap()
	res,err := c.db().Table("lv_order o").LeftJoin("lv_order_modify om","om.order_uuid = o.uuid").Where("o.uuid=?",params["uuid"]).One()
	if err != nil && err != sql.ErrNoRows{
		r.Response.WriteJson(resp.Fail("服务器未响应，请稍后再试！"))
		r.Exit()
	}
	status := gconv.Int(params["status"])
	if status > 2 {
		r.Response.WriteJson(resp.Fail("操作异常，请稍后再试！"))
		r.Exit()
	}
	ordermap := res.Map()

	tx,err := c.db().Begin()
	if err != nil {
		r.Response.WriteJson(resp.Fail("服务器未响应,请稍后再试！"))
		r.Exit()
	}
	_,err = tx.Table("lv_order_modify").Where("order_uuid=?",params["uuid"]).Data(g.Map{
		"oper_status":status,
		"hotel_comment":params["comment"],
	}).OmitEmpty().Update()
	if err != nil {
		tx.Rollback()
		r.Response.WriteJson(resp.Fail("操作失败"+err.Error()))
		r.Exit()
	}
	message := "您的订单信息修改审核通过！"
	if status == 2 {
		//酒店方拒绝,订单数据需还原
		_,err := c.db().Table("lv_order").Where("uuid=?", params["uuid"]).Data(g.Map{
			gconv.String(ordermap["field"]):ordermap["value"],
		}).Update()
		if err != nil {
			tx.Rollback()
			r.Response.WriteJson(resp.Fail("操作失败"+err.Error()))
			r.Exit()
		}
		message = "您的订单信息修改审核已被拒绝！"
	}
	tx.Commit()
	go service.TsModifyOrderMsgToTravel(params["uuid"],message)
	r.Response.WriteJson(resp.Succ("操作成功！"))
}