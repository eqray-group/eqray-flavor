<!-- 登录日志页面：只读，数据来自 base 认证组件 -->
<template>
  <div class="login-log-page art-full-height">
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
  import { fetchGetLoginLogList } from '@/api/system/login-log'
  import { ElCard, ElTag } from 'element-plus'

  defineOptions({ name: 'SystemLoginLog' })

  type LoginLogItem = Api.SystemManage.LoginLogItem

  const searchForm = ref<Api.SystemManage.LoginLogSearchParams>({
    username: undefined,
    type: undefined,
    result: undefined
  })

  const searchItems = [
    {
      label: '登录账号',
      key: 'username',
      type: 'input',
      placeholder: '请输入登录账号',
      clearable: true
    },
    {
      label: '类型',
      key: 'type',
      type: 'select',
      props: {
        placeholder: '全部',
        options: [
          { label: '登录', value: 1 },
          { label: '登出', value: 2 }
        ]
      }
    },
    {
      label: '结果',
      key: 'result',
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
      apiFn: fetchGetLoginLogList,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        { prop: 'username', label: '登录账号', minWidth: 140 },
        {
          prop: 'type',
          label: '类型',
          width: 90,
          formatter: (row: LoginLogItem) =>
            h(ElTag, { type: row.type === 1 ? 'primary' : 'info' }, () =>
              row.type === 1 ? '登录' : '登出'
            )
        },
        {
          prop: 'result',
          label: '结果',
          width: 90,
          formatter: (row: LoginLogItem) =>
            h(ElTag, { type: row.result === 1 ? 'success' : 'danger' }, () =>
              row.result === 1 ? '成功' : '失败'
            )
        },
        {
          prop: 'userIp',
          label: 'IP',
          width: 140,
          formatter: (row: LoginLogItem) => row.userIp || '-'
        },
        {
          prop: 'userAgent',
          label: '客户端',
          minWidth: 260,
          showOverflowTooltip: true,
          formatter: (row: LoginLogItem) => row.userAgent || '-'
        },
        { prop: 'createdAt', label: '时间', width: 170 }
      ]
    }
  })

  const handleSearch = (params: Api.SystemManage.LoginLogSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }
</script>
