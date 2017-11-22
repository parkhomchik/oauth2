package model

import "github.com/jinzhu/gorm"

type User struct {
	//ID       int `gorm:"primary_key"`
	gorm.Model
	Name     string
	Login    string `gorm:"not null;unique"`
	Password string
	Scopes   []Scope `gorm:"many2many:user_scopes;"`
}
