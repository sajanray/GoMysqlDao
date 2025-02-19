package config

import (
	"errors"
	"fmt"
	JTStools "github.com/sajanray/GoJsonToStruct"
	"os"
)

// 配置文件
type DatabaseConf struct {
	Mysql MysqlMasterSlave `json:"mysql"`
	Redis RedisMasterSlave `json:"redis"`
}

// ParseDbConf 解析数据库配置文件
func ParseDbConf(filePath string) (dbConf DatabaseConf, err error) {
	//读取配置文件
	confStr, err := os.ReadFile(filePath)
	if err != nil {
		return
	}

	//转换成结构体
	m := JTStools.NewMapToStruct()
	m.Transform(&dbConf, string(confStr))
	if !m.Success {
		err = errors.New(fmt.Sprint("parse config file", filePath, "failed:", m.GetErrmsg()))
		return
	}

	//从库公共部分合并到每个节点
	combineMysqlConf(&dbConf)
	combineRedisConf(&dbConf)
	return
}
