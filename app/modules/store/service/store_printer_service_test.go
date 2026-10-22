package service

import (
	"encoding/json"
	"strings"
	"testing"

	baseEnums "eqray-flavor/app/modules/base/enums"
	"eqray-flavor/app/modules/store/enums"
	"eqray-flavor/app/modules/store/model"
)

// TestPrinterEnums 验证打印设备类型与接入方式的展示文案和合法性。
func TestPrinterEnums(t *testing.T) {
	types := []struct {
		value enums.PrinterTypeEnum
		name  string
	}{
		{enums.PrinterTypeCashier, "前台小票"},
		{enums.PrinterTypeKitchen, "后厨"},
		{enums.PrinterTypeLabel, "标签"},
		{enums.PrinterTypeCall, "叫号"},
	}
	for _, item := range types {
		if !item.value.IsValid() {
			t.Fatalf("设备类型 %d 应为合法值", item.value)
		}
		if item.value.String() != item.name {
			t.Fatalf("设备类型 %d 文案错误：期望 %q，实际 %q", item.value, item.name, item.value.String())
		}
	}
	if enums.PrinterTypeEnum(0).IsValid() {
		t.Fatal("设备类型 0 不合法")
	}

	modes := []struct {
		value enums.PrinterAccessModeEnum
		name  string
	}{
		{enums.PrinterAccessCloud, "云打印机"},
		{enums.PrinterAccessLAN, "局域网"},
		{enums.PrinterAccessLocal, "本地客户端"},
	}
	for _, item := range modes {
		if !item.value.IsValid() {
			t.Fatalf("接入方式 %d 应为合法值", item.value)
		}
		if item.value.String() != item.name {
			t.Fatalf("接入方式 %d 文案错误：期望 %q，实际 %q", item.value, item.name, item.value.String())
		}
	}
	if enums.PrinterAccessModeEnum(0).IsValid() {
		t.Fatal("接入方式 0 不合法")
	}
}

// TestNormalizePrinterInput 验证设备字段清洗、默认份数与非法配置拦截。
func TestNormalizePrinterInput(t *testing.T) {
	in := PrinterSaveInput{
		StoreID:    100,
		Name:       " 前台小票机 ",
		DeviceNo:   " FP-001 ",
		Type:       enums.PrinterTypeCashier,
		AccessMode: enums.PrinterAccessCloud,
		Config:     ` {"vendor":"feie"} `,
		Secret:     "  secret  ",
		Status:     baseEnums.CommonStatusEnabled,
		Remark:     " 收银台 ",
	}
	if err := normalizePrinterInput(&in); err != nil {
		t.Fatalf("期望校验通过，实际返回错误：%v", err)
	}
	if in.Name != "前台小票机" || in.DeviceNo != "FP-001" || in.Remark != "收银台" {
		t.Fatalf("文本字段清洗不符合预期：%+v", in)
	}
	if in.Config != `{"vendor":"feie"}` || in.Copies != 1 {
		t.Fatalf("配置清洗或默认份数不符合预期：%+v", in)
	}
	if in.Secret != "  secret  " {
		t.Fatalf("密钥不应被裁剪空白：%q", in.Secret)
	}

	invalidConfig := PrinterSaveInput{
		StoreID: 100, Name: "前台", DeviceNo: "FP-002",
		Type: enums.PrinterTypeCashier, AccessMode: enums.PrinterAccessCloud,
		Config: `{"vendor":`, Status: baseEnums.CommonStatusEnabled,
	}
	if err := normalizePrinterInput(&invalidConfig); err == nil {
		t.Fatal("非法 JSON 配置应校验失败")
	}

	invalidCopies := PrinterSaveInput{
		StoreID: 100, Name: "前台", DeviceNo: "FP-003",
		Type: enums.PrinterTypeCashier, AccessMode: enums.PrinterAccessCloud,
		Copies: printerMaxCopies + 1, Status: baseEnums.CommonStatusEnabled,
	}
	if err := normalizePrinterInput(&invalidCopies); err == nil {
		t.Fatal("超出范围的打印份数应校验失败")
	}

	blankName := PrinterSaveInput{
		StoreID: 100, Name: "  ", DeviceNo: "FP-004",
		Type: enums.PrinterTypeCashier, AccessMode: enums.PrinterAccessCloud,
		Status: baseEnums.CommonStatusEnabled,
	}
	if err := normalizePrinterInput(&blankName); err == nil {
		t.Fatal("空设备名称应校验失败")
	}
}

// TestMaskPrinterSecret 验证密钥脱敏只保留末四位，短密钥不泄露任何字符。
func TestMaskPrinterSecret(t *testing.T) {
	cases := []struct {
		secret string
		mask   string
	}{
		{"", ""},
		{"1234", "****"},
		{"12345", "****2345"},
		{"abcdefgh", "****efgh"},
	}
	for _, item := range cases {
		if got := maskPrinterSecret(item.secret); got != item.mask {
			t.Fatalf("密钥 %q 脱敏错误：期望 %q，实际 %q", item.secret, item.mask, got)
		}
	}
}

// TestPrinterListRowHidesSecret 验证列表 JSON 不包含密钥密文，只暴露脱敏值。
func TestPrinterListRowHidesSecret(t *testing.T) {
	row := PrinterListRow{
		StorePrinterModel: model.StorePrinterModel{
			Name:       "前台小票机",
			DeviceNo:   "FP-001",
			Secret:     "ciphertext-must-not-leak",
			SecretMask: "****abcd",
		},
		HasSecret: true,
	}

	data, err := json.Marshal(row)
	if err != nil {
		t.Fatalf("序列化失败：%v", err)
	}
	text := string(data)
	if strings.Contains(text, "ciphertext-must-not-leak") {
		t.Fatalf("列表 JSON 泄露了密钥：%s", text)
	}
	if !strings.Contains(text, `"secretMask":"****abcd"`) || !strings.Contains(text, `"hasSecret":true`) {
		t.Fatalf("列表 JSON 缺少脱敏字段：%s", text)
	}
}
