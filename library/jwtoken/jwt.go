package jwtoken

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gogf/gf/os/glog"
	"github.com/gogf/gf/os/gtime"
)

const(
	SIGNED_KEY = "dc_jingyu_key"
)


//func GenToken() (string,error){
//	// Create a new token object, specifying signing method and the claims
//	// you would like it to contain.
//	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
//		"foo": "bar",
//		"nbf": time.Date(2015, 10, 10, 12, 0, 0, 0, time.UTC).Unix(),
//	})
//
//	// Sign and get the complete encoded token as a string using the secret
//	tokenString, err := token.SignedString([]byte(SIGNED_KEY))
//
//	return tokenString,err
//}



//获取签名算法为HS256的token
func GetHStoken(arg map[string]interface{}) string {
	claims := jwt.MapClaims{
		"exp":gtime.Now().Add(gtime.D*1).Unix(),
		"iat":gtime.Now().Unix(),
	}
	for k,v := range arg{
		claims[k] = v
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	//加密算法是HS256时，这里的SignedString必须是[]byte（）类型
	ss, err := token.SignedString([]byte(SIGNED_KEY))
	if err != nil {
		glog.Error("token生成签名错误,err=%v", err)
		return ""
	}
	return ss
}



//解析签名算法为HS256的token
func ParseHStoken(tokenString string) (jwt.MapClaims,error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(SIGNED_KEY),nil
	})
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims,nil
	} else {
		glog.Error(err)
		return nil,err
	}

}
