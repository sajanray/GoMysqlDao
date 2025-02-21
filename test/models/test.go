package models

import (
	"github.com/sajanray/GoMysqlDao"
	"sync"
)

// 单例模式 保证全局只有一个实例
var (
	testInstance *Test
	onceTest     sync.Once
)

// Test 模型
type Test struct {
	GoMysqlDao.MysqlDao // 继承MysqlDao
}

func GetGlobalScope() GoMysqlDao.GlobalScopeConf {
	where := GoMysqlDao.NewMysqlWhereColl()
	where.Add("mobile", "!=", "")
	where.Add("create_time", ">", "2024-12-12")
	where.WhereJoinStr = "OR"
	data := make(map[string]interface{})
	data["create_time"] = "2025-01-01 00:00:00"
	return GoMysqlDao.GlobalScopeConf{
		WithGlobalScopeWhere:       true,
		GlobalScopeWhere:           where,
		GlobalScopeInsertField:     data,
		WithGlobalScopeInsertField: true,
	}
}

// GetTestInstance 获取Test模型实例
func GetTestInstance() *Test {
	onceTest.Do(func() {
		testInstance = &Test{
			GoMysqlDao.MysqlDao{
				Pk:              "id",
				TableName:       "test",
				GlobalScopeConf: GetGlobalScope(),
			},
		}
	})
	return testInstance
}
