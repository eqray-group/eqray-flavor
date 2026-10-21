<!-- 操作日志页面：只读，数据由 OperateLog 中间件写入 -->
<template>
  <div class="operate-log-page art-full-height">
    <ArtSearchBar
      v-model="searchForm"
      :items="searchItems"
      :rules="{}"
      @reset="resetSearchParams"
      @search="handleSearch"
    />

    <ElCard class="art-table-card">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData" />

      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { useTable } from '@/hooks/core/useTable'
  import { fetchGetOperateLogList } from '@/api/system/operate-log'
  import { ElCard, ElTag } from 'element-plus'

  defineOptions({ name: 'SystemOperateLog' })

  type OperateLogItem = Api.SystemManage.OperateLogItem

  const searchForm = ref<Api.SystemManage.OperateLogSearchParams>({
    type: undefined,
    subType: undefined,
    keyword: undefined,
    success: undefined
  })

  const searchItems = [
    {
      label: '操作模块',
      key: 'type',
      type: 'input',
      placeholder: '如：用户管理',
      clearable: true
    },
    {
      label: '操作动作',
      key: 'subType',
      type: 'input',
      placeholder: '如：新增用户',
      clearable: true
    },
    {
      label: '描述关键字',
      key: 'keyword',
      type: 'input',
      placeholder: '请输入操作描述',
      clearable: true
    },
    {
      label: '结果',
      key: 'success',
      type: 'select',
      props: {
        placeholder: '全部',
        options: [
          { label: '成功', value: 1 },
          { label: '失败', value: 0 }
        ]
      }
    }
  ]

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
    refreshData
  } = useTable({
    core: {
      apiFn: fetchGetOperateLogList,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        {
          prop: 'type',
          label: '操作模块',
          minWidth: 120,
          formatter: (row: OperateLogItem) => row.type || '-'
        },
        {
          prop: 'subType',
          label: '操作动作',
          minWidth: 130,
          formatter: (row: OperateLogItem) => row.subType || '-'
        },
        {
          prop: 'action',
          label: '操作描述',
          minWidth: 200,
          showOverflowTooltip: true
        },
        {
          prop: 'success',
          label: '结果',
          width: 90,
          formatter: (row: OperateLogItem) =>
            h(ElTag, { type: row.success === 1 ? 'success' : 'danger' }, () =>
              row.success === 1 ? '成功' : '失败'
            )
        },
        {
          prop: 'requestMethod',
          label: '方法',
          width: 90,
          formatter: (row: OperateLogItem) => row.requestMethod || '-'
        },
        {
          prop: 'requestUrl',
          label: '请求地址',
          minWidth: 240,
          showOverflowTooltip: true
        },
        {
          prop: 'requestIp',
          label: 'IP',
          width: 140,
          formatter: (row: OperateLogItem) => row.requestIp || '-'
        },
        {
          prop: 'duration',
          label: '耗时(ms)',
          width: 100,
          formatter: (row: OperateLogItem) => `${row.duration ?? 0}`
        },
        { prop: 'createdAt', label: '时间', width: 170 }
      ]
    }
  })

  const handleSearch = (params: Api.SystemManage.OperateLogSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }
</script>
