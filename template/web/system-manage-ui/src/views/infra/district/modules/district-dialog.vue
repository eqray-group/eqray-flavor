<!-- 行政区划新增 / 编辑弹窗：层级与父级编码由后端按上级推导 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑行政区划' : '新增行政区划'"
    width="560px"
    align-center
    :close-on-click-modal="false"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="96px">
      <ElFormItem label="上级区划" prop="pid">
        <ElTreeSelect
          v-model="form.pid"
          :data="parentOptions"
          :disabled="isEdit"
          check-strictly
          clearable
          :render-after-expand="false"
          placeholder="不选表示顶级（省）"
          style="width: 100%"
        />
      </ElFormItem>
      <div v-if="isEdit" class="-mt-4 mb-3 pl-24 text-xs text-g-600">
        上级与行政编码创建后不可修改，如需调整请删除后重建
      </div>

      <ElFormItem label="区划名称" prop="areaName">
        <ElInput v-model="form.areaName" placeholder="如：杭州市" maxlength="50" />
      </ElFormItem>

      <ElFormItem label="简称" prop="shortName">
        <ElInput v-model="form.shortName" placeholder="如：杭州，留空则与全称一致" maxlength="50" />
      </ElFormItem>

      <ElFormItem label="行政编码" prop="areaCode">
        <ElInput
          v-model="form.areaCode"
          :disabled="isEdit"
          placeholder="如：330100，需与第三方数据对齐"
          maxlength="50"
        />
      </ElFormItem>

      <ElFormItem label="行政级别">
        <ElInput :model-value="levelHint" disabled />
      </ElFormItem>

      <ElFormItem label="排序" prop="sort">
        <ElInputNumber v-model="form.sort" :min="0" :max="9999" controls-position="right" />
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
  import { fetchCreateDistrict, fetchUpdateDistrict } from '@/api/infra/district'

  defineOptions({ name: 'DistrictDialog' })

  type DistrictItem = Api.Infra.DistrictItem
  type DistrictTreeNode = Api.Infra.DistrictTreeNode
  type DialogType = 'add' | 'edit'

  interface Props {
    visible: boolean
    /** add：新增；edit：编辑 */
    type: DialogType
    /** 编辑时传入当前行 */
    districtData?: DistrictItem | null
    /** 区划树，用于上级选择 */
    districtTree: DistrictTreeNode[]
  }

  const props = withDefaults(defineProps<Props>(), {
    districtData: null
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

  const formRef = ref<FormInstance>()
  const submitting = ref(false)

  const form = reactive({
    pid: '0',
    areaName: '',
    shortName: '',
    areaCode: '',
    sort: 1
  })

  const rules: FormRules = {
    areaName: [{ required: true, message: '请输入区划名称', trigger: 'blur' }],
    areaCode: [{ required: true, message: '请输入行政编码', trigger: 'blur' }]
  }

  // 级别文案：与后端 1 省 / 2 市 / 3 区县 / 4 街道 对应
  const LEVEL_LABELS: Record<number, string> = { 1: '省', 2: '市', 3: '区县', 4: '街道' }

  // 上级候选：排除自身与自身子树，避免把区划挂到自己的下级
  const parentOptions = computed(() => {
    const excludeId = isEdit.value ? (props.districtData?.id ?? '') : ''
    const build = (nodes: DistrictTreeNode[]): any[] =>
      nodes
        .filter((node) => node.id !== excludeId)
        .map((node) => ({
          value: node.id,
          label: `${node.areaName}（${LEVEL_LABELS[node.level] ?? `L${node.level}`}）`,
          children: build(node.children ?? [])
        }))

    return build(props.districtTree)
  })

  // 层级预览：新增按所选上级 +1，编辑保持原值
  const levelHint = computed(() => {
    if (isEdit.value) {
      const level = props.districtData?.level ?? 0
      return `${level}（${LEVEL_LABELS[level] ?? '未知'}）`
    }

    if (!form.pid || form.pid === '0') return '1（省）'

    const find = (nodes: DistrictTreeNode[]): DistrictTreeNode | undefined => {
      for (const node of nodes) {
        if (node.id === form.pid) return node
        const hit = find(node.children ?? [])
        if (hit) return hit
      }
      return undefined
    }
    const parent = find(props.districtTree)
    const level = (parent?.level ?? 0) + 1
    return `${level}（${LEVEL_LABELS[level] ?? '未知'}）`
  })

  // 打开弹窗时按当前行回填，并清掉上一次的校验提示
  watch(
    () => [props.visible, props.districtData] as const,
    ([visible]) => {
      if (!visible) return

      const data = props.districtData
      Object.assign(form, {
        pid: data?.pid ?? '0',
        areaName: data?.areaName ?? '',
        shortName: data?.shortName ?? '',
        areaCode: data?.areaCode ?? '',
        sort: data?.sort ?? 1
      })

      nextTick(() => formRef.value?.clearValidate())
    },
    { immediate: true }
  )

  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    const payload: Api.Infra.DistrictSaveParams = {
      pid: form.pid && form.pid !== '0' ? Number(form.pid) : 0,
      areaName: form.areaName.trim(),
      shortName: form.shortName.trim(),
      areaCode: form.areaCode.trim(),
      sort: form.sort
    }

    submitting.value = true
    try {
      if (isEdit.value && props.districtData) {
        await fetchUpdateDistrict(props.districtData.id, payload)
        ElMessage.success('保存成功')
      } else {
        await fetchCreateDistrict(payload)
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
