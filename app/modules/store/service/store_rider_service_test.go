package service

import (
	"strings"
	"testing"

	baseEnums "eqray-flavor/app/modules/base/enums"
	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/store/enums"
	"eqray-flavor/app/modules/store/model"
	"eqray-flavor/app/modules/store/repository"
)

// TestRiderStatusEnum 验证骑手状态的文案、合法性与接单 / 停用判断。
func TestRiderStatusEnum(t *testing.T) {
	cases := []struct {
		status enums.RiderStatusEnum
		name   string
	}{
		{enums.RiderStatusIdle, "空闲"},
		{enums.RiderStatusAccepting, "接单中"},
		{enums.RiderStatusDelivering, "配送中"},
		{enums.RiderStatusResting, "休息"},
		{enums.RiderStatusDisabled, "停用"},
	}
	for _, item := range cases {
		if !item.status.IsValid() {
			t.Fatalf("状态 %d 应为合法值", item.status)
		}
		if item.status.String() != item.name {
			t.Fatalf("状态 %d 文案错误：期望 %q，实际 %q", item.status, item.name, item.status.String())
		}
	}
	if enums.RiderStatusEnum(0).IsValid() {
		t.Fatal("状态 0 不合法")
	}
	if !enums.RiderStatusAccepting.CanAccept() {
		t.Fatal("只有接单中状态允许抢单")
	}
	if enums.RiderStatusIdle.CanAccept() || enums.RiderStatusDisabled.CanAccept() {
		t.Fatal("空闲 / 停用状态不允许抢单")
	}
	if !enums.RiderStatusResting.BlocksOrders() || !enums.RiderStatusDisabled.BlocksOrders() {
		t.Fatal("休息 / 停用状态应阻止有单骑手切换")
	}
}

// TestNormalizeRiderCreateInput 验证自动开通账号时姓名、手机号与并发上限的校验。
func TestNormalizeRiderCreateInput(t *testing.T) {
	in := RiderCreateInput{StoreID: 100, Name: " 张三 ", Mobile: " 13800138000 ", Status: enums.RiderStatusIdle}
	if err := normalizeRiderCreateInput(&in); err != nil {
		t.Fatalf("期望校验通过，实际返回错误：%v", err)
	}
	if in.Name != "张三" || in.Mobile != "13800138000" || in.MaxOrders != riderDefaultMaxOrders {
		t.Fatalf("字段清洗不符合预期：%+v", in)
	}

	noName := RiderCreateInput{StoreID: 100, Mobile: "13800138000"}
	if err := normalizeRiderCreateInput(&noName); err == nil {
		t.Fatal("自动开通账号时缺少姓名应校验失败")
	}

	badMobile := RiderCreateInput{StoreID: 100, Name: "张三", Mobile: "12345"}
	if err := normalizeRiderCreateInput(&badMobile); err == nil {
		t.Fatal("手机号格式非法应校验失败")
	}

	overLimit := RiderCreateInput{StoreID: 100, Name: "张三", Mobile: "13800138000", MaxOrders: 21}
	if err := normalizeRiderCreateInput(&overLimit); err == nil {
		t.Fatal("并发上限超过 20 应校验失败")
	}

	existingUser := RiderCreateInput{StoreID: 100, UserID: 200, MaxOrders: 5, Status: enums.RiderStatusResting}
	if err := normalizeRiderCreateInput(&existingUser); err != nil {
		t.Fatalf("指定已有账号时不应要求姓名与手机号：%v", err)
	}
}

// TestNormalizeRiderUpdateInput 验证编辑骑手时的状态、并发上限与备注清洗。
func TestNormalizeRiderUpdateInput(t *testing.T) {
	in := RiderUpdateInput{MaxOrders: 5, Status: enums.RiderStatusAccepting, Remark: " 午高峰 "}
	if err := normalizeRiderUpdateInput(&in); err != nil {
		t.Fatalf("期望校验通过，实际返回错误：%v", err)
	}
	if in.Remark != "午高峰" {
		t.Fatalf("备注未清洗：%q", in.Remark)
	}

	if err := normalizeRiderUpdateInput(&RiderUpdateInput{MaxOrders: 0, Status: enums.RiderStatusIdle}); err == nil {
		t.Fatal("并发上限为 0 应校验失败")
	}
	if err := normalizeRiderUpdateInput(&RiderUpdateInput{MaxOrders: 3, Status: enums.RiderStatusEnum(9)}); err == nil {
		t.Fatal("非法状态应校验失败")
	}
}

// TestEmployeeStatusForRider 验证停用骑手会同步禁用员工关联，其余状态保持启用。
func TestEmployeeStatusForRider(t *testing.T) {
	if employeeStatusForRider(enums.RiderStatusDisabled) != baseEnums.CommonStatusDisabled {
		t.Fatal("停用骑手应同步禁用门店员工")
	}
	for _, status := range []enums.RiderStatusEnum{
		enums.RiderStatusIdle,
		enums.RiderStatusAccepting,
		enums.RiderStatusDelivering,
		enums.RiderStatusResting,
	} {
		if employeeStatusForRider(status) != baseEnums.CommonStatusEnabled {
			t.Fatalf("状态 %d 不应禁用门店员工", status)
		}
	}
}

// TestGenerateRiderPassword 验证初始密码长度、字符集与随机性。
func TestGenerateRiderPassword(t *testing.T) {
	first, err := generateRiderPassword()
	if err != nil {
		t.Fatalf("生成初始密码失败：%v", err)
	}
	if len(first) != riderPasswordLength {
		t.Fatalf("初始密码长度应为 %d，实际 %d", riderPasswordLength, len(first))
	}
	for _, char := range first {
		if !strings.ContainsRune(riderPasswordAlphabet, char) {
			t.Fatalf("初始密码包含不在字符集内的字符：%q", char)
		}
	}

	second, err := generateRiderPassword()
	if err != nil {
		t.Fatalf("生成初始密码失败：%v", err)
	}
	if first == second {
		t.Fatal("两次生成的初始密码不应相同")
	}
}

// TestFilterEmployeesByScope 验证关键字命中账号后再按可见门店过滤员工。
func TestFilterEmployeesByScope(t *testing.T) {
	service := &StoreRiderService{}
	employees := []model.StoreEmployeeModel{
		{StoreID: 100, UserID: 1},
		{StoreID: 200, UserID: 1},
	}
	serviceFilter := repository.StoreRiderFilter{RestrictStores: true, StoreIDs: []types.ID{100}}
	got := service.filterEmployeesByScope(employees, serviceFilter)
	if len(got) != 1 || got[0].StoreID != 100 {
		t.Fatalf("可见门店过滤结果不符合预期：%+v", got)
	}

	contextFilter := repository.StoreRiderFilter{StoreID: 200}
	got = service.filterEmployeesByScope(employees, contextFilter)
	if len(got) != 1 || got[0].StoreID != 200 {
		t.Fatalf("门店上下文过滤结果不符合预期：%+v", got)
	}
}
