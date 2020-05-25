package mysql

import (
	"fmt"

	"github.com/micro-plat/hydra/conf/plat/db"
)

//MySQL mysql数据库连接信息
type MySQL struct {
	*db.DB
}

//New 构建oracle连接信息
func New(connString string, opts ...db.Option) *MySQL {
	return &MySQL{
		DB: db.New("mysql", connString, opts...),
	}
}

//NewBy 构建oracle连接信息
func NewBy(uName string, pwd string, serverIP string, dbName string, opts ...db.Option) *MySQL {
	return New(fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8", uName, pwd, serverIP, dbName), opts...)
}