<!-- 会员等级页面 -->
<template>
  <div class="member-level-page art-full-height">
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
            <ElButton v-auth="'member:level:create'" @click="showDialog('add')" v-ripple>
              新增等级
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

      <LevelDialog
        v-model:visible="dialogVisible"
        :type="dialogType"
        :level-data="currentLevel"
        @success="handleDialogSuccess"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import { fetchDeleteLevel, fetchGetLevelList } from '@/api/member/level'
  import { ElButton, ElMessage, ElMessageBox, ElSpace, ElTag } from 'element-plus'
  import LevelDialog from './modules/level-dialog.vue'

  defineOptions({ name: 'MemberLevel' })

  type LevelItem = Api.Member.LevelItem
  type DialogType = 'add' | 'edit'

  const { hasAuth } = useAuth()

  const searchForm = ref<Api.Member.LevelSearchParams>({
    keyword: undefined,
    status: undefined
  })

  const searchItems = [
    {
      label: '等级名称',
      key: 'keyword',
      type: 'input',
      placeholder: '请输入等级名称',
      clearable: true
    },
    {
      label: '状态',
      key: 'status',
      type: 'select',
      props: {
        placeholder: '全部',
        clearable: true,
        options: [
          { label: '启用', value: 1 },
          { label: '禁用', value: 0 }
        ]
      }
    }
  ]

  const dialogVisible = ref(false)
  const dialogType = ref<DialogType>('add')
  const currentLevel = ref<LevelItem | null>(null)

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
      apiFn: fetchGetLevelList,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        {
          prop: 'level',
          label: '等级',
          width: 100,
          formatter: (row: LevelItem) =>
            h(ElTag, { type: 'warning', effect: 'plain' }, () => `Lv.${row.level}`)
        },
        { prop: 'name', label: '等级名称', minWidth: 140, showOverflowTooltip: true },
        {
          prop: 'experience',
          label: '升级经验',
          width: 120,
          formatter: (row: LevelItem) => row.experience
        },
        {
          prop: 'discountPercent',
          label: '折扣',
          width: 100,
          formatter: (row: LevelItem) =>
            row.discountPercent === 100 ? '无折扣' : `${row.discountPercent}%`
        },
        {
          prop: 'description',
          label: '等级说明',
          minWidth: 200,
          showOverflowTooltip: true,
          formatter: (row: LevelItem) => row.description || '-'
        },
        { prop: 'sort', label: '排序', width: 80 },
        {
          prop: 'status',
          label: '状态',
          width: 90,
          formatter: (row: LevelItem) =>
            h(ElTag, { type: row.status === 1 ? 'success' : 'info' }, () =>
              row.status === 1 ? '启用' : '禁用'
            )
        },
        {
          prop: 'createdAt',
          label: '创建时间',
          width: 170
        },
        {
          prop: 'operation',
          label: '操作',
          width: 170,
          fixed: 'right',
          formatter: (row: LevelItem) =>
            h('div', { class: 'flex items-center' }, [
              ...(hasAuth('member:level:update')
                ? [
                    h(ArtButtonTable, {
                      type: 'edit',
                      text: '编辑',
                      onClick: () => showDialog('edit', row)
                    })
                  ]
                : []),
              ...(hasAuth('member:level:delete')
                ? [
                    h(ArtButtonTable, {
                      type: 'delete',
                      text: '删除',
                      onClick: () => deleteLevel(row)
                    })
                  ]
                : [])
            ])
        }
      ]
    }
  })

  const handleSearch = (params: Api.Member.LevelSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }

  const showDialog = (type: DialogType, row?: LevelItem): void => {
    dialogType.value = type
    currentLevel.value = type === 'edit' ? (row ?? null) : null
    nextTick(() => {
      dialogVisible.value = true
    })
  }

  const handleDialogSuccess = (): void => {
    if (dialogType.value === 'add') {
      refreshCreate()
      return
    }
    refreshUpdate()
  }

  const deleteLevel = async (row: LevelItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(`确定要删除等级「${row.name}」吗？`, '删除会员等级', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch {
      return
    }

    try {
      await fetchDeleteLevel(row.id)
      ElMessage.success('删除成功')
      refreshRemove()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }
</script>
