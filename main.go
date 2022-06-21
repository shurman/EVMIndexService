package main

import (
	"fmt"
	"context"
	//"time"
	"math/big"
	//"reflect"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	//"github.com/ethereum/go-ethereum/params"
	//"gorm.io/driver/mysql"
	//"gorm.io/gorm"
)

/*const (
	UserName     string = "root"
	Password     string = "123qweasd"
	Addr         string = "127.0.0.1"
	Port         int    = 3306
	Database     string = "evm_data"
	MaxLifetime  int    = 10
	MaxOpenConns int    = 10
	MaxIdleConns int    = 10
)

type TxInfo struct {
	Txhash	string	`gorm:"type:varchar(70) NOT NULL primary key;"`
	Block	uint64 `gorm:"type:bigint(20) UNSIGNED NOT NULL;"`
	Txfrom	string	`gorm:"type:varchar(65) NOT NULL;"`
	Txto	string	`gorm:"type:varchar(65) NOT NULL;"`
	Value	string	`gorm:"type:varchar(30) NOT NULL;"`
	Nonce	uint64	`gorm:"type:bigint(20) UNSIGNED NOT NULL;"`
	Data	[]byte	`gorm:"type:blob;"`
}

type Blocks struct {
	Num	uint64	`gorm:"type:bigint(20) UNSIGNED NOT NULL primary key;"`
	Hash	string	`gorm:"type:varchar(70) NOT NULL;"`
	Timestamp	uint64	`gorm:"type:bigint(20) UNSIGNED NOT NULL;"`
	ParentHash	string	`gorm:"type:varchar(68) NOT NULL;"`
	Txs	[]*TxInfo	`gorm:"foreignKey:Block;references:Num;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

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
}*/

var ctx = context.TODO()
var client *ethclient.Client
var e error
func getlastblocknum() *big.Int {
	bnum, err := client.BlockNumber(ctx)
        if err != nil {
                panic(err)
        }
	return new(big.Int).SetUint64(bnum)
}

func main() {
	db := connectdb()
	if db == nil {
		panic("No DB instance")
		return
	}

	//fmt.Println("Create Blocks")
	db.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&Block{}, &TxInfo{})

	
	url := fmt.Sprintf("%s://%s:%d", "https", "data-seed-prebsc-2-s3.binance.org", 8545)
	client, e = ethclient.Dial(url)
	if e != nil {
		panic(e)
	}
	fmt.Println("Connected to RPC")

	chainID, err := client.NetworkID(ctx)
	if err != nil {
		panic(err)
	}

	block_num := big.NewInt(20396000)
	for {
		last_block_num := getlastblocknum()

		//fmt.Println(reflect.TypeOf(client))
		fmt.Printf("Start from %s... Current latest: %s\n", block_num.String(), last_block_num.String());

		for ; block_num.Cmp(last_block_num) <= 0; block_num.Add(block_num, big.NewInt(1)) {
			fmt.Printf("Parsing No.%s Block  ", block_num.String())

			block, err := client.BlockByNumber(ctx, block_num)
        		if err != nil {
                		panic(err)
	        	}
			fmt.Printf("%s  tx count:%d\n", block.Hash().Hex(), len(block.Transactions()))

			newblock := Block{Num:block.NumberU64(), Hash:block.Hash().Hex(), Timestamp:block.Time(), ParentHash:block.ParentHash().Hex()}
			result := db.Create(&newblock)
        		if result.Error != nil {
                		panic("Create failt")
		        }
        		if result.RowsAffected != 1 {
                		panic("RowsAffected Number failt")
		        }


			txs := make([]*TxInfo, 0)
			if len(block.Transactions()) > 0 {
				for _, tx := range block.Transactions() {
					//fmt.Println(tx.Hash().Hex())

					msg, err := tx.AsMessage(types.NewEIP155Signer(chainID), big.NewInt(1))
					if err != nil {
						panic(err)
					}
					//fmt.Println(msg.From().Hex())

					var txto string
					if tx.To() == nil {
						txto = "N/A"
					} else {
						txto = tx.To().Hex()
					}
					//fmt.Println(txto)

					txobj := &TxInfo{Txhash:tx.Hash().Hex(), Txfrom:msg.From().Hex(), Txto:txto, Value:tx.Value().String(), Nonce:tx.Nonce(), Data:"0x"+hex.EncodeToString(tx.Data()), Block:block.NumberU64()}

					txs = append(txs, txobj)
				}

				//db.Debug().Model(&newblock).Association("Txs").Append(txs)
				db.Create(&txs)
			}
		}
	}

	var fblock Block
	db.Debug().First(&fblock)
}

