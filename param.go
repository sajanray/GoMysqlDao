package GoMysqlDao

type OneOption struct {
	Where     *MysqlWhereColl //where条件
	Fields    string          //要查询的字段，多个字段用逗号分隔，默认查询所有
	DstModel  interface{}     //模型，告诉处理逻辑该用什么数据结构去容纳查询的结构，只是一个模具而已
	ForUpdate bool
	Write     bool
	TableName string
	Pk        string
}

type MoreOption struct {
	Where     *MysqlWhereColl
	Fields    string
	DstModel  interface{}
	ForUpdate bool
	Write     bool
	TableName string
	Pk        string
	CalcCount bool //是否计算总数
}

type BuildSqlOption struct {
	Where     *MysqlWhereColl
	Fields    string
	TableName string
	Pk        string
	ForUpdate bool
	Write     bool
	CalcCount bool //是否计算总数
}

type BuildSqlReturn struct {
	Sql         string
	CalcSql     string
	ParseResult ParseWhereReturn
}
