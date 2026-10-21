package migrations

import (
	"eqray-flavor/app/facades"
	"eqray-flavor/app/modules/base/support"
	"eqray-flavor/app/modules/base/types"
	infraModel "eqray-flavor/app/modules/infra/model"
)

type M20261020000032SeedInfraSms struct{}

// Signature The unique signature for the migration.
func (r *M20261020000032SeedInfraSms) Signature() string {
	return "20261020000032_seed_infra_sms"
}

// Up 写入默认的本地日志渠道与登录验证码模板，保证开箱即可用；
// 生产环境在后台「基础设施 → 短信渠道 / 短信模板」改用阿里云 / 腾讯云即可。
func (r *M20261020000032SeedInfraSms) Up() error {
	if !facades.Schema().HasTable("infra_sms_channel") ||
		!facades.Schema().HasTable("infra_sms_template") {
		return nil
	}

	now := types.Now()

	var channel infraModel.InfraSmsChannelModel
	if err := facades.DB().Table(infraModel.InfraSmsChannelTable).
		Where("code", infraModel.InfraSmsChannelDebugLog).
		WhereNull("deleted_at").
		First(&channel); err != nil {
		return err
	}
	if channel.ID == 0 {
		channel.ID = support.NewID()
		channel.Name = "本地日志（开发）"
		channel.Code = infraModel.InfraSmsChannelDebugLog
		channel.Status = 1
		channel.CreatedAt = now
		channel.UpdatedAt = now
		if _, err := facades.DB().Table(infraModel.InfraSmsChannelTable).Insert(map[string]any{
			"id":         channel.ID,
			"name":       channel.Name,
			"code":       channel.Code,
			"status":     channel.Status,
			"creator":    "system",
			"created_at": now,
			"updater":    "system",
			"updated_at": now,
		}); err != nil {
			return err
		}
	}

	var template infraModel.InfraSmsTemplateModel
	if err := facades.DB().Table(infraModel.InfraSmsTemplateTable).
		Where("code", "member_sms_login").
		WhereNull("deleted_at").
		First(&template); err != nil {
		return err
	}
	if template.ID == 0 {
		if _, err := facades.DB().Table(infraModel.InfraSmsTemplateTable).Insert(map[string]any{
			"id":              support.NewID(),
			"channel_id":      channel.ID,
			"code":            "member_sms_login",
			"name":            "会员登录验证码",
			"type":            infraModel.InfraSmsTemplateTypeCode,
			"api_template_id": "",
			"content":         "【会员登录】您的验证码是 {code}，5 分钟内有效，请勿泄露。",
			"params":          "code",
			"status":          1,
			"creator":         "system",
			"created_at":      now,
			"updater":         "system",
			"updated_at":      now,
		}); err != nil {
			return err
		}
	}

	return nil
}

// Down 软删默认渠道与登录模板。
func (r *M20261020000032SeedInfraSms) Down() error {
	if !facades.Schema().HasTable("infra_sms_channel") ||
		!facades.Schema().HasTable("infra_sms_template") {
		return nil
	}

	now := types.Now()
	if _, err := facades.DB().Table(infraModel.InfraSmsChannelTable).
		Where("code", infraModel.InfraSmsChannelDebugLog).
		WhereNull("deleted_at").
		Update(map[string]any{"deleted_at": now, "updater": "system", "updated_at": now}); err != nil {
		return err
	}

	_, err := facades.DB().Table(infraModel.InfraSmsTemplateTable).
		Where("code", "member_sms_login").
		WhereNull("deleted_at").
		Update(map[string]any{"deleted_at": now, "updater": "system", "updated_at": now})

	return err
}
