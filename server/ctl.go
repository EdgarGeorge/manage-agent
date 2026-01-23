package server

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func FEnum() ([]map[string]any, error) {

	return []map[string]interface{}{
		{
			"val":   STOCK_A,
			"cname": "A股",
			"name":  "stock_a",
		},
		{
			"val":   STOCK_M,
			"cname": "美股",
			"name":  "stock_m",
		},
		{
			"val":   HONGLI,
			"cname": "红利品类",
			"name":  "hongli",
		},
		{
			"val":   BOND,
			"cname": "债券",
			"name":  "bond",
		},
		{
			"val":   DEBT,
			"cname": "债权",
			"name":  "debt",
		},
		{
			"val":   Cash,
			"cname": "现金",
			"name":  "cash",
		},
	}, nil
}

func initProfitSlice(currentValue float64, tmp []map[string]any, field string) ([]float64, error) {
	if len(tmp) == 0 {
		return nil, fmt.Errorf("tmp 切片为空，无法获取 %s 的上一条本金", field)
	}
	lastTmp := tmp[len(tmp)-1]
	lastSlice, ok := lastTmp[field].([]float64)
	if !ok {
		return nil, fmt.Errorf("tmp 中 %s 字段类型错误，期望 []float64", field)
	}
	if len(lastSlice) < 2 {
		return nil, fmt.Errorf("tmp 中 %s 切片长度不足，无法获取本金", field)
	}
	return []float64{currentValue, lastSlice[1], 0.0}, nil
}

func QueryByTime(startTimeStr, endTimeStr string) ([]map[string]any, error) {
	res := make([]map[string]any, 0)
	worthModel := WorthModel{}
	flowRecordModel := FlowRecordModel{}

	worthList, err := worthModel.QueryByTime(startTimeStr, endTimeStr)
	if err != nil {
		return res, err
	}

	if len(worthList) == 0 {
		return res, nil
	}

	tmp := []map[string]any{}
	// 记录第一次现值记录为 初始本金
	tmp = append(tmp, map[string]any{
		"time":    worthList[0].Time,
		"cash":    []float64{worthList[0].Cash, worthList[0].Cash, 0.0}, // 现值，本金，利润
		"stock_a": []float64{worthList[0].StockA, worthList[0].StockA, 0.0},
		"stock_m": []float64{worthList[0].StockM, worthList[0].StockM, 0.0},
		"hongli":  []float64{worthList[0].Hongli, worthList[0].Hongli, 0.0},
		"bond":    []float64{worthList[0].Bond, worthList[0].Bond, 0.0},
		"debt":    []float64{worthList[0].Debt, worthList[0].Debt, 0.0},
	})

	// 查找初始时间后的资金变动记录
	flowRecordList, err := flowRecordModel.QueryByTime(worthList[0].Time, endTimeStr)
	if err != nil {
		return res, err
	}

	fields := []string{"cash", "stock_a", "stock_m", "hongli", "bond", "debt"}

	index := 0
	flowLen := len(flowRecordList)
	for _, worth := range worthList[1:] {
		currTime := worth.Time
		// 1. 初始化当前利润表
		currProfit := map[string]any{"time": currTime}
		for _, field := range fields {
			var currentValue float64
			switch field {
			case "cash":
				currentValue = worth.Cash
			case "stock_a":
				currentValue = worth.StockA
			case "stock_m":
				currentValue = worth.StockM
			case "hongli":
				currentValue = worth.Hongli
			case "bond":
				currentValue = worth.Bond
			case "debt":
				currentValue = worth.Debt
			}
			slice, err := initProfitSlice(currentValue, tmp, field)
			if err != nil {
				return res, err
			}
			currProfit[field] = slice
		}

		// 2. 应用资金变动记录
		for index < flowLen && flowRecordList[index].Time < currTime {
			record := flowRecordList[index]
			if slice, ok := currProfit[record.Type].([]float64); ok && len(slice) >= 2 {
				slice[1] += record.Value // 本金 += 变动值
			}
			index++
		}

		// 3. 批量计算所有字段的利润
		for _, field := range fields {
			if slice, ok := currProfit[field].([]float64); ok && len(slice) >= 3 {
				slice[2] = slice[0] - slice[1] // 利润 = 现值 - 本金
			}
		}

		// 4. 将当前利润数据添加到 tmp（根据业务需求）
		tmp = append(tmp, currProfit)
	}
	return tmp, nil
}

