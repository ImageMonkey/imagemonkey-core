package imagemonkeydb

import (
	"errors"
	datastructures "github.com/bbernhard/imagemonkey-core/datastructures"
	"github.com/getsentry/raven-go"
	log "github.com/sirupsen/logrus"
	"time"
	"context"
)

func (p *ImageMonkeyDatabase) GetUserInfo(username string) (datastructures.UserInfo, error) {
	var userInfo datastructures.UserInfo
	var removeLabelPermission bool = false
	var unlockImageDescriptionPermission bool = false
	var unlockImage bool = false
	var canMonitorSystem bool = false
	var canAcceptTrendingLabel bool = false
	var canAccessPgStat bool = false

	userInfo.Name = ""
	userInfo.Created = 0
	userInfo.ProfilePicture = ""
	userInfo.IsModerator = false
	userInfo.Permissions = nil

	rows, err := p.db.Query(context.TODO(),
                             `SELECT a.name, COALESCE(a.profile_picture, ''), a.created, a.is_moderator,
                              COALESCE(p.can_remove_label, false) as remove_label_permission,
                              COALESCE(p.can_unlock_image_description, false) as unlock_image_description,
                              COALESCE(p.can_unlock_image, false) as unlock_image,
                              COALESCE(p.can_monitor_system, false) as can_monitor_system,
							  COALESCE(p.can_accept_trending_label, false) as can_accept_trending_label,
							  COALESCE(p.can_access_pg_stat, false) as can_access_pg_stat
                              FROM account a 
                              LEFT JOIN account_permission p ON p.account_id = a.id 
                              WHERE a.name = $1`, username)
	if err != nil {
		log.Error("[User Info] Couldn't get user info: ", err.Error())
		raven.CaptureError(err, nil)
		return userInfo, err
	}

	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&userInfo.Name, &userInfo.ProfilePicture, &userInfo.Created,
			&userInfo.IsModerator, &removeLabelPermission,
			&unlockImageDescriptionPermission, &unlockImage, &canMonitorSystem, &canAcceptTrendingLabel, 
			&canAccessPgStat)

		if err != nil {
			log.Error("[User Info] Couldn't scan user info: ", err.Error())
			raven.CaptureError(err, nil)
			return userInfo, err
		}

		if userInfo.IsModerator {
			permissions := &datastructures.UserPermissions{CanRemoveLabel: removeLabelPermission,
				CanUnlockImageDescription: unlockImageDescriptionPermission,
				CanUnlockImage:            unlockImage,
				CanMonitorSystem:          canMonitorSystem,
				CanAcceptTrendingLabel:    canAcceptTrendingLabel,
				CanAccessPgStat:           canAccessPgStat,
			}
			userInfo.Permissions = permissions
		}
	}

	return userInfo, nil
}

func (p *ImageMonkeyDatabase) UserExists(username string) (bool, error) {
	var numOfExistingUsers int32
	err := p.db.QueryRow(context.TODO(), "SELECT count(*) FROM account u WHERE u.name = $1", username).Scan(&numOfExistingUsers)
	if err != nil {
		log.Error("[User exists] Couldn't get num of existing users: ", err.Error())
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
	err := p.db.QueryRow(context.TODO(), "SELECT count(*) FROM account u WHERE u.email = $1", email).Scan(&numOfExistingUsers)
	if err != nil {
		log.Error("[Email exists] Couldn't get num of existing users: ", err.Error())
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
	err := p.db.QueryRow(context.TODO(), "SELECT hashed_password FROM account u WHERE u.name = $1", username).Scan(&hashedPassword)
	if err != nil {
		log.Error("[Hashed Password] Couldn't get hashed password for user: ", err.Error())
		raven.CaptureError(err, nil)
		return "", err
	}

	return hashedPassword, nil
}

func (p *ImageMonkeyDatabase) CreateUser(username string, hashedPassword []byte, email string) error {
	var insertedId int64

	type DefaultImageCollection struct {
		Name        string
		Description string
	}

	created := int64(time.Now().Unix())

	defaultImageCollections := []DefaultImageCollection{
		DefaultImageCollection{Name: MyDonations, Description: "My donations"},
		DefaultImageCollection{Name: MyOpenTasks, Description: "My open tasks"},
	}

	tx, err := p.db.Begin(context.TODO())
	if err != nil {
		log.Error("[Creating User] Couldn't begin transaction: ", err.Error())
		raven.CaptureError(err, nil)
		return err
	}

	insertedId = 0
	err = tx.QueryRow(context.TODO(),
                       `INSERT INTO account(name, hashed_password, email, created, is_moderator) 
                        VALUES($1, $2, $3, $4, $5) RETURNING id`,
		username, hashedPassword, email, created, false).Scan(&insertedId)
	if err != nil {
		tx.Rollback(context.TODO())
		log.Error("[Creating User] Couldn't create user: ", err.Error())
		raven.CaptureError(err, nil)
		return err
	}

	if insertedId == 0 {
		tx.Rollback(context.TODO())
		return errors.New("nothing inserted")
	}

	for _, defaultImageCollection := range defaultImageCollections {
		err = p._addImageCollectionInTransaction(tx, username, defaultImageCollection.Name, defaultImageCollection.Description)
		if err != nil {
			//transaction already rolled back
			log.Error("[Creating User] Couldn't add default image collection: ", err.Error())
			raven.CaptureError(err, nil)
			return err
		}
	}

	err = tx.Commit(context.TODO())
	if err != nil {
		log.Error("[Creating User] Couldn't commit transaction: ", err.Error())
		raven.CaptureError(err, nil)
		return err
	}

	return nil
}

func (p *ImageMonkeyDatabase) ChangeProfilePicture(username string, uuid string) (string, error) {
	var existingProfilePicture string

	existingProfilePicture = ""

	tx, err := p.db.Begin(context.TODO())
	if err != nil {
		log.Error("[Change Profile Picture] Couldn't begin transaction: ", err.Error())
		raven.CaptureError(err, nil)
		return existingProfilePicture, err
	}

	err = tx.QueryRow(context.TODO(),
                       `SELECT COALESCE(a.profile_picture, '') FROM account a WHERE a.name = $1`, username).Scan(&existingProfilePicture)
	if err != nil {
		log.Error("[Change Profile Picture] Couldn't get existing profile picture: ", err.Error())
		raven.CaptureError(err, nil)
		return existingProfilePicture, err
	}

	_, err = tx.Exec(context.TODO(),
                       `UPDATE account SET profile_picture = $1 WHERE name = $2`, uuid, username)
	if err != nil {
		log.Error("[Change Profile Picture] Couldn't change profile picture: ", err.Error())
		raven.CaptureError(err, nil)
		return existingProfilePicture, err
	}

	err = tx.Commit(context.TODO())
	if err != nil {
		log.Error("[Change Profile Picture] Couldn't commit transaction: ", err.Error())
		raven.CaptureError(err, nil)
		return existingProfilePicture, err
	}

	return existingProfilePicture, nil
}
