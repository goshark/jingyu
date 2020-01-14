package sms

import (
	"errors"
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/encoding/gjson"
	"github.com/gogf/gf/os/gcache"
	"github.com/gogf/gf/os/glog"
	"github.com/gogf/gf/util/gconv"
	"math/rand"
	"time"
)

//模板编码
const (
	//LOGIN_CODE = "SMS_181540808"
	REG_CODE = "SMS_181540805"
	SETPWD_CODE = "SMS_181540808"
)


const (
	SIGN  = "景煜旅游"
	//缓存前缀
	LOGIN_TYPE = "login_code_"
	LICENSE_TYPE = "license_code_"
	SETPWD_TYPE = "resetpwd_code_"
	REG_TYPE = "reg_code_"
)

type SmsSendApi struct{
	AccessKey_ID string//ali的key
	Access_Key_Secret string//ali 的 secret
	VerificationCode string//短信验证码
	EffectiveTime int//验证码有效时间  单位/分钟 默认1
	Sms_Tpl_Code string//模板编码
	Sms_Tpl_Param map[string]interface{}//模板参数JSON字符串
}


func New() *SmsSendApi{
	return &SmsSendApi{
		AccessKey_ID:"LTAI4FhoXWs4p1V1bN5kkkSe",
		Access_Key_Secret:"g326dltQZZC5NjwMSO9wQsOExsO1ll",
		EffectiveTime:1,
		VerificationCode:"000000",
	}
}

/*
	发送消息到手机
*/
func SendSmsMsg(mobile,msg interface{}){
	SMS := New()
	SMS.Sms_Tpl_Code = gconv.String(msg)
	SMS.Sms_Tpl_Param = g.Map{}
	SMS.sendSmsCode(gconv.String(mobile))
}



/*
	code_type,验证码类型 1注册,2登录,3认证
	mobile 手机号
*/
func (SMS *SmsSendApi) Smssend(code_type int,mobile string) error{
	//验证码类型
	ct := getCodeType(code_type)
	if ct == ""{
		return errors.New("未知的验证码类型！"+gconv.String(code_type))
	}
	if gcache.Contains(ct) {//查缓存
		return errors.New("验证码发送频繁,请稍后再试！")
	}

	SMS.EffectiveTime = 1
	SMS.VerificationCode = CreateCaptcha()
	if err := SMS.setTpl(code_type);err != nil {//设置模板
		return err
	}
	if err := SMS.sendSmsCode(mobile);err != nil {//发送短信
		return err
	}
	//缓存1分钟

	gcache.Set(ct+mobile, SMS.VerificationCode, gconv.Duration(SMS.EffectiveTime*1000*60))
	return nil


}

func (SMS *SmsSendApi) setTpl(code_type int) error{
	SMS.Sms_Tpl_Param = g.Map{"code":SMS.VerificationCode}
	if code_type == 1{//reg
		SMS.Sms_Tpl_Code = REG_CODE
	}else if code_type == 2{
		SMS.Sms_Tpl_Code = SETPWD_CODE
	}else{
		return errors.New("Unkown code_type of SMS!")
	}
	return nil
}

func getCodeType(code_type int)string{
	switch code_type {
	case 1:
		return REG_TYPE
	case 2:
		return SETPWD_TYPE
	default:
		return ""
	}
}


/*
	验证码验证
*/

func CheckedSmsCode(code_type,mobile,code string) error{
	key := code_type+mobile
	if gcache.Contains(key) {
		cache_code := gcache.Get(key)
		if cache_code.(string) == code{
			//gcache.Remove(key) //验证通过清除
			return nil
		}else{
			return errors.New("验证码错误,请重新输入！")
		}
	}else{
		return errors.New("验证码已过期,请重新获取！")
	}

}

/*
 	生成验证码
*/
func CreateCaptcha() string {
	return fmt.Sprintf("%06v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(1000000))
}


/*
	发送
*/
func (SMS *SmsSendApi) sendSmsCode(mobile string)error{
	client, err := dysmsapi.NewClientWithAccessKey("cn-hangzhou", SMS.AccessKey_ID, SMS.Access_Key_Secret)
	request := dysmsapi.CreateSendSmsRequest()
	request.Scheme = "https"
	request.PhoneNumbers = mobile
	request.SignName = SIGN
	request.TemplateCode = SMS.Sms_Tpl_Code
	param,_ := gjson.Encode(SMS.Sms_Tpl_Param)
	request.TemplateParam = string(param)
	response, err := client.SendSms(request)
	if err != nil {
		return err
	}

	glog.Println("ali-resp",response)
	if response.Code != "OK"{
		return errors.New(response.Message)
	}
	return nil
}