func QueryProfit(c *gin.Context) (map[string]any, error) {
	res := map[string]any{}
	startDate := c.DefaultQuery("start_date", "")
	endDate := c.DefaultQuery("end_date", "")
	if startDate == "" || endDate == "" {
		return res, fmt.Errorf("param error")
	}

	if _, err := time.Parse("2006-01-02", startDate); err != nil {
		return res, fmt.Errorf("start_date 格式错误，需为 YYYY-MM-DD，当前值：%s", startDate)
	}
	if _, err := time.Parse("2006-01-02", endDate); err != nil {
		return res, fmt.Errorf("end_date 格式错误，需为 YYYY-MM-DD，当前值：%s", endDate)
	}

	startTimeStr := startDate + " 00:00:00"
	endTimeStr := endDate + " 23:59:59"
	tmp, err := QueryByTime(startTimeStr, endTimeStr)
	if err != nil {
		return res, err
	}

	if len(tmp) != 0 {
		res = tmp[len(tmp)-1]
	}
	return res, nil
}

func QueryHistoryProfit(c *gin.Context) ([]map[string]any, error) {
	res := []map[string]any{}
	startDate := c.DefaultQuery("start_date", "")
	endDate := c.DefaultQuery("end_date", "")
	fType := c.DefaultQuery("type", "")

	if startDate == "" || endDate == "" || fType == "" {
		return res, fmt.Errorf("param error")
	}

	if _, err := time.Parse("2006-01-02", startDate); err != nil {
		return res, fmt.Errorf("start_date 格式错误，需为 YYYY-MM-DD，当前值：%s", startDate)
	}
	if _, err := time.Parse("2006-01-02", endDate); err != nil {
		return res, fmt.Errorf("end_date 格式错误，需为 YYYY-MM-DD，当前值：%s", endDate)
	}

	startTimeStr := startDate + " 00:00:00"
	endTimeStr := endDate + " 23:59:59"
	tmp, err := QueryByTime(startTimeStr, endTimeStr)
	if len(tmp) == 0 {
		return res, nil
	}
	if err != nil {
		return res, err
	}
	fields := []string{"cash", "stock_a", "stock_m", "hongli", "bond", "debt"}

	switch fType {
	case "cash", "stock_a", "stock_m", "hongli", "bond", "debt":
		// 提取时序变化
		for _, cur_profit := range tmp {
			res = append(res, map[string]any{
				"time":      cur_profit["time"],
				"worth":     cur_profit[fType].([]float64)[0],
				"principal": cur_profit[fType].([]float64)[1],
				"profit":    cur_profit[fType].([]float64)[0] - cur_profit[fType].([]float64)[1],
			})
		}

	case "all":
		for _, cur_profit := range tmp {
			cur_worth, cur_principal := 0.0, 0.0
			for _, field := range fields {
				cur_worth += cur_profit[field].([]float64)[0]
				cur_principal += cur_profit[field].([]float64)[1]
			}
			res = append(res, map[string]any{
				"time":      cur_profit["time"],
				"worth":     cur_worth,
				"principal": cur_principal,
				"profit":    cur_worth - cur_principal,
			})
		}
	default:
		return res, fmt.Errorf("错误的参数值: %s", fType)
	}

	return res, nil

}

