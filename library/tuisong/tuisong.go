package tuisong

//第三方推送集成(个推)
import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/encoding/gjson"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/gcache"
	"github.com/gogf/gf/os/glog"
	"github.com/gogf/gf/os/gmutex"
	"github.com/gogf/gf/os/gtime"
	"github.com/gogf/gf/util/gconv"
	"github.com/gogf/gf/util/grand"
)

//填写对应客户端应用配置


//var (
//	//商户端
//	AppID        = "rUEYM7eiaz6tQrBdrdHeW9"
//	AppKey       = "QRNHfwOfJu95TR7ZHvVZ2"
//	AppSecret    = "dR5apw97c1A014TE9y2bE6"
//	MasterSecret = "SsODKzu2Hp87DYwPwWrka7"
//)

/*
	推送配置
*/
type TS_CONFIG struct{
	ServerType int `orm:"server"`//0:客户端;1:商户端
	AppID string `orm:"app_id"`
	AppKey string	`orm:"app_key"`
	AppSecret string	`orm:"app_secret"`
	MasterSecret string	`orm:"master_secret"`
}


var (
	ts_config = new(TS_CONFIG)
)

/*
	获取推送配置信息,服务类型,"client=1,business=2"
*/
func getTsConfig(server_type int)*TS_CONFIG{
	if server_type != 1 && server_type != 0{
		glog.Error("未知的推送服务类型!")
		return nil
	}
	ts_config.ServerType = server_type
	dbres,err := g.DB().Table("ts_settings").Where("server=?",ts_config.ServerType).One()
	if err != nil {
		glog.Error(err)
		return nil
	}
	if err := dbres.ToStruct(ts_config);err != nil {
		glog.Error(err)
		return nil
	}
	glog.Println("推送配置",ts_config)
	return ts_config
}

/*
完整API由 (TS_SERVER+参数(appid)+TS_API_*)组成
*/
const (
	TS_SERVER   = "https://restapi.getui.com/v1/"
	TS_API_PUSH = "/push_single" //推送API末尾部分
	TS_API_SIGN = "/auth_sign"   //鉴权API末尾部分
	TS_API_SAVE = "/save_list_body"   //鉴权API末尾部分
	TS_API_LIST = "/push_list"   //鉴权API末尾部分
)

const (
	// 通知栏消息布局样式
	T_SYSTEM   = 0 //0:系统样式
	T_GETUI    = 1 //1:个推样式
	T_JUSTPIC  = 4 //4:纯图样式(背景图样式)
	T_FULLOPEN = 6 //6:展开通知样式
	//消息类型
	NOTYPOPLOAD  = "notypopload"  //下载模板
	NOTIFICATION = "notification" //普通消息模板
	LINK         = "link"         //链接模板
	TRANSMISSION = "transmission"

)

//模板结构参考第三方推送(个推)开发文档;http://docs.getui.com/getui/server/rest/template/
type TsTemplate struct {
	Message      `json:"message,omitempty"` //消息配置
	Transmission interface{}	`json:"transmission,omitempty"`//透传模板
	Notification interface{}                `json:"notification,omitempty"` //普通通知模板
	Link         interface{}                `json:"link,omitempty"`         //链接模板
	Notypopload  interface{}                `json:"notypopload,omitempty"`  //文件下载模板
	Cid          interface{}                     `json:"cid,omitempty"`          //客户端ID,与alias二选其一
	Alias        interface{}                     `json:"alias,omitempty"`        //客户端别名与cid二选其一
	Requestid    string                     `json:"requestid,omitempty"`    //请求唯一标识号
}

