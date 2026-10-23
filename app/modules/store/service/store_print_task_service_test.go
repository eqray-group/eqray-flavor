package service

import "testing"

// TestRenderPrintContent 校验模板占位符替换：常规字段、金额换算与未知占位符原样保留。
func TestRenderPrintContent(t *testing.T) {
	in := PrintTaskOrderInput{
		OrderNo:     "A20261023001",
		PickupCode:  "A18",
		TableName:   "A区 03",
		TypeText:    "堂食",
		ChannelText: "小程序",
		Remark:      "少辣",
		PayAmount:   12345,
		CreatedAt:   "2026-10-23 12:00:00",
	}

	template := "门店：{{storeName}}\n单号：{{orderNo}}\n桌台：{{tableName}}\n类型：{{orderType}}/{{channel}}\n" +
		"备注：{{remark}}\n应付：{{payAmount}}\n{{unknown}}"

	got := renderPrintContent(template, in, "示例门店")

	want := "门店：示例门店\n单号：A20261023001\n桌台：A区 03\n类型：堂食/小程序\n" +
		"备注：少辣\n应付：123.45\n{{unknown}}"
	if got != want {
		t.Fatalf("渲染结果不符\n实际: %q\n期望: %q", got, want)
	}
}

// TestRenderPrintItems 校验明细渲染：带规格时拼在名称后，备注另起一行，末行不留空行。
func TestRenderPrintItems(t *testing.T) {
	items := []PrintTaskItem{
		{Name: "宫保鸡丁", SkuName: "大份", Quantity: 2, Subtotal: 7200, Remark: "少辣"},
		{Name: "米饭", Quantity: 1, Subtotal: 200},
	}

	got := renderPrintItems(items)
	want := "宫保鸡丁(大份) x2  72.00\n  备注：少辣\n米饭 x1  2.00"
	if got != want {
		t.Fatalf("明细渲染不符\n实际: %q\n期望: %q", got, want)
	}

	if renderPrintItems(nil) != "" {
		t.Fatalf("空明细应渲染为空串")
	}
}

// TestFormatPrintYuan 校验分转元口径。
func TestFormatPrintYuan(t *testing.T) {
	cases := map[int64]string{0: "0.00", 100: "1.00", 12345: "123.45"}
	for cent, want := range cases {
		if got := formatPrintYuan(cent); got != want {
			t.Fatalf("formatPrintYuan(%d) = %q，期望 %q", cent, got, want)
		}
	}
}
