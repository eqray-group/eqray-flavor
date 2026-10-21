<!-- API 访问日志页面：只读，由全局请求日志中间件写入 -->
<template>
  <div class="api-access-log-page art-full-height">
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
  import { fetchGetApiAccessLog } from '@/api/infra/api-log'
  import { ElCard, ElTag } from 'element-plus'

  defineOptions({ name: 'InfraApiAccessLog' })

  type AccessLogItem = Api.Infra.ApiAccessLogItem

  const searchForm = ref<Api.Infra.ApiLogSearchParams>({
    requestUrl: undefined,
    resultCode: undefined
  })

  const searchItems = [
    {
      label: '请求地址',
      key: 'requestUrl',
      type: 'input',
      placeholder: '如：/api/v1/system/users',
      clearable: true
    },
    {
      label: '结果',
      key: 'resultCode',
      type: 'select',
      props: {
        placeholder: '全部',
        options: [
          { label: '成功', value: 0 },
          { label: '失败', value: 500 }
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
      apiFn: fetchGetApiAccessLog,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        {
          prop: 'requestMethod',
          label: '方法',
          width: 90,
          formatter: (row: AccessLogItem) => row.requestMethod || '-'
        },
        {
          prop: 'requestUrl',
          label: '请求地址',
          minWidth: 260,
          showOverflowTooltip: true
        },
        {
          prop: 'requestParams',
          label: '请求参数',
          minWidth: 200,
          showOverflowTooltip: true,
          formatter: (row: AccessLogItem) => row.requestParams || '-'
        },
        {
          prop: 'resultCode',
          label: '结果',
          width: 90,
          formatter: (row: AccessLogItem) =>
            h(ElTag, { type: row.resultCode === 0 ? 'success' : 'danger' }, () =>
              row.resultCode === 0 ? '成功' : String(row.resultCode)
            )
        },
        {
          prop: 'userIp',
          label: 'IP',
          width: 140,
          formatter: (row: AccessLogItem) => row.userIp || '-'
        },
        {
          prop: 'duration',
          label: '耗时(ms)',
          width: 100,
          formatter: (row: AccessLogItem) => `${row.duration ?? 0}`
        },
        { prop: 'createdAt', label: '时间', width: 170 }
      ]
    }
  })

  const handleSearch = (params: Api.Infra.ApiLogSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }
</script>