//通知栏消息布局样式
type Style struct {
	Type         int    `json:"type"`                   // 通知栏消息布局样式
	Text         interface{} `json:"text,omitempty"`         //通知内容
	Title        string `json:"title,omitempty"`        //通知标题
	Logo         string `json:"logo,omitempty"`         //通知的图标名称,包含后缀名（需要在客户端开发时嵌入），如“push.png”
	Logourl      string `json:"logourl,omitempty"`      //通知图标URL地址
	Is_ring      bool   `json:"is_ring,omitempty"`      //是否响铃,默认响铃
	Is_vibrate   bool   `json:"is_vibrate,omitempty"`   //是否震动,默认振动
	Is_clearable bool   `json:"is_clearable,omitempty"` //是否可清除,默认可清除
	Banner_url   string `json:"banner_url,omitempty"`   //通过url方式指定动态banner图片作为通知背景图,纯图类型通知
	Big_style    string `json:"big_style,omitempty"`    //通知展示样式,枚举值包括 1,2,3
	/*
		big_style	必传属性					展开样式说明
		1		big_image_url				通知展示大图样式，参数是大图的URL地址
		2		big_text					通知展示文本+长文本样式，参数是长文本
		3		big_image_url,banner_url	通知展示大图+小图样式，参数是大图URL和小图URL
	*/
	Big_image_url string `json:"big_image_url,omitempty"` //通知大图URL地址
	Big_text      string `json:"big_text,omitempty"`      //通知展示文本+长文本样式，参数是长文本

}

//简单通知模板
type Notification struct {
	Style                interface{} `json:"style,omitempty"`                //通知栏消息布局样式
	Transmission_type    bool        `json:"transmission_type,omitempty"`    //收到消息是否立即启动应用，true为立即启动，false则广播等待启动，默认是否
	Transmission_content string      `json:"transmission_content,omitempty"` //透传内容
	Duration_begin       string      `json:"duration_begin,omitempty"`       //设定展示开始时间，格式为yyyy-MM-dd HH:mm:ss
	Duration_end         string      `json:"duration_end,omitempty"`         //设定展示结束时间，格式为yyyy-MM-dd HH:mm:ss
}

//点开通知打开网页模板
type Link struct {
	Style          interface{} `json:"style,omitempty"`          //通知栏消息布局样式
	Duration_begin string      `json:"duration_begin,omitempty"` //设定展示开始时间，格式为yyyy-MM-dd HH:mm:ss
	Duration_end   string      `json:"duration_end,omitempty"`   //设定展示结束时间，格式为yyyy-MM-dd HH:mm:ss
	Url            string      `json:"url,omitempty"`            //打开网址;当使用link作为推送模板时，当客户收到通知时，在通知栏会下是一条含图标、标题等的通知，用户点击时，可以打开您指定的网页。
}

//透传模板
type Transmission struct{
	Transmission_Type interface{} `json:"transmission_type,omitempty"`
	Transmission_Content interface{} `json:"transmission_content,omitempty"`
	Duration_Begin string `json:"duration_begin,omitempty"`
	Duration_End string `json:"duration_end,omitempty"`
}

//点击通知弹窗下载模板
type Notypopload struct {
	Style          interface{} `json:"style,omitempty"`          //通知栏消息布局样式
	Notyicon       string      `json:"notyicon,omitempty"`       //是	通知栏图标
	Notytitle      string      `json:"notytitle,omitempty"`      //是	通知标题
	Notycontent    string      `json:"notycontent,omitempty"`    //是	通知内容
	Poptitle       string      `json:"poptitle,omitempty"`       //是	弹出框标题
	Popcontent     string      `json:"popcontent,omitempty"`     //是	弹出框内容
	Popimage       string      `json:"popimage,omitempty"`       //是	弹出框图标
	Popbutton1     string      `json:"popbutton1,omitempty"`     //是	弹出框左边按钮名称
	Popbutton2     string      `json:"popbutton2,omitempty"`     //是	弹出框右边按钮名称
	Loadicon       string      `json:"loadicon,omitempty"`       //否	现在图标
	Loadtitle      string      `json:"loadtitle,omitempty"`      //否	下载标题
	Loadurl        string      `json:"loadurl,omitempty"`        //是	下载文件地址
	Is_autoinstall bool        `json:"is_autoinstall,omitempty"` //否	是否自动安装，默认值false
	Is_actived     bool        `json:"is_actived,omitempty"`     //否	安装完成后是否自动启动应用程序，默认值false
	Androidmark    string      `json:"androidmark,omitempty"`    //否	安卓标识
	Symbianmark    string      `json:"symbianmark,omitempty"`    //否	塞班标识
	Iphonemark     string      `json:"iphonemark,omitempty"`     //否	苹果标志
	Duration_begin string      `json:"duration_begin,omitempty"` //否	设定展示开始时间，格式为yyyy-MM-dd HH:mm:ss
	Duration_end   string      `json:"duration_end,omitempty"`   //否	设定展示结束时间，格式为yyyy-MM-dd HH:mm:ss
}

