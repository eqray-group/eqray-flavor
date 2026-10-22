package service

import (
	"encoding/json"
	"strings"
	"testing"

	"eqray-flavor/app/modules/store/enums"
)

// validStoreInput 返回一份可通过基础校验的门店入参，测试中按需覆盖字段。
func validStoreInput() StoreSaveInput {
	return StoreSaveInput{
		Type:      enums.StoreTypeShop,
		Code:      " S001 ",
		Name:      " 测试门店 ",
		Status:    enums.StoreStatusOpen,
		OpenTime:  "09:00",
		CloseTime: "23:00",
	}
}

// TestNormalizeStoreInput 验证门店入参的字符串清洗、枚举与时间校验。
func TestNormalizeStoreInput(t *testing.T) {
	t.Run("清洗字符串并修正负数排序", func(t *testing.T) {
		in := validStoreInput()
		in.Sort = -1

		if err := normalizeStoreInput(&in); err != nil {
			t.Fatalf("期望校验通过，实际返回错误：%v", err)
		}
		if in.Code != "S001" || in.Name != "测试门店" {
			t.Fatalf("字符串未清洗：code=%q name=%q", in.Code, in.Name)
		}
		if in.Sort != 0 {
			t.Fatalf("负数排序应归零，实际为 %d", in.Sort)
		}
	})

	t.Run("跨夜与全天营业合法", func(t *testing.T) {
		for _, times := range [][2]string{{"18:00", "02:00"}, {"00:00", "00:00"}} {
			in := validStoreInput()
			in.OpenTime, in.CloseTime = times[0], times[1]
			if err := normalizeStoreInput(&in); err != nil {
				t.Fatalf("营业时间 %s~%s 应合法，实际返回：%v", times[0], times[1], err)
			}
		}
	})

	t.Run("非法营业时间被拒绝", func(t *testing.T) {
		in := validStoreInput()
		in.CloseTime = "25:00"
		if err := normalizeStoreInput(&in); err == nil {
			t.Fatal("25:00 应被判为非法营业时间")
		}
	})

	t.Run("非法门店类型与状态被拒绝", func(t *testing.T) {
		in := validStoreInput()
		in.Type = enums.StoreTypeEnum(9)
		if err := normalizeStoreInput(&in); err == nil {
			t.Fatal("非法门店类型应被拒绝")
		}

		in = validStoreInput()
		in.Status = enums.StoreStatusEnum(9)
		if err := normalizeStoreInput(&in); err == nil {
			t.Fatal("非法经营状态应被拒绝")
		}
	})

	t.Run("编码与名称不能为空", func(t *testing.T) {
		in := validStoreInput()
		in.Code = "   "
		if err := normalizeStoreInput(&in); err == nil {
			t.Fatal("空编码应被拒绝")
		}

		in = validStoreInput()
		in.Name = "   "
		if err := normalizeStoreInput(&in); err == nil {
			t.Fatal("空名称应被拒绝")
		}
	})

	t.Run("经纬度越界被拒绝", func(t *testing.T) {
		in := validStoreInput()
		in.Longitude = 181
		if err := normalizeStoreInput(&in); err == nil {
			t.Fatal("经度 181 应被拒绝")
		}

		in = validStoreInput()
		in.Latitude = -91
		if err := normalizeStoreInput(&in); err == nil {
			t.Fatal("纬度 -91 应被拒绝")
		}
	})

	t.Run("非法配送规则被拒绝", func(t *testing.T) {
		in := validStoreInput()
		in.DeliveryRule = `{"radius":-1}`
		if err := normalizeStoreInput(&in); err == nil {
			t.Fatal("负配送半径应被拒绝")
		}
	})
}

// TestNormalizeDeliveryRule 验证配送规则 JSON 的解析、校验与统一序列化。
func TestNormalizeDeliveryRule(t *testing.T) {
	t.Run("空串视为未配置", func(t *testing.T) {
		got, err := normalizeDeliveryRule("  ")
		if err != nil {
			t.Fatalf("空串不应报错：%v", err)
		}
		if got != "" {
			t.Fatalf("空串应原样返回，实际为 %q", got)
		}
	})

	t.Run("合法规则被统一序列化", func(t *testing.T) {
		raw := `{"radius":3000,"minAmount":2000,"deliveryFee":500,"packingFee":100,"etaMinutes":40}`
		got, err := normalizeDeliveryRule(raw)
		if err != nil {
			t.Fatalf("合法规则不应报错：%v", err)
		}

		var rule DeliveryRule
		if err := json.Unmarshal([]byte(got), &rule); err != nil {
			t.Fatalf("序列化结果不是合法 JSON：%v", err)
		}
		if rule.Radius != 3000 || rule.MinAmount != 2000 || rule.DeliveryFee != 500 ||
			rule.PackingFee != 100 || rule.EtaMinutes != 40 {
			t.Fatalf("规则字段不匹配：%+v", rule)
		}
	})

	t.Run("负数金额被拒绝", func(t *testing.T) {
		if _, err := normalizeDeliveryRule(`{"deliveryFee":-1}`); err == nil {
			t.Fatal("负配送费应被拒绝")
		}
	})

	t.Run("非法 JSON 被拒绝", func(t *testing.T) {
		_, err := normalizeDeliveryRule(`{bad json}`)
		if err == nil {
			t.Fatal("非法 JSON 应被拒绝")
		}
		if !strings.Contains(err.Error(), "配送规则") {
			t.Fatalf("错误文案应说明配送规则问题，实际为：%v", err)
		}
	})
}

// TestParseDeliveryRule 验证库中 JSON 的解析行为，空串返回零值。
func TestParseDeliveryRule(t *testing.T) {
	rule, err := ParseDeliveryRule("")
	if err != nil {
		t.Fatalf("空串不应报错：%v", err)
	}
	if rule != (DeliveryRule{}) {
		t.Fatalf("空串应返回零值，实际为 %+v", rule)
	}

	if _, err := ParseDeliveryRule(`{"radius":"abc"}`); err == nil {
		t.Fatal("字段类型错误应被拒绝")
	}
}
