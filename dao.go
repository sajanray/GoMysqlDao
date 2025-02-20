package GoMysqlDao

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/sajanray/GoJsonToStruct"
	"log"
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

// UErrTableNameNotFound 自定义错误
var UErrTableNameNotFound = errors.New("table name not found")
var UErrDstModelAssertStruct = errors.New("dstModel is need struct")
var UErrGetConnectError = errors.New("get mysql connect error")

// MysqlDao 操作数据库 基础结构体
type MysqlDao struct {
	//主键
	Pk string
	//表名
	TableName string
	//自身的mysql连接池
	LocalConnectPool *MysqlConnectPool
	//事务配置
	dbTx *sql.Tx
	//事务是否开始
	isBegin bool
}

// 获取连接池 区分自身连接池还是全局连接池
func (md *MysqlDao) getPool() *MysqlConnectPool {
	if md.LocalConnectPool == nil {
		return GlobalConnectPool
	} else {
		return md.LocalConnectPool
	}
}

// Free 释放数据库连接,一般不需要调用,连接其它数据库实例，临时用一次后释放可以调用该方法
func (md *MysqlDao) Free() {
	if md.LocalConnectPool != nil {
		md.LocalConnectPool.Free()
	}
}

// Begin 开启事务
func (md *MysqlDao) Begin() {
	//捕获异常
	defer func() {
		if err := recover(); err != nil {
			pc, file, no, ok := runtime.Caller(1)
			log.Printf("BEGIN_ERROR:%s %s:%d ptr:%v ok:%v", err, file, no, pc, ok)
		}
	}()

	var err error
	db := md.getPool().Write()
	if db != nil {
		md.dbTx, err = db.Begin()
		if err != nil {
			panic(err)
		}
		md.isBegin = true
	} else {
		panic(UErrGetConnectError)
	}
}

// Commit 提交事务
func (md *MysqlDao) Commit() {
	//捕获异常
	defer func() {
		md.isBegin = false
		if err := recover(); err != nil {
			pc, file, no, ok := runtime.Caller(1)
			log.Printf("ROLLBACK_ERROR:%s %s:%d ptr:%v ok:%v", err, file, no, pc, ok)
		}
	}()

	err := md.dbTx.Commit()
	if err != nil {
		md.Rollback()
		panic(err)
	}
}

// Rollback 回滚事务
func (md *MysqlDao) Rollback() {
	//捕获异常
	defer func() {
		md.isBegin = false
		if err := recover(); err != nil {
			pc, file, no, ok := runtime.Caller(1)
			log.Printf("ROLLBACK_ERROR:%s %s:%d ptr:%v ok:%v", err, file, no, pc, ok)
		}
	}()

	err := md.dbTx.Rollback()
	if err != nil {
		panic(err)
	}
}

// BuildWhere 构建一个where条件集合
// 格式要求：1、冒号开头的两个参数为一组；2、非冒号开头的三个参数为一组；3、最后一个参数可以是1个参数为一组
func (md *MysqlDao) BuildWhere(params ...interface{}) *MysqlWhereColl {
	length := len(params)
	if length == 0 {
		return NewMysqlWhereColl()
	}
	where := NewMysqlWhereColl()
	var i = 0
	var end = 0
	for {
		if i >= length {
			break
		}
		if strings.HasPrefix((params[i]).(string), ":") {
			end = i + 2
			where.Add(params[i:end]...)

		} else {
			end = i + 3
			if end > length {
				end = length
			}
			where.Add(params[i:end]...)
		}
		i = end
	}
	return where
}

// One 查询一条记录
// result 返回的结果，当DstModel=nil是返回map，当DstModel时struct时返回结构体，在处理result是请使用断言处理
// err 错误信息
func (md *MysqlDao) One(option OneOption) (result interface{}, err error) {
	//捕获异常
	defer func() {
		if err := recover(); err != nil {
			pc, file, no, ok := runtime.Caller(1)
			log.Printf("ERROR:%s %s:%d ptr:%v ok:%v", err, file, no, pc, ok)
		}
	}()

	//构造sql语句
	buildSqlReturn, err := md.BuildSelectSql(&BuildSqlOption{
		Where:     option.Where,
		Fields:    option.Fields,
		TableName: option.TableName,
		Pk:        option.Pk,
		ForUpdate: option.ForUpdate,
		Write:     option.Write,
	})
	if err != nil {
		return
	}

	//执行查询
	tmp, err := md.Query(&buildSqlReturn.Sql, buildSqlReturn.ParseResult.Param, option.DstModel, option.Write)
	if err == nil && tmp != nil && len(tmp) > 0 {
		result = tmp[0]
	}
	return
}