//消息配置结构
type Message struct {
	Appkey              string `json:"appkey,omitempty"`              //注册应用时生成的appkey
	Is_offline          bool   `json:"is_offline,omitempty"`          //是否离线推送
	Offline_expire_time int    `json:"offline_expire_time,omitempty"` //消息离线存储有效期，单位：ms
	Push_network_type   int    `json:"push_network_type,omitempty"`   //选择推送消息使用网络类型，0：不限制，1：wifi
	Msgtype             string `json:"msgtype,omitempty"`             //消息应用类型，可选项：notification、link、notypopload、transmission
}


//多发
type PushList struct {
	Cid              []string `json:"cid,omitempty"`              //注册应用时生成的appkey
	Taskid          string   `json:"taskid,omitempty"`          //是否离线推送
	Need_detail bool    `json:"need_detail,omitempty"` //消息离线存储有效期，单位：ms
}

/*
	群消息
*/
func NewList(content string, msgtype ...string)TsTemplate{
	mtype := NOTIFICATION
	if len(msgtype) > 0 {
		mtype = msgtype[0]
	}
	ts_tmp := TsTemplate{
		Message: Message{
			Appkey:              ts_config.AppKey,
			Is_offline:          true,
			Offline_expire_time: 10000000,
			Msgtype:             mtype,
		},
	}
	ts_tmp.refresh(mtype,content)
	return ts_tmp
}
/*
	返回默认模板结构,可直接发送也可重构
*/
func New(cid, content interface{}, msgtype ...string) (TsTemplate, error) {
	mtype := NOTIFICATION
	if len(msgtype) > 0 {
		mtype = msgtype[0]
	}
	ts_tmp := TsTemplate{
		Message: Message{
			Appkey:              ts_config.AppKey,
			Is_offline:          true,
			Offline_expire_time: 10000000,
			Msgtype:             mtype,
		},
		Cid:       cid,
		Requestid: grand.Digits(grand.N(15, 30)),
	}
	ts_tmp.refresh(mtype,content)
	return ts_tmp, nil
}

