package support

import (
	"errors"

	mysqlDriver "github.com/go-sql-driver/mysql"
)

// mysqlDuplicateEntry 是 MySQL 唯一键冲突的错误号。
const mysqlDuplicateEntry = 1062

// IsDuplicate 判断错误是否为 MySQL 唯一键冲突（1062）。
// Repository 捕获后应转换成「该字段已存在」的业务错误码，禁止把驱动原始错误透传给前端（docs 6.8(5)）。
func IsDuplicate(err error) bool {
	var me *mysqlDriver.MySQLError
	return errors.As(err, &me) && me.Number == mysqlDuplicateEntry
}
