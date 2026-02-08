package server

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// FEnum 从数据库获取所有资金类型
// FEnum 从数据库获取所有未删除的资金类型
func FEnum() ([]map[string]any, error) {
	ftype := FTypeModel{}
	types, err := ftype.GetAllTypes()
	if err != nil {
		return nil, fmt.Errorf("查询资金类型失败: %w", err)
	}

	result := make([]map[string]any, 0, len(types))
	for _, t := range types {
		result = append(result, map[string]interface{}{
			"id":    t.ID,
			"name":  t.Name,
			"cname": t.Cname,
			"order": t.Order,
		})
	}
	return result, nil
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

// QueryByTime 查询指定时间范围内的利润数据（支持动态类型）
func QueryByTime(startTimeStr, endTimeStr string) ([]map[string]any, error) {
	res := make([]map[string]any, 0)
	worthModel := WorthModel{}
	flowRecordModel := FlowRecordModel{}

	// 获取所有资金类型
	ftype := FTypeModel{}
	types, err := ftype.GetAllTypes()
	if err != nil {
		return res, fmt.Errorf("获取资金类型失败: %w", err)
	}

	// 构建类型名称集合（用于筛选未删除的类型）
	typeNameSet := make(map[string]bool)
	fields := make([]string, 0, len(types))
	for _, t := range types {
		typeNameSet[t.Name] = true
		fields = append(fields, t.Name)
	}

	worthList, err := worthModel.QueryByTime(startTimeStr, endTimeStr)
	if err != nil {
		return res, err
	}

	if len(worthList) == 0 {
		return res, nil
	}

	// 构建第一个记录的初始值（从 TypeWorths 中获取）
	firstWorth := worthList[0]
	firstRecord := map[string]any{"time": firstWorth.Time}

	// 构建类型名称到值的映射（只包含未删除的类型）
	firstValueMap := make(map[string]float64)
	for _, tw := range firstWorth.TypeWorths {
		if typeNameSet[tw.TypeName] {
			firstValueMap[tw.TypeName] = tw.Value
		}
	}

	// 初始化所有类型的值
	for _, field := range fields {
		value := firstValueMap[field]
		firstRecord[field] = []float64{value, value, 0.0} // 现值，本金，利润
	}

	tmp := []map[string]any{firstRecord}

	// 查找初始时间后的资金变动记录
	flowRecordList, err := flowRecordModel.QueryByTime(firstWorth.Time, endTimeStr)
	if err != nil {
		return res, err
	}

	index := 0
	flowLen := len(flowRecordList)
	for _, worth := range worthList[1:] {
		currTime := worth.Time
		// 1. 初始化当前利润表
		currProfit := map[string]any{"time": currTime}

		// 构建当前记录的类型值映射（只包含未删除的类型）
		currValueMap := make(map[string]float64)
		for _, tw := range worth.TypeWorths {
			if typeNameSet[tw.TypeName] {
				currValueMap[tw.TypeName] = tw.Value
			}
		}

		for _, field := range fields {
			currentValue := currValueMap[field]
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

		// 4. 将当前利润数据添加到 tmp
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

	// 获取所有资金类型
	ftype := FTypeModel{}
	types, err := ftype.GetAllTypes()
	if err != nil {
		return res, fmt.Errorf("获取资金类型失败: %w", err)
	}

	fields := make([]string, 0, len(types))
	for _, t := range types {
		fields = append(fields, t.Name)
	}

	switch fType {
	case "all":
		// 汇总所有类型
		for _, cur_profit := range tmp {
			cur_worth, cur_principal := 0.0, 0.0
			for _, field := range fields {
				if fieldData, ok := cur_profit[field].([]float64); ok && len(fieldData) >= 2 {
					cur_worth += fieldData[0]
					cur_principal += fieldData[1]
				}
			}
			res = append(res, map[string]any{
				"time":      cur_profit["time"],
				"worth":     cur_worth,
				"principal": cur_principal,
				"profit":    cur_worth - cur_principal,
			})
		}
	default:
		// 检查是否是有效的类型名称
		typeExists := false
		for _, t := range types {
			if t.Name == fType {
				typeExists = true
				break
			}
		}
		if !typeExists {
			return res, fmt.Errorf("错误的参数值: %s", fType)
		}

		// 提取指定类型的时序变化
		for _, cur_profit := range tmp {
			if fieldData, ok := cur_profit[fType].([]float64); ok && len(fieldData) >= 2 {
				res = append(res, map[string]any{
					"time":      cur_profit["time"],
					"worth":     fieldData[0],
					"principal": fieldData[1],
					"profit":    fieldData[0] - fieldData[1],
				})
			}
		}
	}

	return res, nil
}

// WorthJSON 现值记录 JSON 结构（支持动态类型）
type WorthJSON struct {
	ID         uint            `json:"id"`          // 记录ID
	Time       string          `json:"time"`        // 现值时刻
	TypeWorths []TypeWorthJSON `json:"type_worths"` // 各类型价值列表
}

// TypeWorthJSON 类型价值 JSON 结构
type TypeWorthJSON struct {
	TypeName string  `json:"type_name"` // 类型名称（如 cash、stock_a）
	Value    float64 `json:"value"`     // 价值数值
}

// FTypeJSON 资金类型 JSON 结构
type FTypeJSON struct {
	Name  string `json:"name"`  // 类型名称（英文）
	Cname string `json:"cname"` // 类型中文名称
	Order int    `json:"order"` // 排序顺序
}

// CreateWorthCtl 创建现值记录
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

	// 创建主记录
	worthModel := WorthModel{
		Time:       currentTime,
		SyncStatus: 0, // 新创建的数据默认为未同步
	}

	// 转换类型价值数据（直接使用类型名，不验证是否存在，允许已删除的类型）
	typeWorths := make([]TypeWorthModel, 0, len(worthJson.TypeWorths))
	for _, tw := range worthJson.TypeWorths {
		typeWorths = append(typeWorths, TypeWorthModel{
			Model:    gorm.Model{}, // 确保 ID 为 0，让 GORM 自动生成
			TypeName: tw.TypeName,
			Value:    tw.Value,
		})
	}
	worthModel.TypeWorths = typeWorths

	if err := worthModel.Create(); err != nil {
		return err
	}
	return nil
}

// FType 相关控制器

// GetFTypesCtl 获取所有资金类型（返回与 FEnum 相同的数据格式）
func GetFTypesCtl(c *gin.Context) ([]map[string]any, error) {
	ftype := FTypeModel{}
	types, err := ftype.GetAllTypes()
	if err != nil {
		return nil, fmt.Errorf("查询资金类型失败: %w", err)
	}

	// 转换为与 FEnum 相同的数据格式
	result := make([]map[string]any, 0, len(types))
	for _, t := range types {
		result = append(result, map[string]interface{}{
			"id":    t.ID,
			"name":  t.Name,
			"cname": t.Cname,
			"order": t.Order,
		})
	}
	return result, nil
}

// CreateFTypeCtl 创建资金类型
func CreateFTypeCtl(c *gin.Context) error {
	var ftypeJson FTypeJSON
	if err := c.ShouldBindJSON(&ftypeJson); err != nil {
		return fmt.Errorf("参数错误: %w", err)
	}

	if ftypeJson.Name == "" {
		return fmt.Errorf("类型名称不能为空")
	}
	if ftypeJson.Cname == "" {
		return fmt.Errorf("类型中文名称不能为空")
	}

	// 检查是否存在（包括软删除的记录）
	var existing FTypeModel
	err := Mysql.Unscoped().Where("name = ?", ftypeJson.Name).First(&existing).Error

	if err == nil {
		// 记录已存在
		if existing.DeletedAt.Valid {
			// 如果是软删除的记录，恢复它
			// 恢复软删除：将 DeletedAt 设置为 NULL
			if err := Mysql.Unscoped().Model(&existing).Update("deleted_at", nil).Error; err != nil {
				return fmt.Errorf("恢复资金类型失败: %w", err)
			}
			// 更新其他字段
			if err := Mysql.Model(&existing).Updates(map[string]interface{}{
				"cname": ftypeJson.Cname,
				"order": ftypeJson.Order,
			}).Error; err != nil {
				return fmt.Errorf("更新资金类型失败: %w", err)
			}
			return nil
		} else {
			// 记录已存在且未删除
			return fmt.Errorf("资金类型 '%s' 已存在", ftypeJson.Name)
		}
	} else if err != gorm.ErrRecordNotFound {
		// 其他错误
		return fmt.Errorf("查询资金类型失败: %w", err)
	}

	// 记录不存在，创建新记录
	ftype := FTypeModel{
		Name:  ftypeJson.Name,
		Cname: ftypeJson.Cname,
		Order: ftypeJson.Order,
	}

	if err := ftype.Create(); err != nil {
		return fmt.Errorf("创建资金类型失败: %w", err)
	}
	return nil
}

// UpdateFTypeCtl 更新资金类型
func UpdateFTypeCtl(c *gin.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return fmt.Errorf("类型ID不能为空")
	}

	var ftypeJson FTypeJSON
	if err := c.ShouldBindJSON(&ftypeJson); err != nil {
		return fmt.Errorf("参数错误: %w", err)
	}

	var ftype FTypeModel
	if err := Mysql.First(&ftype, idStr).Error; err != nil {
		return fmt.Errorf("资金类型不存在: %w", err)
	}

	ftype.Name = ftypeJson.Name
	ftype.Cname = ftypeJson.Cname
	ftype.Order = ftypeJson.Order

	if err := ftype.Update(); err != nil {
		return fmt.Errorf("更新资金类型失败: %w", err)
	}
	return nil
}

// DeleteFTypeCtl 删除资金类型（允许直接删除，查询时会自动过滤已删除的类型）
func DeleteFTypeCtl(c *gin.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return fmt.Errorf("类型ID不能为空")
	}

	var ftype FTypeModel
	if err := Mysql.First(&ftype, idStr).Error; err != nil {
		return fmt.Errorf("资金类型不存在: %w", err)
	}

	// 直接删除，不检查使用情况（使用软删除，查询时会自动过滤）
	if err := ftype.Delete(); err != nil {
		return fmt.Errorf("删除资金类型失败: %w", err)
	}
	return nil
}