func (ts_tmp *TsTemplate) refresh(mtype string,content interface{}){
	switch mtype {
	case NOTIFICATION:
		ts_tmp.Message.Msgtype = mtype
		ts_tmp.Notification = Notification{
			Style: Style{
				Type:         T_SYSTEM, //系统样式
				Text:         content,
				Title:        "分时住",
				Is_clearable: true, //可清除
				Is_ring:      true, //响铃
				Is_vibrate:   true, //震动
			},
			Transmission_type: true, //点击时默认打开应用
		}
	case LINK:
		ts_tmp.Message.Msgtype = mtype
		ts_tmp.Link = Link{
			Style: Style{
				Type:         T_SYSTEM, //系统样式
				Text:         content,
				Title:        "分时住",
				Is_clearable: true, //可清除
				Is_ring:      true, //响铃
				Is_vibrate:   true, //震动
			},
			Url: content.(string),
		}
	case NOTYPOPLOAD:
		ts_tmp.Message.Msgtype = mtype
		ts_tmp.Notypopload = Notypopload{
			Style: Style{
				Type:         T_SYSTEM, //系统样式
				Text:         "通知内容",
				Title:        "分时住",
				Is_clearable: true, //可清除
				Is_ring:      true, //响铃
				Is_vibrate:   true, //震动
			},
			Notyicon:       "noty.png",
			Notytitle:      "请填写通知标题",
			Notycontent:    "请填写通知内容",
			Poptitle:       "请填写弹出框标题",
			Popcontent:     "请填写弹出框内容",
			Popimage:       "image.png",
			Popbutton1:     "leftButton",
			Popbutton2:     "rightButton",
			Loadicon:       "",
			Loadtitle:      "请填写下载标题",
			Loadurl:        "请填写下载文件地址",
			Is_autoinstall: false,
			Is_actived:     false,
		}
	case TRANSMISSION:
		ts_tmp.Message.Msgtype = mtype
		ts_tmp.Transmission = Transmission{
			Transmission_Content:content,
			Transmission_Type:true,
		}
	default:
		glog.Error("未知的消息推送类型！:", mtype)
	}
}

/*
	发送前钩子初始参数准备
*/
func (tmp *TsTemplate)sendBeforeHook(server_type int)interface{}{
	ts_config = getTsConfig(server_type)
	if ts_config == nil {
		glog.Error("推送配置获取失败！")
		return nil
	}

	tmp.Appkey = ts_config.AppKey
	token_name := "auth_token_"+gconv.String(server_type)
	auth_token := gcache.Get(token_name)

	if auth_token == nil { //过期重新获取
		auth_token = getAuthToken(token_name)
		return auth_token
	}

	return auth_token
}

/*
	推送消息流程
	1.调用鉴权接口,获取auth_token
	2.调用推送接口
*/
func (tmp *TsTemplate) Send(server_type int) {

	auth_token := tmp.sendBeforeHook(server_type)
	if auth_token == nil {
		glog.Error("推送失败,token获取失败")
	}
	params, err := gjson.Encode(tmp)
	if err != nil {
		glog.Error("json解析失败！", err)
		return
	}

	//该URL为单推URL,如需群推查阅文档修改此处和对应消息模板内容即可
	url := TS_SERVER + ts_config.AppID + TS_API_PUSH
	http_client := ghttp.NewClient()
	http_client.SetHeader("Content-Type", "application/json;charset=utf-8")
	http_client.SetHeader("authtoken", gconv.String(auth_token))
	r, err := http_client.Post(url, params)
	if err != nil {
		glog.Error(err)
		return
	}
	defer r.Close()
	resp := gjson.New(r.ReadAll()).ToMap()
	if resp["result"] == "ok" {
		glog.Println("推送发送成功", resp)
	} else {
		glog.Error("推送发送失败", resp)
	}
}



/*TsTemplate
	推送消息流程
	1.调用鉴权接口,获取auth_token
	2.调用保存消息接口
	3.调用多发消息接口
*/
func (tmp *TsTemplate) SendList(cidList []string,server_type int) {
	auth_token := tmp.sendBeforeHook(server_type)
	if auth_token == nil {
		glog.Error("推送失败,token获取失败")
	}
	params, err := gjson.Encode(tmp)
	if err != nil {
		glog.Error("json解析失败！", err)
		return
	}
	//该URL为多推URL,
	url := TS_SERVER + ts_config.AppID + TS_API_SAVE
	http_client := ghttp.NewClient()
	http_client.SetHeader("Content-Type", "application/json;charset=utf-8")
	http_client.SetHeader("authtoken", gconv.String(auth_token))
	r, err := http_client.Post(url, params)
	if err != nil {
		glog.Error(err)
		return
	}
	defer r.Close()
	resp := gjson.New(r.ReadAll()).ToMap()
	if resp["result"] == "ok" {
		pushList := &PushList{
			Cid:cidList,
			Taskid:resp["taskid"].(string),
			Need_detail:true,
		}
		params, err := gjson.Encode(pushList)
		if err != nil {
			glog.Error("json解析失败！", err)
			return
		}

		url := TS_SERVER + ts_config.AppID + TS_API_LIST
		r, err := http_client.Post(url, params)
		if err != nil {
			glog.Error(err)
			return
		}
		resp := gjson.New(r.ReadAll()).ToMap()
		if resp["result"] == "ok" {
			glog.Println("推送发送成功", resp)
		}else{
			glog.Println("推送发送失败", resp)
		}

	} else {
		glog.Error("推送消息保存失败", resp)
	}

}

