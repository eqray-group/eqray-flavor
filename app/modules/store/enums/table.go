package enums

// TableStatusEnum 是桌台的占用状态。
// 取值必须与数据字典 table_status 保持一致，改动时两边同步。
type TableStatusEnum int8

const (
	TableStatusIdle     TableStatusEnum = 1 // 空闲
	TableStatusDining   TableStatusEnum = 2 // 使用中
	TableStatusCleaning TableStatusEnum = 3 // 待清台
)

// String 返回桌台状态的中文描述，便于列表展示与日志定位。
func (s TableStatusEnum) String() string {
	switch s {
	case TableStatusIdle:
		return "空闲"
	case TableStatusDining:
		return "使用中"
	case TableStatusCleaning:
		return "待清台"
	default:
		return "未知状态"
	}
}

// IsValid 判断桌台状态取值是否合法。
func (s TableStatusEnum) IsValid() bool {
	switch s {
	case TableStatusIdle, TableStatusDining, TableStatusCleaning:
		return true
	default:
		return false
	}
}
