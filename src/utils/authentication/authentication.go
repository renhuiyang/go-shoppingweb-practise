package authentication

import (
	"github.com/coopernurse/gorp"
	jwt "github.com/dgrijalva/jwt-go"
	//"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"net/http"
	"utils/redis"
	"time"
	. "utils/types"
	"github.com/golang/glog"
)

const (
	expireOffset = 3600
	secretKey    = "HGJKUYXCQWSDFDSPPMCNV"
)

func GetParsKey(token *jwt.Token) (interface{}, error) {
	return []byte(secretKey), nil
}

func GenerateToken(userTel string) (string, error) {
	token := jwt.New(jwt.GetSigningMethod("HS256"))
	token.Claims["exp"] = time.Now().Add(time.Hour * time.Duration(600)).Unix()
	token.Claims["iat"] = time.Now().Unix()
	token.Claims["sub"] = userTel
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		panic(err)
		return "", err
	}
	return tokenString, nil
}

func Authenticate(user * User, dbmap *gorp.DbMap) bool {
	var db_user User
	err := dbmap.SelectOne(&db_user, "SELECT * FROM CUSTOMER WHERE CUS_TEL=?", user.PhoneNumber)
	//count,err := dbmap.SelectInt("SELECT count(*) FROM CUSTOMER WHERE CUS_TEL=?",user.PhoneNumber)
	if err != nil {
		return false
	}
    glog.V(4).Infof("User Info:%v",db_user)
	return user.Password == db_user.Password
}
func getTokenRemainingValidity(timestamp interface{}) int {
	if validity, ok := timestamp.(float64); ok {
		tm := time.Unix(int64(validity), 0)
		remainer := tm.Sub(time.Now())
		if remainer > 0 {
			return int(remainer.Seconds() + expireOffset)
		}
	}
	return expireOffset
}

func Logout(tokenString string, token *jwt.Token) error {
	redisConn := redis.Connect()
	return redisConn.SetValue(tokenString, tokenString, getTokenRemainingValidity(token.Claims["exp"]))
}

func RequireTokenAuthentication(r render.Render, req *http.Request) {
	token, err := jwt.Parse(req.Header.Get("Authorization"), GetParsKey)

	if err == nil && token.Valid && !IsInBlacklist(req.Header.Get("Authorization")){
		return
	} else {
		r.JSON(http.StatusUnauthorized, map[string]string{"message": "Authen Fail"})
	}
}
func IsInBlacklist(token string) bool {
	redisConn := redis.Connect()
	redisToken, _ := redisConn.GetValue(token)

	if redisToken == nil {
		return false
	}

	return true
}


