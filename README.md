# EVM Index Service (Not finished)

## Introduction
Implement EVM-based chain information indexer service for practice.

### Services
1. Block Indexer which is to download block & trasaction data to database
2. API Service to retrieve data from database

## Installation
1. Build docker image with `docker build -t evmindex . --no-cache`
1. Start image: `docker run -itd --name evmindex -p 8000:8000`
1. [To be done]
1. Config Setup [TBD]

## Trial Execution
1. `cd /go/src`
2. Start Service
  * Block Indexer: `go run /go/src/main.go /go/src/dbstruct.go -g [int] -s [int] -b [int]`
    * `-g`: Number of Goroutine (Default:3)
    * `-s`: Block is defined as stable after s block confirm (Default:20)
    * `-b`: Download from number b block (Default:20000000)
  * API Service: `go run /go/src/srvmain.go /go/src/dbstruct.go`
    * Default listen tp 8000 port

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
* Authentication with access token in header

1. Get Recent N Blocks Information
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
1. Get Specific Block Information
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
1. Get Specific Block Information
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
1. Initialize DB instance for each goroutine
2. API service loaded to much not needed data
3. Run services as daemon