// More 查询多条数据
// result 返回的结果，当dstModel=nil是返回切片map，当dstModel时struct时返回切片结构体，在处理result是请使用断言处理
// err 错误信息
func (md *MysqlDao) More(option MoreOption) (result []interface{}, totals int64, err error) {
	defer func() {
		if err := recover(); err != nil {
			pc, file, no, ok := runtime.Caller(1)
			log.Printf("ERROR:%s %s:%d ptr:%v ok:%v", err, file, no, pc, ok)
		}
	}()

	//构造sql语句
	buildSqlReturn, err := md.BuildSelectSql(&BuildSqlOption{
		Where:     option.Where,
		Fields:    option.Fields,
		TableName: option.TableName,
		Pk:        option.Pk,
		ForUpdate: option.ForUpdate,
		Write:     option.Write,
		CalcCount: option.CalcCount,
	})
	if err != nil {
		return
	}

	//查询数据
	result, err = md.Query(&buildSqlReturn.Sql, buildSqlReturn.ParseResult.Param, option.DstModel, option.Write)
	if err != nil {
		return
	}

	//查询总数
	if option.CalcCount {
		rowsRet := make([]interface{}, 0)
		rowsRet, err = md.Query(&buildSqlReturn.CalcSql, buildSqlReturn.ParseResult.Param, nil, option.Write)
		if err != nil {
			return
		}
		res, _ := rowsRet[0].(map[string]interface{})
		n1, _ := res["counts"].(string)
		totals, err = strconv.ParseInt(n1, 10, 0)
	}
	return
}

// Query 执行查询 用于执行select语句
// sqlStr	sql语句
// params	sql语句里面的占位符参数
// dstModel	返回结果数据结构 nil:返回map  struct:返回结构体
// isMaster	是否强制使用主库查询
// result 	返回的结果，当dstModel=nil是返回切片map，当dstModel时struct时返回切片结构体，在处理result是请使用断言处理
// err		错误信息
func (md *MysqlDao) Query(sqlStr *string, params []interface{}, dstModel interface{}, isMaster bool) (result []interface{}, err error) {
	//日志记录执行的sql
	pc, file, no, ok := runtime.Caller(1)
	log.Printf("INFO debugSQL:%s params:%v %s:%d ptr:%v ok:%v", *sqlStr, params, file, no, pc, ok)

	//查询数据
	var rows *sql.Rows
	if !md.isBegin {
		rows, err = md.getPool().SqlDB(isMaster).Query(*sqlStr, params...)
	} else {
		rows, err = md.dbTx.Query(*sqlStr, params...)
	}

	if err != nil {
		return
	}
	mapData, err := md.ParseRowsToMap(rows) //统一转换成map
	if err != nil {
		return
	}

	if dstModel != nil {
		dstModelReflect, err := md.assertStruct(dstModel)
		if err != nil {
			return result, err
		}
		jts := JTStools.NewMapToStruct()

		for i, mapV := range mapData {
			if i == 0 { //第一个元素存放在dstModel 这块内存空间，提高内存使用效率
				jts.Transform(dstModel, mapV)
				result = append(result, dstModel)
			} else {
				//根据模具创建对象
				dstItem := reflect.New(dstModelReflect).Interface()
				jts.Transform(dstItem, mapV)
				result = append(result, dstItem)
			}
		}
	} else {
		result = mapData
	}

	return result, err
}

// Exec 使用主库 Exec执行一次命令（包括查询、删除、更新、插入等），不返回任何执行结果。参数args表示query中的占位参数。
// sqlStr	sql语句
// params	sql语句里面的占位符参数
// result	返回的结果
// err		错误信息
func (md *MysqlDao) Exec(sqlStr *string, params []interface{}) (result sql.Result, err error) {
	//查询数据
	if !md.isBegin {
		result, err = md.getPool().Write().Exec(*sqlStr, params...)
	} else {
		result, err = md.dbTx.Exec(*sqlStr, params...)
	}
	return
}

// Insert 插入数据
// sqlStr	sql语句
// args 		第一个参数设定为表名
// newId 	返回插入之后数据的主键ID
// err		错误信息
func (md *MysqlDao) Insert(data map[string]interface{}, args ...string) (newId int64, err error) {
	defer func() {
		if err := recover(); err != nil {
			pc, file, no, ok := runtime.Caller(1)
			log.Printf("ERROR:%s %s:%d ptr:%v ok:%v", err, file, no, pc, ok)
		}
	}()

	//表名
	var tableName string
	if len(args) > 0 {
		tableName = args[0]
	} else {
		tableName = md.TableName
	}
	if len(tableName) == 0 {
		err = UErrTableNameNotFound
		return
	}

	//抽取字段名和值
	var field []string
	var value []interface{}
	var ph []string
	for k, v := range data {
		field = append(field, k)
		value = append(value, v)
		ph = append(ph, "?")
	}
	sqlStr := fmt.Sprintf("INSERT INTO %s (%s) VALUE (%s)", tableName, strings.Join(field, ","), strings.Join(ph, ","))

	//日志记录执行的sql
	pc, file, no, ok := runtime.Caller(1)
	log.Printf("INFO debugSQL:%s value:%v %s:%d ptr:%v ok:%v", sqlStr, value, file, no, pc, ok)

	//执行插入
	var ret sql.Result
	if !md.isBegin {
		ret, err = md.getPool().Write().Exec(sqlStr, value...)
	} else {
		ret, err = md.dbTx.Exec(sqlStr, value...)
	}
	if err != nil {
		return
	}
	newId, err = ret.LastInsertId()
	return
}

