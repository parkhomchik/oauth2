package model

import "github.com/jinzhu/gorm"

type Scope struct {
	//ID   int    `gorm:"primary_key"`
	gorm.Model
	Name string `gorm:"not null;unique"`
}
