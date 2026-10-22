package service

import (
	"testing"

	"eqray-flavor/app/modules/base/types"
)

// TestStoreScopeCacheEncoding 验证可见门店缓存的编码与解析，覆盖全量、受限与空集合。
func TestStoreScopeCacheEncoding(t *testing.T) {
	t.Run("全量范围", func(t *testing.T) {
		raw := encodeStoreScope(StoreScope{All: true})
		if raw != "a" {
			t.Fatalf("全量编码应为 a，实际为 %q", raw)
		}

		scope, ok := decodeStoreScope(raw)
		if !ok || !scope.All {
			t.Fatalf("全量缓存解析失败：ok=%v scope=%+v", ok, scope)
		}
	})

	t.Run("受限门店集合", func(t *testing.T) {
		raw := encodeStoreScope(StoreScope{IDs: []types.ID{11, 22}})
		if raw != "r:11,22" {
			t.Fatalf("受限编码应为 r:11,22，实际为 %q", raw)
		}

		scope, ok := decodeStoreScope(raw)
		if !ok || scope.All || len(scope.IDs) != 2 || scope.IDs[0] != 11 || scope.IDs[1] != 22 {
			t.Fatalf("受限缓存解析失败：ok=%v scope=%+v", ok, scope)
		}
	})

	t.Run("受限但无门店", func(t *testing.T) {
		raw := encodeStoreScope(StoreScope{})
		if raw != "r:" {
			t.Fatalf("空受限编码应为 r:，实际为 %q", raw)
		}

		scope, ok := decodeStoreScope(raw)
		if !ok || scope.All || len(scope.IDs) != 0 {
			t.Fatalf("空受限缓存解析失败：ok=%v scope=%+v", ok, scope)
		}
	})

	t.Run("非法缓存值", func(t *testing.T) {
		for _, raw := range []string{"", "x", "r:a", "r:1,,2", "r:-1"} {
			if _, ok := decodeStoreScope(raw); ok {
				t.Fatalf("非法缓存值 %q 不应解析成功", raw)
			}
		}
	})
}

// TestStoreScopeContains 验证可见门店集合的成员判断。
func TestStoreScopeContains(t *testing.T) {
	scope := StoreScope{IDs: []types.ID{7, 9}}
	if !scope.Contains(7) || scope.Contains(8) {
		t.Fatalf("可见门店成员判断错误：%+v", scope)
	}
}
