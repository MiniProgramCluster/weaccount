package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"weaccount/internal/conf"
	"weaccount/utils/db"
	"weaccount/utils/log"

	"github.com/golang-jwt/jwt/v5"
)

type WxLoginResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

func AuthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var reqBody struct {
		AppID string `json:"appid"`
		Code  string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.Logger().Error().Err(err).Msg("Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	logCtx := log.Logger().With().Str("appid", reqBody.AppID).Logger()
	appConf := conf.App(reqBody.AppID)
	if appConf == nil {
		logCtx.Error().Msg("appid not found")
		http.Error(w, "Invalid appid", http.StatusUnauthorized)
		return
	}
	var wxResp WxLoginResponse
	wxLoginURL := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		appConf.AppID, appConf.AppSecret, reqBody.Code)
	resp, err := http.Get(wxLoginURL)
	if err != nil {
		logCtx.Error().Err(err).Msg("Failed to connect to WeChat API")
		http.Error(w, "Failed to connect to WeChat API", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logCtx.Error().Err(err).Msg("Failed to read response")
		http.Error(w, "Failed to read response", http.StatusInternalServerError)
		return
	}
	if errx := json.Unmarshal(body, &wxResp); errx != nil {
		logCtx.Error().Err(errx).Msg("Failed to parse WeChat response")
		http.Error(w, "Failed to parse WeChat response", http.StatusInternalServerError)
		return
	}
	if wxResp.ErrCode != 0 {
		logCtx.Error().Int("errcode", wxResp.ErrCode).Str("errmsg", wxResp.ErrMsg).Msg("WeChat login failed")
		http.Error(w, wxResp.ErrMsg, http.StatusUnauthorized)
		return
	}
	openID := wxResp.OpenID
	unionID := wxResp.UnionID
	sessionKey := wxResp.SessionKey
	logCtx = logCtx.With().Str("openid", openID).Str("unionid", unionID).Logger()
	sql := "insert into users (appid, openid, unionid, session_key) values (?, ?, ?, ?) on duplicate key update session_key = ?"
	_, err = db.Instance().Exec(sql, appConf.AppID, openID, unionID, sessionKey, sessionKey)
	if err != nil {
		logCtx.Error().Err(err).Msg("Failed to insert user")
		http.Error(w, "Failed to insert user", http.StatusInternalServerError)
		return
	}
	var uid uint64
	err = db.Instance().QueryRow("select uid from users where openid = ?", openID).Scan(&uid)
	if err != nil {
		logCtx.Error().Err(err).Msg("Failed to select user")
		http.Error(w, "Failed to select user", http.StatusInternalServerError)
		return
	}

	var rspBody struct {
		Token           string `json:"token"`
		ReferralCode    string `json:"referralCode"`
		TokenExpireTime int64  `json:"tokenExpireTime"`
	}
	rspBody.TokenExpireTime = conf.Token().LifeTime + time.Now().Unix()
	var tokenClamis struct {
		UID   uint64 `json:"uid"`
		AppID string `json:"appid"`
		Type  string `json:"type"`
		jwt.RegisteredClaims
	}
	tokenClamis.UID = uid
	tokenClamis.AppID = appConf.AppID
	tokenClamis.Type = "auth"
	tokenClamis.ExpiresAt = jwt.NewNumericDate(time.Unix(rspBody.TokenExpireTime, 0))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClamis)
	rspBody.Token, err = token.SignedString([]byte(conf.Token().Secret))
	if err != nil {
		logCtx.Error().Err(err).Msg("Failed to sign token")
		http.Error(w, "Failed to sign token", http.StatusInternalServerError)
		return
	}
	rspBody.ReferralCode = fmt.Sprintf("%d", uid)
	json.NewEncoder(w).Encode(rspBody)
}
