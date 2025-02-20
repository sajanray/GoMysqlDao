package test

import (
	"github.com/sajanray/GoMysqlDao"
	"testing"
)

func TestParseWhere1(t *testing.T) {
	where := GoMysqlDao.NewMysqlWhereColl()
	where.Debug = true
	where.Add("name", "tom")
	w := where.ParseWhere()

	r1 := w.SqlWHere[0] != "(name = ?)"
	r2 := w.Param[0] != "tom"
	r3 := w.SqlWhereToString("") != "(name = ?)"
	if r1 || r2 || r3 {
		t.Fatal("解析 = 失败")
	} else {
		t.Logf("解析 = 成功")
	}
}

func TestParseWhere2(t *testing.T) {
	where := GoMysqlDao.NewMysqlWhereColl()
	where.Debug = true
	where.Add("name", "=", "tom")
	w := where.ParseWhere()

	r1 := w.SqlWHere[0] != "(name = ?)"
	r2 := w.Param[0] != "tom"
	r3 := w.SqlWhereToString("") != "(name = ?)"
	if r1 || r2 || r3 {
		t.Fatal("解析 = 失败")
	} else {
		t.Logf("解析 = 成功")
	}
}

func TestParseWhere3(t *testing.T) {
	where := GoMysqlDao.NewMysqlWhereColl()
	where.Debug = true
	where.Add("name", 0)
	w := where.ParseWhere()

	r1 := w.SqlWHere[0] != "(name = ?)"
	r2 := w.Param[0] != 0
	r3 := w.SqlWhereToString("") != "(name = ?)"
	if r1 || r2 || r3 {
		t.Fatal("解析 = 失败")
	} else {
		t.Logf("解析 = 成功")
	}
}

func TestParseWhere4(t *testing.T) {
	where := GoMysqlDao.NewMysqlWhereColl()
	where.Debug = true
	where.Add("name", "=", 1)
	w := where.ParseWhere()

	r1 := w.SqlWHere[0] != "(name = ?)"
	r2 := w.Param[0] != 1
	r3 := w.SqlWhereToString("") != "(name = ?)"
	if r1 || r2 || r3 {
		t.Fatal("解析 = 失败")
	} else {
		t.Logf("解析 = 成功")
	}
}

func TestParseWhere5(t *testing.T) {
	where := GoMysqlDao.NewMysqlWhereColl()
	where.Debug = true
	where.Add("age", ">", 18)
	w := where.ParseWhere()

	r1 := w.SqlWHere[0] != "(age > ?)"
	r2 := w.Param[0] != 18
	r3 := w.SqlWhereToString("") != "(age > ?)"
	if r1 || r2 || r3 {
		t.Fatal("解析 > 失败")
	} else {
		t.Logf("解析 > 成功")
	}
}

func TestParseWhere6(t *testing.T) {
	where := GoMysqlDao.NewMysqlWhereColl()
	where.Debug = true
	where.Add("age", ">=", 18)
	w := where.ParseWhere()

	r1 := w.SqlWHere[0] != "(age >= ?)"
	r2 := w.Param[0] != 18
	r3 := w.SqlWhereToString("") != "(age >= ?)"
	if r1 || r2 || r3 {
		t.Fatal("解析 >= 失败")
	} else {
		t.Logf("解析 >= 成功")
	}
}

func TestParseWhere7(t *testing.T) {
	where := GoMysqlDao.NewMysqlWhereColl()
	where.Debug = true
	where.Add("age", "<", 18)
	w := where.ParseWhere()

	r1 := w.SqlWHere[0] != "(age < ?)"
	r2 := w.Param[0] != 18
	r3 := w.SqlWhereToString("") != "(age < ?)"
	if r1 || r2 || r3 {
		t.Fatal("解析 < 失败")
	} else {
		t.Logf("解析 < 成功")
	}
}

func TestParseWhere8(t *testing.T) {
	where := GoMysqlDao.NewMysqlWhereColl()
	where.Debug = true
	where.Add("age", "<=", 18)
	w := where.ParseWhere()

	r1 := w.SqlWHere[0] != "(age <= ?)"
	r2 := w.Param[0] != 18
	r3 := w.SqlWhereToString("") != "(age <= ?)"
	if r1 || r2 || r3 {
		t.Fatal("解析 <= 失败")
	} else {
		t.Logf("解析 <= 成功")
	}
}

func TestParseWhere9(t *testing.T) {
	where := GoMysqlDao.NewMysqlWhereColl()
	where.Debug = true
	where.Add(":order", "id DESC")
	w := where.ParseWhere()

	r1 := w.Order != " ORDER BY id DESC"
	r2 := w.Param != nil
	if r1 || r2 {
		t.Fatal("解析 order 失败")
	} else {
		t.Logf("解析 order 成功")
	}
}

