package db

import (
	"fmt"
	"strings"

	"github.com/parkhomchik/oauth2/model"
	uuid "github.com/satori/go.uuid"
)

func (dbm *DBManager) GetUserByID(id uuid.UUID) (model.User, error) {
	var user model.User
	err := dbm.DB.Where("id = ?", id).Find(&user).Error
	return user, err
}

func (dbm *DBManager) GetUsers() (users []model.User) {
	dbm.DB.Find(&users)
	return
}

func (dbm *DBManager) GetScopesUser(id uuid.UUID) []model.Scope {
	var userScopeID []model.UserScopes
	var scopes []model.Scope
	dbm.DB.Where("user_id = ?", id).Find(&userScopeID)
	var scopeIDs []uuid.UUID
	for _, uid := range userScopeID {
		scopeIDs = append(scopeIDs, uid.ScopeID)
	}
	dbm.DB.Where("id in (?)", scopeIDs).Find(&scopes)
	return scopes
}

func (dbm *DBManager) UserScope(userID uuid.UUID, role string) error {
	var user model.User
	dbm.DB.Where("ID = ?", userID).First(&user)
	var scope model.Scope
	dbm.DB.Where("name = ?", role).First(&scope)
	var userScope model.UserScopes
	return dbm.DB.Where("user_id = ? AND scope_id = ?", user.ID, scope.ID).First(&userScope).Error
}

func (dbm *DBManager) RegistrationUser(user model.User, roles string) (model.User, error) {
	if err := dbm.DB.Create(&user).Error; err != nil {
		return user, err
	}
	if roles != "" {
		var userScope model.UserScopes
		scopes := strings.Split(roles, ",")
		for _, s := range scopes {
			scope, err := dbm.GetScopeByName(s)
			if err == nil {
				userScope.UserID = user.ID
				userScope.ScopeID = scope.ID
				err = dbm.PostScopesUser(userScope)
				if err != nil {
					fmt.Println("Error user post scope: ", err)
				}
			}
		}
	}
	/*
		initPortalDB()
		var staff model.Staff
		staff.Name = user.Name

		err := PortalDb.Create(&staff).Error
		PortalDb.Close()
	*/
	return user, nil
}

func (dbm *DBManager) SetUserInfo(userInf model.User) (model.User, error) {
	var user model.User
	if err := dbm.DB.Where("id = ?", userInf.ID).First(&user).Error; err != nil {
		return user, err
	}
	user = userInf
	if err := dbm.DB.Save(&userInf).Error; err != nil {
		return user, err
	}
	return user, nil
}

func (dbm *DBManager) DeleteUser(user model.User) error {
	return dbm.DB.Delete(&user).Error
}

func (dbm *DBManager) DeleteUserScope(user model.User) error {
	var userScopes model.UserScopes
	if err := dbm.DB.Where("user_id = ?", user.ID).First(&userScopes).Error; err != nil {
		return err
	}
	return dbm.DB.Where("user_id = ?", userScopes.UserID).Delete(&userScopes).Error
}

func (dbm *DBManager) Login(username, password string) (model.User, error) {
	var user model.User
	err := dbm.DB.Where("login = ? AND password = ?", username, password).First(&user).Error
	return user, err
}

func (dbm *DBManager) PostScopesUser(userScope model.UserScopes) error {
	return dbm.DB.Save(&userScope).Error
}
