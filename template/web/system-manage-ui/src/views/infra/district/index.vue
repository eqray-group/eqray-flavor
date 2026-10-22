<!-- 行政区划管理：省 / 市 / 区县 / 街道数据维护 -->
<template>
  <div class="district-page art-full-height">
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
            <ElButton v-auth="'infra:district:create'" @click="showDialog('add')" v-ripple>
              新增区划
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

      <DistrictDialog
        v-model:visible="dialogVisible"
        :type="dialogType"
        :district-data="currentDistrict"
        :district-tree="districtTree"
        @success="handleDialogSuccess"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import {
    fetchDeleteDistrict,
    fetchGetDistrictList,
    fetchGetDistrictTree
  } from '@/api/infra/district'
  import { ElButton, ElMessage, ElMessageBox, ElSpace, ElTag } from 'element-plus'
  import DistrictDialog from './modules/district-dialog.vue'

  defineOptions({ name: 'InfraDistrict' })

  type DistrictItem = Api.Infra.DistrictItem
  type DialogType = 'add' | 'edit'

  const { hasAuth } = useAuth()

  // 级别标签：1 省 / 2 市 / 3 区县 / 4 街道，超出范围按 L{n} 展示
  const LEVEL_CONFIG: Record<
    number,
    { type: 'primary' | 'success' | 'warning' | 'info'; text: string }
  > = {
    1: { type: 'primary', text: '省' },
    2: { type: 'success', text: '市' },
    3: { type: 'warning', text: '区县' },
    4: { type: 'info', text: '街道' }
  }

  const searchForm = ref<Api.Infra.DistrictSearchParams>({
    keyword: undefined,
    level: undefined
  })

  const searchItems = [
    {
      label: '关键字',
      key: 'keyword',
      type: 'input',
      span: 6,
      props: { placeholder: '名称 / 简称 / 行政编码', clearable: true }
    },
    {
      label: '行政级别',
      key: 'level',
      type: 'select',
      props: {
        placeholder: '全部',
        clearable: true,
        options: [
          { label: '省', value: 1 },
          { label: '市', value: 2 },
          { label: '区县', value: 3 },
          { label: '街道', value: 4 }
        ]
      }
    }
  ]

  const dialogVisible = ref(false)
  const dialogType = ref<DialogType>('add')
  const currentDistrict = ref<DistrictItem | null>(null)

  // 区划树：既是上级选择的数据源，也用来把 pid 翻译成上级名称
  const districtTree = ref<Api.Infra.DistrictTreeNode[]>([])
  const parentNameMap = computed(() => {
    const map = new Map<string, string>()
    const walk = (nodes: Api.Infra.DistrictTreeNode[]) => {
      nodes.forEach((node) => {
        map.set(node.id, node.areaName)
        walk(node.children ?? [])
      })
    }
    walk(districtTree.value)
    return map
  })

  // 加载区划树，失败时错误提示由请求层统一弹出
  const loadTree = async (): Promise<void> => {
    try {
      districtTree.value = await fetchGetDistrictTree()
    } catch {
      districtTree.value = []
    }
  }

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
      apiFn: fetchGetDistrictList,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        {
          prop: 'areaName',
          label: '区划名称',
          minWidth: 180,
          formatter: (row: DistrictItem) =>
            h('div', { class: 'flex flex-col' }, [
              h('span', row.areaName),
              row.shortName && row.shortName !== row.areaName
                ? h('span', { class: 'text-xs text-g-600' }, row.shortName)
                : null
            ])
        },
        { prop: 'areaCode', label: '行政编码', width: 130 },
        {
          prop: 'level',
          label: '级别',
          width: 90,
          formatter: (row: DistrictItem) => {
            const conf = LEVEL_CONFIG[row.level] ?? { type: 'info' as const, text: `L${row.level}` }
            return h(ElTag, { type: conf.type, effect: 'plain' }, () => conf.text)
          }
        },
        {
          prop: 'pid',
          label: '上级区划',
          minWidth: 160,
          formatter: (row: DistrictItem) =>
            row.pid === '0'
              ? '—'
              : (parentNameMap.value.get(row.pid) ?? (row.parentAreaCode || row.pid))
        },
        { prop: 'sort', label: '排序', width: 80 },
        { prop: 'createdAt', label: '创建时间', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          width: 170,
          fixed: 'right',
          formatter: (row: DistrictItem) =>
            h('div', { class: 'flex items-center' }, [
              ...(hasAuth('infra:district:update')
                ? [
                    h(ArtButtonTable, {
                      type: 'edit',
                      text: '编辑',
                      onClick: () => showDialog('edit', row)
                    })
                  ]
                : []),
              ...(hasAuth('infra:district:delete')
                ? [
                    h(ArtButtonTable, {
                      type: 'delete',
                      text: '删除',
                      onClick: () => deleteDistrict(row)
                    })
                  ]
                : [])
            ])
        }
      ]
    }
  })

  onMounted(() => {
    loadTree()
  })

  // 搜索：把搜索栏条件并入列表参数后重新请求
  const handleSearch = (params: Api.Infra.DistrictSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }

  // 打开新增 / 编辑弹窗；编辑时回填当前行
  const showDialog = (type: DialogType, row?: DistrictItem): void => {
    dialogType.value = type
    currentDistrict.value = type === 'edit' ? (row ?? null) : null
    nextTick(() => {
      dialogVisible.value = true
    })
  }

  // 弹窗提交成功：新增与编辑走不同的刷新策略，并同步区划树
  const handleDialogSuccess = (): void => {
    loadTree()
    if (dialogType.value === 'add') {
      refreshCreate()
      return
    }
    refreshUpdate()
  }

  // 删除区划：仍有下级时后端会拒绝，错误提示由请求层统一弹出
  const deleteDistrict = async (row: DistrictItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(`确定要删除区划「${row.areaName}」吗？`, '删除区划', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch {
      return
    }

    try {
      await fetchDeleteDistrict(row.id)
      ElMessage.success('删除成功')
      loadTree()
      refreshRemove()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }
</script>
