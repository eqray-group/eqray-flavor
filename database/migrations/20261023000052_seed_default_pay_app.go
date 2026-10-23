package migrations

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
)

// M20261023000052SeedDefaultPayApp 初始化默认支付应用与 POS 渠道占位。
// 背景：本项目全局限定一套支付，POS 扫码收款要求"恰好一个启用的支付应用"，
// 但空库没有任何应用，收银端会直接报"尚未配置支付应用"。
// 做法：仅在库中没有任何应用时补一条默认应用，并挂上微信扫码 / 支付宝条码两个渠道，
// 状态先置为关闭——真实商户号与密钥必须由管理员填写，脚本不伪造凭证。
type M20261023000052SeedDefaultPayApp struct{}

// Signature The unique signature for the migration.
func (r *M20261023000052SeedDefaultPayApp) Signature() string {
	return "20261023000052_seed_default_pay_app"
}

// Up 补默认支付应用与两个 POS 渠道占位，已有数据时整体跳过。
func (r *M20261023000052SeedDefaultPayApp) Up() error {
	if !facades.Schema().HasTable("pay_app") || !facades.Schema().HasTable("pay_channel") {
		return nil
	}

	existing, err := facades.DB().Table("pay_app").WhereNull("deleted_at").Count()
	if err != nil {
		return err
	}
	// 已有应用说明环境已配置过，不覆盖管理员的任何设置
	if existing > 0 {
		return nil
	}

	now := time.Now()
	appID := support.NewID()

	appKey, err := randomHex(16)
	if err != nil {
		return err
	}
	appSecret, err := randomHex(16)
	if err != nil {
		return err
	}

	if _, err := facades.DB().Table("pay_app").Insert(map[string]any{
		"id":         appID,
		"name":       "默认支付应用",
		"status":     0, // 0 开启（yudao 口径）
		"remark":     "自动创建的默认应用；全项目共用这一套支付",
		"app_key":    appKey,
		"app_secret": appSecret,
		"creator":    "system",
		"created_at": now,
		"updater":    "system",
		"updated_at": now,
	}); err != nil {
		return err
	}

	// 两个 POS 渠道先禁用：填好商户号 / 密钥 / 证书后再开启，避免带空配置调渠道
	channels := []struct {
		code   string
		remark string
	}{
		{code: "wx_native", remark: "微信扫码（POS 展示二维码）；填写商户号、APIv3 密钥与证书后开启"},
		{code: "alipay_bar", remark: "支付宝条码（收银员扫顾客付款码）；填写 appId、私钥与公钥后开启"},
	}
	for _, channel := range channels {
		if _, err := facades.DB().Table("pay_channel").Insert(map[string]any{
			"id":         support.NewID(),
			"app_id":     appID,
			"code":       channel.code,
			"status":     1, // 1 关闭
			"remark":     channel.remark,
			"fee_rate":   0,
			"config":     "{}",
			"creator":    "system",
			"created_at": now,
			"updater":    "system",
			"updated_at": now,
		}); err != nil {
			return err
		}
	}

	return nil
}

// randomHex 生成 n 字节的十六进制随机串，用于应用标识与密钥。
func randomHex(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return hex.EncodeToString(buf), nil
}

// Down 仅在应用名与备注都匹配本迁移写入的默认值时删除，避免误删管理员配置。
func (r *M20261023000052SeedDefaultPayApp) Down() error {
	if !facades.Schema().HasTable("pay_app") {
		return nil
	}

	row := struct {
		ID string `db:"id"`
	}{}
	if err := facades.DB().Table("pay_app").
		Select("id").
		Where("name", "默认支付应用").
		WhereNull("deleted_at").
		First(&row); err != nil {
		return err
	}
	if row.ID == "" {
		return nil
	}
	id := row.ID

	if facades.Schema().HasTable("pay_channel") {
		if _, err := facades.DB().Table("pay_channel").Where("app_id", id).Delete(); err != nil {
			return err
		}
	}

	_, err := facades.DB().Table("pay_app").Where("id", id).Delete()

	return err
}
