package service

import (
	"database/sql"
	"fmt"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/glog"
	"github.com/gogf/gf/util/gconv"
	"jingyu/library/Queue"
)


/*
	旅行社发询价推送
	根据订单信息取出筛选条件(区域,地址,星级)
	推送询价消息给酒店端
*/
func TsOrderMsgToHotel(order_uuid interface{}){
	dbres,err := g.DB().Table("lv_order").Where("uuid=?",order_uuid).One()
	if err != nil && err != sql.ErrNoRows{
		glog.Error(err)
		return
	}
	if dbres == nil {
		return
	}
	ordermap := dbres.Map()
	//根据订单信息取出筛选条件(区域,地址,星级)
	conditions := g.Map{
		"hotel.city":ordermap["city"],
		"hotel.star":ordermap["star"],
		"hotel.address":ordermap["address"],
	}
	res,err := g.DB().Table("jiu_hotel hotel,user as user").Where(conditions).And("user.e_id = hotel.id").Fields("user.uuid").FindAll()
	if err != nil && err != sql.ErrNoRows{
		glog.Error(err)
		return
	}
	user_list := g.SliceStr{}
	for _,v := range res.List(){
		user_list = append(user_list,gconv.String(v["uuid"]))
	}
	if (len(user_list) == 0){
		return
	}
	message := fmt.Sprintf("您有新的询价订单！")
	//查询符合以上条件的酒店
	msg := Queue.NewMsg()
	msg.To = user_list//符合以上条件的酒店负责人uuid列表
	msg.Source_Uuid = gconv.String(order_uuid)
	msg.Category = "order"
	msg.SetContent(ordermap,gconv.String(msg.Category),message)
	Queue.GetInstance().Push(msg)
}

/*
	系统执行改变订单状态,推送给旅行社
*/
func TsSysOrderMsgToTravel(order_uuid interface{},message string){
	dbres,err := g.DB().Table("lv_order").Where("uuid=?",order_uuid).One()
	if err != nil && err != sql.ErrNoRows{
		glog.Error(err)
		return
	}
	if dbres == nil {
		glog.Error("获取订单信息失败！")
		return
	}

	ordermap := dbres.Map()

	//获取旅行社端的推送uuid
	uuid :=  ordermap["createby_uuid"]
	msg := Queue.NewMsg()
	msg.To = uuid
	msg.Category = "order"
	msg.SetContent(ordermap,gconv.String(msg.Category),message)
	Queue.GetInstance().Push(msg)
}
/*
	酒店接单推送给旅行社
*/
func TsOrderMsgToTravel(order_uuid interface{},hotel_uuid interface{}){
	dbres,err := g.DB().Table("lv_order").Where("uuid=?",order_uuid).Value("createby_uuid")
	if err != nil && err != sql.ErrNoRows{
		glog.Error(err)
		return
	}
	if dbres == nil {
		glog.Error("获取订单创建人失败")
		return
	}
	res,err := g.DB().Table("jiu_hotel").Where("uuid=?",hotel_uuid).One()
	if err != nil && err != sql.ErrNoRows{
		glog.Error(err)
		return
	}
	if res == nil {
		glog.Error("未获取到接单酒店的基本信息！")
		return
	}
	message := fmt.Sprintf("您有新的酒店接单！")
	hotelmap := res.Map()
	hotelmap["order_uuid"] = order_uuid
	//获取旅行社端的推送uuid
	uuid :=  dbres.String()
	msg := Queue.NewMsg()
	msg.To = uuid
	msg.Category = "hotel"
	glog.Println("推送hotelmap",hotelmap)
	msg.SetContent(hotelmap,gconv.String(msg.Category),message)
	Queue.GetInstance().Push(msg)
}


/*
	旅行社确认酒店 推送给酒店
*/
func TsConfirmOrderMsgToHotel(order_uuid interface{}){
	dbres,err := g.DB().Table("lv_order").Where("uuid=?",order_uuid).One()
	if err != nil && err != sql.ErrNoRows{
		glog.Error(err)
		return
	}
	if dbres == nil {
		return
	}
	ordermap := dbres.Map()
	//根据订单信息取出筛选条件(区域,地址,星级)
	hotel_uuid := ordermap["hotel_uuid"]
	res,err := g.DB().Table("jiu_hotel hotel,user as user").Where("hotel.uuid=?",hotel_uuid).And("user.e_id = hotel.id").Fields("user.uuid").FindAll()
	if err != nil && err != sql.ErrNoRows{
		glog.Error(err)
		return
	}
	user_list := g.SliceStr{}
	for _,v := range res.List(){
		user_list = append(user_list,gconv.String(v["uuid"]))
	}
	message := fmt.Sprintf("您有已预定的订单！")
	//查询符合以上条件的酒店
	msg := Queue.NewMsg()
	msg.To = user_list//符合以上条件的酒店负责人uuid列表
	msg.Source_Uuid = gconv.String(order_uuid)
	msg.Category = "order"
	msg.SetContent(ordermap,gconv.String(msg.Category),message)
	Queue.GetInstance().Push(msg)
}


/*
	旅行社修改订单信息 发送给酒店确认审核
*/
func TsModifyOrderMsgToHotel(order_uuid,message interface{}){
	dbres,err := g.DB().Table("lv_order").Where("uuid=?",order_uuid).One()
	if err != nil && err != sql.ErrNoRows{
		glog.Error(err)
		return
	}
	if dbres == nil {
		return
	}
	ordermap := dbres.Map()
	hotel_uuid := ordermap["hotel_uuid"]
	res,err := g.DB().Table("jiu_hotel hotel,user as user").Where("hotel.uuid=?",hotel_uuid).And("user.e_id = hotel.id").Fields("user.uuid").One()
	if err != nil && err != sql.ErrNoRows{
		glog.Error(err)
		return
	}
	uuid := res.Map()["uuid"]
	//查询符合以上条件的酒店
	msg := Queue.NewMsg()
	msg.To = uuid//符合以上条件的酒店负责人uuid列表
	msg.Source_Uuid = gconv.String(order_uuid)
	msg.Category = "order_modify"
	msg.SetContent(ordermap,gconv.String(msg.Category),gconv.String(message))
	Queue.GetInstance().Push(msg)
}

/*
	旅行社修改订单信息 发送给酒店确认审核
*/
func TsModifyOrderMsgToTravel(order_uuid,message interface{}){
	dbres,err := g.DB().Table("lv_order").Where("uuid=?",order_uuid).One()
	if err != nil && err != sql.ErrNoRows{
		glog.Error(err)
		return
	}
	if dbres == nil {
		return
	}
	ordermap := dbres.Map()
	uuid := ordermap["createby_uuid"]
	//查询符合以上条件的酒店
	msg := Queue.NewMsg()
	msg.To = uuid//符合以上条件的酒店负责人uuid列表
	msg.Source_Uuid = gconv.String(order_uuid)
	msg.Category = "order_modify"
	msg.SetContent(ordermap,gconv.String(msg.Category),gconv.String(message))
	Queue.GetInstance().Push(msg)
}

