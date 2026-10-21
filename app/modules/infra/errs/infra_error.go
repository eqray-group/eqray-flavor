// Package errs 定义 infra 模块的业务错误码，编码规则见 docs 8.3(2)。
package errs

import baseErrs "eqray-flavor/app/modules/base/errs"

// infra 模块业务错误码：1-001-{子模块}-{错误码}，
// config 占用 001，file 占用 002，file-config 占用 003，sms 占用 004。
var (
	InfraFileUploadFailed   = baseErrs.ErrorCode{Code: 1_001_002_001, Msg: "文件上传失败"}
	InfraFileTypeNotAllowed = baseErrs.ErrorCode{Code: 1_001_002_002, Msg: "不支持的文件类型，仅支持 jpg / jpeg / png / gif / webp / bmp"}
	InfraFileTooLarge       = baseErrs.ErrorCode{Code: 1_001_002_003, Msg: "文件不能超过 %vMB"}
	InfraFileNotExists      = baseErrs.ErrorCode{Code: 1_001_002_004, Msg: "文件【%v】不存在"}

	InfraFileConfigNotExists       = baseErrs.ErrorCode{Code: 1_001_003_001, Msg: "文件配置【%v】不存在"}
	InfraFileConfigMasterForbidden = baseErrs.ErrorCode{Code: 1_001_003_002, Msg: "主配置【%v】不允许删除"}
	InfraStorageNotSupported       = baseErrs.ErrorCode{Code: 1_001_003_003, Msg: "暂不支持的存储类型【%v】"}

	InfraConfigNotExists        = baseErrs.ErrorCode{Code: 1_001_001_001, Msg: "参数【%v】不存在"}
	InfraConfigKeyExists        = baseErrs.ErrorCode{Code: 1_001_001_002, Msg: "参数键【%v】已存在"}
	InfraConfigBuiltinForbidden = baseErrs.ErrorCode{Code: 1_001_001_003, Msg: "系统内置参数【%v】不允许删除"}

	// sms 占用 004 子模块：短信渠道、模板与发送
	InfraSmsChannelNotExists   = baseErrs.ErrorCode{Code: 1_001_004_001, Msg: "短信渠道【%v】不存在或已停用"}
	InfraSmsTemplateNotExists  = baseErrs.ErrorCode{Code: 1_001_004_002, Msg: "短信模板【%v】不存在或已停用"}
	InfraSmsChannelCodeExists  = baseErrs.ErrorCode{Code: 1_001_004_003, Msg: "短信渠道代码【%v】已存在"}
	InfraSmsTemplateCodeExists = baseErrs.ErrorCode{Code: 1_001_004_004, Msg: "短信模板代码【%v】已存在"}
	InfraSmsSendFailed         = baseErrs.ErrorCode{Code: 1_001_004_005, Msg: "短信发送失败：%v"}
	InfraSmsChannelInUse       = baseErrs.ErrorCode{Code: 1_001_004_006, Msg: "渠道【%v】下仍有模板，不可删除"}
)
