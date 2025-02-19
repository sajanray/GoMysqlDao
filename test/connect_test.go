package test

import (
	"github.com/sajanray/GoMysqlDao"
	"testing"
)

var confFile = "./db.conf"

// 测试数据库连接
func TestInitMysqlPool(t *testing.T) {
	GoMysqlDao.InitMysqlPool(confFile)
	if ret, err := GoMysqlDao.GlobalConnectPool.IsConnectedWrite(); ret {
		t.Logf("数据库连接成功")
		//GoMysqlDao.GlobalConnectPool.Free()
	} else {
		t.Fatal("数据库连接失败:", err)
	}
}

func TestNewMysqlPool(t *testing.T) {
	pool := GoMysqlDao.NewMysqlPool(confFile)
	if ret, err := pool.IsConnectedWrite(); ret {
		t.Logf("数据库连接成功")
		pool.Free()
	} else {
		t.Fatal("数据库连接失败:", err)
	}
}
