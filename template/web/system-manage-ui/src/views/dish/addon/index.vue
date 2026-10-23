<!-- 加料管理页面：维护可复用的加料组（加料、甜度、温度等）与其下的加料项 -->
<template>
  <div class="dish-addon-page art-full-height">
    <ArtSearchBar
      v-model="searchForm"
      :items="searchItems"
      :rules="{}"
      @reset="resetSearchParams"
      @search="handleSearch"
    />

    <ElCard class="art-table-card">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElSpace wrap>
            <ElButton v-auth="'dish:addon:create'" @click="showDialog('add')" v-ripple>
              新增加料组
            </ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />

      <AddonDialog
        v-model:visible="dialogVisible"
        :type="dialogType"
        :addon-data="currentAddon"
        @success="handleDialogSuccess"
      />

      <AddonItemDialog
        v-model:visible="itemDialogVisible"
        :group-data="currentAddon"
        @success="refreshUpdate"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import { fetchDeleteAddonGroup, fetchGetAddonGroupList } from '@/api/dish/addon'
  import { ElButton, ElMessage, ElMessageBox, ElSpace, ElTag } from 'element-plus'
  import AddonDialog from './modules/addon-dialog.vue'
  import AddonItemDialog from './modules/addon-item-dialog.vue'

  defineOptions({ name: 'DishAddon' })

  type AddonGroupItem = Api.Dish.AddonGroupItem
  type DialogType = 'add' | 'edit'

  const { hasAuth } = useAuth()

  const searchForm = ref<Api.Dish.AddonSearchParams>({
    keyword: ''
  })

  const searchItems = [
    {
      label: '加料组名称',
      key: 'keyword',
      type: 'input',
      span: 8,
      props: {
        placeholder: '请输入加料组名称',
        clearable: true
      }
    }
  ]

  const dialogVisible = ref(false)
  const dialogType = ref<DialogType>('add')
  const currentAddon = ref<AddonGroupItem | null>(null)
  const itemDialogVisible = ref(false)

  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    getData,
    searchParams,
    resetSearchParams,
    handleSizeChange,
    handleCurrentChange,
    refreshData,
    refreshCreate,
    refreshUpdate,
    refreshRemove
  } = useTable({
    core: {
      apiFn: fetchGetAddonGroupList,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        {
          prop: 'name',
          label: '加料组名称',
          minWidth: 160,
          showOverflowTooltip: true
        },
        {
          prop: 'required',
          label: '是否必选',
          width: 110,
          formatter: (row: AddonGroupItem) =>
            h(
              ElTag,
              { type: row.required === 1 ? 'warning' : 'info', effect: 'plain', size: 'small' },
              () => (row.required === 1 ? '必选' : '可选')
            )
        },
        {
          prop: 'selectRange',
          label: '选择数量',
          width: 130,
          formatter: (row: AddonGroupItem) =>
            `${row.minSelect} ~ ${row.maxSelect > 0 ? row.maxSelect : '不限'}`
        },
        {
          prop: 'itemCount',
          label: '加料项',
          width: 100
        },
        {
          prop: 'spuCount',
          label: '引用菜品',
          width: 110
        },
        {
          prop: 'sort',
          label: '排序',
          width: 90
        },
        {
          prop: 'status',
          label: '状态',
          width: 100,
          formatter: (row: AddonGroupItem) =>
            h(
              ElTag,
              { type: row.status === 1 ? 'success' : 'info', effect: 'plain', size: 'small' },
              () => (row.status === 1 ? '启用' : '禁用')
            )
        },
        {
          prop: 'remark',
          label: '备注',
          minWidth: 160,
          showOverflowTooltip: true
        },
        { prop: 'createdAt', label: '创建时间', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          width: 260,
          fixed: 'right',
          formatter: (row: AddonGroupItem) =>
            h('div', { class: 'flex items-center' }, [
              ...(hasAuth('dish:addon:list')
                ? [
                    h(ArtButtonTable, {
                      type: 'view',
                      text: '加料项',
                      onClick: () => showItemDialog(row)
                    })
                  ]
                : []),
              ...(hasAuth('dish:addon:update')
                ? [
                    h(ArtButtonTable, {
                      type: 'edit',
                      text: '编辑',
                      onClick: () => showDialog('edit', row)
                    })
                  ]
                : []),
              ...(hasAuth('dish:addon:delete')
                ? [
                    h(ArtButtonTable, {
                      type: 'delete',
                      text: '删除',
                      onClick: () => deleteAddon(row)
                    })
                  ]
                : [])
            ])
        }
      ]
    }
  })

  // 搜索条件变化后回到第一页重新加载加料组列表
  const handleSearch = (params: Api.Dish.AddonSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }

  // 打开新增 / 编辑加料组弹窗
  const showDialog = (type: DialogType, row?: AddonGroupItem): void => {
    dialogType.value = type
    currentAddon.value = type === 'edit' ? (row ?? null) : null
    nextTick(() => {
      dialogVisible.value = true
    })
  }

  // 打开加料项管理弹窗
  const showItemDialog = (row: AddonGroupItem): void => {
    currentAddon.value = row
    itemDialogVisible.value = true
  }

  // 弹窗提交成功后按新增 / 编辑分别刷新列表
  const handleDialogSuccess = (): void => {
    if (dialogType.value === 'add') {
      refreshCreate()
      return
    }
    refreshUpdate()
  }

  // 删除加料组：仍被菜品引用时由后端拒绝
  const deleteAddon = async (row: AddonGroupItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(
        `确定要删除加料组「${row.name}」吗？其下加料项会一并删除。`,
        '删除加料组',
        {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }
      )
    } catch {
      return
    }

    try {
      await fetchDeleteAddonGroup(row.id)
      ElMessage.success('删除成功')
      refreshRemove()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }
</script>