// Update 更新数据
// args 第一个参数设定为表名
func (md *MysqlDao) Update(upData *MysqlWhereColl, upWhere *MysqlWhereColl, args ...string) (effectRows int64, err error) {
	defer func() {
		if err := recover(); err != nil {
			pc, file, no, ok := runtime.Caller(1)
			log.Printf("ERROR:%s %s:%d ptr:%v ok:%v", err, file, no, pc, ok)
		}
	}()

	//表名
	var tableName string
	if len(args) > 0 {
		tableName = args[0]
	} else {
		tableName = md.TableName
	}
	if len(tableName) == 0 {
		err = UErrTableNameNotFound
		return
	}

	//抽取字段名和值
	updateSetData := md.ParseUpdateSetData(upData)
	whereReturn := upWhere.ParseWhere()
	param := append(updateSetData.Param, whereReturn.Param...)
	var sqlStr string
	var whereStr string
	if len(whereReturn.SqlWHere) > 0 {
		whereStr = " WHERE " + whereReturn.SqlWhereToString(fmt.Sprintf(" %s ", upWhere.WhereJoinStr))
	}
	sqlStr = fmt.Sprintf("UPDATE %s SET %s%s", tableName, updateSetData.SqlWhereToString(", "), whereStr)

	//日志
	pc, file, no, ok := runtime.Caller(1)
	log.Printf("INFO debugSQL:%s param:%v %s:%d ptr:%v ok:%v", sqlStr, param, file, no, pc, ok)

	//执行SQL
	var ret sql.Result
	if !md.isBegin {
		ret, err = md.getPool().Write().Exec(sqlStr, param...)
	} else {
		ret, err = md.dbTx.Exec(sqlStr, param...)
	}
	if err != nil {
		return
	}

	//返回影响行数
	effectRows, err = ret.RowsAffected()
	return
}

// Delete 删除数据
// args 第一个参数设定为表名
func (md *MysqlDao) Delete(delWhere *MysqlWhereColl, args ...string) (effectRows int64, err error) {
	defer func() {
		if err := recover(); err != nil {
			pc, file, no, ok := runtime.Caller(1)
			log.Printf("ERROR:%s %s:%d ptr:%v ok:%v", err, file, no, pc, ok)
		}
	}()

	//表名
	var tableName string
	if len(args) > 0 {
		tableName = args[0]
	} else {
		tableName = md.TableName
	}
	if len(tableName) == 0 {
		err = UErrTableNameNotFound
		return
	}

	//抽取字段名和值
	whereReturn := delWhere.ParseWhere()
	var whereStr string
	if len(whereReturn.SqlWHere) > 0 {
		whereStr = " WHERE " + whereReturn.SqlWhereToString(fmt.Sprintf(" %s ", delWhere.WhereJoinStr))
	}
	sqlStr := fmt.Sprintf("DELETE FROM %s%s", tableName, whereStr)

	//日志
	pc, file, no, ok := runtime.Caller(1)
	log.Printf("INFO debugSQL:%s param:%v %s:%d ptr:%v ok:%v", sqlStr, whereReturn.Param, file, no, pc, ok)

	var ret sql.Result
	if !md.isBegin {
		ret, err = md.getPool().Write().Exec(sqlStr, whereReturn.Param...)
	} else {
		ret, err = md.dbTx.Exec(sqlStr, whereReturn.Param...)
	}

	if err != nil {
		return
	}

	effectRows, err = ret.RowsAffected()
	return
}

