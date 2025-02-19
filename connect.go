package GoMysqlDao

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/sajanray/GoMysqlDao/config"
	"log"
	"math/rand"
	"runtime"
	"time"
)

// GlobalConnectPool 全局数据库连接池
var GlobalConnectPool *MysqlConnectPool

// MysqlConnectPool 连接池结构体
type MysqlConnectPool struct {
	//写库连接池
	writePool *sql.DB
	//读库连接池
	readPool map[int]*sql.DB
	//是否是第一次连接
	isFirstConnect bool
	//数据库配置
	dbConf config.DatabaseConf
	//数据库配置文件
	dbConfFile string
}

// InitMysqlPool 初始化全局mysql连接池
func InitMysqlPool(confFile string) {
	if GlobalConnectPool != nil {
		GlobalConnectPool.isFirstConnect = false
		GlobalConnectPool.tryConnect()
	} else {
		GlobalConnectPool = NewMysqlPool(confFile)
	}
}

// NewMysqlPool 初始化局部变量的数据库连接池 一般要配合dao.LocalConnectPool来临时使用，使用完之后尽量dao.Free()
func NewMysqlPool(confFile string) *MysqlConnectPool {
	pool := new(MysqlConnectPool)
	pool.dbConfFile = confFile
	pool.isFirstConnect = true
	pool.readPool = make(map[int]*sql.DB)
	pool.tryConnect()
	return pool
}

// 尝试链接数据库
func (mcp *MysqlConnectPool) tryConnect() {
	//第一次连接 需要读取配置文件
	if mcp.isFirstConnect {
		//读取数据库配置文件
		conf, err := config.ParseDbConf(mcp.dbConfFile)
		if err != nil {
			panic("读取解析数据库配置文件失败:" + err.Error())
		}
		if len(conf.Mysql.Master) == 0 || len(conf.Mysql.Slave) == 0 {
			panic("主库和从库配置不能同时为空")
		}
		mcp.dbConf = conf
		mcp.isFirstConnect = false
	}

	//尝试连接主库
	var needConnect bool
	masterConf := mcp.dbConf.Mysql.Master[0]
	if mcp.writePool == nil {
		needConnect = true
	} else {
		if err := mcp.writePool.Ping(); err != nil {
			mcp.writePool = nil
			needConnect = true
			log.Printf("ping主库" + masterConf.Host + "失败：" + err.Error())
			//尝试关闭不可用连接
			_ = mcp.writePool.Close()
		}
	}
	if needConnect {
		writePool, err := mcp.connectItem(&masterConf)
		if err == nil {
			mcp.writePool = writePool
		} else {
			log.Printf("connect主库" + masterConf.Host + "失败：" + err.Error())
		}
	}

	//尝试连接从库
	for i, slaveConf := range mcp.dbConf.Mysql.Slave {
		slaveConn, ok := mcp.readPool[i]
		if ok {
			err := slaveConn.Ping()
			if err == nil {
				continue
			} else {
				_ = slaveConn.Close()
			}
		}

		readPool, err := mcp.connectItem(&slaveConf)
		if err == nil {
			mcp.readPool[i] = readPool
			continue
		}

		if ok {
			delete(mcp.readPool, i) //移除对象池
		}
		log.Printf("connect从库" + slaveConf.Host + "失败：" + err.Error())
	}

	//检测从库可用连接对象
	slaveLen := len(mcp.dbConf.Mysql.Slave)
	readlen := len(mcp.readPool)
	if (slaveLen > 0) && (slaveLen != readlen) {
		log.Printf("总共有%d个从库配置,有%d个从库未连接上,请及时处理", slaveLen, readlen)
	}
}

// 连接数据库
func (mcp *MysqlConnectPool) connectItem(mysqlItem *config.MysqlItem) (*sql.DB, error) {
	hostUrl := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s",
		mysqlItem.User,
		mysqlItem.Pwd,
		mysqlItem.Host,
		mysqlItem.Port,
		mysqlItem.DbName,
		mysqlItem.CharSet)
	dbPool, err := sql.Open("mysql", hostUrl)
	if err != nil {
		return nil, err
	}

	if mysqlItem.MaxIdleConns > 0 {
		dbPool.SetMaxIdleConns(mysqlItem.MaxIdleConns)
	}

	if mysqlItem.MaxOpenConns > 0 {
		dbPool.SetMaxOpenConns(mysqlItem.MaxOpenConns)
	}

	if mysqlItem.MaxLifetime > 0 {
		dbPool.SetConnMaxLifetime(time.Minute * time.Duration(mysqlItem.MaxLifetime))
	}

	return dbPool, nil
}

// 写连接
func (mdb *MysqlConnectPool) Write() *sql.DB {
	return mdb.writePool
}

// 读连接
func (mdb *MysqlConnectPool) Read() *sql.DB {
	readLen := len(mdb.readPool)
	if readLen == 0 {
		log.Println("无可用从库连接句柄，已切换到可写库备用，请及时处理")
		return mdb.writePool
	}

	num := rand.Intn(readLen)
	if _, ok := mdb.readPool[num]; ok {
		return mdb.readPool[num]
	} else {
		log.Println("取从库连接句柄失败，已切换到可写库备用，请及时处理")
		return mdb.writePool
	}
}

// 获取连接
func (mdb *MysqlConnectPool) SqlDB(master bool) *sql.DB {
	if !master {
		return mdb.Read()
	} else {
		return mdb.Write()
	}
}

// 释放数据连接
func (mdb *MysqlConnectPool) Free() {
	//捕获异常
	defer func() {
		if err := recover(); err != nil {
			pc, file, no, ok := runtime.Caller(1)
			log.Printf("FREE_ERROR:%s %s:%d ptr:%v ok:%v", err, file, no, pc, ok)
		}
	}()

	//释放主库连接
	if mdb.writePool != nil {
		_ = mdb.writePool.Close()
		mdb.writePool = nil
	}

	//释放从库连接
	for k, v := range mdb.readPool {
		_ = v.Close()
		delete(mdb.readPool, k)
	}

	//下次重新读取配置文件进行连接
	mdb.isFirstConnect = true
}

// 是否连接成功主库
func (mdb *MysqlConnectPool) IsConnectedWrite() (bool, error) {
	err := mdb.Write().Ping()
	if err != nil {
		return false, err
	}
	return true, nil
}

// 读库是否连接成功
func (mdb *MysqlConnectPool) IsConnectedRead() (bool, error) {
	if mdb.readPool == nil {
		return false, errors.New("读库暂无链接")
	}
	for _, v := range mdb.readPool {
		err := v.Ping()
		if err != nil {
			return false, err
		}
	}
	return true, nil
}
