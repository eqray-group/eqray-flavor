// Package errs 定义 store 模块的业务错误码。
// 编码规则：1_{模块段}_{子模块}_{序号}，store 模块段为 004。
package errs

import baseErrs "eqray-flavor/app/modules/base/errs"

// store 模块业务错误码：1-004-{子模块}-{序号}。
// 子模块占用：store 门店 001、table 桌台 002、employee 员工 003、printer 打印 004、rider 骑手 005。
var (
	// StoreNotExists 目标门店不存在或已删除
	StoreNotExists = baseErrs.ErrorCode{Code: 1_004_001_001, Msg: "门店【%v】不存在"}
	// StoreCodeExists 门店编码重复，编码全局唯一
	StoreCodeExists = baseErrs.ErrorCode{Code: 1_004_001_002, Msg: "门店编码【%v】已存在"}
	// StoreParentNotExists 指定的上级门店不存在
	StoreParentNotExists = baseErrs.ErrorCode{Code: 1_004_001_003, Msg: "上级节点【%v】不存在"}
	// StoreParentTypeInvalid 上级节点类型不符合层级约束，参数为「上级类型」与「当前类型」
	StoreParentTypeInvalid = baseErrs.ErrorCode{Code: 1_004_001_004, Msg: "【%v】下不能创建【%v】"}
	// StoreParentCycle 上级节点指向自身或自身下级，会形成环
	StoreParentCycle = baseErrs.ErrorCode{Code: 1_004_001_005, Msg: "上级节点不能是自身或下级节点"}
	// StoreHasChildren 仍有子节点时禁止删除，避免组织树断裂
	StoreHasChildren = baseErrs.ErrorCode{Code: 1_004_001_006, Msg: "门店【%v】下仍有子节点，不可删除"}
	// StoreTypeHasChildren 有子节点时不能降级为门店，门店是最末级
	StoreTypeHasChildren = baseErrs.ErrorCode{Code: 1_004_001_007, Msg: "【%v】下仍有子节点，不能改为门店类型"}
	// StoreHeadquarterExists 品牌总部全库唯一，已存在总部根节点时禁止再建
	StoreHeadquarterExists = baseErrs.ErrorCode{Code: 1_004_001_008, Msg: "品牌总部已存在，全库只能有一个总部根节点"}
	// StoreHeadquarterProtected 总部根节点不允许删除或改为其他类型
	StoreHeadquarterProtected = baseErrs.ErrorCode{Code: 1_004_001_009, Msg: "品牌总部是根节点，不能删除或改为其他类型"}
	// StoreContextInvalid 请求头或查询串携带的门店标识格式非法
	StoreContextInvalid = baseErrs.ErrorCode{Code: 1_004_001_010, Msg: "门店标识格式不正确"}
	// StoreContextNotExists 请求上下文指定的门店不存在或已删除
	StoreContextNotExists = baseErrs.ErrorCode{Code: 1_004_001_011, Msg: "门店不存在或已删除"}
	// StoreAccessDenied 当前用户无权访问请求上下文指定的门店
	StoreAccessDenied = baseErrs.ErrorCode{Code: 1_004_001_012, Msg: "无权访问该门店数据"}
	// StoreContextMismatch 请求体门店与请求头门店上下文不一致
	StoreContextMismatch = baseErrs.ErrorCode{Code: 1_004_001_013, Msg: "请求门店与当前门店上下文不一致"}
	// StoreContextTypeInvalid 门店上下文只能选择门店类型节点，不能指向总部或大区
	StoreContextTypeInvalid = baseErrs.ErrorCode{Code: 1_004_001_014, Msg: "当前门店上下文只能选择门店类型节点"}
	// StoreNotShop 门店级业务只能操作 type=3 的门店节点
	StoreNotShop = baseErrs.ErrorCode{Code: 1_004_001_015, Msg: "只能操作门店类型节点"}

	// StoreTableAreaNotExists 桌台区域不存在或已删除
	StoreTableAreaNotExists = baseErrs.ErrorCode{Code: 1_004_002_001, Msg: "桌台区域不存在"}
	// StoreTableAreaExists 同一门店内区域名称重复
	StoreTableAreaExists = baseErrs.ErrorCode{Code: 1_004_002_002, Msg: "区域名称【%v】已存在"}
	// StoreTableAreaInUse 区域下仍有桌台时禁止删除
	StoreTableAreaInUse = baseErrs.ErrorCode{Code: 1_004_002_003, Msg: "该区域下仍有【%v】张桌台，请先删除或移动桌台"}
	// StoreTableAreaStoreImmutable 区域不能跨门店移动
	StoreTableAreaStoreImmutable = baseErrs.ErrorCode{Code: 1_004_002_004, Msg: "桌台区域不能跨门店移动"}
	// StoreTableNotExists 桌台不存在或已删除
	StoreTableNotExists = baseErrs.ErrorCode{Code: 1_004_002_005, Msg: "桌台不存在"}
	// StoreTableCodeExists 桌码重复，桌码全局唯一
	StoreTableCodeExists = baseErrs.ErrorCode{Code: 1_004_002_006, Msg: "桌码【%v】已存在"}
	// StoreTableNameExists 同一门店内桌号重复
	StoreTableNameExists = baseErrs.ErrorCode{Code: 1_004_002_007, Msg: "桌号【%v】已存在"}
	// StoreTableAreaMismatch 桌台所属区域与门店不一致
	StoreTableAreaMismatch = baseErrs.ErrorCode{Code: 1_004_002_008, Msg: "所选区域不属于当前门店"}
	// StoreTableStatusInvalid 桌台状态取值或流转不合法
	StoreTableStatusInvalid = baseErrs.ErrorCode{Code: 1_004_002_009, Msg: "桌台状态流转不合法"}
	// StoreTableNotIdle 只有空闲桌台可以开台
	StoreTableNotIdle = baseErrs.ErrorCode{Code: 1_004_002_010, Msg: "桌台不是空闲状态，无法开台"}
	// StoreTableBusy 使用中的桌台禁止删除
	StoreTableBusy = baseErrs.ErrorCode{Code: 1_004_002_011, Msg: "桌台正在使用中，请先清台"}
	// StoreTableStoreImmutable 桌台不能跨门店移动
	StoreTableStoreImmutable = baseErrs.ErrorCode{Code: 1_004_002_012, Msg: "桌台不能跨门店移动"}
	// StoreTableBatchExists 批量生成时桌号或桌码已存在
	StoreTableBatchExists = baseErrs.ErrorCode{Code: 1_004_002_013, Msg: "批量生成失败：桌号【%v】或桌码已存在"}
	// StoreTableExists 桌号或桌码已被占用
	StoreTableExists = baseErrs.ErrorCode{Code: 1_004_002_014, Msg: "桌号或桌码已存在"}

	// StoreEmployeeNotExists 门店员工关联记录不存在或已删除
	StoreEmployeeNotExists = baseErrs.ErrorCode{Code: 1_004_003_001, Msg: "门店员工记录不存在"}
	// StoreEmployeeExists 同一账号已经分配到同一门店
	StoreEmployeeExists = baseErrs.ErrorCode{Code: 1_004_003_002, Msg: "该账号已分配到当前门店"}
	// StoreEmployeeRoleInvalid 门店员工角色取值不合法
	StoreEmployeeRoleInvalid = baseErrs.ErrorCode{Code: 1_004_003_003, Msg: "门店角色取值不合法"}
	// StoreEmployeeUserNotExists 指定的后台账号不存在
	StoreEmployeeUserNotExists = baseErrs.ErrorCode{Code: 1_004_003_004, Msg: "后台账号【%v】不存在"}
	// StoreEmployeeStoreTypeInvalid 员工只能分配到门店类型节点
	StoreEmployeeStoreTypeInvalid = baseErrs.ErrorCode{Code: 1_004_003_005, Msg: "只能将员工分配到门店类型节点"}
	// StoreEmployeeStatusInvalid 门店员工启停状态取值不合法
	StoreEmployeeStatusInvalid = baseErrs.ErrorCode{Code: 1_004_003_006, Msg: "门店员工状态取值不合法"}

	// StorePrinterNotExists 打印设备不存在或已删除
	StorePrinterNotExists = baseErrs.ErrorCode{Code: 1_004_004_001, Msg: "打印设备不存在"}
	// StorePrinterDeviceExists 同一门店内设备号已经绑定
	StorePrinterDeviceExists = baseErrs.ErrorCode{Code: 1_004_004_002, Msg: "设备号【%v】已绑定，请勿重复添加"}
	// StorePrinterTypeInvalid 打印设备类型取值不合法
	StorePrinterTypeInvalid = baseErrs.ErrorCode{Code: 1_004_004_003, Msg: "打印设备类型取值不合法"}
	// StorePrinterAccessModeInvalid 打印设备接入方式取值不合法
	StorePrinterAccessModeInvalid = baseErrs.ErrorCode{Code: 1_004_004_004, Msg: "打印设备接入方式取值不合法"}
	// StorePrinterStatusInvalid 打印设备启停状态取值不合法
	StorePrinterStatusInvalid = baseErrs.ErrorCode{Code: 1_004_004_005, Msg: "打印设备状态取值不合法"}
	// StorePrinterConfigInvalid 打印设备非敏感配置不是合法 JSON
	StorePrinterConfigInvalid = baseErrs.ErrorCode{Code: 1_004_004_006, Msg: "打印配置必须是合法的 JSON"}
	// StorePrinterSecretInvalid 打印设备密钥长度超出限制
	StorePrinterSecretInvalid = baseErrs.ErrorCode{Code: 1_004_004_007, Msg: "打印设备密钥不能超过 512 字节"}
	// StorePrinterCopiesInvalid 默认打印份数超出允许范围
	StorePrinterCopiesInvalid = baseErrs.ErrorCode{Code: 1_004_004_008, Msg: "默认打印份数需在 1 ~ 9 之间"}
	// StorePrinterStoreImmutable 打印设备不能跨门店移动
	StorePrinterStoreImmutable = baseErrs.ErrorCode{Code: 1_004_004_009, Msg: "打印设备不能跨门店移动"}
	// StorePrintTemplateNotExists 打印模板不存在或已删除
	StorePrintTemplateNotExists = baseErrs.ErrorCode{Code: 1_004_004_010, Msg: "打印模板不存在"}
	// StorePrintTemplateKindExists 同一门店下设备类型与打印类型的模板重复
	StorePrintTemplateKindExists = baseErrs.ErrorCode{Code: 1_004_004_011, Msg: "【%v】的【%v】模板已存在，请直接编辑原模板"}
	// StorePrintTemplatePrintTypeInvalid 打印模板的业务打印类型取值不合法
	StorePrintTemplatePrintTypeInvalid = baseErrs.ErrorCode{Code: 1_004_004_012, Msg: "打印模板的打印类型取值不合法"}
	// StorePrintTemplatePaperWidthInvalid 打印模板纸宽超出允许范围
	StorePrintTemplatePaperWidthInvalid = baseErrs.ErrorCode{Code: 1_004_004_013, Msg: "模板纸宽需在 40 ~ 200 mm 之间"}
	// StorePrintTemplateContentRequired 打印模板正文为空
	StorePrintTemplateContentRequired = baseErrs.ErrorCode{Code: 1_004_004_014, Msg: "模板内容不能为空"}
	// StorePrintTemplateStoreImmutable 打印模板不能跨门店移动
	StorePrintTemplateStoreImmutable = baseErrs.ErrorCode{Code: 1_004_004_015, Msg: "打印模板不能跨门店移动"}
	// StorePrintTemplateStatusInvalid 打印模板启停状态取值不合法
	StorePrintTemplateStatusInvalid = baseErrs.ErrorCode{Code: 1_004_004_016, Msg: "打印模板状态取值不合法"}

	// StoreRiderNotExists 骑手档案不存在或已删除
	StoreRiderNotExists = baseErrs.ErrorCode{Code: 1_004_005_001, Msg: "骑手档案不存在"}
	// StoreRiderExists 同一门店员工已经存在骑手档案
	StoreRiderExists = baseErrs.ErrorCode{Code: 1_004_005_002, Msg: "该员工已有骑手档案"}
	// StoreRiderStatusInvalid 骑手状态取值不合法
	StoreRiderStatusInvalid = baseErrs.ErrorCode{Code: 1_004_005_003, Msg: "骑手状态取值不合法"}
	// StoreRiderUserNotExists 指定的后台账号不存在
	StoreRiderUserNotExists = baseErrs.ErrorCode{Code: 1_004_005_004, Msg: "后台账号【%v】不存在"}
	// StoreRiderUserDisabled 指定的后台账号已被停用
	StoreRiderUserDisabled = baseErrs.ErrorCode{Code: 1_004_005_005, Msg: "后台账号【%v】已被停用，不能绑定为骑手"}
	// StoreRiderEmployeeRoleInvalid 指定账号在门店内的角色不是骑手
	StoreRiderEmployeeRoleInvalid = baseErrs.ErrorCode{Code: 1_004_005_006, Msg: "该账号在当前门店的员工角色不是骑手，请先调整员工角色"}
	// StoreRiderMaxOrdersInvalid 接单并发上限超出允许范围
	StoreRiderMaxOrdersInvalid = baseErrs.ErrorCode{Code: 1_004_005_007, Msg: "接单并发上限需在 1 ~ 20 之间"}
	// StoreRiderHasActiveOrders 仍有进行中的配送单，不能休息、停用或删除
	StoreRiderHasActiveOrders = baseErrs.ErrorCode{Code: 1_004_005_008, Msg: "骑手仍有【%v】个进行中的配送单，请先完成或取消"}
	// StoreRiderLimitExceeded 当前进行中订单已达到并发上限
	StoreRiderLimitExceeded = baseErrs.ErrorCode{Code: 1_004_005_009, Msg: "骑手当前进行中订单已达并发上限"}
	// StoreRiderStatusNotAccepting 骑手不在接单中状态，不能抢单
	StoreRiderStatusNotAccepting = baseErrs.ErrorCode{Code: 1_004_005_010, Msg: "骑手不在接单中状态，无法接单"}
	// StoreRiderEmployeeNotExists 骑手关联的门店员工记录不存在
	StoreRiderEmployeeNotExists = baseErrs.ErrorCode{Code: 1_004_005_011, Msg: "骑手关联的门店员工不存在"}
	// StoreRiderStoreMismatch 骑手档案与关联员工不属于同一门店
	StoreRiderStoreMismatch = baseErrs.ErrorCode{Code: 1_004_005_012, Msg: "骑手档案与门店员工不属于同一门店"}
	// StoreRiderEmployeeManaged 员工已关联骑手档案，需到骑手管理页维护
	StoreRiderEmployeeManaged = baseErrs.ErrorCode{Code: 1_004_005_013, Msg: "该员工已有骑手档案，请到骑手管理页维护"}
	// StoreRiderOperationConflict 操作过程中骑手状态或负载发生变化，需要刷新后重试
	StoreRiderOperationConflict = baseErrs.ErrorCode{Code: 1_004_005_014, Msg: "骑手档案状态已变化，请刷新后重试"}

	// StorePrintTaskNotExists 打印任务不存在或已删除
	StorePrintTaskNotExists = baseErrs.ErrorCode{Code: 1_004_006_001, Msg: "打印任务不存在"}
	// StorePrintTaskStatusInvalid 打印任务当前状态不允许该操作
	StorePrintTaskStatusInvalid = baseErrs.ErrorCode{Code: 1_004_006_002, Msg: "打印任务当前状态【%v】不允许该操作"}
	// StorePrintTaskAccessDenied 无权操作该门店的打印任务
	StorePrintTaskAccessDenied = baseErrs.ErrorCode{Code: 1_004_006_003, Msg: "无权操作该门店的打印任务"}
)