type FlowRecordJson struct {
	ID    uint    `json:"id"` // 记录ID
	Time  string  `json:"time"`
	Type  string  `json:"type"`
	Value float64 `json:"value"`
}

func CreateFlowRecordCtl(c *gin.Context) error {
	var flowRecordJson FlowRecordJson
	if err := c.ShouldBindJSON(&flowRecordJson); err != nil {
		return err
	}

	currentTime := flowRecordJson.Time
	if currentTime == "" {
		loc, _ := time.LoadLocation("Asia/Shanghai")
		currentTime = time.Now().In(loc).Format("2006-01-02 15:04:05")
	}

	modelList := []FlowRecordModel{}
	switch flowRecordJson.Type {
	case "cash":
		// ?? 批量变动，导致数据校验出错，暂时不用
		// if worth+flowRecordJson.Value < 0 {
		// 	return fmt.Errorf("现值不能为负，%s 当前值 %f, 变化值 %f", flowRecordJson.Type, worth, flowRecordJson.Value)
		// }
		modelList = append(modelList, FlowRecordModel{
			Time:  currentTime,
			Type:  flowRecordJson.Type,
			Value: flowRecordJson.Value,
		})

	default:

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

func QueryWorthCtl(c *gin.Context) ([]WorthJSON, error) {
	pageIndexStr := c.DefaultQuery("page_index", "1")
	pageLimitStr := c.DefaultQuery("page_limit", "10")

	pageIndex, err := strconv.Atoi(pageIndexStr)
	if err != nil || pageIndex <= 0 {
		pageIndex = 1
	}
	pageLimit, err := strconv.Atoi(pageLimitStr)
	if err != nil || pageLimit <= 0 {
		pageLimit = 10
	}

	worthModel := WorthModel{}
	worthList, err := worthModel.Query(pageIndex, pageLimit)
	if err != nil {
		return nil, err
	}

	res := make([]WorthJSON, 0, len(worthList))
	for _, worth := range worthList {
		typeWorths := make([]TypeWorthJSON, 0, len(worth.TypeWorths))
		for _, tw := range worth.TypeWorths {
			typeWorths = append(typeWorths, TypeWorthJSON{
				TypeName: tw.TypeName,
				Value:    tw.Value,
			})
		}

		res = append(res, WorthJSON{
			ID:         worth.ID,
			Time:       worth.Time,
			TypeWorths: typeWorths,
		})
	}

	return res, nil
}

func QueryFlowRecordCtl(c *gin.Context) ([]FlowRecordJson, error) {
	pageIndexStr := c.DefaultQuery("page_index", "1")
	pageLimitStr := c.DefaultQuery("page_limit", "10")
	pageIndex, err := strconv.Atoi(pageIndexStr)
	if err != nil || pageIndex <= 0 {
		pageIndex = 1
	}
	pageLimit, err := strconv.Atoi(pageLimitStr)
	if err != nil || pageLimit <= 0 {
		pageLimit = 10
	}
	flowRecordModel := FlowRecordModel{}
	recordList, err := flowRecordModel.Query(pageIndex, pageLimit)
	if err != nil {
		return nil, err
	}

	res := make([]FlowRecordJson, 0, len(recordList))
	for _, record := range recordList {
		res = append(res, FlowRecordJson{
			ID:    record.ID,
			Time:  record.Time,
			Type:  record.Type,
			Value: record.Value,
		})
	}

	return res, nil

}

// DeleteWorthCtl 删除现值记录
func DeleteWorthCtl(id uint) error {
	// idStr := c.Param("id")
	// if idStr == "" {
	// 	return fmt.Errorf("ID不能为空")
	// }

	// id, err := strconv.Atoi(idStr)
	// if err != nil {
	// 	return fmt.Errorf("无效的ID格式: %w", err)
	// }

	return Mysql.Transaction(func(tx *gorm.DB) error {
		// 1. 删除关联的子表记录 (TypeWorthModel)
		if err := tx.Where("worth_id = ?", id).Delete(&TypeWorthModel{}).Error; err != nil {
			return fmt.Errorf("删除关联价值记录失败: %w", err)
		}

		// 2. 删除主记录 (WorthModel)
		if err := tx.Delete(&WorthModel{}, id).Error; err != nil {
			return fmt.Errorf("删除现值主记录失败: %w", err)
		}

		return nil
	})
}

// DeleteFlowRecordCtl 删除流水记录
func DeleteFlowRecordCtl(id uint) error {
	// idStr := c.Param("id")
	// if idStr == "" {
	// 	return fmt.Errorf("ID不能为空")
	// }

	// id, err := strconv.Atoi(idStr)
	// if err != nil {
	// 	return fmt.Errorf("无效的ID格式: %w", err)
	// }

	// 执行软删除
	if err := Mysql.Delete(&FlowRecordModel{}, id).Error; err != nil {
		return fmt.Errorf("删除流水记录失败: %w", err)
	}

	return nil
}
