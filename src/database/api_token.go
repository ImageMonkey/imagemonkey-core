package imagemonkeydb

import (
    "github.com/getsentry/raven-go"
    log "github.com/Sirupsen/logrus"
    "github.com/dgrijalva/jwt-go"
    "../datastructures"
    "time"
    "errors"
)

func (p *ImageMonkeyDatabase) GetApiTokens(username string) ([]datastructures.APIToken, error) {
    var apiTokens []datastructures.APIToken
    rows, err := p.db.Query(`SELECT token, issued_at, description, revoked 
                           FROM api_token a
                           JOIN account a1 ON a1.id = a.account_id
                           WHERE a1.name = $1`, username)
    if err != nil {
        log.Debug("[Get API Tokens] Couldn't get rows: ", err.Error())
        raven.CaptureError(err, nil)
        return apiTokens, err
    }

    defer rows.Close() 

    for rows.Next() {
        var apiToken datastructures.APIToken
        err = rows.Scan(&apiToken.Token, &apiToken.IssuedAtUnixTimestamp, &apiToken.Description, &apiToken.Revoked)
        if err != nil {
            log.Debug("[Get API Tokens] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return apiTokens, err
        }

        apiTokens = append(apiTokens, apiToken)
    }

    return apiTokens, nil
}

func (p *ImageMonkeyDatabase) IsApiTokenRevoked(token string) (bool, error) {
    var revoked bool = false
    rows, err := p.db.Query("SELECT revoked FROM api_token WHERE token = $1", token)
    if err != nil {
        log.Error("[Is API Token revoked] Couldn't determine whether API token is revoked: ", err.Error())
        raven.CaptureError(err, nil)
        return false, err
    }
    defer rows.Close()

    if rows.Next() {
        err = rows.Scan(&revoked)
        if err != nil {
            log.Error("[Is API Token revoked] Couldn't scan row: ", err.Error())
            raven.CaptureError(err, nil)
            return false, err
        }

        return revoked, nil
    }

    return revoked, errors.New("[Is API Token revoked] Invalid result set")
}

func (p *ImageMonkeyDatabase) GenerateApiToken(jwtSecret string, username string, description string) (datastructures.APIToken, error) {
    type MyCustomClaims struct {
        Username string `json:"username"`
        Created int64 `json:"created"`
        jwt.StandardClaims
    }

    var apiToken datastructures.APIToken

    issuedAt := time.Now()
    expiresAt := issuedAt.Add(time.Hour * 24 * 365 * 10) //10 years

    claims := MyCustomClaims {
                  username,
                  issuedAt.Unix(),
                  jwt.StandardClaims{
                        ExpiresAt: expiresAt.Unix(),
                        Issuer: "imagemonkey-api",
                  },
              }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

    tokenString, err := token.SignedString([]byte(jwtSecret))
    if err != nil {
        return apiToken, err
    }


    _, err = p.db.Exec(`INSERT INTO api_token(account_id, issued_at, description, revoked, token, expires_at)
                        SELECT id, $2, $3, $4, $5, $6 FROM account WHERE name = $1`, 
                        username, issuedAt.Unix(), description, false, tokenString, expiresAt.Unix())
    if err != nil {
        log.Debug("[Generate API Token] Couldn't insert entry: ", err.Error())
        raven.CaptureError(err, nil)
        return apiToken, err
    }

    apiToken.Description = description
    apiToken.Token = tokenString
    apiToken.IssuedAtUnixTimestamp = issuedAt.Unix()

    return apiToken, nil
}

func (p *ImageMonkeyDatabase) RevokeApiToken(username string, apiToken string) (bool, error) {
    var modifiedId int64
    err := p.db.QueryRow(`UPDATE api_token AS a 
                       SET revoked = true
                       FROM account AS acc 
                       WHERE acc.id = a.account_id AND acc.name = $1 AND a.token = $2
                       RETURNING a.id`, username, apiToken).Scan(&modifiedId)
    if err != nil {
        log.Debug("[Revoke API Token] Couldn't revoke token: ", err.Error())
        raven.CaptureError(err, nil)
        return false, err
    }

    return true, nil
}