type WorthJSON struct {
	Time   string  `json:"time"`    // 现值时刻
	Cash   float64 `json:"cash"`    // 现金
	StockA float64 `json:"stock_a"` // A股
	StockM float64 `json:"stock_m"` // M股
	Hongli float64 `json:"hongli"`  // 红利
	Bond   float64 `json:"bond"`    // 债券
	Debt   float64 `json:"debt"`    // 债权
}

func CreateWorthCtl(c *gin.Context) error {
	var worthJson WorthJSON
	if err := c.ShouldBindJSON(&worthJson); err != nil {
		return err
	}

	currentTime := worthJson.Time
	if currentTime == "" {
		loc, _ := time.LoadLocation("Asia/Shanghai")
		currentTime = time.Now().In(loc).Format("2006-01-02 15:04:05")
	}

	worthModel := WorthModel{
		Time:   currentTime,
		Cash:   worthJson.Cash,
		StockA: worthJson.StockA,
		StockM: worthJson.StockM,
		Hongli: worthJson.Hongli,
		Bond:   worthJson.Bond,
		Debt:   worthJson.Debt,
	}
	if err := worthModel.Create(); err != nil {
		return err
	}
	return nil

}

type FlowRecordJson struct {
	Time  string  `json:"time"`
	Type  string  `json:"type"`
	Value float64 `json:"value"`
}

func CreateFlowRecordCtl(c *gin.Context) error {
	var flowRecordJson FlowRecordJson
	if err := c.ShouldBindJSON(&flowRecordJson); err != nil {
		return err
	}
	// var worthModel WorthModel
	// lastWorth, err := worthModel.GetLatestWorth()
	// if err != nil {
	// 	return err
	// }

	// worth := 0.0

	// switch flowRecordJson.Type {
	// case "cash":
	// 	worth = lastWorth.Cash
	// case "stock_a":
	// 	worth = lastWorth.StockA
	// case "stock_m":
	// 	worth = lastWorth.StockM
	// case "hongli":
	// 	worth = lastWorth.Hongli
	// case "bond":
	// 	worth = lastWorth.Bond
	// case "debt":
	// 	worth = lastWorth.Debt
	// }

	currentTime := flowRecordJson.Time
	if currentTime == "" {
		loc, _ := time.LoadLocation("Asia/Shanghai")
		currentTime = time.Now().In(loc).Format("2006-01-02 15:04:05")
	}

	modelList := []FlowRecordModel{}
	switch flowRecordJson.Type {
	case "cash":
		// ?? 不支持批量变动，导致数据校验出错
		// if worth+flowRecordJson.Value < 0 {
		// 	return fmt.Errorf("现值不能为负，%s 当前值 %f, 变化值 %f", flowRecordJson.Type, worth, flowRecordJson.Value)
		// }
		modelList = append(modelList, FlowRecordModel{
			Time:  currentTime,
			Type:  flowRecordJson.Type,
			Value: flowRecordJson.Value,
		})

	case "stock_a", "stock_m", "hongli", "bond", "debt":

		// if worth+flowRecordJson.Value < 0 {
		// 	return fmt.Errorf("现值不能为负，%s 当前值 %f, 变化值 %f", flowRecordJson.Type, worth, flowRecordJson.Value)
		// }
		// if lastWorth.Cash-flowRecordJson.Value < 0 {
		// 	return fmt.Errorf("现值不能为负，%s 当前值 %f, 变化值 %f", flowRecordJson.Type, lastWorth.Cash, -flowRecordJson.Value)
		// }
		modelList = append(modelList, FlowRecordModel{
			Time:  currentTime,
			Type:  flowRecordJson.Type,
			Value: flowRecordJson.Value,
		})
		modelList = append(modelList, FlowRecordModel{
			Time:  currentTime,
			Type:  "cash",
			Value: -flowRecordJson.Value,
		})
	}

	flowRecordModel := FlowRecordModel{}
	if err := flowRecordModel.CreateInBatch(modelList); err != nil {
		return err
	}
	return nil
}
