package imagemonkeydb

import (
    "github.com/getsentry/raven-go"
    log "github.com/sirupsen/logrus"
    "errors"
)

func (p *ImageMonkeyDatabase) AddAccessToken(username string, accessToken string, expirationTime int64) error {
    var insertedId int64

    insertedId = 0
    err := p.db.QueryRow(`INSERT INTO access_token(user_id, token, expiration_time)
                        SELECT id, $2, $3 FROM account a WHERE a.name = $1 RETURNING id`, username, accessToken, expirationTime).Scan(&insertedId)
    if err != nil {
        log.Debug("[Add Access Token] Couldn't add access token: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    if insertedId == 0 {
        log.Debug("[Add Access Token] Nothing inserted")
        return errors.New("Nothing inserted")
    }

    return nil
}

func (p *ImageMonkeyDatabase) RemoveAccessToken(accessToken string) error {
    _, err := p.db.Exec(`DELETE FROM access_token WHERE token = $1`, accessToken)
    if err != nil {
        log.Debug("[Remove Access Token] Couldn't remove access token: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    return nil
}

func (p *ImageMonkeyDatabase) AccessTokenExists(accessToken string) bool {
    var numOfAccessTokens int32

    numOfAccessTokens = 0
    err := p.db.QueryRow("SELECT count(*) FROM access_token WHERE token = $1", accessToken).Scan(&numOfAccessTokens)
    if err != nil {
        log.Debug("[Add Access Token] Couldn't add access token: ", err.Error())
        raven.CaptureError(err, nil)
        return false
    }

    if numOfAccessTokens == 0 {
        return false
    }

    return true
}
