package task

import (
	"database/sql"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/glog"
	"github.com/gogf/gf/os/gtime"
	"github.com/gogf/gf/os/gtimer"
	"github.com/gogf/gf/util/gconv"
	"time"
)



/*
	酒店自动接单
	检测系统自动接单开关，0关1开
	根据询价但匹配有条件的酒店
	匹配开启自动接单的酒店
	匹配满足自动接单条件的酒店
	创建抢单数据
*/
func AutoReceiveOrder(order_uuid interface{}){
	dbres,err := g.DB().Table("lv_order").Where("uuid=?",order_uuid).One()
	if err != nil && err != sql.ErrNoRows{
		glog.Error(err)
		return
	}
	if dbres == nil {
		return
	}
	//系统自动接单开关
	sysauto,err := g.DB().Table("sys_settings").Where("key=?","auto_receive").FindValue("`value`")
	isauto := sysauto.Int()
	if sysauto == nil || isauto == 0{
		return
	}
	ordermap := dbres.Map()
	//根据订单信息取出筛选条件(区域,地址,星级)
	conditions := g.Map{
		"hotel.city":ordermap["city"],
		"hotel.star":ordermap["star"],
		"hotel.address":ordermap["address"],
		"hotel.auto_receive":1,//用户自动接单开关
		"hotel.receive_minprice <= ?":gconv.Int(ordermap["room_price"]),
		"hotel.receive_maxprice >= ?":gconv.Int(ordermap["room_price"]),
	}
	res,err := g.DB().Table("jiu_hotel hotel,user as user").Where(conditions).And("user.e_id = hotel.id").Fields("hotel.*,user.uuid as user_uuid").FindAll()

	if err != nil && err != sql.ErrNoRows{
		glog.Error(err)
		return
	}

	for _,v := range res.List(){
		g.DB().Table("lv_order_receive").Data(g.Map{
			"create_time":gtime.Datetime(),
			"createby_uuid":v["user_uuid"],
			"order_uuid":ordermap["uuid"],
			"hotel_uuid":v["uuid"],
			"isauto":1,//自动接单;默认是手动接单0
			"e_id":v["id"],
		}).Save()

	}


}
/*
	定时任务,对于已确认的订单,根据订单时间,自动修改订单状态
//0:待确认，1:已确认，2：进行中,3:已完成，4：已失效
*/
func AutoChangeOrderStatus(order_uuid interface{}) {

	dbres,err := g.DB().Table("lv_order").Where("uuid=?",order_uuid).One()
	if err != nil && err != sql.ErrNoRows{
		glog.Error(err)
		return
	}
	if dbres == nil {
		return
	}
	//ordermap := dbres.Map()

	//获取订单最早开始时间
	res,err := g.DB().Table("lv_ordertime").Where("order_uuid=?",order_uuid).Order("start_time asc").One()
	if err != nil && err != sql.ErrNoRows{
		glog.Error(err)
		return
	}
	start_t:= gconv.String(res.Map()["start_time"])
	//获取订单最晚结束时间
	res,err = g.DB().Table("lv_ordertime").Where("order_uuid=?",order_uuid).Order("end_time desc").One()
	if err != nil && err != sql.ErrNoRows{
		glog.Error(err)
		return
	}
	end_t := gconv.String(res.Map()["end_time"])
	start_time := gtime.NewFromStr(start_t)
	end_time := gtime.NewFromStr(end_t)

	for {
		if gtime.Now().After(start_time){//订单过了开始时间,修改订单状态为进行中
			if gtime.Now().After(end_time){
				//过了结束时间,修改订单状态为已完成
				_,err := g.DB().Table("lv_order").Where("uuid=?",order_uuid).Data(g.Map{
					"status":3,//已完成
					"remarks":"该询价订单已完成使用！",
				}).Update()
				if err != nil {
					glog.Error(err)
					return
				}
				//msg := fmt.Sprintf("您在%s发布的询价订单已完成使用！",gtime.NewFromStr(gconv.String(ordermap["create_time"])).Format("m-d h"))
				//service.TsSysOrderMsgToTravel(order_uuid,msg)
				return
			}
			_,err = g.DB().Table("lv_order").Where("uuid=?",order_uuid).Data(g.Map{
				"status":2,//进行中
				"remarks":"该询价订单正在进行中！",
			}).Update()
			if err != nil {
				glog.Error(err)
				return
			}
			//msg := fmt.Sprintf("您在%s发布的询价订单正在进行中！",gtime.NewFromStr(gconv.String(ordermap["create_time"])).Format("m-d h"))
			//service.TsSysOrderMsgToTravel(order_uuid,msg)
			return
		}
		time.Sleep(gtime.S*15)//15秒一次
	}
}



/*
	订单创建后的一个自动取消任务
	目的：15分钟后,对于没有接单的询价订单，系统自动设置为取消状态

*/
func SysCancelOrder(order_uuid interface{}) {
	gtimer.SetTimeout(gtime.M*15, func() {
		dbres,err := g.DB().Table("lv_order").Where("uuid=?",order_uuid).One()
		if err != nil && err != sql.ErrNoRows{
			glog.Error(err)
			return
		}
		ordermap := dbres.Map()
		if gconv.Int(ordermap["status"]) != 0{//状态已改,退出任务
			return
		}

		//如果酒店半小时内没接单,取消该订单
		_,err =  g.DB().Table("lv_order").Where("uuid=?",order_uuid).Data(g.Map{
			"status":4,
			"remarks":"订单已超时15分钟,已被系统取消！",
		}).Update()
		if err != nil{
			glog.Error(err)
			return
		}
		//msg := fmt.Sprintf("您在%s发布的询价订单,15分内没有酒店接单,系统已自动取消该订单！",ordermap["create_time"])

		//service.TsSysOrderMsgToTravel(order_uuid,msg)

	})
}

