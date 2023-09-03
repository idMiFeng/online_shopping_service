package errno

import "errors"

var (
	ErrQueryFailed = errors.New("query db failed")
	ErrQueryEmpty  = errors.New("query empty") // 查询结果为空

	ErrUnderstock          = errors.New("understock")
	ErrReducestockFailed   = errors.New("reduce stock failed")   // 库存扣减失败
	ErrRollbackstockFailed = errors.New("rollback stock failed") // 回滚库存失败
)
