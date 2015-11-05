package authService

import (
	"utils/authentication"
	"github.com/coopernurse/gorp"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/binding"
	"github.com/martini-contrib/render"
	"net/http"
	"strconv"
	. "utils/types"
)

func PostLogin(r render.Render, dbmap *gorp.DbMap, res http.ResponseWriter, u User, e binding.Errors) {
	if authentication.Authenticate(&u, dbmap) {
		token, err := authentication.GenerateToken(strconv.FormatInt(u.PhoneNumber, 10))
		if err != nil {
			r.JSON(http.StatusBadRequest, map[string]string{"message": "Generate Toker for user fail"})
			return
		} else {
			r.JSON(200, map[string]string{"token": token})
			return
		}
	}

	r.JSON(http.StatusBadRequest, map[string]string{"message": "User not exists!"})
}

func GetToken(r render.Render, dbmap *gorp.DbMap, res http.ResponseWriter, u User, e binding.Errors) {
	token, err := authentication.GenerateToken(strconv.FormatInt(u.PhoneNumber, 10))
	if err != nil {
		r.JSON(http.StatusBadRequest, map[string]string{"message": "Generate Toker for user fail"})
		return
	} else {
		r.JSON(200, map[string]string{"message": token})
		return
	}
}

func Logout(r render.Render, params martini.Params, req *http.Request) {
	token, err := jwt.Parse(req.Header.Get("Authorization"), authentication.GetParsKey)

	if err == nil && token.Valid {
		tokenString := req.Header.Get("Authorization")
		authentication.Logout(tokenString, token)
	}

	r.JSON(200, map[string]string{"message": "SUCCESS"})
}
