package repository

import (
	"eqray-flavor/app/modules/base/types"
)

// idValues 把雪花 ID 切片转成 WhereIn 需要的 []any，空集合由调用方提前返回。
func idValues(ids []types.ID) []any {
	values := make([]any, 0, len(ids))
	for _, id := range ids {
		values = append(values, id)
	}

	return values
}

// uniqueIDs 按传入顺序去重，用于批量查询与批量写入前的清洗。
func uniqueIDs(ids []types.ID) []types.ID {
	list := make([]types.ID, 0, len(ids))
	seen := make(map[types.ID]bool, len(ids))
	for _, id := range ids {
		if id <= 0 || seen[id] {
			continue
		}
		seen[id] = true
		list = append(list, id)
	}

	return list
}