func TestParseWhere10(t *testing.T) {
	where := GoMysqlDao.NewMysqlWhereColl()
	where.Debug = true
	where.Add(":group", "status")
	w := where.ParseWhere()

	r1 := w.Group != " GROUP BY status"
	r2 := w.Param != nil
	if r1 || r2 {
		t.Fatal("解析 group 失败")
	} else {
		t.Logf("解析 group 成功")
	}
}

func TestParseWhere11(t *testing.T) {
	where := GoMysqlDao.NewMysqlWhereColl()
	where.Debug = true
	where.Add(":limit", "10,20")
	w := where.ParseWhere()

	r1 := w.Limit != " LIMIT 10,20"
	r2 := w.Param != nil
	if r1 || r2 {
		t.Fatal("解析 limit 失败")
	} else {
		t.Logf("解析 limit 成功")
	}
}

func TestParseWhere12(t *testing.T) {
	where := GoMysqlDao.NewMysqlWhereColl()
	where.Debug = true
	where.Add(":limit", 2)
	w := where.ParseWhere()

	r1 := w.Limit != " LIMIT 2"
	r2 := w.Param != nil
	if r1 || r2 {
		t.Fatal("解析 limit 失败")
	} else {
		t.Logf("解析 limit 成功")
	}
}

func TestParseWhere13(t *testing.T) {
	where := GoMysqlDao.NewMysqlWhereColl()
	where.Debug = true
	where.Add(":having", "c > 10")
	w := where.ParseWhere()

	r1 := w.Having != " HAVING c > 10"
	r2 := w.Param != nil
	if r1 || r2 {
		t.Fatal("解析 having 失败")
	} else {
		t.Logf("解析 having 成功")
	}
}

func TestParseWhere14(t *testing.T) {
	where := GoMysqlDao.NewMysqlWhereColl()
	where.Debug = true
	where.Add(":having", 10)
	w := where.ParseWhere()

	r1 := w.Having != " HAVING 10"
	r2 := w.Param != nil
	if r1 || r2 {
		t.Fatal("解析 having 失败")
	} else {
		t.Logf("解析 having 成功")
	}
}

func TestParseWhere15(t *testing.T) {
	where := GoMysqlDao.NewMysqlWhereColl()
	where.Debug = true
	where.Add(":where", "is_delete = 0")
	w := where.ParseWhere()

	r1 := w.SqlWHere[0] != "(is_delete = 0)"
	r2 := w.Param != nil
	if r1 || r2 {
		t.Fatal("解析 where 失败")
	} else {
		t.Logf("解析 where 成功")
	}
}

func TestParseWhere16(t *testing.T) {
	where := GoMysqlDao.NewMysqlWhereColl()
	where.Debug = true
	//var ids []int = make([]int, 5)
	//ids = append(ids, 1, 2, 3, 4, 5)
	where.Add("id", "IN", 1, 2, 3, 4, 5)
	w := where.ParseWhere()

	r1 := w.SqlWHere[0] != "(id IN (?,?,?,?,?))"
	r2 := len(w.Param) != 5
	if r1 || r2 {
		t.Fatal("解析 IN 失败")
	} else {
		t.Logf("解析 IN 成功")
	}
}

func TestParseWhere17(t *testing.T) {
	where := GoMysqlDao.NewMysqlWhereColl()
	where.Debug = true
	where.Add("id", "IN", "123")
	w := where.ParseWhere()

	r1 := w.SqlWHere[0] != "(id IN (?))"
	r2 := len(w.Param) != 1
	if r1 || r2 {
		t.Fatal("解析 IN 失败")
	} else {
		t.Logf("解析 IN 成功")
	}
}

func TestParseWhere18(t *testing.T) {
	where := GoMysqlDao.NewMysqlWhereColl()
	where.Debug = true
	where.Add("id", "IN", 100)
	w := where.ParseWhere()

	r1 := w.SqlWHere[0] != "(id IN (?))"
	r2 := len(w.Param) != 1
	if r1 || r2 {
		t.Fatal("解析 IN 失败")
	} else {
		t.Logf("解析 IN 成功")
	}
}

func TestParseWhere19(t *testing.T) {
	where := GoMysqlDao.NewMysqlWhereColl()
	where.Debug = true
	var ids = make([]int, 0)
	where.Add("id", "IN", ids)
	w := where.ParseWhere()

	r1 := w.SqlWHere[0] != "(id IN (-900009))"
	r2 := w.Param != nil
	if r1 || r2 {
		t.Fatal("解析 IN 失败")
	} else {
		t.Logf("解析 IN 成功")
	}
}

func TestParseWhere20(t *testing.T) {
	where := GoMysqlDao.NewMysqlWhereColl()
	where.Debug = true
	where.Add("id", "NOT IN", 1, 2, 3, 4, 5)
	w := where.ParseWhere()
	r1 := w.SqlWHere[0] != "(id NOT IN (?,?,?,?,?))"
	r2 := len(w.Param) != 5
	if r1 || r2 {
		t.Fatal("解析 NOT IN 失败")
	} else {
		t.Logf("解析 NOT IN 成功")
	}
}
