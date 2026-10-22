<!-- 门店新增 / 编辑弹窗：组织层级、营业时间与配送规则一次性维护 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑门店' : '新增门店'"
    width="760px"
    align-center
    :close-on-click-modal="false"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="104px">
      <ElDivider content-position="left">基本信息</ElDivider>

      <ElRow :gutter="16">
        <ElCol :span="12">
          <ElFormItem label="节点类型" prop="type">
            <ElRadioGroup v-model="form.type" :disabled="isEdit && storeData?.type === 1">
              <ElRadioButton :value="1" :disabled="hasHeadquarter && form.type !== 1">
                总部
              </ElRadioButton>
              <ElRadioButton :value="2">大区</ElRadioButton>
              <ElRadioButton :value="3">门店</ElRadioButton>
            </ElRadioGroup>
          </ElFormItem>
        </ElCol>
        <ElCol :span="12">
          <ElFormItem label="上级节点" prop="parentId">
            <ElTreeSelect
              v-model="form.parentId"
              :data="parentOptions"
              :disabled="form.type === 1"
              check-strictly
              clearable
              :render-after-expand="false"
              placeholder="总部为根节点，无需上级"
              style="width: 100%"
            />
          </ElFormItem>
        </ElCol>
      </ElRow>
      <div class="-mt-3 mb-3 pl-26 text-xs text-g-600">
        品牌总部全库唯一；总部根节点不能删除或改为其他类型。
      </div>

      <ElRow :gutter="16">
        <ElCol :span="12">
          <ElFormItem label="门店编码" prop="code">
            <ElInput
              v-model="form.code"
              :disabled="isEdit"
              placeholder="字母、数字、下划线或短横线"
              maxlength="32"
            />
          </ElFormItem>
        </ElCol>
        <ElCol :span="12">
          <ElFormItem label="门店名称" prop="name">
            <ElInput v-model="form.name" placeholder="请输入门店名称" maxlength="64" />
          </ElFormItem>
        </ElCol>
      </ElRow>

      <ElRow :gutter="16">
        <ElCol :span="12">
          <ElFormItem label="门店简称" prop="shortName">
            <ElInput v-model="form.shortName" placeholder="小票与看板展示用" maxlength="32" />
          </ElFormItem>
        </ElCol>
        <ElCol :span="12">
          <ElFormItem label="经营状态" prop="status">
            <ElSelect v-model="form.status" style="width: 100%">
              <ElOption label="营业" :value="1" />
              <ElOption label="休息" :value="2" />
              <ElOption label="装修" :value="3" />
              <ElOption label="停业" :value="4" />
            </ElSelect>
          </ElFormItem>
        </ElCol>
      </ElRow>

      <ElRow :gutter="16">
        <ElCol :span="12">
          <ElFormItem label="联系人" prop="contactName">
            <ElInput v-model="form.contactName" placeholder="请输入联系人" maxlength="32" />
          </ElFormItem>
        </ElCol>
        <ElCol :span="12">
          <ElFormItem label="联系电话" prop="contactMobile">
            <ElInput v-model="form.contactMobile" placeholder="请输入联系电话" maxlength="20" />
          </ElFormItem>
        </ElCol>
      </ElRow>

      <ElRow :gutter="16">
        <ElCol :span="12">
          <ElFormItem label="所在地区" prop="districtId">
            <ArtDistrictCascader v-model="form.districtId" placeholder="省 / 市 / 区县" />
          </ElFormItem>
        </ElCol>
        <ElCol :span="12">
          <ElFormItem label="详细地址" prop="address">
            <ElInput v-model="form.address" placeholder="街道、门牌号" maxlength="255" />
          </ElFormItem>
        </ElCol>
      </ElRow>

      <ElFormItem label="门店坐标" prop="longitude">
        <ArtMapPicker
          v-model:longitude="form.longitude"
          v-model:latitude="form.latitude"
          @pick="handleMapPick"
        />
      </ElFormItem>
      <div class="-mt-4 mb-3 pl-26 text-xs text-g-600">
        地图选点后自动带出经纬度；详细地址为空时按选点地址回填
      </div>

      <ElDivider content-position="left">营业与配送</ElDivider>

      <ElRow :gutter="16">
        <ElCol :span="12">
          <ElFormItem label="开始营业" prop="openTime">
            <ElTimePicker
              v-model="form.openTime"
              type="time"
              format="HH:mm"
              value-format="HH:mm"
              placeholder="09:00"
              style="width: 100%"
            />
          </ElFormItem>
        </ElCol>
        <ElCol :span="12">
          <ElFormItem label="结束营业" prop="closeTime">
            <ElTimePicker
              v-model="form.closeTime"
              type="time"
              format="HH:mm"
              value-format="HH:mm"
              placeholder="23:00"
              style="width: 100%"
            />
          </ElFormItem>
        </ElCol>
      </ElRow>
      <div class="-mt-3 mb-3 pl-26 text-xs text-g-600">
        结束时间早于开始时间表示跨夜营业，两者相同表示全天营业
      </div>

      <ElRow :gutter="16">
        <ElCol :span="8">
          <ElFormItem label="配送半径" prop="radius">
            <ElInputNumber
              v-model="form.radius"
              :min="0"
              :controls="false"
              placeholder="米，0 不限"
              style="width: 100%"
            />
          </ElFormItem>
        </ElCol>
        <ElCol :span="8">
          <ElFormItem label="起送价" prop="minAmountYuan">
            <ElInputNumber
              v-model="form.minAmountYuan"
              :min="0"
              :precision="2"
              :controls="false"
              placeholder="元"
              style="width: 100%"
            />
          </ElFormItem>
        </ElCol>
        <ElCol :span="8">
          <ElFormItem label="配送费" prop="deliveryFeeYuan">
            <ElInputNumber
              v-model="form.deliveryFeeYuan"
              :min="0"
              :precision="2"
              :controls="false"
              placeholder="元"
              style="width: 100%"
            />
          </ElFormItem>
        </ElCol>
      </ElRow>

      <ElRow :gutter="16">
        <ElCol :span="8">
          <ElFormItem label="打包费" prop="packingFeeYuan">
            <ElInputNumber
              v-model="form.packingFeeYuan"
              :min="0"
              :precision="2"
              :controls="false"
              placeholder="每份，元"
              style="width: 100%"
            />
          </ElFormItem>
        </ElCol>
        <ElCol :span="8">
          <ElFormItem label="预计时长" prop="etaMinutes">
            <ElInputNumber
              v-model="form.etaMinutes"
              :min="0"
              :controls="false"
              placeholder="分钟"
              style="width: 100%"
            />
          </ElFormItem>
        </ElCol>
        <ElCol :span="8">
          <ElFormItem label="自动接单" prop="autoAccept">
            <ElSwitch
              v-model="form.autoAccept"
              :active-value="1"
              :inactive-value="0"
              active-text="自动"
              inactive-text="人工"
            />
          </ElFormItem>
        </ElCol>
      </ElRow>

      <ElDivider content-position="left">其他</ElDivider>

      <ElRow :gutter="16">
        <ElCol :span="8">
          <ElFormItem label="开放线上" prop="onlineEnabled">
            <ElSwitch
              v-model="form.onlineEnabled"
              :active-value="1"
              :inactive-value="0"
              active-text="开放"
              inactive-text="关闭"
            />
          </ElFormItem>
        </ElCol>
        <ElCol :span="8">
          <ElFormItem label="开启外卖" prop="deliveryEnabled">
            <ElSwitch
              v-model="form.deliveryEnabled"
              :active-value="1"
              :inactive-value="0"
              active-text="开启"
              inactive-text="关闭"
            />
          </ElFormItem>
        </ElCol>
        <ElCol :span="8">
          <ElFormItem label="排序" prop="sort">
            <ElInputNumber
              v-model="form.sort"
              :min="0"
              :max="9999"
              controls-position="right"
              style="width: 100%"
            />
          </ElFormItem>
        </ElCol>
      </ElRow>
      <div class="-mt-3 mb-3 pl-26 text-xs text-g-600">
        外卖开关关闭时顾客端不展示外卖入口；配送规则仍可提前配置。
      </div>

      <ElFormItem label="备注" prop="remark">
        <ElInput
          v-model="form.remark"
          type="textarea"
          :rows="2"
          maxlength="255"
          show-word-limit
          placeholder="选填"
        />
      </ElFormItem>
    </ElForm>

    <template #footer>
      <ElButton @click="dialogVisible = false">取消</ElButton>
      <ElButton type="primary" :loading="submitting" @click="handleSubmit">提交</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { ElMessage } from 'element-plus'
  import type { FormInstance, FormRules } from 'element-plus'
  import { fetchCreateStore, fetchUpdateStore } from '@/api/store/store'
  import ArtDistrictCascader from '@/components/core/forms/art-district-cascader/index.vue'
  import ArtMapPicker from '@/components/core/forms/art-map-picker/index.vue'

  defineOptions({ name: 'StoreDialog' })

  type StoreItem = Api.Store.StoreItem
  type StoreTreeNode = Api.Store.StoreTreeNode
  type DialogType = 'add' | 'edit'

  interface Props {
    visible: boolean
    /** add：新增；edit：编辑 */
    type: DialogType
    /** 编辑时传入当前行 */
    storeData?: StoreItem | null
    /** 组织树，用于上级节点选择 */
    storeTree: StoreTreeNode[]
    /** 新增时的上级节点，用于预填上级；为空表示新建根节点 */
    parentNode?: StoreTreeNode | null
    /** 新增时的默认节点类型：1 总部 / 2 大区 / 3 门店 */
    defaultType?: number
  }

  const props = withDefaults(defineProps<Props>(), {
    storeData: null,
    parentNode: null,
    defaultType: 3
  })

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void
    (e: 'success'): void
  }>()

  const dialogVisible = computed({
    get: () => props.visible,
    set: (value: boolean) => emit('update:visible', value)
  })

  const isEdit = computed(() => props.type === 'edit')

  // 品牌总部全库唯一：已存在时新增弹窗禁用「总部」选项
  const hasHeadquarter = computed(() => {
    const walk = (nodes: StoreTreeNode[]): boolean =>
      nodes.some((node) => node.type === 1 || walk(node.children ?? []))
    return walk(props.storeTree)
  })

  const formRef = ref<FormInstance>()
  const submitting = ref(false)

  // 配送规则在库里是 JSON、金额以分存储；表单按元录入，提交时再换算
  const form = reactive({
    parentId: '0',
    type: 3,
    code: '',
    name: '',
    shortName: '',
    status: 1,
    contactName: '',
    contactMobile: '',
    // 区县 ID：级联组件回传最后一级，空串表示未选择
    districtId: '',
    address: '',
    longitude: 0,
    latitude: 0,
    openTime: '09:00',
    closeTime: '23:00',
    radius: 0,
    minAmountYuan: 0,
    deliveryFeeYuan: 0,
    packingFeeYuan: 0,
    etaMinutes: 0,
    autoAccept: 0,
    onlineEnabled: 1,
    deliveryEnabled: 0,
    sort: 0,
    remark: ''
  })

  const rules: FormRules = {
    type: [{ required: true, message: '请选择节点类型', trigger: 'change' }],
    code: [{ required: true, message: '请输入门店编码', trigger: 'blur' }],
    name: [{ required: true, message: '请输入门店名称', trigger: 'blur' }],
    status: [{ required: true, message: '请选择经营状态', trigger: 'change' }],
    openTime: [{ required: true, message: '请选择开始营业时间', trigger: 'change' }],
    closeTime: [{ required: true, message: '请选择结束营业时间', trigger: 'change' }]
  }

  // 节点类型文案，与后端 StoreTypeEnum 对应
  const TYPE_LABELS: Record<number, string> = { 1: '总部', 2: '大区', 3: '门店' }

  // 收集指定节点及其整棵子树的 ID：编辑时这些节点都不能被选为上级，否则会形成层级环。
  // 入参：nodes 组织树、targetId 当前编辑节点 ID；返回需要排除的 ID 集合（找不到时为空集合）。
  const collectSubtreeIds = (nodes: StoreTreeNode[], targetId: string): Set<string> => {
    const ids = new Set<string>()
    const markAll = (node: StoreTreeNode): void => {
      ids.add(node.id)
      ;(node.children ?? []).forEach(markAll)
    }
    const walk = (list: StoreTreeNode[]): void => {
      for (const node of list) {
        if (node.id === targetId) {
          markAll(node)
          return
        }
        walk(node.children ?? [])
      }
    }
    walk(nodes)
    return ids
  }

  // 上级节点候选：排除自身与自身子树，并禁止把门店当作父级
  const parentOptions = computed(() => {
    const excluded = props.storeData
      ? collectSubtreeIds(props.storeTree, props.storeData.id)
      : new Set<string>()
    const build = (nodes: StoreTreeNode[]): any[] =>
      nodes
        .filter((node) => !excluded.has(node.id))
        .map((node) => ({
          value: node.id,
          label: `${node.name}（${TYPE_LABELS[node.type] ?? '未知'}）`,
          disabled: node.type === 3,
          children: build(node.children ?? [])
        }))

    return build(props.storeTree)
  })

  // 分转元，用于回填配送规则表单
  const fenToYuan = (fen: number): number => Math.round(fen) / 100

  // 元转分：用四舍五入避免 0.1 + 0.2 这类浮点误差带进库
  const yuanToFen = (yuan: number): number => Math.round((yuan || 0) * 100)

  // 把库中的配送规则 JSON 回填到表单，解析失败时保持默认值
  const applyDeliveryRule = (raw: string): void => {
    let rule: Partial<Api.Store.DeliveryRule> = {}
    if (raw) {
      try {
        rule = JSON.parse(raw)
      } catch {
        rule = {}
      }
    }

    form.radius = rule.radius ?? 0
    form.minAmountYuan = fenToYuan(rule.minAmount ?? 0)
    form.deliveryFeeYuan = fenToYuan(rule.deliveryFee ?? 0)
    form.packingFeeYuan = fenToYuan(rule.packingFee ?? 0)
    form.etaMinutes = rule.etaMinutes ?? 0
  }

  // 组装配送规则 JSON；全部为 0 时发空串，表示未配置（下单按不限范围、费用 0 处理）
  const buildDeliveryRule = (): string => {
    const rule: Api.Store.DeliveryRule = {
      radius: Math.max(0, Math.round(form.radius || 0)),
      minAmount: yuanToFen(form.minAmountYuan),
      deliveryFee: yuanToFen(form.deliveryFeeYuan),
      packingFee: yuanToFen(form.packingFeeYuan),
      etaMinutes: Math.max(0, Math.round(form.etaMinutes || 0))
    }

    if (
      !rule.radius &&
      !rule.minAmount &&
      !rule.deliveryFee &&
      !rule.packingFee &&
      !rule.etaMinutes
    ) {
      return ''
    }

    return JSON.stringify(rule)
  }

  // 每次打开弹窗时按当前行回填表单，并清掉上一次的校验提示
  watch(
    () => [props.visible, props.storeData, props.parentNode, props.defaultType] as const,
    ([visible]) => {
      if (!visible) return

      const data = props.storeData
      Object.assign(form, {
        // 新增时按父节点预填上级，编辑时沿用原值；雪花 ID 保持字符串
        parentId: data?.parentId ?? (props.parentNode?.id || '0'),
        type: data?.type ?? props.defaultType,
        code: data?.code ?? '',
        name: data?.name ?? '',
        shortName: data?.shortName ?? '',
        status: data?.status ?? 1,
        contactName: data?.contactName ?? '',
        contactMobile: data?.contactMobile ?? '',
        districtId: data?.districtId && data.districtId !== '0' ? String(data.districtId) : '',
        address: data?.address ?? '',
        longitude: data?.longitude ?? 0,
        latitude: data?.latitude ?? 0,
        openTime: data?.openTime || '09:00',
        closeTime: data?.closeTime || '23:00',
        autoAccept: data?.autoAccept ?? 0,
        onlineEnabled: data?.onlineEnabled ?? 1,
        deliveryEnabled: data?.deliveryEnabled ?? 0,
        sort: data?.sort ?? 0,
        remark: data?.remark ?? ''
      })
      applyDeliveryRule(data?.deliveryRule ?? '')

      nextTick(() => formRef.value?.clearValidate())
    },
    { immediate: true }
  )

  // 地图选点回调：仅在详细地址为空时回填，避免覆盖手工填写的门牌号
  const handleMapPick = (payload: {
    longitude: number
    latitude: number
    address: string
  }): void => {
    if (!form.address.trim() && payload.address) {
      form.address = payload.address
    }
  }

  // 切换为总部时清空上级节点，避免提交出非法层级
  watch(
    () => form.type,
    (type) => {
      if (type === 1) form.parentId = '0'
    }
  )

  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    // 雪花 ID 必须保持字符串，转 Number 会丢精度导致上级节点查不到
    const parentId = form.parentId && form.parentId !== '0' ? form.parentId : '0'
    if (form.type !== 1 && parentId === '0') {
      ElMessage.warning('大区与门店必须选择上级节点')
      return
    }

    const payload: Api.Store.StoreSaveParams = {
      parentId,
      type: form.type,
      code: form.code.trim(),
      name: form.name.trim(),
      shortName: form.shortName.trim(),
      status: form.status,
      contactName: form.contactName.trim(),
      contactMobile: form.contactMobile.trim(),
      // 区县雪花 ID 保持字符串，不转 Number
      districtId: form.districtId || '0',
      address: form.address.trim(),
      longitude: form.longitude || 0,
      latitude: form.latitude || 0,
      openTime: form.openTime,
      closeTime: form.closeTime,
      deliveryRule: buildDeliveryRule(),
      autoAccept: form.autoAccept,
      onlineEnabled: form.onlineEnabled,
      deliveryEnabled: form.deliveryEnabled,
      sort: form.sort,
      remark: form.remark.trim()
    }

    submitting.value = true
    try {
      if (isEdit.value && props.storeData) {
        await fetchUpdateStore(props.storeData.id, payload)
        ElMessage.success('保存成功')
      } else {
        await fetchCreateStore(payload)
        ElMessage.success('新增成功')
      }

      dialogVisible.value = false
      emit('success')
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      submitting.value = false
    }
  }
</script>
