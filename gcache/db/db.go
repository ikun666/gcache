package db

import (
	"log"

	"github.com/ikun666/gcache/conf"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Student struct {
	gorm.Model
	Name  string `json:"name"`
	Score string `json:"score"`
}

var DB *gorm.DB

func Init() {
	var err error
	DB, err = gorm.Open(mysql.Open(conf.GConfig.DNS), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}

	err = DB.AutoMigrate(&Student{})
	if err != nil {
		log.Fatalln(err)
	}
}
