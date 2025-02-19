package GoMysqlDao

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

// SQL条件集合
type MysqlWhereColl struct {
	//表名
	TableName string
	//where集合
	WhereColl []mysqlWhereItem
	//返回承载数据的结构体
	ReturnStruct func() interface{}
	//使用写库
	UseWrite bool
	//返回Map
	ReturnMap bool
	//是否调试模式
	Debug bool
}

// where条件节点
type mysqlWhereItem struct {
	//字段
	Field string
	//操作符
	Op string
	//操作值
	ItemValue mysqlWhereItemValue
}

// where条件节点值
type mysqlWhereItemValue struct {
	IntValue       int
	StringValue    string
	SliceValue     interface{}
	ValueType      reflect.Kind
	SliceValueType reflect.Kind
}

// 解析后的SQL WHERE条件
type ParseWhereReturn struct {
	SqlWHere []string
	Param    []interface{}
	Order    string
	Group    string
	Having   string
	Limit    string
	Basesql  string
}

// 实例化一个MysqlWhereColl集合
func NewMysqlWhereColl() *MysqlWhereColl {
	return &MysqlWhereColl{
		WhereColl: make([]mysqlWhereItem, 0),
	}
}

// 设置表名
func (mwc *MysqlWhereColl) SetTableName(tableName string) {
	mwc.TableName = tableName
}

// 添加where条件
func (mwc *MysqlWhereColl) Add(args ...interface{}) {
	//参数校验
	argsLen := len(args)
	if argsLen == 0 {
		if mwc.Debug {
			log.Println("args cannot be empty")
		}
		panic("args cannot be empty")
		return
	}

	//第1个参数处理
	var field string
	fType := reflect.TypeOf(args[0]).Kind() //第一个参数类型
	if fType == reflect.String {
		field = fmt.Sprintf("%s", args[0])
	} else if fType == reflect.Int {
		field = fmt.Sprintf("%d", args[0])
	} else {
		if mwc.Debug {
			log.Println("args[0] mast be int or string")
		}
		panic("args[0] mast be int or string")
		return
	}

	//第2 第3个参数处理
	var op string
	var value interface{}
	//以冒号开头的参数
	if strings.HasPrefix(field, ":") {
		if argsLen != 2 {
			log.Println("The length of the ARG parameter must be 2")
			panic("The length of the ARG parameter must be 2")
		}
		value = args[1]
		//if argsLen == 1 {
		//} else if argsLen == 2 {
		//	op = (args[1]).(string)
		//} else if argsLen == 3 {
		//	op = (args[1]).(string)
		//	value = args[2]
		//} else {
		//	op = (args[1]).(string)
		//	value = args[2:]
		//}
	} else {
		switch argsLen {
		case 1:
			value = field
			field = ":where"
		case 2:
			op = "="
			value = args[1]
		case 3:
			op = (args[1]).(string)
			value = args[2]
		default:
			op = (args[1]).(string)
			value = args[2:]
		}
	}

	//为每一个字段值创建一个节点保存
	var param = mysqlWhereItem{
		Field: field,
		Op:    op,
	}
	if value != nil {
		miv := mysqlWhereItemValue{}
		miv.ValueType = reflect.TypeOf(value).Kind()
		if miv.ValueType == reflect.Int {
			miv.IntValue = value.(int)
		} else if miv.ValueType == reflect.String {
			miv.StringValue = value.(string)
		} else if miv.ValueType == reflect.Slice {
			miv.SliceValue = value
			miv.SliceValueType = reflect.TypeOf(value).Elem().Kind()
		}
		param.ItemValue = miv
	}
	mwc.WhereColl = append(mwc.WhereColl, param)
}

