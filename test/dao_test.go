package test

import (
	"fmt"
	"github.com/sajanray/GoMysqlDao"
	"github.com/sajanray/GoMysqlDao/test/models"
	"math/rand"
	"testing"
)

func init() {
	fmt.Println("初始化数据库连接")
	GoMysqlDao.InitMysqlPool(confFile)
}

type Test struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	Mobile     string `json:"mobile"`
	CreateTime string `json:"create_time"`
}

func TestModel(t *testing.T) {
	test := models.GetTestInstance()
	where := test.BuildWhere("id=1")
	one, err := test.One(where, "*", nil)
	fmt.Println("再次调用GetInstance")
	test2 := models.GetUserInstance()
	fmt.Println(test2)

	if err != nil {
		t.Fatal(err.Error())
		return
	} else {
		t.Log(one)
	}

}

// 快接构建where条件
func TestBuildwhere(t *testing.T) {
	test := models.GetTestInstance()
	where := test.BuildWhere(
		":limit", "1",
		"id", ">=", 1,
		":order", "id DESC",
		"id=1")
	one, err := test.One(where, "*", nil)
	if err != nil {
		t.Fatal(err.Error())
		return
	} else {
		t.Log(one)
	}
}

// 插入测试
func TestInsert(t *testing.T) {
	num := rand.Intn(10000) + 1000
	dao := GoMysqlDao.MysqlDao{}
	dao.LocalConnectPool = GoMysqlDao.NewMysqlPool(confFile)
	dao.TableName = "test"
	data := make(map[string]interface{}, 0)
	data["name"] = fmt.Sprintf("zhang_%d", num)
	data["mobile"] = fmt.Sprintf("1381039%d", num)
	id, err := dao.Insert(data)
	if err != nil {
		t.Fatal(err.Error())
	} else {
		t.Logf("id:%d", id)
	}
}

// 查询测试
func TestOne(t *testing.T) {
	dao := GoMysqlDao.MysqlDao{}
	dao.LocalConnectPool = GoMysqlDao.NewMysqlPool(confFile)
	dao.TableName = "test"
	where := GoMysqlDao.NewMysqlWhereColl()
	where.SetTableName("test")
	where.Add("id", ">", 1)

	test := new(Test)
	_, err := dao.One(where, "*", test)
	if err != nil {
		t.Fatal(err.Error())
	} else {
		t.Logf("%+v", test)
	}
}

func TestOne2(t *testing.T) {
	dao := GoMysqlDao.MysqlDao{}
	dao.LocalConnectPool = GoMysqlDao.NewMysqlPool(confFile)
	dao.TableName = "test"
	where := GoMysqlDao.NewMysqlWhereColl()
	where.SetTableName("test")
	where.Add(":sql", "select a.* from test as a")
	where.Add("a.id", ">", 1)

	test := new(Test)
	_, err := dao.One(where, "", test)
	if err != nil {
		t.Fatal(err.Error())
	} else {
		t.Logf("%+v", test)
	}
}

// 构建SQL测试
func TestBuildSql(t *testing.T) {
	dao := GoMysqlDao.MysqlDao{}
	dao.TableName = "test"

	where := GoMysqlDao.NewMysqlWhereColl()
	where.SetTableName("test")
	where.Add("id", 1)
	where.Add("age", ">", 20)
	where.Add("name", "=", "zhangshan")
	where.Add("words", "LIKE", "%中%")
	where.Add(":where", "create_time Between 2025 AND 2026")
	where.Add(":group", "status")
	where.Add(":order", "status")
	where.Add(":limit", "10,20")
	where.Add(":having", "c>1")
	sql, _, err := dao.BuildSelectSql(where, "*")
	if err != nil {
		t.Fatal(err.Error())
	} else {
		t.Logf(sql)
	}
}

// 多条查询测试
func TestMore(t *testing.T) {
	dao := GoMysqlDao.MysqlDao{}
	dao.LocalConnectPool = GoMysqlDao.NewMysqlPool(confFile)
	dao.TableName = "test"
	where := GoMysqlDao.NewMysqlWhereColl()
	where.Add(":limit", 2)

	test := new(Test)
	rows, err := dao.More(where, "*", test)
	if err != nil {
		t.Fatal(err.Error())
	} else {
		t.Logf("%+v", rows)
	}
}

// 更新测试
func TestUpdate(t *testing.T) {
	where := GoMysqlDao.NewMysqlWhereColl()
	where.Add("id", 1)

	up := GoMysqlDao.NewMysqlWhereColl()
	up.Add("name", "lisi")
	up.Add("mobile", "13989898787")

	dao := GoMysqlDao.MysqlDao{
		TableName: "test",
	}

	update, err := dao.Update(up, where)
	if err != nil {
		return
	} else {
		t.Logf("影响行数：%+v", update)
	}
}

// 删除测试
func TestDelete(t *testing.T) {
	where := GoMysqlDao.NewMysqlWhereColl()
	where.Add("id", 0)

	dao := GoMysqlDao.MysqlDao{
		TableName: "test",
	}

	update, err := dao.Delete(where)
	if err != nil {
		return
	} else {
		t.Logf("影响行数：%+v", update)
	}
}

// exec测试
func TestExec(t *testing.T) {
	dao := GoMysqlDao.MysqlDao{
		TableName: "test",
	}
	sql := "UPDATE test SET name='zhangshan' WHERE id=1"
	exec, err := dao.Exec(&sql, nil)
	if err != nil {
		return
	} else {
		t.Log(exec)
	}
}

func TestQuery(t *testing.T) {
	dao := GoMysqlDao.MysqlDao{}
	params := make([]interface{}, 0)
	params = append(params, 1)
	sql := "SELECT * FROM test WHERE id=?"
	query, err := dao.Query(&sql, params, nil, false)
	if err != nil {
		return
	} else {
		t.Log(query)
	}
}

// 测试事务
func TestTransaction(t *testing.T) {
	dao := GoMysqlDao.MysqlDao{
		TableName: "test",
	}
	defer func(dao *GoMysqlDao.MysqlDao) {
		if err := recover(); err != nil {
			fmt.Println(err)
			dao.Rollback()
		}
		dao.Commit()
	}(&dao)

	dao.Begin()
	sql1 := "UPDATE test SET name='zhangshan11' WHERE id=1"
	sql2 := "UPDATE test SET name='zhangshan22' WHERE id=2"
	exec, err := dao.Exec(&sql1, nil)
	exec, err = dao.Exec(&sql2, nil)
	if err != nil {
		return
	} else {
		t.Log(exec)
	}
}
