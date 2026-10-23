package support

import (
	"strings"
	"testing"
)

// TestBuildCSV 校验导出文件带 BOM、中文原样保留、含逗号的值被正确转义。
func TestBuildCSV(t *testing.T) {
	data, err := BuildCSV([]string{"订单号", "备注"}, [][]string{
		{"A001", "少辣,加急"},
		{"A002", "不要香菜"},
	})
	if err != nil {
		t.Fatalf("BuildCSV 返回错误: %v", err)
	}

	if len(data) < 3 || string(data[:3]) != "\xEF\xBB\xBF" {
		t.Fatalf("导出文件缺少 UTF-8 BOM")
	}

	content := string(data[3:])
	lines := strings.Split(strings.TrimRight(content, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("期望 3 行（表头 + 2 条数据），实际 %d 行: %q", len(lines), lines)
	}
	if lines[0] != "订单号,备注" {
		t.Fatalf("表头不符合预期: %q", lines[0])
	}
	// 含逗号的字段必须被引号包裹，否则列会错位
	if lines[1] != `A001,"少辣,加急"` {
		t.Fatalf("含逗号字段未正确转义: %q", lines[1])
	}
	if lines[2] != "A002,不要香菜" {
		t.Fatalf("普通行不符合预期: %q", lines[2])
	}
}

// TestFormatYuan 校验分转元的口径：整元补两位小数、负数保留符号。
func TestFormatYuan(t *testing.T) {
	cases := []struct {
		cent int64
		want string
	}{
		{cent: 0, want: "0.00"},
		{cent: 100, want: "1.00"},
		{cent: 12345, want: "123.45"},
		{cent: -500, want: "-5.00"},
	}

	for _, c := range cases {
		if got := FormatYuan(c.cent); got != c.want {
			t.Fatalf("FormatYuan(%d) = %q，期望 %q", c.cent, got, c.want)
		}
	}
}
