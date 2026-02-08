package server

import (
	"net/http"
	"strconv"

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

// 资金类型管理视图

func CreateFTypeView(c *gin.Context) {
	err := CreateFTypeCtl(c)
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

func UpdateFTypeView(c *gin.Context) {
	err := UpdateFTypeCtl(c)
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

func DeleteFTypeView(c *gin.Context) {
	err := DeleteFTypeCtl(c)
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

// 现值记录查询与删除

func QueryWorthView(c *gin.Context) {
	res, err := QueryWorthCtl(c)
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

func DeleteWorthView(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg":  "无效的ID格式",
			"code": http.StatusBadRequest,
		})
		return
	}

	err = DeleteWorthCtl(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg":  err.Error(),
			"code": http.StatusInternalServerError,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":  "success",
		"code": http.StatusOK,
	})
}

// 流水记录查询与删除

func QueryFlowRecordView(c *gin.Context) {
	res, err := QueryFlowRecordCtl(c)
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

func DeleteFlowRecordView(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg":  "无效的ID格式",
			"code": http.StatusBadRequest,
		})
		return
	}

	err = DeleteFlowRecordCtl(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg":  err.Error(),
			"code": http.StatusInternalServerError,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":  "success",
		"code": http.StatusOK,
	})
}
