package service

import (
	"testing"

	baseEnums "eqray-flavor/app/modules/base/enums"
	"eqray-flavor/app/modules/store/enums"
)

// TestPrintTypeEnum 验证打印类型枚举的中文文案与合法性。
func TestPrintTypeEnum(t *testing.T) {
	cases := []struct {
		value enums.PrintTypeEnum
		name  string
	}{
		{enums.PrintTypeCheckout, "结账单"},
		{enums.PrintTypePreBill, "预结单"},
		{enums.PrintTypeKitchen, "厨打"},
		{enums.PrintTypeReturn, "退菜"},
		{enums.PrintTypeLabel, "标签"},
		{enums.PrintTypeRefund, "退款单"},
		{enums.PrintTypeShift, "交班"},
		{enums.PrintTypeDailyBill, "日结"},
	}
	for _, item := range cases {
		if !item.value.IsValid() {
			t.Fatalf("打印类型 %d 应为合法值", item.value)
		}
		if item.value.String() != item.name {
			t.Fatalf("打印类型 %d 文案错误：期望 %q，实际 %q", item.value, item.name, item.value.String())
		}
	}
	if enums.PrintTypeEnum(0).IsValid() {
		t.Fatal("打印类型 0 不合法")
	}
}

// TestNormalizePrintTemplateInput 验证模板字段清洗、正文空白保留与纸宽校验。
func TestNormalizePrintTemplateInput(t *testing.T) {
	content := "  门店：{{storeName}}\n\n明细..."
	in := PrintTemplateSaveInput{
		StoreID:     100,
		Name:        " 前台 80mm 结账单 ",
		PrinterType: enums.PrinterTypeCashier,
		PrintType:   enums.PrintTypeCheckout,
		PaperWidth:  80,
		Content:     content,
		Status:      baseEnums.CommonStatusEnabled,
		Remark:      " 默认模板 ",
	}
	if err := normalizePrintTemplateInput(&in); err != nil {
		t.Fatalf("期望校验通过，实际返回错误：%v", err)
	}
	if in.Name != "前台 80mm 结账单" || in.Remark != "默认模板" {
		t.Fatalf("文本字段清洗不符合预期：%+v", in)
	}
	if in.Content != content {
		t.Fatalf("模板正文不应被裁剪：%q", in.Content)
	}

	blankContent := in
	blankContent.Content = " \n\t "
	if err := normalizePrintTemplateInput(&blankContent); err == nil {
		t.Fatal("空模板正文应校验失败")
	}

	invalidWidth := in
	invalidWidth.PaperWidth = printTemplateMaxPaperWidth + 1
	if err := normalizePrintTemplateInput(&invalidWidth); err == nil {
		t.Fatal("超出范围的纸宽应校验失败")
	}

	invalidKind := in
	invalidKind.PrintType = enums.PrintTypeEnum(99)
	if err := normalizePrintTemplateInput(&invalidKind); err == nil {
		t.Fatal("非法打印类型应校验失败")
	}
}

// TestNextPrintTemplateVersion 验证正文未变化保持版本，变化后递增且异常旧版本兜底为 1。
func TestNextPrintTemplateVersion(t *testing.T) {
	if got := nextPrintTemplateVersion(3, "same", "same"); got != 3 {
		t.Fatalf("正文未变化时版本应保持 3，实际 %d", got)
	}
	if got := nextPrintTemplateVersion(3, "old", "new"); got != 4 {
		t.Fatalf("正文变化时版本应递增为 4，实际 %d", got)
	}
	if got := nextPrintTemplateVersion(0, "old", "new"); got != 2 {
		t.Fatalf("异常旧版本应按 1 兜底后递增为 2，实际 %d", got)
	}
}
