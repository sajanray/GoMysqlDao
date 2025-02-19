package test

import (
	"github.com/sajanray/GoMysqlDao/config"
	"testing"
)

var confFile = "../db.conf"

func TestParseDbConf(t *testing.T) {
	_, err := config.ParseDbConf(confFile)
	if err != nil {
		t.Fatal("解析配置文件失败:", err.Error())
	} else {
		t.Logf("解析配置文件成功")
	}
}
