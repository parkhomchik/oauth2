package db

import "github.com/parkhomchik/oauth2/model"

func (dbm *DBManager) GetScopeByName(name string) (scope model.Scope, err error) {
	err = dbm.DB.Where("name = ?", name).Find(&scope).Error
	return
}
