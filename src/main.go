package main

import (
	"fmt"
	"context"
	"sync"
	"flag"
	"time"
	"math/big"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	//"github.com/ethereum/go-ethereum/params"
	//"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var gr_count = flag.Int("g", 20, "Number of Goroutine")
var i_st_range = flag.Int64("s", 20, "Block is defined as stable after s block confirm")
var start_block_num = flag.Int64("b", 20400000, "Download from number b block")
var avg_block_time = flag.Int64("t", 3, "Average block confirmation time")
var rpc_url = flag.String("u","https://data-seed-prebsc-2-s3.binance.org:8545", "RPC URL, Default URL is BSC Testnet RPC")

var db *gorm.DB
var ctx = context.TODO()
var chainID *big.Int
var e error

var wg *sync.WaitGroup
var numlock sync.Mutex

var latest_block_num *big.Int
var block_num_stable *big.Int
var block_num_flag *big.Int

func main() {
	flag.Parse()

	//Connect to DB and tables
	db = connectdb()
	if db == nil {
		panic("No DB instance")
		return
	}
	db.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&Block{}, &TxInfo{}, &Log{})

	//Connect to RPC
	client, err := ethclient.Dial(*rpc_url)
	if err != nil {
		panic(err)
		return
	}
	fmt.Printf("Connected to RPC %s\n", *rpc_url)

	//Get Chain ID
	chainID, e = client.NetworkID(ctx)
	if e != nil {
		panic(e)
		return
	}

	block_num_flag = big.NewInt(*start_block_num)
	stable_range := big.NewInt(-*i_st_range)

	wg = new(sync.WaitGroup)
	for {
		latest_block_num = GetLatestBlockNum(client)
		block_num_stable = new(big.Int).Add(latest_block_num, stable_range)

		fmt.Printf("Indexer starts from %s to %s with %d goroutine(s). (stable after %d blocks)\n", block_num_flag.String(), latest_block_num.String(), *gr_count, *i_st_range)

		//Use GoRoutine to download block data
		for i:=0; i < *gr_count; i++ {
			wg.Add(1)
			go GetBlockInfo(i, wg)
		}
		wg.Wait()

		//fmt.Println("Waiting for next session...")
		time.Sleep(time.Duration((*i_st_range) * (*avg_block_time)) * time.Second)

		block_num_flag = block_num_stable
	}

}

func GetLatestBlockNum(client *ethclient.Client) *big.Int {
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
	if block_num_stable.Cmp(block_num_flag) < 0 {
		isstable = false
	}else {
		isstable = true
	}

	return tmpnum, isstable
}

func GetBlockInfo(id int, wg *sync.WaitGroup){
	defer wg.Done()

	client, err := ethclient.Dial(*rpc_url)
	if err != nil {
		panic(err)
		return
	}

	for {
		//Get next block number to download
		numlock.Lock()
		block_num, isstable := GetNextParseNum()
		numlock.Unlock()
		if block_num == nil {
			return
		}

		//Check exist. If exists, check update
		fblock := new(Block)
		db.Find(&fblock, block_num.String())
		if fblock.Num != 0 && fblock.Stable {
			fmt.Printf("[%2d] Block %s (%t) exists and no need to update. Skip\n", id, block_num.String(), isstable)
			continue
		}

		block, err := client.BlockByNumber(ctx, block_num)
		if err != nil {
			panic(err)
			return
		}

		if fblock.Hash == block.Hash().Hex() {
			if isstable {
				fmt.Printf("[%2d] Block %s (%t) flags to stable\n", id, block_num.String(), isstable)
				db.Model(&Block{}).Where("num = ?", fblock.Num).Update("stable", true)
			}else{
				fmt.Printf("[%2d] Block %s (%t) exists but not stable yet\n", id, block_num.String(), isstable)
			}
			continue
		}else if fblock.Num != 0{
			fmt.Printf("[%2d] Block %s(%t) need to be updated w/ %d TXs\n", id, block_num.String(), isstable, len(block.Transactions()))
			db.Delete(&Block{}, block.NumberU64())
		}else {
			fmt.Printf("[%2d] Block %s (%t) Parsing w/ %d TXs\n", id, block_num.String(), isstable, len(block.Transactions()))
		}


		txs := make([]*TxInfo, 0)
		logs := make([]*Log, 0)
		var logslock sync.Mutex
		c_wg := new(sync.WaitGroup)
		for _, tx := range block.Transactions() {
			msg, err := tx.AsMessage(types.NewEIP155Signer(chainID), big.NewInt(1))
			if err != nil {
				panic(err)
				return
			}

			var txto string
			if tx.To() == nil {
				txto = "N/A"
			} else {
				txto = tx.To().Hex()
			}

			c_wg.Add(1)
			go func(){
				defer c_wg.Done()
				receipt, err := client.TransactionReceipt(ctx, tx.Hash())
				if err != nil {
					panic(err)
					return
				}
				for _, lg := range receipt.Logs {
					lgobj := &Log{Idx:lg.Index, Data:hex.EncodeToString(lg.Data), Tx:tx.Hash().Hex()}
					logslock.Lock()
					logs = append(logs, lgobj)
					logslock.Unlock()
				}
			}()

			txobj := &TxInfo{Txhash:tx.Hash().Hex(), Txfrom:msg.From().Hex(), Txto:txto, Value:tx.Value().String(), Nonce:tx.Nonce(), Data:"0x"+hex.EncodeToString(tx.Data()), Block:block.NumberU64()}

			txs = append(txs, txobj)
		}
		c_wg.Wait()

		newblock := Block{Num:block.NumberU64(), Hash:block.Hash().Hex(), Timestamp:block.Time(), ParentHash:block.ParentHash().Hex()}
		newblock.Stable = isstable

		db.Create(&newblock)
		if len(txs) > 0 {
			db.Create(&txs)
		}
		if len(logs) > 0 {
			db.Create(&logs)
		}
	}
}
