package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateFlowRecordView(c *gin.Context) {
	err := CreateFlowRecordCtl(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg":  err.Error(),
			"code": http.StatusBadRequest,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":  "success",
		"code": http.StatusOK,
	})
}

func CreateWorthView(c *gin.Context) {
	err := CreateWorthCtl(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg":  err.Error(),
			"code": http.StatusBadRequest,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":  "success",
		"code": http.StatusOK,
	})
}

func QueryHistoryProfitView(c *gin.Context) {
	res, err := QueryHistoryProfit(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg":  err.Error(),
			"code": http.StatusBadRequest,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":  "success",
		"code": http.StatusOK,
		"data": res,
	})
}

func QyeryProfitView(c *gin.Context) {
	res, err := QueryProfit(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg":  err.Error(),
			"code": http.StatusBadRequest,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":  "success",
		"code": http.StatusOK,
		"data": res,
	})
}

func FEnumView(c *gin.Context) {
	res, err := FEnum()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg":  err.Error(),
			"code": http.StatusBadRequest,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":  "success",
		"code": http.StatusOK,
		"data": res,
	})
}
