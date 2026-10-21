<!-- 字典数据面板：展示指定字典类型下的数据项 -->
<template>
  <div class="dict-data-panel">
    <div class="dict-data-header">
      <div class="dict-data-title">
        <span class="dict-data-name">{{ dictType?.name || '字典数据' }}</span>
        <ElTag v-if="dictType" type="info" size="small">{{ dictType.type }}</ElTag>
      </div>
      <ElSpace>
        <ElButton
          v-auth="'system:dict:create'"
          type="primary"
          :disabled="!dictType"
          v-ripple
          @click="openForm()"
        >
          新增字典数据
        </ElButton>
        <ElButton :loading="loading" :disabled="!dictType" @click="loadData">刷新</ElButton>
      </ElSpace>
    </div>

    <div class="dict-data-body">
      <ElEmpty v-if="!dictType" description="请先在左侧选择字典类型" />
      <ElTable v-else v-loading="loading" :data="list" border stripe height="100%">
        <ElTableColumn prop="label" label="显示文本" min-width="130" />
        <ElTableColumn prop="value" label="键值" min-width="110" />
        <ElTableColumn prop="sort" label="排序" width="80" />
        <ElTableColumn label="标签样式" width="110">
          <template #default="{ row }">
            <ElTag :type="tagType(row.colorType)">{{ row.colorType || 'default' }}</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="状态" width="90">
          <template #default="{ row }">
            <ElTag :type="row.status === 1 ? 'success' : 'info'">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="remark" label="备注" min-width="140" show-overflow-tooltip>
          <template #default="{ row }">{{ row.remark || '-' }}</template>
        </ElTableColumn>
        <ElTableColumn label="操作" width="170" fixed="right">
          <template #default="{ row }">
            <div class="flex items-center">
              <ArtButtonTable
                v-if="hasAuth('system:dict:update')"
                type="edit"
                text="编辑"
                @click="openForm(row)"
              />
              <ArtButtonTable
                v-if="hasAuth('system:dict:delete')"
                type="delete"
                text="删除"
                @click="handleDelete(row)"
              />
            </div>
          </template>
        </ElTableColumn>
      </ElTable>
    </div>

    <div v-if="dictType" class="dict-data-pagination">
      <ElPagination
        v-model:current-page="pageNo"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        @size-change="handleSizeChange"
        @current-change="loadData"
      />
    </div>

    <ElDialog
      v-model="formVisible"
      :title="editing ? '编辑字典数据' : '新增字典数据'"
      width="520px"
      align-center
      append-to-body
      :close-on-click-modal="false"
    >
      <ElForm ref="formRef" :model="form" :rules="rules" label-width="90px">
        <ElFormItem label="显示文本" prop="label">
          <ElInput v-model="form.label" placeholder="如：启用" maxlength="64" />
        </ElFormItem>
        <ElFormItem label="键值" prop="value">
          <ElInput v-model="form.value" placeholder="如：1" maxlength="64" />
        </ElFormItem>
        <ElFormItem label="排序" prop="sort">
          <ElInputNumber v-model="form.sort" :min="0" :max="9999" controls-position="right" />
        </ElFormItem>
        <ElFormItem label="标签样式" prop="colorType">
          <ElSelect v-model="form.colorType" placeholder="请选择标签样式">
            <ElOption
              v-for="item in colorOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="状态" prop="status">
          <ElSwitch
            v-model="form.status"
            :active-value="1"
            :inactive-value="0"
            active-text="启用"
            inactive-text="禁用"
          />
        </ElFormItem>
        <ElFormItem label="备注" prop="remark">
          <ElInput
            v-model="form.remark"
            type="textarea"
            :rows="3"
            maxlength="500"
            show-word-limit
            placeholder="选填"
          />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="formVisible = false">取消</ElButton>
        <ElButton type="primary" :loading="submitting" @click="handleSubmit">提交</ElButton>
      </template>
    </ElDialog>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useAuth } from '@/hooks/core/useAuth'
  import type { FormInstance, FormRules } from 'element-plus'
  import {
    fetchCreateDictData,
    fetchDeleteDictData,
    fetchGetDictDataList,
    fetchUpdateDictData
  } from '@/api/system/dict'

  defineOptions({ name: 'DictDataPanel' })

  type DictTypeItem = Api.SystemManage.DictTypeItem
  type DictDataItem = Api.SystemManage.DictDataItem

  const { hasAuth } = useAuth()

  interface Props {
    dictType?: DictTypeItem | null
  }

  const props = withDefaults(defineProps<Props>(), { dictType: null })

  const colorOptions = [
    { label: '默认', value: 'default' },
    { label: '主要', value: 'primary' },
    { label: '成功', value: 'success' },
    { label: '信息', value: 'info' },
    { label: '警告', value: 'warning' },
    { label: '危险', value: 'danger' }
  ]

  const loading = ref(false)
  const list = ref<DictDataItem[]>([])
  const pageNo = ref(1)
  const pageSize = ref(20)
  const total = ref(0)

  const loadData = async (): Promise<void> => {
    if (!props.dictType?.type) {
      list.value = []
      total.value = 0
      return
    }

    loading.value = true
    try {
      const res = await fetchGetDictDataList({
        dictType: props.dictType.type,
        pageNo: pageNo.value,
        pageSize: pageSize.value
      })
      list.value = res.list
      total.value = res.total
    } catch {
      list.value = []
      total.value = 0
    } finally {
      loading.value = false
    }
  }

  const handleSizeChange = (): void => {
    pageNo.value = 1
    loadData()
  }

  // 切换左侧字典类型时重新加载数据
  watch(
    () => props.dictType?.type,
    () => {
      pageNo.value = 1
      loadData()
    },
    { immediate: true }
  )

  const formVisible = ref(false)
  const formRef = ref<FormInstance>()
  const submitting = ref(false)
  const editing = ref<DictDataItem | null>(null)

  const form = reactive({
    label: '',
    value: '',
    sort: 0,
    colorType: 'default',
    status: 1,
    remark: ''
  })

  const rules: FormRules = {
    label: [
      { required: true, message: '请输入显示文本', trigger: 'blur' },
      { max: 64, message: '显示文本不能超过 64 个字符', trigger: 'blur' }
    ],
    value: [
      { required: true, message: '请输入键值', trigger: 'blur' },
      { max: 64, message: '键值不能超过 64 个字符', trigger: 'blur' }
    ],
    remark: [{ max: 500, message: '备注不能超过 500 个字符', trigger: 'blur' }]
  }

  const openForm = (row?: DictDataItem): void => {
    editing.value = row || null
    Object.assign(form, {
      label: row?.label ?? '',
      value: row?.value ?? '',
      sort: row?.sort ?? 0,
      colorType: row?.colorType ?? 'default',
      status: row?.status ?? 1,
      remark: row?.remark ?? ''
    })
    formVisible.value = true
    nextTick(() => formRef.value?.clearValidate())
  }

  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value || !props.dictType?.type) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    const payload: Api.SystemManage.DictDataSaveParams = {
      label: form.label.trim(),
      value: form.value.trim(),
      sort: form.sort,
      dictType: props.dictType.type,
      status: form.status,
      colorType: form.colorType,
      remark: form.remark.trim()
    }

    submitting.value = true
    try {
      if (editing.value) {
        await fetchUpdateDictData(editing.value.id, payload)
        ElMessage.success('保存成功')
      } else {
        await fetchCreateDictData(payload)
        ElMessage.success('新增成功')
      }
      formVisible.value = false
      loadData()
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      submitting.value = false
    }
  }

  const handleDelete = async (row: DictDataItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(`确定要删除字典项「${row.label}」吗？`, '删除字典数据', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch {
      return
    }

    try {
      await fetchDeleteDictData(row.id)
      ElMessage.success('删除成功')
      loadData()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }

  const tagType = (colorType: string): 'primary' | 'success' | 'info' | 'warning' | 'danger' => {
    const map: Record<string, 'primary' | 'success' | 'info' | 'warning' | 'danger'> = {
      primary: 'primary',
      success: 'success',
      info: 'info',
      warning: 'warning',
      danger: 'danger'
    }
    return map[colorType] ?? 'info'
  }
</script>

<style scoped>
  .dict-data-panel {
    display: flex;
    flex-direction: column;
    height: 100%;
    min-height: 0;
  }

  .dict-data-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    flex-wrap: wrap;
    margin-bottom: 12px;
  }

  .dict-data-title {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
  }

  .dict-data-name {
    font-size: 15px;
    font-weight: 600;
    color: var(--el-text-color-primary);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .dict-data-body {
    flex: 1;
    min-height: 0;
  }

  .dict-data-pagination {
    display: flex;
    justify-content: flex-end;
    margin-top: 12px;
  }
</style>
