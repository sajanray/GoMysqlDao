package models

import (
	"fmt"
	"github.com/sajanray/GoMysqlDao"
	"sync"
)

var (
	userInstance *User
	onceUser     sync.Once
)

type User struct {
	GoMysqlDao.MysqlDao
}

func GetUserInstance() *User {
	onceUser.Do(func() {
		fmt.Println("初始化了User Model")
		userInstance = &User{
			GoMysqlDao.MysqlDao{
				Pk:        "id",
				TableName: "test",
			},
		}
	})
	return userInstance
}
