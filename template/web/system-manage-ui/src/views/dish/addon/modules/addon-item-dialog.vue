<!-- 加料项管理抽屉：在指定加料组内维护加料项的名称、加价与排序 -->
<template>
  <ElDrawer
    v-model="drawerVisible"
    :title="`加料项管理 · ${groupData?.name || ''}`"
    size="720px"
    :close-on-click-modal="false"
    destroy-on-close
  >
    <div class="flex h-full flex-col">
      <!-- 内联编辑区：新增与编辑共用，editingId 为空表示新增 -->
      <div class="mb-4 rounded-lg bg-[var(--el-fill-color-lighter)] p-4">
        <div class="mb-2 text-sm font-medium">
          {{ editingId ? '编辑加料项' : '新增加料项' }}
        </div>
        <div class="grid grid-cols-2 gap-x-3 gap-y-2">
          <ElInput v-model="form.name" placeholder="加料项名称，如加珍珠" maxlength="64" />
          <ElInputNumber
            v-model="form.price"
            :min="0"
            :precision="2"
            :step="1"
            :controls="false"
            class="w-full"
            placeholder="加价（元），0 表示免费"
          />
          <ElInputNumber
            v-model="form.sort"
            :min="0"
            :max="9999"
            :controls="false"
            class="w-full"
            placeholder="排序"
          />
          <ElSelect v-model="form.status" class="w-full">
            <ElOption label="启用" :value="1" />
            <ElOption label="禁用" :value="0" />
          </ElSelect>
          <ElInput
            v-model="form.remark"
            class="col-span-2"
            placeholder="备注，选填"
            maxlength="255"
          />
        </div>

        <div class="mt-3 flex items-center gap-2">
          <ElButton type="primary" :loading="submitting" @click="handleSubmit">
            {{ editingId ? '保存修改' : '添加' }}
          </ElButton>
          <ElButton v-if="editingId" @click="resetForm">取消编辑</ElButton>
          <span class="text-xs text-g-600">加料项在组内名称唯一，加价单位为元</span>
        </div>
      </div>

      <div class="min-h-0 flex-1">
        <ArtTable
          :loading="loading"
          :data="items"
          :columns="columns"
          :show-table-header="false"
          height="calc(100vh - 420px)"
          row-key="id"
          empty-text="该加料组还没有加料项"
        />
      </div>
    </div>
  </ElDrawer>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { ElMessage, ElMessageBox, ElTag } from 'element-plus'
  import {
    fetchCreateAddonItem,
    fetchDeleteAddonItem,
    fetchGetAddonGroupDetail,
    fetchUpdateAddonItem
  } from '@/api/dish/addon'
  import { useAuth } from '@/hooks/core/useAuth'

  defineOptions({ name: 'DishAddonItemDialog' })

  type AddonGroupItem = Api.Dish.AddonGroupItem
  type AddonItem = Api.Dish.AddonItem

  interface Props {
    visible: boolean
    groupData?: AddonGroupItem | null
  }

  const props = withDefaults(defineProps<Props>(), {
    groupData: null
  })

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void
    (e: 'success'): void
  }>()

  const drawerVisible = computed({
    get: () => props.visible,
    set: (value: boolean) => emit('update:visible', value)
  })

  const { hasAuth } = useAuth()

  const loading = ref(false)
  const submitting = ref(false)
  const items = ref<AddonItem[]>([])
  const editingId = ref('')

  const form = reactive({
    name: '',
    price: null as number | null,
    sort: 0,
    status: 1,
    remark: ''
  })

  // 分转元，空值返回 null 以便输入框显示占位文案
  const centToYuan = (cent: number): number | null =>
    cent > 0 ? Number((cent / 100).toFixed(2)) : null

  // 元转分，空值按 0 处理表示免费
  const yuanToCent = (yuan: number | null): number => {
    if (yuan === null || Number.isNaN(yuan)) return 0
    return Math.round(yuan * 100)
  }

  // 重置内联表单为新增状态
  const resetForm = (): void => {
    editingId.value = ''
    Object.assign(form, { name: '', price: null, sort: 0, status: 1, remark: '' })
  }

  // 拉取加料组详情，用返回的加料项刷新表格
  const loadItems = async (): Promise<void> => {
    if (!props.groupData) return

    loading.value = true
    try {
      const detail = await fetchGetAddonGroupDetail(props.groupData.id)
      items.value = detail.items || []
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      loading.value = false
    }
  }

  const columns = computed(() => [
    { prop: 'name', label: '加料项', minWidth: 140, showOverflowTooltip: true },
    {
      prop: 'price',
      label: '加价（元）',
      width: 110,
      formatter: (row: AddonItem) => (row.price / 100).toFixed(2)
    },
    { prop: 'sort', label: '排序', width: 80 },
    {
      prop: 'status',
      label: '状态',
      width: 90,
      formatter: (row: AddonItem) =>
        h(
          ElTag,
          { type: row.status === 1 ? 'success' : 'info', effect: 'plain', size: 'small' },
          () => (row.status === 1 ? '启用' : '禁用')
        )
    },
    { prop: 'remark', label: '备注', minWidth: 130, showOverflowTooltip: true },
    {
      prop: 'operation',
      label: '操作',
      width: 150,
      formatter: (row: AddonItem) =>
        h('div', { class: 'flex items-center' }, [
          ...(hasAuth('dish:addon:update')
            ? [
                h(ArtButtonTable, {
                  type: 'edit',
                  text: '编辑',
                  onClick: () => startEdit(row)
                })
              ]
            : []),
          ...(hasAuth('dish:addon:delete')
            ? [
                h(ArtButtonTable, {
                  type: 'delete',
                  text: '删除',
                  onClick: () => deleteItem(row)
                })
              ]
            : [])
        ])
    }
  ])

  // 打开抽屉时加载加料项并重置内联表单
  watch(
    () => props.visible,
    (visible) => {
      if (!visible) return
      resetForm()
      loadItems()
    },
    { immediate: true }
  )

  // 把行数据回填到内联表单，进入编辑态
  const startEdit = (row: AddonItem): void => {
    editingId.value = row.id
    Object.assign(form, {
      name: row.name,
      price: centToYuan(row.price),
      sort: row.sort,
      status: row.status,
      remark: row.remark
    })
  }

  // 提交加料项：新增调用组接口，编辑调用加料项接口
  const handleSubmit = async (): Promise<void> => {
    if (!props.groupData) return

    if (!form.name.trim()) {
      ElMessage.error('请输入加料项名称')
      return
    }

    const payload: Api.Dish.AddonItemSaveParams = {
      name: form.name.trim(),
      price: yuanToCent(form.price),
      sort: form.sort,
      status: form.status,
      remark: form.remark
    }

    submitting.value = true
    try {
      if (editingId.value) {
        await fetchUpdateAddonItem(editingId.value, payload)
        ElMessage.success('保存成功')
      } else {
        await fetchCreateAddonItem(props.groupData.id, payload)
        ElMessage.success('添加成功')
      }

      resetForm()
      await loadItems()
      emit('success')
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      submitting.value = false
    }
  }

  // 删除加料项：确认后删除并刷新列表
  const deleteItem = async (row: AddonItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(`确定要删除加料项「${row.name}」吗？`, '删除加料项', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch {
      return
    }

    try {
      await fetchDeleteAddonItem(row.id)
      ElMessage.success('删除成功')
      await loadItems()
      emit('success')
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }
</script>
