package model

import "github.com/jinzhu/gorm"

type Client struct {
	//ID     int `gorm:"primary_key"`
	gorm.Model
	Secret string
	Domain string
	UserID string
	Scope  []Scope `gorm:"many2many:client_scopes;"`
}
