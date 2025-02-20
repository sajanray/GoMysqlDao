package GoMysqlDao

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

// MysqlWhereColl SQL条件集合
type MysqlWhereColl struct {
	//where条件集合
	WhereItems []mysqlWhereItem
	//是否调试模式
	Debug bool
	//where条件之间的连接符
	WhereJoinStr string
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

	WhereCollValue MysqlWhereColl
}

// ParseWhereReturn 解析后的SQL WHERE条件
type ParseWhereReturn struct {
	SqlWHere []string
	Param    []interface{}
	Order    string
	Group    string
	Having   string
	Limit    string
	BaseSql  string
}

// NewMysqlWhereColl 实例化一个MysqlWhereColl集合
func NewMysqlWhereColl() *MysqlWhereColl {
	return &MysqlWhereColl{
		WhereItems:   make([]mysqlWhereItem, 0),
		WhereJoinStr: "AND",
	}
}

// Add 添加where条件
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
	} else if fType == reflect.Pointer {
		if reflect.TypeOf(args[0]).Elem().String() == reflect.TypeOf(mwc).Elem().String() {
			pWhereColl, ok := (args[0]).(*MysqlWhereColl)
			if ok {
				var param = mysqlWhereItem{
					ItemValue: mysqlWhereItemValue{
						ValueType:      reflect.Struct,
						WhereCollValue: *pWhereColl,
					},
				}
				mwc.WhereItems = append(mwc.WhereItems, param)
			}
		}
		return
	} else if fType == reflect.Struct {
		if reflect.TypeOf(args[0]).String() == reflect.TypeOf(mwc).Elem().String() {
			whereColl, ok := (args[0]).(MysqlWhereColl)
			if ok {
				var param = mysqlWhereItem{
					ItemValue: mysqlWhereItemValue{
						ValueType:      reflect.Struct,
						WhereCollValue: whereColl,
					},
				}
				mwc.WhereItems = append(mwc.WhereItems, param)
			}
		}
		return
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
		switch miv.ValueType {
		case reflect.Int:
			miv.IntValue = value.(int)
		case reflect.String:
			miv.StringValue = value.(string)
		case reflect.Slice:
			miv.SliceValue = value
			miv.SliceValueType = reflect.TypeOf(value).Elem().Kind() //切片元素的数据类型
		default:
		}
		param.ItemValue = miv
	}
	mwc.WhereItems = append(mwc.WhereItems, param)
}

func (mwc *MysqlWhereColl) ParseWhere() (whereReturn ParseWhereReturn) {
	for _, item := range mwc.WhereItems {
		if item.ItemValue.ValueIsStruct() {
			ret := item.ItemValue.WhereCollValue.ParseWhere()
			if len(ret.SqlWHere) > 0 {
				joinStr := " " + item.ItemValue.WhereCollValue.WhereJoinStr + " "
				whereStr := fmt.Sprintf("(%s)", ret.SqlWhereToString(joinStr))
				whereReturn.SqlWHere = append(whereReturn.SqlWHere, whereStr)
				if len(ret.Param) > 0 {
					whereReturn.Param = append(whereReturn.Param, ret.Param...)
				}
			}
			continue
		}
		if !strings.HasPrefix(item.Field, ":") {
			//拼装SQL条件
			whereString := fmt.Sprintf("(%s %s %s)", item.Field, item.Op, item.GetPlaceHolder())
			whereReturn.SqlWHere = append(whereReturn.SqlWHere, whereString)

			//按占位符顺序拼装参数
			if item.ItemValue.ValueIsInt() {
				whereReturn.Param = append(whereReturn.Param, item.ItemValue.IntValue)
			} else if item.ItemValue.ValueIsString() {
				whereReturn.Param = append(whereReturn.Param, item.ItemValue.StringValue)
			} else if item.ItemValue.ValueIsSlice() {
				//占位符
				if item.ItemValue.SliceValueIsInt() {
					for _, iv := range item.ItemValue.SliceValue.([]int) {
						whereReturn.Param = append(whereReturn.Param, iv)
					}
				} else if item.ItemValue.SliceValueIsString() {
					for _, iv := range item.ItemValue.SliceValue.([]string) {
						whereReturn.Param = append(whereReturn.Param, iv)
					}
				} else if item.ItemValue.SliceValueIsInterface() {
					for _, iv := range item.ItemValue.SliceValue.([]interface{}) {
						whereReturn.Param = append(whereReturn.Param, iv)
					}
				}
			}
			continue
		}

		if item.FieldIsWhere() {
			whereReturn.SqlWHere = append(whereReturn.SqlWHere, fmt.Sprintf("(%s)", item.ItemValue.StringValue))
			continue
		}

		if item.FieldIsSql() {
			whereReturn.BaseSql = item.ItemValue.StringValue
			continue
		}

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
	}

	return whereReturn
}

// GetPlaceHolder 获取占位符
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
		if mwi.ItemValue.SliceValueIsInt() {
			length = len(mwi.ItemValue.SliceValue.([]int))
		} else if mwi.ItemValue.SliceValueIsString() {
			length = len(mwi.ItemValue.SliceValue.([]string))
		} else if mwi.ItemValue.SliceValueIsInterface() {
			length = len(mwi.ItemValue.SliceValue.([]interface{}))
		}
		if length > 0 {
			return fmt.Sprintf("(%s)", strings.TrimRight(strings.Repeat("?,", length), ","))
		} else {
			return "(-900009)"
		}
	}
	return ""
}

// FieldIsGroup 判断是否是group
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

func (miv *mysqlWhereItemValue) ValueIsStruct() bool {
	return miv.ValueType == reflect.Struct
}

func (miv *mysqlWhereItemValue) SliceValueIsInt() bool {
	return miv.SliceValueType == reflect.Int
}

func (miv *mysqlWhereItemValue) SliceValueIsString() bool {
	return miv.SliceValueType == reflect.String
}

func (miv *mysqlWhereItemValue) SliceValueIsInterface() bool {
	return miv.SliceValueType == reflect.Interface
}

func (miv *mysqlWhereItemValue) ValueIsSlice() bool {
	return miv.ValueType == reflect.Slice
}

func (pwr *ParseWhereReturn) SqlWhereToString(separate string) string {
	if len(pwr.SqlWHere) > 0 {
		return strings.Join(pwr.SqlWHere, separate)
	} else {
		return ""
	}
}
