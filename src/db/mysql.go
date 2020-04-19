package db

import (
	"cmnfunc"
	"database/sql"
)

func InitServerDB() error {
	if err := initDB(cmnfunc.Cfg["dbuser"], cmnfunc.Cfg["dbpwd"],
		cmnfunc.Cfg["dbaddr"], cmnfunc.Cfg["dbname"]); err != nil {
		return err
	}
	return nil
}
var DBCli *sql.DB

func initDB(user string, pwd string, dbaddr string, dbname string) error {
	/*DSN数据源名称
	  [username[:password]@][protocol[(address)]]/dbname[?param1=value1&paramN=valueN]
	  user@unix(/path/to/socket)/dbname
	  user:password@tcp(localhost:5555)/dbname?charset=utf8&autocommit=true
	  user:password@tcp([de:ad:be:ef::ca:fe]:80)/dbname?charset=utf8mb4,utf8
	  user:password@/dbname
	  无数据库: user:password@/
	*/
	s := user + ":" + pwd + "@tcp(" + dbaddr + ")/" + dbname + "?charset=utf8mb4"
	db, err := sql.Open("mysql", s)
	if err != nil {
		return err
	}
	err = db.Ping()
	if err != nil {
		return err
	}
	DBCli = db
	return nil
}