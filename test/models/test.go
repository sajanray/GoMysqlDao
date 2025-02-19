package models

import (
	"fmt"
	"github.com/sajanray/GoMysqlDao"
	"sync"
)

var (
	testInstance *Test
	onceTest     sync.Once
)

type Test struct {
	GoMysqlDao.MysqlDao
}

func GetTestInstance() *Test {
	onceTest.Do(func() {
		fmt.Println("初始化了Test Model")
		testInstance = &Test{
			GoMysqlDao.MysqlDao{
				Pk:        "id",
				TableName: "test",
			},
		}
	})
	return testInstance
}
