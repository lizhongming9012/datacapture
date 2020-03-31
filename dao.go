package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"log"
)

var db *gorm.DB

type Zsk struct {
	ID     uint   `gorm:"primary_key"`
	Zlcode string `json:"zlcode"`
	Zlflag string `json:"zlflag"`
	Title  string `json:"title" gorm:"size:300000;COMMENT:'标题'"`
	Type   string `json:"type" gorm:"size:300000;COMMENT:'所属类别'"`
	Key    string `json:"key" gorm:"COMMENT:'关键字'"`
	Uptime string `json:"uptime" gorm:"COMMENT:'更新日期'"`
	Answer string `json:"answer" gorm:"size:300000;COMMENT:'问题解答'"`
}

func SetupDataBase() {
	sdb, err := gorm.Open(c.Dbinfo.Type,
		fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local",
			c.Dbinfo.User,
			c.Dbinfo.Pwd,
			c.Dbinfo.Uri,
			c.Dbinfo.DataName))
	if err != nil {
		log.Fatalf("SetupDataBase err: %v", err)
	}

	sdb.SingularTable(true)
	sdb.DB().SetMaxIdleConns(10)
	sdb.DB().SetMaxOpenConns(100)

	if !sdb.HasTable("zsk") {
		sdb.CreateTable(Zsk{})
	} else {
		sdb.AutoMigrate(Zsk{})
	}
	db = sdb
}

func IsDataExist(zlcode string) bool {
	var zsk Zsk
	if err := db.Table("zsk").Where("zlcode=?", zlcode).First(&zsk).Error; err != nil {
		return false
	}
	return true
}

func IsNewData(zlcode, uptime string) bool {
	var zsk Zsk
	if err := db.Table("zsk").Where("zlcode=? and uptime=?", zlcode, uptime).First(&zsk).Error; err != nil {
		return false
	}
	return true
}

func AddData(data interface{}) error {
	if err := db.Table("zsk").Create(data).Error; err != nil {
		return err
	}
	return nil
}

func DeleteData(zlcode string) error {
	if err := db.Table("zsk").Where("zlcode=?", zlcode).Delete(Zsk{}).Error; err != nil {
		return err
	}
	return nil
}

func DeleteNullData() error {
	if err := db.Table("zsk").Where("answer like '%单元格编号:%' or answer = '详见附件'").
		Delete(Zsk{}).Error; err != nil {
		return err
	}
	return nil
}
