<!-- 打印模板新增 / 编辑弹窗：维护机型、打印类型、纸宽、正文与启停状态 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑打印模板' : '新增打印模板'"
    width="900px"
    align-center
    :close-on-click-modal="false"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="100px">
      <ElFormItem label="所属门店" prop="storeId">
        <ElSelect
          v-model="form.storeId"
          class="w-full"
          filterable
          :disabled="isEdit || !!storeContext.currentStoreId"
          :loading="storeLoading"
          placeholder="请选择门店"
          no-data-text="暂无可操作门店"
        >
          <ElOption
            v-for="store in storeOptions"
            :key="store.id"
            :label="store.name"
            :value="store.id"
          />
        </ElSelect>
      </ElFormItem>

      <ElFormItem label="模板名称" prop="name">
        <ElInput v-model="form.name" placeholder="如前台 80mm 结账单" maxlength="64" />
      </ElFormItem>

      <div class="grid grid-cols-2 gap-x-4">
        <ElFormItem label="设备类型" prop="printerType">
          <ElSelect v-model="form.printerType" class="w-full">
            <ElOption
              v-for="item in printerTypeOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </ElSelect>
        </ElFormItem>

        <ElFormItem label="打印类型" prop="printType">
          <ElSelect v-model="form.printType" class="w-full" @change="handlePrintTypeChange">
            <ElOption
              v-for="item in printTypeOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </ElSelect>
        </ElFormItem>
      </div>

      <div class="grid grid-cols-2 gap-x-4">
        <ElFormItem label="纸宽" prop="paperWidth">
          <ElInputNumber
            v-model="form.paperWidth"
            :min="40"
            :max="200"
            :step="1"
            style="width: 100%"
          />
          <div class="mt-1 text-xs text-g-600">单位毫米，常用 58 / 76 / 80。</div>
        </ElFormItem>

        <ElFormItem label="状态" prop="status">
          <ElRadioGroup v-model="form.status">
            <ElRadio :value="1">启用</ElRadio>
            <ElRadio :value="0">禁用</ElRadio>
          </ElRadioGroup>
        </ElFormItem>
      </div>

      <ElFormItem label="模板正文" prop="content">
        <div class="template-editor">
          <!-- 左：正文编辑区 -->
          <div class="template-editor-main">
            <div class="mb-2 flex flex-wrap items-center justify-between gap-2">
              <span class="text-xs text-g-600">
                用 <code class="template-code">{{ TIP_PLACEHOLDER }}</code> 占位， 菜品明细用
                <code class="template-code">{{ TIP_ITEM_START }}</code> …
                <code class="template-code">{{ TIP_ITEM_END }}</code> 循环
              </span>
              <ElButton size="small" text type="primary" @click="applyDefaultTemplate">
                插入默认模板
              </ElButton>
            </div>

            <ElInput
              ref="contentInputRef"
              v-model="form.content"
              type="textarea"
              :rows="14"
              maxlength="65535"
              class="template-textarea"
              placeholder="点击右侧变量插入占位符，或点「插入默认模板」生成一份可直接使用的正文"
            />
          </div>

          <!-- 右：可用变量面板，点击插入到光标处 -->
          <div class="template-vars">
            <div v-for="group in variableGroups" :key="group.label" class="mb-3 last:mb-0">
              <div class="mb-1 text-xs text-g-600">{{ group.label }}</div>
              <div class="flex flex-wrap gap-1">
                <span
                  v-for="item in group.items"
                  :key="item.token"
                  class="var-chip"
                  :title="item.token"
                  @click="insertVariable(item.token)"
                >
                  {{ item.label }}
                </span>
              </div>
            </div>
          </div>
        </div>

        <div class="mt-1 text-xs text-g-600">
          <template v-if="isEdit">
            当前版本 v{{ props.templateData?.version ?? 1 }}；正文发生变化时保存后版本号自动 +1。
          </template>
          <template v-else>新建模板版本号从 v1 开始。</template>
          同一门店下同一设备类型与打印类型只能有一个模板；占位符在打印时由渲染器替换成实际值。
        </div>
      </ElFormItem>

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
  import { ElInput, ElMessage } from 'element-plus'
  import type { FormInstance, FormRules } from 'element-plus'
  import { fetchCreatePrintTemplate, fetchUpdatePrintTemplate } from '@/api/store/print-template'
  import { fetchGetVisibleStores } from '@/api/store/store'
  import { useStoreContextStore } from '@/store/modules/store-context'

  defineOptions({ name: 'PrintTemplateDialog' })

  type PrintTemplateItem = Api.Store.PrintTemplateItem
  type DialogType = 'add' | 'edit'

  interface Props {
    visible: boolean
    type: DialogType
    templateData?: PrintTemplateItem | null
  }

  const props = withDefaults(defineProps<Props>(), {
    templateData: null
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
  const storeContext = useStoreContextStore()

  // 设备类型：与后端 printer_type 枚举一致
  const printerTypeOptions = [
    { label: '前台小票', value: 1 },
    { label: '后厨', value: 2 },
    { label: '标签', value: 3 },
    { label: '叫号', value: 4 }
  ]

  // 打印类型：与后端 print_type 枚举一致
  const printTypeOptions = [
    { label: '结账单', value: 1 },
    { label: '预结单', value: 2 },
    { label: '厨打', value: 3 },
    { label: '退菜', value: 4 },
    { label: '标签', value: 5 },
    { label: '退款单', value: 6 },
    { label: '交班', value: 7 },
    { label: '日结', value: 8 }
  ]

  // 语法提示文案：模板里直接写双层大括号会被 Vue 当成插值解析，统一放在常量里
  const TIP_PLACEHOLDER = '{{变量}}'
  const TIP_ITEM_START = '{{#items}}'
  const TIP_ITEM_END = '{{/items}}'

  // 可用占位符：按业务对象分组，点击后插入到正文光标处
  const variableGroups = [
    {
      label: '门店',
      items: [
        { label: '门店名称', token: '{{store.name}}' },
        { label: '门店编码', token: '{{store.code}}' },
        { label: '门店电话', token: '{{store.mobile}}' },
        { label: '门店地址', token: '{{store.address}}' }
      ]
    },
    {
      label: '订单',
      items: [
        { label: '订单号', token: '{{order.no}}' },
        { label: '取餐号', token: '{{order.pickupNo}}' },
        { label: '就餐人数', token: '{{order.guests}}' },
        { label: '下单时间', token: '{{order.createdAt}}' },
        { label: '订单备注', token: '{{order.remark}}' },
        { label: '操作人', token: '{{cashier}}' }
      ]
    },
    {
      label: '桌台',
      items: [
        { label: '桌号', token: '{{table.name}}' },
        { label: '区域', token: '{{table.area}}' }
      ]
    },
    {
      label: '金额',
      items: [
        { label: '合计', token: '{{order.totalAmount}}' },
        { label: '优惠', token: '{{order.discountAmount}}' },
        { label: '实付', token: '{{order.payAmount}}' },
        { label: '支付方式', token: '{{order.payMethod}}' }
      ]
    },
    {
      label: '菜品循环',
      items: [
        { label: '明细开始', token: '{{#items}}' },
        { label: '菜名', token: '{{name}}' },
        { label: '数量', token: '{{qty}}' },
        { label: '规格', token: '{{spec}}' },
        { label: '小计', token: '{{amount}}' },
        { label: '菜品备注', token: '{{remark}}' },
        { label: '明细结束', token: '{{/items}}' }
      ]
    },
    {
      label: '其他',
      items: [{ label: '打印时间', token: '{{printTime}}' }]
    }
  ]

  // 各打印类型的默认正文示例；语法定稿见 docs/design/09-设备接入与打印对接.md §3.1，
  // 渲染器（T-PRINT-01）按同一套 {{变量}} 与 {{#items}} 规则解释
  const DEFAULT_TEMPLATES: Record<number, string> = {
    1: `{{store.name}}
结账单
--------------------------------
单号：{{order.no}}
桌号：{{table.name}}（{{table.area}}）
人数：{{order.guests}}
时间：{{order.createdAt}}
--------------------------------
{{#items}}
{{name}} x{{qty}}            {{amount}}
{{/items}}
--------------------------------
合计：{{order.totalAmount}}
优惠：{{order.discountAmount}}
实付：{{order.payAmount}}
支付方式：{{order.payMethod}}
备注：{{order.remark}}
--------------------------------
谢谢光临，欢迎再次光临`,
    2: `{{store.name}}
预结单
--------------------------------
桌号：{{table.name}}（{{table.area}}）
人数：{{order.guests}}
时间：{{order.createdAt}}
--------------------------------
{{#items}}
{{name}} x{{qty}}            {{amount}}
{{/items}}
--------------------------------
合计：{{order.totalAmount}}
（预结单仅用于核对，不作结算凭证）`,
    3: `{{store.name}} 厨打单
--------------------------------
单号：{{order.no}}
桌号：{{table.name}}
时间：{{order.createdAt}}
--------------------------------
{{#items}}
{{name}} x{{qty}}
规格：{{spec}}
备注：{{remark}}
{{/items}}
--------------------------------
订单备注：{{order.remark}}`,
    4: `{{store.name}} 退菜单
--------------------------------
单号：{{order.no}}
桌号：{{table.name}}
退菜时间：{{printTime}}
--------------------------------
{{#items}}
{{name}} x{{qty}}
原因：{{remark}}
{{/items}}`,
    5: `{{order.pickupNo}}
{{table.name}}
{{printTime}}`,
    6: `{{store.name}} 退款单
--------------------------------
原单号：{{order.no}}
退款时间：{{printTime}}
退款金额：{{order.payAmount}}
支付方式：{{order.payMethod}}
操作人：{{cashier}}`,
    7: `{{store.name}} 交班单
--------------------------------
交班时间：{{printTime}}
操作人：{{cashier}}
订单数：{{order.count}}
营业总额：{{order.totalAmount}}
优惠金额：{{order.discountAmount}}
实收金额：{{order.payAmount}}`,
    8: `{{store.name}} 日结单
--------------------------------
统计时间：{{printTime}}
订单数：{{order.count}}
营业总额：{{order.totalAmount}}
优惠金额：{{order.discountAmount}}
实收金额：{{order.payAmount}}`
  }

  const formRef = ref<FormInstance>()
  const contentInputRef = ref<InstanceType<typeof ElInput>>()
  const submitting = ref(false)
  const storeLoading = ref(false)
  const storeOptions = ref<Api.Store.VisibleStore[]>([])

  const form = reactive({
    storeId: '',
    name: '',
    printerType: 1,
    printType: 1,
    paperWidth: 80,
    content: '',
    status: 1,
    remark: ''
  })

  const rules: FormRules = {
    storeId: [{ required: true, message: '请选择门店', trigger: 'change' }],
    name: [{ required: true, message: '请输入模板名称', trigger: 'blur' }],
    printerType: [{ required: true, message: '请选择设备类型', trigger: 'change' }],
    printType: [{ required: true, message: '请选择打印类型', trigger: 'change' }],
    paperWidth: [{ required: true, message: '请填写纸宽', trigger: 'change' }],
    content: [{ required: true, message: '请输入模板正文', trigger: 'blur' }],
    status: [{ required: true, message: '请选择模板状态', trigger: 'change' }]
  }

  // 拉取当前用户可见门店；具体门店上下文或编辑场景下锁定门店
  const loadStores = async (): Promise<void> => {
    storeLoading.value = true
    try {
      let list = await fetchGetVisibleStores()
      if (storeContext.currentStoreId) {
        list = list.filter((store) => store.id === storeContext.currentStoreId)
      } else if (
        isEdit.value &&
        props.templateData &&
        !list.some((store) => store.id === props.templateData?.storeId)
      ) {
        list.unshift({
          id: props.templateData.storeId,
          name: props.templateData.storeName || '当前门店',
          code: '',
          status: 1
        })
      }
      storeOptions.value = list

      if (!list.some((store) => store.id === form.storeId)) {
        form.storeId = list[0]?.id ?? ''
      }
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      storeLoading.value = false
    }
  }

  // 取模板正文对应的原生 textarea，用于把占位符插到光标处
  const resolveTextarea = (): HTMLTextAreaElement | null => {
    const root = contentInputRef.value?.$el as HTMLElement | undefined
    return root?.querySelector('textarea') ?? null
  }

  // 把占位符插入光标位置（无光标时追加到末尾），插入后焦点与光标留在占位符之后
  const insertVariable = (token: string): void => {
    const el = resolveTextarea()
    if (!el) {
      form.content += token
      return
    }

    const start = el.selectionStart ?? form.content.length
    const end = el.selectionEnd ?? start
    form.content = form.content.slice(0, start) + token + form.content.slice(end)

    const caret = start + token.length
    nextTick(() => {
      el.focus()
      el.setSelectionRange(caret, caret)
    })
  }

  // 用当前打印类型的默认正文覆盖编辑器内容，便于在此基础上微调
  const applyDefaultTemplate = (): void => {
    form.content = DEFAULT_TEMPLATES[form.printType] ?? DEFAULT_TEMPLATES[1]
  }

  // 新增时切换打印类型：正文仍为默认示例（或为空）时同步换成对应类型的默认模板
  const handlePrintTypeChange = (value: number): void => {
    if (isEdit.value && form.content.trim()) return
    const next = DEFAULT_TEMPLATES[value]
    if (!next) return
    const isUntouched =
      !form.content.trim() || Object.values(DEFAULT_TEMPLATES).includes(form.content)
    if (isUntouched) form.content = next
  }

  // 每次打开弹窗重置表单并加载门店选项
  watch(
    () => props.visible,
    async (visible) => {
      if (!visible) return

      const data = props.templateData
      Object.assign(form, {
        storeId: data?.storeId ?? storeContext.currentStoreId ?? '',
        name: data?.name ?? '',
        printerType: data?.printerType ?? 1,
        printType: data?.printType ?? 1,
        paperWidth: data?.paperWidth ?? 80,
        content: data?.content ?? DEFAULT_TEMPLATES[data?.printType ?? 1] ?? '',
        status: data?.status ?? 1,
        remark: data?.remark ?? ''
      })
      storeOptions.value = []

      await loadStores()
      nextTick(() => formRef.value?.clearValidate())
    },
    { immediate: true }
  )

  // 提交打印模板；版本号由后端按正文是否变化自动维护
  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    const payload: Api.Store.PrintTemplateSaveParams = {
      storeId: form.storeId,
      name: form.name,
      printerType: form.printerType,
      printType: form.printType,
      paperWidth: form.paperWidth,
      content: form.content,
      status: form.status,
      remark: form.remark
    }

    submitting.value = true
    try {
      if (isEdit.value && props.templateData) {
        await fetchUpdatePrintTemplate(props.templateData.id, payload)
        ElMessage.success('保存成功')
      } else {
        await fetchCreatePrintTemplate(payload)
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

<style scoped>
  /* 正文编辑区：左侧编辑器 + 右侧变量面板，等宽字体便于对齐小票内容 */
  .template-editor {
    display: flex;
    gap: 12px;
    width: 100%;
  }

  .template-editor-main {
    flex: 1;
    min-width: 0;
  }

  .template-vars {
    flex: none;
    width: 220px;
    max-height: 348px;
    padding: 10px 12px;
    overflow: auto;
    background-color: var(--el-fill-color-lighter);
    border-radius: 8px;
  }

  /* 变量胶囊：点击插入占位符 */
  .var-chip {
    padding: 2px 8px;
    font-size: 12px;
    color: var(--el-text-color-regular);
    cursor: pointer;
    background-color: var(--el-bg-color);
    border: 1px solid var(--el-border-color);
    border-radius: 4px;
  }

  .var-chip:hover {
    color: var(--el-color-primary);
    border-color: var(--el-color-primary-light-5);
  }

  /* 正文输入框使用等宽字体，方便手动对齐列宽 */
  .template-textarea :deep(.el-textarea__inner) {
    font-family: Consolas, Menlo, monospace;
    font-size: 13px;
    line-height: 1.6;
  }

  /* 行内占位符提示 */
  .template-code {
    padding: 1px 4px;
    font-family: Consolas, Menlo, monospace;
    background-color: var(--el-fill-color-light);
    border-radius: 4px;
  }
</style>
