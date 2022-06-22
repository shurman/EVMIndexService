package main

import (
	"fmt"
	"context"
	"sync"
	//"time"
	"math/big"
	//"reflect"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	//"github.com/ethereum/go-ethereum/params"
	//"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB
var ctx = context.TODO()
var client *ethclient.Client
var chainID *big.Int
var e error

var wg *sync.WaitGroup
var sqllock sync.Mutex
var numlock sync.Mutex

var latest_block_num *big.Int
var block_num_stable *big.Int
var block_num_flag *big.Int

func main() {
	wg = new(sync.WaitGroup)
	db = connectdb()

	if db == nil {
		panic("No DB instance")
		return
	}

	//fmt.Println("Create Blocks")
	db.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&Block{}, &TxInfo{}, &Log{})

	
	url := fmt.Sprintf("%s://%s:%d", "https", "data-seed-prebsc-2-s3.binance.org", 8545)
	client, e = ethclient.Dial(url)
	if e != nil {
		panic(e)
		return
	}
	fmt.Println("Connected to RPC")

	chainID, e = client.NetworkID(ctx)
	if e != nil {
		panic(e)
		return
	}

	block_num_flag = big.NewInt(20427600)
	stable_range := big.NewInt(-20)
	gr_count := 3

	for {
		latest_block_num = GetLatestBlockNum()
		block_num_stable = new(big.Int).Add(latest_block_num, stable_range)

		//fmt.Println(reflect.TypeOf(client))
		fmt.Printf("Start from %s... Current latest: %s\n", block_num_flag.String(), latest_block_num.String())

		for i:=0; i<gr_count; i++ {
			wg.Add(1)
			go GetBlockInfo(i, wg)
		}
		wg.Wait()

		break
	}

}

func GetLatestBlockNum() *big.Int {
	bnum, err := client.BlockNumber(ctx)
        if err != nil {
                panic(err)
        }
	return new(big.Int).SetUint64(bnum)
}

func GetNextParseNum() (*big.Int, bool) {
	if block_num_flag.Cmp(latest_block_num) > 0 {
		return nil, false
	}
	tmpnum := new(big.Int).Add(block_num_flag, big.NewInt(0))
	block_num_flag.Add(block_num_flag, big.NewInt(1))

	var isstable bool
	if block_num_stable.Cmp(block_num_flag) > 0 {
		isstable = false
	}else {
		isstable = true
	}

	return tmpnum, isstable
}

func GetBlockInfo(id int, wg *sync.WaitGroup){
	defer wg.Done()

	for {
		numlock.Lock()
		block_num, isstable := GetNextParseNum()
		numlock.Unlock()
		if block_num == nil {
			return
		}
		//fmt.Printf("Parsing No.%s Block  ", block_num.String())

		block, err := client.BlockByNumber(ctx, block_num)
        	if err != nil {
              		panic(err)
	      	}
		fmt.Printf("[%d] Parsing Block %s w/ %d TXs\n", id, block_num.String(), len(block.Transactions()))

		txs := make([]*TxInfo, 0)
		logs := make([]*Log, 0)
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

			receipt, _ := client.TransactionReceipt(ctx, tx.Hash())
			//fmt.Println(receipt.Logs)
			for _, lg := range receipt.Logs {
				lgobj := &Log{Idx:lg.Index, Data:hex.EncodeToString(lg.Data), Tx:tx.Hash().Hex()}
				logs = append(logs, lgobj)
			}

			txobj := &TxInfo{Txhash:tx.Hash().Hex(), Txfrom:msg.From().Hex(), Txto:txto, Value:tx.Value().String(), Nonce:tx.Nonce(), Data:"0x"+hex.EncodeToString(tx.Data()), Block:block.NumberU64()}

			txs = append(txs, txobj)
		}

		newblock := Block{Num:block.NumberU64(), Hash:block.Hash().Hex(), Timestamp:block.Time(), ParentHash:block.ParentHash().Hex()}
		newblock.Stable = isstable
		sqllock.Lock()
		db.Delete(&Block{}, block.NumberU64())
		db.Create(&newblock)
		if len(block.Transactions()) > 0 {
			db.Create(&txs)
			db.Create(&logs)
		}
		sqllock.Unlock()
	}
}
