package imagemonkeydb

import (
    "../datastructures"
    "github.com/getsentry/raven-go"
    log "github.com/Sirupsen/logrus"
    "time"
    "errors"
)

func (p *ImageMonkeyDatabase) GetUserInfo(username string) (datastructures.UserInfo, error) {
    var userInfo datastructures.UserInfo
    var removeLabelPermission bool
    var unlockImageDescriptionPermission bool
    var unlockImage bool
    removeLabelPermission = false
    unlockImageDescriptionPermission = false
    unlockImage = false


    userInfo.Name = ""
    userInfo.Created = 0
    userInfo.ProfilePicture = ""
    userInfo.IsModerator = false
    userInfo.Permissions = nil

    rows, err := p.db.Query(`SELECT a.name, COALESCE(a.profile_picture, ''), a.created, a.is_moderator,
                              COALESCE(p.can_remove_label, false) as remove_label_permission,
                              COALESCE(p.can_unlock_image_description, false) as unlock_image_description,
                              COALESCE(p.can_unlock_image, false) as unlock_image
                              FROM account a 
                              LEFT JOIN account_permission p ON p.account_id = a.id 
                              WHERE a.name = $1`, username)
    if err != nil {
        log.Debug("[User Info] Couldn't get user info: ", err.Error())
        raven.CaptureError(err, nil)
        return userInfo, err
    }

    defer rows.Close()

    if rows.Next() {
        err = rows.Scan(&userInfo.Name, &userInfo.ProfilePicture, &userInfo.Created, 
                        &userInfo.IsModerator, &removeLabelPermission, 
                        &unlockImageDescriptionPermission, &unlockImage)

        if err != nil {
            log.Debug("[User Info] Couldn't scan user info: ", err.Error())
            raven.CaptureError(err, nil)
            return userInfo, err
        }

        if userInfo.IsModerator {
            permissions := &datastructures.UserPermissions {CanRemoveLabel: removeLabelPermission, 
                                                            CanUnlockImageDescription: unlockImageDescriptionPermission,
                                                            CanUnlockImage: unlockImage}
            userInfo.Permissions = permissions
        }
    }

    return userInfo, nil
}

func (p *ImageMonkeyDatabase) UserExists(username string) (bool, error) {
    var numOfExistingUsers int32
    err := p.db.QueryRow("SELECT count(*) FROM account u WHERE u.name = $1", username).Scan(&numOfExistingUsers)
    if err != nil {
        log.Debug("[User exists] Couldn't get num of existing users: ", err.Error())
        raven.CaptureError(err, nil)
        return false, err
    }

    if numOfExistingUsers > 0 {
        return true, nil
    }
    return false, nil
}

func (p *ImageMonkeyDatabase) EmailExists(email string) (bool, error) {
    var numOfExistingUsers int32
    err := p.db.QueryRow("SELECT count(*) FROM account u WHERE u.email = $1", email).Scan(&numOfExistingUsers)
    if err != nil {
        log.Debug("[Email exists] Couldn't get num of existing users: ", err.Error())
        raven.CaptureError(err, nil)
        return false, err
    }

    if numOfExistingUsers > 0 {
        return true, nil
    }
    return false, nil
}

func (p *ImageMonkeyDatabase) GetHashedPasswordForUser(username string) (string, error) {
    var hashedPassword string
    err := p.db.QueryRow("SELECT hashed_password FROM account u WHERE u.name = $1", username).Scan(&hashedPassword)
    if err != nil {
        log.Debug("[Hashed Password] Couldn't get hashed password for user: ", err.Error())
        raven.CaptureError(err, nil)
        return "", err
    }

    return hashedPassword, nil
}

func (p *ImageMonkeyDatabase) CreateUser(username string, hashedPassword []byte, email string) error {
    var insertedId int64

    created := int64(time.Now().Unix())

    insertedId = 0
    err := p.db.QueryRow(`INSERT INTO account(name, hashed_password, email, created, is_moderator) 
                        VALUES($1, $2, $3, $4, $5) RETURNING id`, 
                        username, hashedPassword, email, created, false).Scan(&insertedId)
    if err != nil {
        log.Debug("[Creating User] Couldn't create user: ", err.Error())
        raven.CaptureError(err, nil)
        return err
    }

    if insertedId == 0 {
        return errors.New("nothing inserted")
    }

    return nil
}

func (p *ImageMonkeyDatabase) ChangeProfilePicture(username string, uuid string) (string, error) {
    var existingProfilePicture string

    existingProfilePicture = ""

    tx, err := p.db.Begin()
    if err != nil {
        log.Debug("[Change Profile Picture] Couldn't begin transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return existingProfilePicture, err
    }

    err = tx.QueryRow(`SELECT COALESCE(a.profile_picture, '') FROM account a WHERE a.name = $1`, username).Scan(&existingProfilePicture)
    if err != nil {
        log.Debug("[Change Profile Picture] Couldn't get existing profile picture: ", err.Error())
        raven.CaptureError(err, nil)
        return existingProfilePicture, err
    }

    _, err = tx.Exec(`UPDATE account SET profile_picture = $1 WHERE name = $2`, uuid, username)
    if err != nil {
        log.Debug("[Change Profile Picture] Couldn't change profile picture: ", err.Error())
        raven.CaptureError(err, nil)
        return existingProfilePicture, err
    }


    err = tx.Commit()
    if err != nil {
        log.Debug("[Change Profile Picture] Couldn't commit transaction: ", err.Error())
        raven.CaptureError(err, nil)
        return existingProfilePicture, err
    }

    return existingProfilePicture, nil
}