// BuildSelectSql 解析where条件，构建sql语句
func (md *MysqlDao) BuildSelectSql(option *BuildSqlOption) (buildSqlReturn BuildSqlReturn, err error) {
	//处理表名
	var tableName string
	if len(option.TableName) > 0 {
		tableName = option.TableName
	} else if len(md.TableName) > 0 {
		tableName = md.TableName
	} else {
		err = UErrTableNameNotFound
		return
	}

	//解析where条件 拼装SQL语句
	buildSqlReturn.ParseResult = option.Where.ParseWhere()
	whereStr := buildSqlReturn.ParseResult.SqlWhereToString(fmt.Sprintf(" %s ", option.Where.WhereJoinStr))
	if len(whereStr) > 0 {
		whereStr = fmt.Sprintf(" WHERE %s", whereStr)
	}
	var baseSql string
	var calcSql string
	if !strings.EqualFold(buildSqlReturn.ParseResult.BaseSql, "") {
		baseSql = buildSqlReturn.ParseResult.BaseSql
		if option.CalcCount {
			calcSql = fmt.Sprintf("SELECT COUNT(*) as counts FROM %s", tableName)
		}
	} else {
		fields := "*"
		if len(option.Fields) > 0 {
			fields = option.Fields
		}
		baseSql = fmt.Sprintf("SELECT %s FROM %s", fields, tableName)
		if option.CalcCount {
			calcSql = fmt.Sprintf("SELECT COUNT(*) as counts FROM %s", tableName)
		}
	}
	buildSqlReturn.Sql = fmt.Sprintf("%s%s%s%s%s%s",
		baseSql,
		whereStr,
		buildSqlReturn.ParseResult.Group,
		buildSqlReturn.ParseResult.Having,
		buildSqlReturn.ParseResult.Order,
		buildSqlReturn.ParseResult.Limit)

	if option.CalcCount {
		buildSqlReturn.CalcSql = fmt.Sprintf("%s%s%s%s",
			calcSql,
			whereStr,
			buildSqlReturn.ParseResult.Group,
			buildSqlReturn.ParseResult.Having)
	}
	return
}

// ParseUpdateSetData 解析update时set数据
func (md *MysqlDao) ParseUpdateSetData(whColl *MysqlWhereColl) (whereReturn ParseWhereReturn) {
	for _, item := range whColl.WhereItems {
		whereReturn.SqlWHere = append(whereReturn.SqlWHere, fmt.Sprintf("%s = ?", item.Field))
		if item.ItemValue.ValueIsInt() {
			whereReturn.Param = append(whereReturn.Param, item.ItemValue.IntValue)
		} else if item.ItemValue.ValueIsString() {
			whereReturn.Param = append(whereReturn.Param, item.ItemValue.StringValue)
		}
	}
	return whereReturn
}

// ParseRowsToMap 解析多行结果为map
// result  在出错或者数据为空的情况下返回nil
func (md *MysqlDao) ParseRowsToMap(rows *sql.Rows) (result []interface{}, err error) {
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
		}
	}(rows)

	//提取列的信息
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		return
	}

	//列的个数
	colsTypLen := len(columnTypes)

	//定义一个扫描数据的载体
	values := make([]sql.RawBytes, colsTypLen)
	dstAddress := md.collectDstAddress(values)
	for rows.Next() {
		err := rows.Scan(dstAddress...)
		if err != nil {
			return result, err
		}

		//切片数组转成Map
		data := make(map[string]interface{}, colsTypLen)
		for i, v := range columnTypes {
			//fmt.Println(v.Name() , v.DatabaseTypeName() , )
			//data[v.Name()] = fmt.Sprintf("%s",values[i])

			//todo 此处应该根据数据的表的字段类型转换成相应的go 类型
			//Common type include "VARCHAR", "TEXT", "NVARCHAR", "DECIMAL", "BOOL", "INT", "BIGINT".
			data[v.Name()] = string(values[i]) //一律转换成string
		}
		//push进去一行
		result = append(result, data)
	}

	if len(result) == 0 {
		err = sql.ErrNoRows
	}

	return
}

// 断言目标是否是结构体
func (md *MysqlDao) assertStruct(dstModel interface{}) (dstModelReflect reflect.Type, err error) {
	//获取模具类型
	dstModelReflect = reflect.TypeOf(dstModel)
	//如果是指针需要.Elem
	if dstModelReflect.Kind() == reflect.Ptr {
		dstModelReflect = dstModelReflect.Elem()
	}
	//如果dst是结构体
	if dstModelReflect.Kind() != reflect.Struct {
		err = UErrDstModelAssertStruct
	}
	return
}

// collectStructAddress 提取结构体每个元素的地址切片
func (md *MysqlDao) collectStructAddress(st interface{}) []interface{} {
	stVal := reflect.ValueOf(st).Elem()
	stTpy := reflect.TypeOf(st)
	numF := stTpy.Elem().NumField()
	var sqlValues []interface{}
	for i := 0; i < numF; i++ {
		sqlValues = append(sqlValues, stVal.Field(i).Addr().Interface())
	}
	return sqlValues
}

// collectDstAddress 提取切片元素的地址
func (md *MysqlDao) collectDstAddress(sqlValues []sql.RawBytes) []interface{} {
	counts := len(sqlValues)
	scanArgs := make([]interface{}, counts)
	for i := 0; i < counts; i++ {
		scanArgs[i] = &sqlValues[i]
	}
	return scanArgs
}
