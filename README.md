# EVM Index Service (Not finished)

## Introduction
Implement EVM-based chain information indexer service for practice.

### Services
1. Block Indexer which is to download block & trasaction data to database
1. API Service to retrieve data from database

## Installation
1. Modify `src/config` or keep it default
1. Build docker image with `docker build -t evmindeximg . --no-cache`
1. Start container: `docker run -itd --name evmindexer -p 8000:8000 evmindeximg`
1. Login container: `docker exec -it evmindexer bash`

## Execution
1. Login created container, `/go/src/run.sh`

## Custom Execution
1. `cd /go/src`
1. Start Service
  * Must run under `/go/src`
  * Block Indexer: `go run main.go dbstruct.go -g [int] -s [int] -b [int]`
    * `-g`: Number of Goroutine (Default: 20)
    * `-s`: Block is defined as stable after s block confirm (Default: 20)
    * `-b`: Download from number b block (Default: 20000000)
  * API Service: `go run srvmain.go dbstruct.go`
    * Listen to `8000` port
1. Alternative
  * Build Go code to executables by `go build [go file]`

## Database schema
* block

| Field | Type | Tag |
| :----: | :----: | :----: |
| num | bigint(20) | Not Null, Primary Key |
| hash | varchar(70) | NN |
| timestamp | bigint(20) | NN |
| parentHash | varchar(70) | NN |
| txs | FK: tx_info.block | -- |
| stable | bool | NN |

* tx_info

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

* log

| Field | Type | Tag |
| :----: | :----: | :----: |
| id | int(10) | NN, PK, AUTO_INCREMENT |
| tx | varchar(70) | NN |
| idx | int(10) | NN |
| data | mediumtext | NN |

## API List

### Get Recent N Blocks Information
  * url: (GET) `/blocks?limit=N`
  * return:
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
  * url: (GET) `/block/{num}`
  * return:
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
  * url: (GET) `/transaction/{tx_hash}`
  * return:
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
