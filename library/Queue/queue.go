package Queue

import (
	"fmt"
	"github.com/gogf/gf/container/gqueue"
	"github.com/gogf/gf/encoding/gjson"
	"github.com/gogf/gf/os/glog"
	"github.com/gogf/gf/os/gtime"
	"github.com/gogf/gf/util/gconv"
	"jingyu/library/sms"
	"jingyu/library/tuisong"
	"sync"
)

var (
	instance *gqueue.Queue
	once sync.Once
)

const (
	SMS = iota	//短信通知
	TS			//第三方推送通知
)

const (
	Notify = iota	//通知
	Boardcast	//广播
	Passthrough	//透传,自定义消息
	Verifycode	//验证码
)

const (
	Client = 0
	Business = 1
)


type TS_Message struct{
	Source_info interface{} `json:"source_info,omitempty"`//消息信息
	Source_type string `json:"source_type,omitempty"`//资源类型，通常为数据表类型
	Source_message string `json:"source_message,omitempty"`//消息内容
}
/*
	消息容器
*/
type Msg struct{
	SendType int	//发送方式,短信,推送
	TemplateType interface{} //消息模板类型
	To interface{}	//消息发送目标
	Content TS_Message	//消息发送内容
	CreateTime interface{}	//消息创建时间,消息入队时间
	From int	//发送方
	ToServer  int //服务端类0:客户端;1:商户端
	Source_Uuid string //消息资源uuid
	Category interface{}	//消息种类,资源种类
}


func NewMsg(v ...interface{})*Msg{
	msg := &Msg{
		SendType:TS,
		CreateTime:gtime.Datetime(),
		TemplateType:tuisong.TRANSMISSION,
		ToServer:Client,
		From:Client,
	}
	if len(v) > 0 && v[0] != nil {
		msg = v[0].(*Msg)
	}
	return msg
}


//发送方
func (m *Msg) SetFrom(v int)*Msg{
	if v == Client{
		m.From = Client
	}else{
		m.From = Business
	}
	return m
}

//发送目标,接收方
func (m *Msg) SetTo(v interface{})*Msg{
	m.To = v
	return m
}

//发送目标,接收方
func (m *Msg) SetToServer(v int)*Msg{
	m.ToServer = v
	return m
}

//发送方式
func (m *Msg) SetSendType(v int)*Msg{
	m.SendType = v
	return m
}

//发送方式
func (m *Msg) SetContent(source_info interface{},source_type string,source_message string)*Msg{
	m.Content.Source_message = source_message
	m.Content.Source_info = source_info
	m.Content.Source_type = source_type
	return m
}


/*
	获取队列实例,单例模式保持全局对象唯一
*/
func GetInstance() *gqueue.Queue {
	once.Do(func() {
		instance = gqueue.New()
	})
	return instance
}







/*
	消息解析派发
*/
func Dispatch(v interface{}){
	fmt.Println("队列消息",v)
	if v == nil {
		fmt.Println("队列消息数据为空")
		return
	}
	msg,ok := v.(*Msg)
	if !ok {
		glog.Error("队列消息解析失败!")
		return
	}

	//插入数据库(业务需求)
	//go recordemsg(msg)
	//
	switch msg.SendType {
	case SMS:
		sms.SendSmsMsg(msg.To,msg.Content)
	case TS:
		cid,_ := tuisong.GetCid(msg.To)
		if len(cid) == 1{
			content,err := msg.getContent()
			if err != nil{
				glog.Error(err)
			}
			ts,err := tuisong.New(cid[0],content,gconv.String(msg.TemplateType))
			if err != nil{
				glog.Error(err)
			}
			glog.Println("msg jiegou",msg)
			ts.Send(msg.ToServer)
		}else{
			content,err := msg.getContent()
			if err != nil{
				glog.Error(err)
			}
			ts := tuisong.NewList(content,gconv.String(msg.TemplateType))
			ts.SendList(gconv.Strings(cid),msg.ToServer)
		}
	default:
		glog.Error("未知的发送方式")

	}
}
func (msg *Msg)getContent()(string,error){
	content,err := gjson.Encode(msg.Content)
	if err != nil {
		return "",err
	}
	glog.Println("发送内容","`"+string(content)+"`")
	return string(content),nil
}

/*
	初始化
*/
func init(){
	fmt.Println("queueing...")
	q := GetInstance()
	go func() {
		fmt.Println("读取队列消息中...")
		for{
			select {
			case v := <- q.C:
				go Dispatch(v)
			}
			fmt.Println("已处理一条队列消息...")
		}
	}()
}

