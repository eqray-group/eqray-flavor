// Package support 提供跨模块复用的基础能力：账号摘要、行政区划、CSV 导出等。
package support

import (
	"bytes"
	"encoding/csv"
	"strconv"
)

// ExportRowLimit 是单次导出允许的最大行数，避免一次拉空整张表。
const ExportRowLimit = 5000

// BuildCSV 把表头与数据行编码成 CSV 字节流，带 UTF-8 BOM，Excel 可直接打开中文。
// 入参：表头与行数据（每行列数需与表头一致）；返回文件字节，编码失败时返回错误。
// 说明：导出统一走这里，避免各模块各写一套 BOM 与转义逻辑。
func BuildCSV(header []string, rows [][]string) ([]byte, error) {
	buffer := &bytes.Buffer{}
	// Excel 打开 UTF-8 CSV 需要 BOM，否则中文乱码
	buffer.WriteString("\xEF\xBB\xBF")

	writer := csv.NewWriter(buffer)
	if err := writer.Write(header); err != nil {
		return nil, err
	}
	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// FormatYuan 把金额（分）转成两位小数的元，供导出表格与展示使用。
func FormatYuan(cent int64) string {
	return strconv.FormatFloat(float64(cent)/100, 'f', 2, 64)
}
