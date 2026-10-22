package service

import (
	"testing"

	baseEnums "eqray-flavor/app/modules/base/enums"
	"eqray-flavor/app/modules/store/enums"
)

// validEmployeeInput 返回一份可通过基础校验的员工入参，测试中按需覆盖字段。
func validEmployeeInput() EmployeeSaveInput {
	return EmployeeSaveInput{
		StoreID: 100,
		UserID:  200,
		Role:    enums.StoreEmployeeRoleWaiter,
		Status:  baseEnums.CommonStatusEnabled,
		Remark:  " 午班 ",
	}
}

// TestNormalizeEmployeeInput 验证员工入参的备注清洗与枚举校验。
func TestNormalizeEmployeeInput(t *testing.T) {
	t.Run("清洗备注", func(t *testing.T) {
		in := validEmployeeInput()
		if err := normalizeEmployeeInput(&in); err != nil {
			t.Fatalf("期望校验通过，实际返回错误：%v", err)
		}
		if in.Remark != "午班" {
			t.Fatalf("备注未清洗：%q", in.Remark)
		}
	})

	t.Run("门店与账号必填", func(t *testing.T) {
		cases := []EmployeeSaveInput{
			{UserID: 200, Role: enums.StoreEmployeeRoleWaiter, Status: baseEnums.CommonStatusEnabled},
			{StoreID: 100, Role: enums.StoreEmployeeRoleWaiter, Status: baseEnums.CommonStatusEnabled},
		}
		for _, in := range cases {
			if err := normalizeEmployeeInput(&in); err == nil {
				t.Fatalf("缺少门店或账号时应校验失败：%+v", in)
			}
		}
	})

	t.Run("角色与状态枚举", func(t *testing.T) {
		invalidRole := validEmployeeInput()
		invalidRole.Role = 6
		if err := normalizeEmployeeInput(&invalidRole); err == nil {
			t.Fatal("非法角色应校验失败")
		}

		invalidStatus := validEmployeeInput()
		invalidStatus.Status = baseEnums.CommonStatusEnum(2)
		if err := normalizeEmployeeInput(&invalidStatus); err == nil {
			t.Fatal("非法状态应校验失败")
		}
	})
}