func (mwc *MysqlWhereColl) ParseWhere() (whereReturn ParseWhereReturn) {
	for _, item := range mwc.WhereColl {
		if item.FieldIsGroup() {
			whereReturn.Group = fmt.Sprintf(" GROUP BY %v", item.ItemValue.StringValue)
			continue
		}

		if item.FieldIsOrder() {
			whereReturn.Order = fmt.Sprintf(" ORDER BY %v", item.ItemValue.StringValue)
			continue
		}

		if item.FieldIsHaving() {
			if item.ItemValue.ValueIsInt() {
				whereReturn.Having = fmt.Sprintf(" HAVING %v", item.ItemValue.IntValue)
			} else if item.ItemValue.ValueIsString() {
				whereReturn.Having = fmt.Sprintf(" HAVING %v", item.ItemValue.StringValue)
			}
			continue
		}

		if item.FieldIsLimit() {
			if item.ItemValue.ValueIsInt() {
				whereReturn.Limit = fmt.Sprintf(" LIMIT %v", item.ItemValue.IntValue)
			} else if item.ItemValue.ValueIsString() {
				whereReturn.Limit = fmt.Sprintf(" LIMIT %v", item.ItemValue.StringValue)
			}
			continue
		}

		if item.FieldIsSql() {
			whereReturn.Basesql = item.ItemValue.StringValue
			continue
		}

		//拼装SQL条件
		var whereString string
		if item.FieldIsWhere() {
			whereString = fmt.Sprintf("(%s)", item.ItemValue.StringValue)
			whereReturn.SqlWHere = append(whereReturn.SqlWHere, whereString)
			continue
		}
		whereString = fmt.Sprintf("(%s %s %s)", item.Field, item.Op, item.GetPlaceHolder())
		whereReturn.SqlWHere = append(whereReturn.SqlWHere, whereString)

		//按占位符顺序拼装参数
		if item.ItemValue.ValueIsInt() {
			whereReturn.Param = append(whereReturn.Param, item.ItemValue.IntValue)
		} else if item.ItemValue.ValueIsString() {
			whereReturn.Param = append(whereReturn.Param, item.ItemValue.StringValue)
		} else if item.ItemValue.ValueIsSlice() {
			//占位符
			if item.ItemValue.SliceValueType == reflect.Int {
				for _, iv := range item.ItemValue.SliceValue.([]int) {
					whereReturn.Param = append(whereReturn.Param, iv)
				}
			} else if item.ItemValue.SliceValueType == reflect.String {
				for _, iv := range item.ItemValue.SliceValue.([]string) {
					whereReturn.Param = append(whereReturn.Param, iv)
				}
			} else if item.ItemValue.SliceValueType == reflect.Interface {
				for _, iv := range item.ItemValue.SliceValue.([]interface{}) {
					whereReturn.Param = append(whereReturn.Param, iv)
				}
			}
		}
	}

	return whereReturn
}

// 获取占位符
func (mwi *mysqlWhereItem) GetPlaceHolder() string {
	if mwi.ItemValue.ValueIsInt() || mwi.ItemValue.ValueIsString() {
		if strings.EqualFold(mwi.Op, "IN") || strings.EqualFold(mwi.Op, "NOT IN") {
			return "(?)"
		} else {
			return "?"
		}
	} else if mwi.ItemValue.ValueIsSlice() {
		var length int
		//占位符
		if mwi.ItemValue.SliceValueType == reflect.Int {
			length = len(mwi.ItemValue.SliceValue.([]int))
		} else if mwi.ItemValue.SliceValueType == reflect.String {
			length = len(mwi.ItemValue.SliceValue.([]string))
		} else if mwi.ItemValue.SliceValueType == reflect.Interface {
			length = len(mwi.ItemValue.SliceValue.([]interface{}))
		}
		if length > 0 {
			var ph []string
			for i := 0; i < length; i++ {
				ph = append(ph, "?")
			}
			return fmt.Sprintf("(%s)", strings.Join(ph, ","))
		} else {
			return "(-900009)"
		}
	}
	return ""
}

// 判断是否是group
func (mwi *mysqlWhereItem) FieldIsGroup() bool {
	return strings.EqualFold(mwi.Field, ":group")
}

func (mwi *mysqlWhereItem) FieldIsOrder() bool {
	return strings.EqualFold(mwi.Field, ":order")
}

func (mwi *mysqlWhereItem) FieldIsHaving() bool {
	return strings.EqualFold(mwi.Field, ":having")
}

func (mwi *mysqlWhereItem) FieldIsLimit() bool {
	return strings.EqualFold(mwi.Field, ":limit")
}

func (mwi *mysqlWhereItem) FieldIsWhere() bool {
	return strings.EqualFold(mwi.Field, ":where")
}

func (mwi *mysqlWhereItem) FieldIsSql() bool {
	return strings.EqualFold(mwi.Field, ":sql")
}

func (miv *mysqlWhereItemValue) ValueIsInt() bool {
	return miv.ValueType == reflect.Int
}

func (miv *mysqlWhereItemValue) ValueIsString() bool {
	return miv.ValueType == reflect.String
}

func (miv *mysqlWhereItemValue) ValueIsSlice() bool {
	return miv.ValueType == reflect.Slice
}

func (pwr *ParseWhereReturn) SqlWhereToString(separate string) string {
	if len(pwr.SqlWHere) > 0 {
		return " " + strings.Join(pwr.SqlWHere, separate)
	} else {
		return ""
	}
}
