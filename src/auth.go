package main

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"errors"
	log "github.com/sirupsen/logrus"
	"strings"
	imagemonkeydb "github.com/bbernhard/imagemonkey-core/database"
	commons "github.com/bbernhard/imagemonkey-core/commons"
	datastructures "github.com/bbernhard/imagemonkey-core/datastructures"
)

type SessionInformation struct {
	Username string
	LoggedIn bool
	IsModerator bool
	UserPermissions *datastructures.UserPermissions `json:"permissions,omitempty"`
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


func _strToToken(tokenString string, jwtSecret string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok { //is algorithm correctly set?
	    	log.Debug("unexcpected signing method")
	    	return nil, errors.New("Unexpected signing method")
		}

	    // hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
	    return []byte(jwtSecret), nil
	})

	return token, err
}

func _isAccessTokenRevoked(db *imagemonkeydb.ImageMonkeyDatabase, accessToken string) bool {
	if db.AccessTokenExists(accessToken) {
		return false
	}

	return true
}

func _parseAccessToken(db *imagemonkeydb.ImageMonkeyDatabase, tokenString string, jwtSecret string) AccessTokenInfo {
	var accessTokenInfo AccessTokenInfo
	accessTokenInfo.Username = ""
	accessTokenInfo.Token = ""
	accessTokenInfo.Valid = false

	if tokenString == "" {
		accessTokenInfo.Empty = true
	} else {
		accessTokenInfo.Empty = false
	}


	token, err := _strToToken(tokenString, jwtSecret)

	if err == nil && token.Valid {
		//token is valid and signed by the backend, check now if the token was revoked
		//or if it is still valid

		if !_isAccessTokenRevoked(db, tokenString) { //still valid - not revoked
			accessTokenInfo.Valid = true
			accessTokenInfo.Token = tokenString
			accessTokenInfo.Username = token.Claims.(jwt.MapClaims)["username"].(string)
		}
	}

	return accessTokenInfo
}

func _parseApiToken(db *imagemonkeydb.ImageMonkeyDatabase, tokenString string, jwtSecret string) APITokenInfo {
	var apiTokenInfo APITokenInfo
	apiTokenInfo.Username = ""
	apiTokenInfo.Token = ""
	apiTokenInfo.Valid = false

	if tokenString == "" {
		apiTokenInfo.Empty = true
	} else {
		apiTokenInfo.Empty = false
	}

	token, err := _strToToken(tokenString, jwtSecret)

	if err == nil && token.Valid {
		//token is valid and signed by the backend, check now if the token was revoked
		//or if it is still valid

		revoked, err := db.IsApiTokenRevoked(tokenString) 
		if err == nil && !revoked { //still valid - not revoked
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
	jwtSecret string
}

func NewSessionCookieHandler(db *imagemonkeydb.ImageMonkeyDatabase, jwtSecret string) *SessionCookieHandler {
    return &SessionCookieHandler{
    	db: db,
		jwtSecret: jwtSecret,
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
    		accessTokenInfo := _parseAccessToken(p.db, tokenString, p.jwtSecret)
    		sessionInformation.LoggedIn = accessTokenInfo.Valid
    		sessionInformation.Username = accessTokenInfo.Username

    		if sessionInformation.Username != "" {
    			userInfo, err := p.db.GetUserInfo(sessionInformation.Username)
    			if err != nil {
    				sessionInformation.IsModerator = false
    				sessionInformation.UserPermissions = &datastructures.UserPermissions{CanRemoveLabel: false,
    													   				   			 CanUnlockImageDescription: false,
    													   				   			 CanUnlockImage: false,
    													   				   			 CanMonitorSystem: false,
																					 CanAcceptTrendingLabel: false,
    													  				 			}
    			} else {
    				sessionInformation.IsModerator = userInfo.IsModerator
    				sessionInformation.UserPermissions = userInfo.Permissions
    			}
    		}

    	}
    }

	return sessionInformation
}


type AuthTokenHandler struct {
	db *imagemonkeydb.ImageMonkeyDatabase
	jwtSecret string
}

func NewAuthTokenHandler(db *imagemonkeydb.ImageMonkeyDatabase, jwtSecret string) *AuthTokenHandler {
    return &AuthTokenHandler{
    	db: db,
		jwtSecret: jwtSecret,
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

   	return _parseAccessToken(p.db, auth[1], p.jwtSecret)
}

func (p *AuthTokenHandler) GetAccessTokenInfoFromUrl(c *gin.Context) AccessTokenInfo {
	token := commons.GetParamFromUrlParams(c, "token", "")

    if token == "" {
    	var accessTokenInfo AccessTokenInfo
		accessTokenInfo.Username = ""
		accessTokenInfo.Token = ""
		accessTokenInfo.Valid = false
		accessTokenInfo.Empty = true
    	return accessTokenInfo
   	}

   	return _parseAccessToken(p.db, token, p.jwtSecret)
}

func (p *AuthTokenHandler) GetAPITokenInfo(c *gin.Context) APITokenInfo {
	apiToken := c.Request.Header.Get("X-Api-Token")

	return _parseApiToken(p.db, apiToken, p.jwtSecret)
}
