package db

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/parkhomchik/oauth2/model"
)

//DBManager для связывания методов БД
type DBManager struct {
	DB       *gorm.DB
	PortalDB *gorm.DB
}

func (dbm *DBManager) InitDB() {
	var configuration model.Configuration
	configuration.Load()
	dbinfo := fmt.Sprintf("host=%s user=%s dbname=%s password=%s sslmode=disable", configuration.DbHost, configuration.DbUser, configuration.DbName, configuration.DbPass)
	var err error
	dbm.DB, err = gorm.Open("postgres", dbinfo)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	dbm.DB.LogMode(true)
	dbm.DB.AutoMigrate(&model.User{}, &model.Client{}, &model.Scope{})
}

func (dbm *DBManager) InitPortalDB() {
	var configuration model.Configuration
	configuration.Load()
	dbinfo := fmt.Sprintf("host=%s user=%s dbname=%s password=%s sslmode=disable", configuration.DbPortalHost, configuration.DbPortalUser, configuration.DbPortalName, configuration.DbPortalPass)
	var err error
	dbm.PortalDB, err = gorm.Open("postgres", dbinfo)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	dbm.PortalDB.LogMode(true)
}