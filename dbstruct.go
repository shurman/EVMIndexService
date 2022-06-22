package main

import (
	"fmt"
	"time"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Log struct {
	Id	uint	`gorm:"type:int(10) UNSIGNED NOT NULL primary key AUTO_INCREMENT" json:"-"`
	Tx	string	`gorm:"type:varchar(70) NOT NULL;" json:"-"`
	Idx	uint	`gorm:"type:int(10) UNSIGNED NOT NULL;" json:"index"`
	Data	string	`gorm:"type:mediumtext" json:"data"`
}

type TxInfo struct {
	Txhash	string	`gorm:"type:varchar(70) NOT NULL primary key;" json:"tx_hash"`
	Block	uint64 `gorm:"type:bigint(20) UNSIGNED NOT NULL;" json:"-"`
	Txfrom	string	`gorm:"type:varchar(65) NOT NULL;" json:"from"`
	Txto	string	`gorm:"type:varchar(65) NOT NULL;" json:"to"`
	Value	string	`gorm:"type:varchar(30) NOT NULL;" json:"value"`
	Nonce	uint64	`gorm:"type:bigint(20) UNSIGNED NOT NULL;" json:"nonce"`
	Data	string	`gorm:"type:mediumtext" json:"data"`
	Logs	[]*Log	`gorm:"foreignKey:Tx;references:Txhash;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"logs"`
}

type Block struct {
	Num	uint64	`gorm:"type:bigint(20) UNSIGNED NOT NULL primary key;" json:"block_num"`
	Hash	string	`gorm:"type:varchar(70) NOT NULL;" json:"block_hash"`
	Timestamp	uint64	`gorm:"type:bigint(20) UNSIGNED NOT NULL;" json:"block_time"`
	ParentHash	string	`gorm:"type:varchar(68) NOT NULL;" json:"parenthash"`
	Txs	[]*TxInfo	`gorm:"foreignKey:Block;references:Num;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"txs,omitempty"`
	Txlist	[]string	`gorm:"-" json:"transactions,omitempty"`
	Stable	bool	`gorm:"type:bool" json:"isstable"`
}

type BlockSet struct {
	Blocks	[]Block	`json:"blocks"`
}


const (
	UserName     string = "root"
	Password     string = "123qweasd"
	Addr         string = "127.0.0.1"
	Port         int    = 3306
	Database     string = "evm_data"
	MaxLifetime  int    = 10
	MaxOpenConns int    = 10
	MaxIdleConns int    = 10
)

func connectdb() *gorm.DB {
	addr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True", UserName, Password, Addr, Port, Database)

	//fmt.Println("Connecting to DB")
	conn, err := gorm.Open(mysql.Open(addr), &gorm.Config{})
	if err != nil {
	fmt.Println("connection to mysql failed:", err)
		return nil
	}

	db, err1 := conn.DB()
	if err1 != nil {
		fmt.Println("gdet db failed:", err)
		return nil
	}
	db.SetConnMaxLifetime(time.Duration(MaxLifetime) * time.Second)
	db.SetMaxIdleConns(MaxIdleConns)
	db.SetMaxOpenConns(MaxOpenConns)

	fmt.Println("Connection to DB Succeeded")

	return conn
}
