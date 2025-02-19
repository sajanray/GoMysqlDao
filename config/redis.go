package config

type RedisMasterSlave struct {
	Master    []RedisItem `json:"master"`
	Slave     []RedisItem `json:"slave"`
	SlaveUser RedisItem   `json:"slave_user"`
}

type RedisItem struct {
	Host     string `json:"host"`
	Pwd      string `json:"pwd"`
	Port     string `json:"port"`
	Database string `json:"database"`
}

// 组合Redis配置
func combineRedisConf(dbConf *DatabaseConf) {
	if len(dbConf.Redis.Slave) > 0 {
		for i, v := range dbConf.Redis.Slave {
			if len(v.Host) == 0 {
				dbConf.Redis.Slave[i].Host = dbConf.Redis.SlaveUser.Host
			}
			if len(v.Pwd) == 0 {
				dbConf.Redis.Slave[i].Pwd = dbConf.Redis.SlaveUser.Pwd
			}
			if len(v.Port) == 0 {
				dbConf.Redis.Slave[i].Port = dbConf.Redis.SlaveUser.Port
			}
			if len(v.Database) == 0 {
				dbConf.Redis.Slave[i].Database = dbConf.Redis.SlaveUser.Database
			}
		}
	}
}
