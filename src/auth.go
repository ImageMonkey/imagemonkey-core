package main

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"errors"
	log "github.com/Sirupsen/logrus"
	"strings"
	imagemonkeydb "./database"
)

type SessionInformation struct {
	Username string
	LoggedIn bool
	IsModerator bool
}

type AccessTokenInfo struct {
	Valid bool
	Token string
	Username string
	Empty bool
}

type APITokenInfo struct {
	Valid bool
	Token string
	Username string
	Empty bool
}


func _strToToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok { //is algorithm correctly set?
	    	log.Debug("unexcpected signing method")
	    	return nil, errors.New("Unexpected signing method")
		}

	    // hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
	    return []byte(JWT_SECRET), nil
	})

	return token, err
}

func _isAccessTokenRevoked(accessToken string) bool {
	if accessTokenExists(accessToken) {
		return false
	}

	return true
}

func _parseAccessToken(tokenString string) AccessTokenInfo {
	var accessTokenInfo AccessTokenInfo
	accessTokenInfo.Username = ""
	accessTokenInfo.Token = ""
	accessTokenInfo.Valid = false

	if tokenString == "" {
		accessTokenInfo.Empty = true
	} else {
		accessTokenInfo.Empty = false
	}


	token, err := _strToToken(tokenString)

	if err == nil && token.Valid {
		//token is valid and signed by the backend, check now if the token was revoked
		//or if it is still valid

		if !_isAccessTokenRevoked(tokenString) { //still valid - not revoked
			accessTokenInfo.Valid = true
			accessTokenInfo.Token = tokenString
			accessTokenInfo.Username = token.Claims.(jwt.MapClaims)["username"].(string)
		}
	}

	return accessTokenInfo
}

func _parseApiToken(tokenString string) APITokenInfo {
	var apiTokenInfo APITokenInfo
	apiTokenInfo.Username = ""
	apiTokenInfo.Token = ""
	apiTokenInfo.Valid = false

	if tokenString == "" {
		apiTokenInfo.Empty = true
	} else {
		apiTokenInfo.Empty = false
	}

	token, err := _strToToken(tokenString)

	if err == nil && token.Valid {
		//token is valid and signed by the backend, check now if the token was revoked
		//or if it is still valid

		revoked, _ := isApiTokenRevoked(tokenString) 
		if !revoked { //still valid - not revoked
			apiTokenInfo.Valid = true
			apiTokenInfo.Token = tokenString
			apiTokenInfo.Username = token.Claims.(jwt.MapClaims)["username"].(string)
		}
	}

	return apiTokenInfo
}

type AuthTokenHandlerInterface interface {
    GetSessionInformation() SessionInformation
}

type SessionCookieHandler struct {
	db *imagemonkeydb.ImageMonkeyDatabase
}

func NewSessionCookieHandler(db *imagemonkeydb.ImageMonkeyDatabase) *SessionCookieHandler {
    return &SessionCookieHandler{
    	db: db,
    } 
}

func (p *SessionCookieHandler) GetSessionInformation(c *gin.Context) SessionInformation {
	var sessionInformation SessionInformation
	sessionInformation.LoggedIn = false
	sessionInformation.Username = ""
	sessionInformation.IsModerator = false

	cookie, err := c.Request.Cookie("imagemonkey")

    if err == nil {
    	tokenString := cookie.Value
    	if tokenString != "" {
    		accessTokenInfo := _parseAccessToken(tokenString)
    		sessionInformation.LoggedIn = accessTokenInfo.Valid
    		sessionInformation.Username = accessTokenInfo.Username

    		if sessionInformation.Username != "" {
    			userInfo, err := p.db.GetUserInfo(sessionInformation.Username)
    			if err != nil {
    				sessionInformation.IsModerator = false
    			} else {
    				sessionInformation.IsModerator = userInfo.IsModerator
    			}
    		}

    	}
    }

	return sessionInformation
}


type AuthTokenHandler struct {
}

func NewAuthTokenHandler() *AuthTokenHandler {
    return &AuthTokenHandler{
    } 
}

func (p *AuthTokenHandler) GetAccessTokenInfo(c *gin.Context) AccessTokenInfo {
	auth := strings.SplitN(c.Request.Header.Get("Authorization"), " ", 2)

    if len(auth) != 2 || auth[0] != "Bearer" {
    	var accessTokenInfo AccessTokenInfo
		accessTokenInfo.Username = ""
		accessTokenInfo.Token = ""
		accessTokenInfo.Valid = false
		accessTokenInfo.Empty = true
    	return accessTokenInfo
   	}

   	return _parseAccessToken(auth[1])
}

func (p *AuthTokenHandler) GetAccessTokenInfoFromUrl(c *gin.Context) AccessTokenInfo {
	token := getParamFromUrlParams(c, "token", "")

    if token == "" {
    	var accessTokenInfo AccessTokenInfo
		accessTokenInfo.Username = ""
		accessTokenInfo.Token = ""
		accessTokenInfo.Valid = false
		accessTokenInfo.Empty = true
    	return accessTokenInfo
   	}

   	return _parseAccessToken(token)
}

func (p *AuthTokenHandler) GetAPITokenInfo(c *gin.Context) APITokenInfo {
	apiToken := c.Request.Header.Get("X-Api-Token")

	return _parseApiToken(apiToken)
}