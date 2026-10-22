package service

import (
	"testing"

	"eqray-flavor/app/modules/base/types"
	"eqray-flavor/app/modules/store/enums"
	"eqray-flavor/app/modules/store/model"
)

// TestTableStatusEnum 验证桌台状态的展示文案与合法性判断。
func TestTableStatusEnum(t *testing.T) {
	cases := []struct {
		status enums.TableStatusEnum
		name   string
	}{
		{enums.TableStatusIdle, "空闲"},
		{enums.TableStatusDining, "使用中"},
		{enums.TableStatusCleaning, "待清台"},
	}
	for _, item := range cases {
		if !item.status.IsValid() {
			t.Fatalf("状态 %d 应为合法值", item.status)
		}
		if item.status.String() != item.name {
			t.Fatalf("状态 %d 文案错误：期望 %q，实际 %q", item.status, item.name, item.status.String())
		}
	}
	if enums.TableStatusEnum(0).IsValid() {
		t.Fatal("状态 0 不合法")
	}
}

// TestNormalizeTableAreaInput 验证区域名称清洗与排序兜底。
func TestNormalizeTableAreaInput(t *testing.T) {
	in := TableAreaSaveInput{StoreID: 100, Name: "  大厅  ", Sort: -2, Remark: " 午市 "}
	if err := normalizeTableAreaInput(&in); err != nil {
		t.Fatalf("期望校验通过，实际返回错误：%v", err)
	}
	if in.Name != "大厅" || in.Remark != "午市" || in.Sort != 0 {
		t.Fatalf("字段清洗不符合预期：%+v", in)
	}

	blank := TableAreaSaveInput{StoreID: 100, Name: "  "}
	if err := normalizeTableAreaInput(&blank); err == nil {
		t.Fatal("空区域名称应校验失败")
	}
}

// TestNormalizeTableInput 验证桌号清洗、默认座位数与非法座位数。
func TestNormalizeTableInput(t *testing.T) {
	in := TableSaveInput{StoreID: 100, AreaID: 200, Name: " A01 ", Code: " S001-A01 ", Remark: " 靠窗 "}
	if err := normalizeTableInput(&in); err != nil {
		t.Fatalf("期望校验通过，实际返回错误：%v", err)
	}
	if in.Name != "A01" || in.Code != "S001-A01" || in.Seats != 4 || in.Remark != "靠窗" {
		t.Fatalf("字段清洗不符合预期：%+v", in)
	}

	invalidSeats := TableSaveInput{StoreID: 100, AreaID: 200, Name: "A01", Seats: 100}
	if err := normalizeTableInput(&invalidSeats); err == nil {
		t.Fatal("座位数超出范围应校验失败")
	}

	blankName := TableSaveInput{StoreID: 100, AreaID: 200, Name: " "}
	if err := normalizeTableInput(&blankName); err == nil {
		t.Fatal("空桌号应校验失败")
	}
}

// TestNormalizeTableBatchInput 验证批量生成参数清洗与数量上限。
func TestNormalizeTableBatchInput(t *testing.T) {
	in := TableBatchInput{StoreID: 100, AreaID: 200, Prefix: " A ", Start: 1, Count: 50}
	if err := normalizeTableBatchInput(&in); err != nil {
		t.Fatalf("期望校验通过，实际返回错误：%v", err)
	}
	if in.Prefix != "A" || in.Seats != 4 {
		t.Fatalf("批量参数清洗不符合预期：%+v", in)
	}

	overLimit := TableBatchInput{StoreID: 100, AreaID: 200, Prefix: "A", Start: 1, Count: 101, Seats: 4}
	if err := normalizeTableBatchInput(&overLimit); err == nil {
		t.Fatal("单次生成超过 100 张应校验失败")
	}
}

// TestBuildTableBatch 验证批量桌号补零与 50 个桌码互不重复。
func TestBuildTableBatch(t *testing.T) {
	store := &model.StoreModel{Code: "S001"}
	store.ID = types.ID(100)

	in := TableBatchInput{StoreID: 100, AreaID: 200, Prefix: "A", Start: 1, Count: 50, Seats: 4}
	list, err := buildTableBatch(store, in)
	if err != nil {
		t.Fatalf("批量生成失败：%v", err)
	}
	if len(list) != 50 {
		t.Fatalf("期望生成 50 张桌台，实际 %d", len(list))
	}
	if list[0].Name != "A01" || list[49].Name != "A50" {
		t.Fatalf("桌号补零不符合预期：首个 %q，末个 %q", list[0].Name, list[49].Name)
	}

	codes := make(map[string]bool, len(list))
	names := make(map[string]bool, len(list))
	for _, item := range list {
		if codes[item.Code] {
			t.Fatalf("桌码重复：%s", item.Code)
		}
		if names[item.Name] {
			t.Fatalf("桌号重复：%s", item.Name)
		}
		codes[item.Code] = true
		names[item.Name] = true
		if item.Code != "S001-"+item.Name {
			t.Fatalf("桌码未按门店编码生成：%s", item.Code)
		}
		if item.Status != enums.TableStatusIdle {
			t.Fatalf("新桌台应为空闲状态：%d", item.Status)
		}
	}

	next, err := buildTableBatch(store, TableBatchInput{StoreID: 100, AreaID: 200, Prefix: "A", Start: 99, Count: 3, Seats: 2})
	if err != nil {
		t.Fatalf("批量生成失败：%v", err)
	}
	if next[0].Name != "A099" || next[2].Name != "A101" {
		t.Fatalf("跨位数补零不符合预期：%q ~ %q", next[0].Name, next[2].Name)
	}
}
