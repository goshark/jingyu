package voice

import (
	"errors"
	"github.com/gogf/gf/encoding/gjson"
	"github.com/gogf/gf/encoding/gurl"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/gcache"
	"github.com/gogf/gf/os/glog"
	"github.com/gogf/gf/util/gconv"
	"github.com/gogf/gf/util/grand"
	"net"
)

const (
	Token_Url = "https://openapi.baidu.com/oauth/2.0/token"
	Api_Key = "U3lQjrvuxLtjTdVGbeN2VN1"
	Secret_Key= "ipCn3knW6uAFc8jdwV6NQGhxwS7XYOa"
	Convert_Url = "https://tsn.baidu.com/text2audio"
	//App_ID = 17545700
)
func mac() string{
	// 获取本机的MAC地址
	interfaces, err := net.Interfaces()
	if err != nil {
		glog.Error(err)
		return ""
	}
	for _, inter := range interfaces {

		mac := inter.HardwareAddr //获取本机MAC地址
		return mac.String()
	}
	return ""
}
func getToken() interface{}{
	client := ghttp.NewClient()
	r,err := client.Post(Token_Url,ghttp.BuildParams(g.Map{
		"grant_type":"client_credentials",
		"client_id":Api_Key,
		"client_secret":Secret_Key,
	}))

	if err != nil {
	glog.Error(err)
	return nil
	}
	defer r.Close()
	resmap := gjson.New(r.ReadAll()).ToMap()
	gcache.Set("voice_token",resmap["access_token"],gconv.Duration(resmap["expires_in"]))
	return resmap["access_token"]
	/*token 格式
		{
			"access_token":"24.b74c910a9b79e0e09c81f748a16e75fd.2592000.1573873257.282335-17545700",
			"expires_in":2592000,
			"refresh_token":"25.b7893774ae170a860aaea801e00967d4.315360000.1886641257.282335-17545700",
			"scope":"audio_voice_assistant_get brain_enhanced_asr audio_tts_post public brain_all_scope picchain_test_picchain_api_scope wise_adapt lebo_resource_base lightservice_public hetu_basic lightcms_map_poi kaidian_kaidian ApsMisTest_Test权限 vis-classify_flower lpq_开放 cop_helloScope ApsMis_fangdi_permission smartapp_snsapi_base iop_autocar oauth_tp_app smartapp_smart_game_openapi oauth_sessionkey smartapp_swanid_verify smartapp_opensource_openapi smartapp_opensource_recapi fake_face_detect_开放Scope",
			"session_key":"9mzdWW2H1tCdeM6MCEY8Q7+rqw2QfGmYlBQBWXXrzXerdrQ+XhPc8l83pYLTKXyu3IKDqfYBAVBAFihBjK2DpYn6F164A==",
			"session_secret":"e4ed786fb22605555d00423d99063dae"
		}
	*/
}
func TextToSpeech(text interface{}) (*ghttp.ClientResponse,error){
	if text == nil {
		return nil,errors.New("参数不能为空")
	}
	token := gcache.Get("voice_token")
	if token == nil {
		token = getToken()
		if token == nil {
			glog.Error("获取token 失败！")
			return nil,errors.New("获取token 失败！")
		}
	}

	params := ghttp.BuildParams(
		g.Map{
			"tex":gurl.Encode(gconv.String(text)),
			"lan":"zh",
			"cuid":grand.Digits(5)+mac(),
			"tok":token,
			"ctp":1,
			"per":4,
		})
	client := ghttp.NewClient()
	r,err := client.Post(Convert_Url,params)
	if err != nil {
		glog.Error(err)
		return nil,err
	}

	return r,nil
}


