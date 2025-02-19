package config

type MysqlMasterSlave struct {
	Master    []MysqlItem `json:"master"`
	Slave     []MysqlItem `json:"slave"`
	SlaveUser MysqlItem   `json:"slave_user"`
}

type MysqlItem struct {
	Host         string `json:"host"`
	User         string `json:"user"`
	Pwd          string `json:"pwd"`
	Port         string `json:"port"`
	DbName       string `json:"db_name"`
	MaxIdleConns int    `json:"max_idle_conns"`
	MaxOpenConns int    `json:"max_open_conns"`
	MaxLifetime  int    `json:"max_lifetime"` //分钟
	CharSet      string `json:"char_set"`
}

// 组合Mysql配置文件
func combineMysqlConf(dbConf *DatabaseConf) {
	//从库公共部分合并到每个节点
	if len(dbConf.Mysql.Slave) > 0 {
		for i, v := range dbConf.Mysql.Slave {
			if len(v.Host) == 0 {
				dbConf.Mysql.Slave[i].Host = dbConf.Mysql.SlaveUser.Host
			}
			if len(v.User) == 0 {
				dbConf.Mysql.Slave[i].User = dbConf.Mysql.SlaveUser.User
			}
			if len(v.Pwd) == 0 {
				dbConf.Mysql.Slave[i].Pwd = dbConf.Mysql.SlaveUser.Pwd
			}
			if len(v.Port) == 0 {
				dbConf.Mysql.Slave[i].Port = dbConf.Mysql.SlaveUser.Port
			}
			if len(v.DbName) == 0 {
				dbConf.Mysql.Slave[i].DbName = dbConf.Mysql.SlaveUser.DbName
			}
			if v.MaxOpenConns == 0 {
				dbConf.Mysql.Slave[i].MaxOpenConns = dbConf.Mysql.SlaveUser.MaxOpenConns
			}
			if v.MaxIdleConns == 0 {
				dbConf.Mysql.Slave[i].MaxIdleConns = dbConf.Mysql.SlaveUser.MaxIdleConns
			}
			if v.MaxLifetime == 0 {
				dbConf.Mysql.Slave[i].MaxLifetime = dbConf.Mysql.SlaveUser.MaxLifetime
			}
			if len(v.CharSet) == 0 {
				dbConf.Mysql.Slave[i].CharSet = dbConf.Mysql.SlaveUser.CharSet
			}
		}
	}
}
