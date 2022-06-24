# EVM Index Service

## Introduction
Implement EVM-based chain information indexer service for practice.

### Services
1. **Block Indexer** which is to download block & trasaction data to database
1. **API Service** to retrieve data from database

## Installation
* Modify `src/config` or keep it default
```bash
NUM_GOROUTINE=20        //Indexer will create 20 goroutines to parse blocks
BLOCK_STABLE_COUNT=20   //Block is flagged as stable after 20 blocks confirm
BLOCK_START=20400000    //Download from number 20400000 block
AVG_BLOCK_TIME=3        //Average block confirmation time(3sec)
RPC_URL=xxxxxxxxx       //Default URL is BSC Testnet RPC
```
* Build docker image with
```bash
$ docker build -t evmindeximg . --no-cache
```

* Start container
```bash
$ docker run -itd --name evmindexer -p 8000:8000 evmindeximg
```

* Login container
```bash
$ docker exec -it evmindexer bash
```

## Execution
* Login created container, and run

### Before Runnning
* **Warning:** Any security protection is not invovled, including DB password (default empty), firewall rule (default open), DDOS, and so on. Security should be concerned before using in production environment.
* Services use local mysql to store data. You may change to other database by configuring in `src/dbstruct.go`
```bash
UserName     string = "root"
Password     string = ""
Addr         string = "127.0.0.1"
Port         int    = 3306
Database     string = "evm_data"
MaxLifetime  int    = 10
MaxOpenConns int    = 100
MaxIdleConns int    = 40
```

### Auto Execution
```bash
$ /go/src/run.sh
```
* Check servce running
```bash
curl localhost:8000
```

### Custom Execution
#### Start Service
  * Must run under `/go/src`
  ```bash
  $ cd /go/src
  ```
  * Block Indexer:
  ```bash
  $ go run main.go dbstruct.go -g [int] -s [int] -b [int] -t [int] -u [int]
    -g: Number of Goroutine (Default: 20)
    -s: Block is flagged as stable after s block confirm (Default: 20)
    -b: Download from number b block (Default: 20000000)
    -t: Average block confirmation time(Default: 3)(sec)
    -u: RPC URL (Default: https://data-seed-prebsc-2-s3.binance.org:8545)
  ```
  * API Service (Listen to `8000` port)
  ```bash
  $ go run srvmain.go dbstruct.go
  ```

#### Alternative
  * Build Go code to executable
  ```bash
  $ go build [go file]
  ```

## Stable / Unstable Block
  Because the Longest Chain principle, the latest block may be replaced. (which implies unstable)

  Block which number is `< latest_block_num - BLOCK_STABLE_COUNT` will be flagged as stable block.

  **Block Indexer**, in each session, starts scanning from the oldest unstable block. When scanning to the latest block, **Block Indexer** sleeps `AVG_BLOCK_TIME * BLOCK_STABLE_COUNT / 3` seconds.
  Then start a new session.

## Database Schema
### block

| Field | Type | Tag |
| :----: | :----: | :----: |
| num | bigint(20) | Not Null, Primary Key |
| hash | varchar(70) | NN |
| timestamp | bigint(20) | NN |
| parentHash | varchar(70) | NN |
| txs | FK: tx_info.block | -- |
| stable | bool | NN |

### tx_info

| Field | Type | Tag |
| :----: | :----: | :----: |
| txhash | varchar(70) | NN PK |
| block | bigint(20) | NN |
| txfrom | varchar(65) | NN |
| txto | varchar(65) | NN |
| value | varchar(30) | -- |
| nonce | bigint(20) | NN |
| data | mediumtext | NN |
| logs | FK: log.tx | NN |

### log

| Field | Type | Tag |
| :----: | :----: | :----: |
| id | int(10) | NN, PK, AUTO_INCREMENT |
| tx | varchar(70) | NN |
| idx | int(10) | NN |
| data | mediumtext | NN |

## API List

### Get Recent N Blocks Information
  * URL: (GET) `/blocks?limit=N`
  * Return:
```
{
  blocks:[
    {
      block_num,
      block_hash,
      block_time,
      parent_hash
    },
    {...}, ... 
  ]
}
```
### Get Specific Block Information
  * URL: (GET) `/block/{num}`
  * Return:
```
{
  block_num,
  block_hash,
  block_time,
  parent_hash,
  transactions:[
    tx_hash, ...
  ]
}
```
### Get Specific Block Information
  * URL: (GET) `/transaction/{tx_hash}`
  * Return:
```
{
  tx_hash,
  from,
  to,
  nonce,
  data,
  value,
  logs: [
    {
      index,
      data,
    }
  ]
}
```

## Need to be optimized
1. API service loaded to much not needed data
1. Run services as daemon
1. Dynamically control number of goroutines
* A `g1-small` machine (1 Skylake vCPU, 1.7GB RAM) on GCP, with 20 goroutines running, costs 20%~30% CPU usage in average (peak 34%) and 1.1% memory. But the machine slows down due to Database I/O performance after increasing goroutines. As the result, with using an individual mysql server, 40 goroutines running is reasonable to run service at around 50% CPU usage.
