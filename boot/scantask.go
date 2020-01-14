package boot

import (
	"database/sql"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/glog"
	"github.com/gogf/gf/os/gtime"
	"github.com/gogf/gf/text/gstr"
	"github.com/gogf/gf/util/gconv"
	"jingyu/app/service/task"
	"time"
)

//扫描查询快到期的订单并提示给用户
func scanorder(){
	glog.Println("扫描任务已开启")
	go autoDealCancelOrder()
	go autoDealFinishOrder()
}

/*
//1，已确认，2进行中，3已完成，4已失效
把创建时间超过15分钟的且未确认的订单，状态修改为已失效
*/
func autoDealCancelOrder(){
	dbres,err := g.DB().Table("lv_order").Where("status=?",0).FindAll()
	if err != nil && err != sql.ErrNoRows{
		glog.Error(err)
		return
	}
	if dbres == nil {
		glog.Println("未扫描到符合条件的订单状态。。1")
		return
	}
	order_list := dbres.List()
	for _,v := range order_list {
		create_time := gtime.NewFromStr(gconv.String(v["create_time"])).Time
		cancel_time := gtime.NewFromTime(create_time.Add(gtime.M*15))
		if gtime.Now().After(cancel_time){//十五分钟后,修改未确认酒店的询价单的状态为已失效
			g.DB().Table("lv_order").Where("uuid=?",v["uuid"]).Data(g.Map{
				"status":4,
				"remarks":"订单已超时15分钟,已被系统取消！",
			}).Update()
		}else{
			//没超过十五分钟的订单,列入计划任务
			go func(v g.Map,cancel_time *gtime.Time) {
				for{
					if gtime.Now().After(cancel_time){
						g.DB().Table("lv_order").Where("uuid=?",v["uuid"]).Data(g.Map{
							"status":4,
							"remarks":"订单已超时15分钟,已被系统取消！",
						}).Update()
						return
					}else{
						time.Sleep(gtime.S*15)//15秒一次
					}
				}

			}(v,cancel_time)
		}
	}
}
/*
	1已确认，2进行中，3已完成，4已失效
把订单状态为1已确认的订单，如果超过开始时间未超过结束时间，修改状态为进行中，如果超过结束时间，修改状态为已完成。
*/
func autoDealFinishOrder(){
	dbres,err := g.DB().Table("lv_order").Where(g.Map{
		"status > 0":nil,
		"status < 3":nil,
	}).FindAll()
	if err != nil && err != sql.ErrNoRows{
		glog.Error(err)
		return
	}
	if dbres == nil {
		glog.Println("未扫描到的订单2")
		return
	}
	order_list := dbres.List()
	for _,v := range order_list {
		go task.AutoChangeOrderStatus(v["uuid"])
	}
}
//格式化Time.Duration类型输出字符串
func FormatTime(d string)string{
	return gstr.ReplaceByMap(d,map[string]string{
		"h":"小时",
		"m":"分",
		"s":"秒",
	})
}