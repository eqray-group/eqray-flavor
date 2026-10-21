<!-- API 错误日志页面：业务失败时由请求日志中间件写入 -->
<template>
  <div class="api-error-log-page art-full-height">
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
  import { useAuth } from '@/hooks/core/useAuth'
  import { fetchGetApiErrorLog, fetchProcessApiErrorLog } from '@/api/infra/api-log'
  import { ElButton, ElCard, ElMessage, ElMessageBox, ElTag } from 'element-plus'

  defineOptions({ name: 'InfraApiErrorLog' })

  type ErrorLogItem = Api.Infra.ApiErrorLogItem

  const { hasAuth } = useAuth()

  const searchForm = ref<Api.Infra.ApiLogSearchParams>({
    requestUrl: undefined,
    keyword: undefined
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
      label: '错误信息',
      key: 'keyword',
      type: 'input',
      placeholder: '请输入错误信息关键字',
      clearable: true
    }
  ]

  const handleProcess = async (row: ErrorLogItem): Promise<void> => {
    try {
      await ElMessageBox.confirm('确认已处理该错误？', '标记处理', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch {
      return
    }

    try {
      await fetchProcessApiErrorLog(row.id)
      ElMessage.success('已标记为处理')
      refreshUpdate()
    } catch {
      // 错误提示已由请求层统一弹出
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
    refreshUpdate
  } = useTable({
    core: {
      apiFn: fetchGetApiErrorLog,
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
          formatter: (row: ErrorLogItem) => row.requestMethod || '-'
        },
        {
          prop: 'requestUrl',
          label: '请求地址',
          minWidth: 240,
          showOverflowTooltip: true
        },
        {
          prop: 'name',
          label: '异常类型',
          width: 140,
          formatter: (row: ErrorLogItem) => row.name || '-'
        },
        {
          prop: 'message',
          label: '错误信息',
          minWidth: 240,
          showOverflowTooltip: true,
          formatter: (row: ErrorLogItem) => row.message || '-'
        },
        {
          prop: 'userIp',
          label: 'IP',
          width: 140,
          formatter: (row: ErrorLogItem) => row.userIp || '-'
        },
        {
          prop: 'processStatus',
          label: '处理状态',
          width: 110,
          formatter: (row: ErrorLogItem) =>
            h(ElTag, { type: row.processStatus === 1 ? 'success' : 'warning' }, () =>
              row.processStatus === 1 ? '已处理' : '未处理'
            )
        },
        { prop: 'createdAt', label: '时间', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          width: 120,
          fixed: 'right',
          formatter: (row: ErrorLogItem) =>
            row.processStatus === 1 || !hasAuth('infra:api-log:update')
              ? h('span', { class: 'text-g-500' }, '-')
              : h(
                  ElButton,
                  { type: 'primary', link: true, size: 'small', onClick: () => handleProcess(row) },
                  () => '标记处理'
                )
        }
      ]
    }
  })

  const handleSearch = (params: Api.Infra.ApiLogSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }
</script>
