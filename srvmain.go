package main

import (
	//"fmt"
	"strconv"
	"github.com/gin-gonic/gin"
	"net/http"
	"gorm.io/gorm"
)

type IndexData struct {
	Title   string
	Content string
}

var db *gorm.DB

func lastblock(c *gin.Context) {
	nlimit := c.Query("limit")

	limit, err := strconv.Atoi(nlimit)
	if err != nil {
		c.JSON(400, nil)
		return
	}

	var fblocks []Block
	db.Order("num DESC").Limit(limit).Find(&fblocks)
	result := BlockSet{Blocks:fblocks}

	c.JSON(200, result)
}
func getblock(c *gin.Context) {
	id := c.Param("id")

	var fblock Block
	db.Preload("Txs").Find(&fblock, id)
	fblock.Txlist = make([]string, 0)
	for _, tx := range fblock.Txs {
		fblock.Txlist = append(fblock.Txlist, tx.Txhash)
	}
	fblock.Txs = nil

	if fblock.Num == 0 {
		c.JSON(404, nil)
		return
	}else {
		c.JSON(200, fblock)
	}
}
func gettx(c *gin.Context) {
	txhash := c.Param("txHash")

	var ftx TxInfo
	db.Preload("Logs").Where(&TxInfo{Txhash:txhash}).Find(&ftx)

	if ftx.Txhash == "" {
		c.JSON(404, nil)
		return
	}else {
		c.JSON(200, ftx)
	}
}

func test(c *gin.Context) {
	data := new(IndexData)
	data.Title = "Homepage"
	data.Content = "test"
	c.HTML(http.StatusOK, "index.html", data)
}

func main() {
	db = connectdb()
	if db == nil {
		panic("No DB instance")
		return
	}
	db.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&Block{}, &TxInfo{}, &Log{})


	server := gin.Default()
	server.LoadHTMLGlob("template/*")
	
	server.GET("/", test)
	server.GET("/blocks", lastblock)
	server.GET("/block/:id", getblock)
	server.GET("/transaction/:txHash", gettx)
	server.RedirectFixedPath = true


	server.Run(":8000")
}