/*
	用途:推送鉴权
	功能:auth_token获取
	调用第三方推送API
	变量:url,params,resp

*/
func getAuthToken(token_name string) interface{} {
	url := TS_SERVER + ts_config.AppID + TS_API_SIGN
	http_client := ghttp.NewClient()
	http_client.SetHeader("Content-Type", "application/json;charset=utf-8")

	timestamp := gtime.Now().TimestampMilli()
	sign_byte := sha256.Sum256([]byte(ts_config.AppKey + gconv.String(timestamp) + ts_config.MasterSecret))
	sign := fmt.Sprintf("%x", sign_byte)

	//构建鉴权参数
	params := map[string]interface{}{
		"sign":      sign,
		"timestamp": timestamp,
		"appkey":    ts_config.AppKey,
	}
	glog.Println("鉴权地址",url)
	glog.Println("鉴权参数",params)
	//参数JSON序列化
	json_params, _ := gjson.New(params).ToJson()

	r, err := ghttp.NewClient().Post(url, json_params)
	if err != nil {
		glog.Error("鉴权请求失败", err)
		return nil
	}
	defer r.Close()
	//成功结果{"result":"ok","expire_time":"1568875773358","auth_token":"98b74e881a76dd5bbc98ec6cab8e650dfa33f24b1313507ba19c8947a320d7f5"}
	resp := gjson.New(r.ReadAll()).ToMap()

	if resp["result"] == "ok" { //存入缓存存到过期时间前1秒
		gcache.Set(token_name, resp["auth_token"], gtime.D-1)
		return resp["auth_token"]
	}
	glog.Error(resp["result"])
	return nil

}

/*
	获取ClientID,该值由APP客户端通过集成第三方推送插件后获取客户机的唯一标识后传给后台与用户信息进行绑定并存储。
*/
func GetCid(uuid interface{}) ([]interface{}, error) {
	mu := gmutex.New()
	mu.Lock()
	defer mu.Unlock()
	g.DB().SetDebug(true)
	dbres, err := g.DB().Table("device_bind").Where("user_uuid IN (?)",uuid).FindAll()
	if err != nil && err != sql.ErrNoRows {
		glog.Error(err)
		return nil, err
	}
	if dbres == nil {
		glog.Error(err)
		return nil, errors.New("未查询到该用户关联的设备id")
	}
	glog.Println("查询值",dbres.List())
	cid := make([]interface{},0)
	for _,v :=range dbres.List(){
		if !tslimit(v["user_uuid"]){//超1天不在线，不发推送
			continue
		}
		cid = append(cid,v["cid"])
	}
	return cid,nil
}


/*
	推送限制
	用户离线或者很长时间未响应服务器,取消推送
*/
func tslimit(uuid interface{}) bool {
	last_oper_time,err := g.DB().Table("heartbeat").Where("user_uuid=?",uuid).FindValue("update_time")
	if err != nil && err != sql.ErrNoRows {
		glog.Error(err)
		return true
	}
	oper_time := last_oper_time.String()
	if oper_time != ""{
		if gtime.Now().After(gtime.NewFromStr(oper_time).Add(gtime.D*1)){
			//当前用户已离线一天
			return false
		}
	}
	return true
}
