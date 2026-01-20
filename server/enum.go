package server

type FType int // 资金类型

const (
	STOCK_A FType = iota + 1 // Index = 1
	STOCK_M                  // Index = 2
	HONGLI
	BOND
	DEBT
	Cash
)

func (f FType) Index() int {
	return int(f)
}
func (f FType) String() string {
	return [...]string{"StockA", "StockM", "Hongli", "Bond", "Debt"}[f]
}
