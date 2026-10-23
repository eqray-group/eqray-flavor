<!-- 支付渠道管理页面：按应用维护微信 / 支付宝 / 模拟支付的渠道参数与费率 -->
<template>
  <div class="pay-channel-page art-full-height">
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
            <ElButton v-auth="'pay:channel:create'" @click="showDialog('add')" v-ripple>
              新增支付渠道
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

      <ChannelDialog
        v-model:visible="dialogVisible"
        :type="dialogType"
        :channel-data="currentChannel"
        :app-options="appOptions"
        @success="handleDialogSuccess"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import { fetchGetPayAppOptions } from '@/api/pay/app'
  import {
    fetchGetPayChannelOptions,
    fetchDeletePayChannel,
    fetchGetPayChannelList
  } from '@/api/pay/channel'
  import { ElButton, ElMessage, ElMessageBox, ElSpace, ElTag } from 'element-plus'
  import ChannelDialog from './modules/channel-dialog.vue'

  defineOptions({ name: 'PayChannel' })

  type ChannelItem = Api.Pay.ChannelItem
  type DialogType = 'add' | 'edit'

  const { hasAuth } = useAuth()

  const statusOptions = [
    { label: '开启', value: 0 },
    { label: '关闭', value: 1 }
  ]

  const statusTagTypes: Record<number, 'success' | 'info'> = {
    0: 'success',
    1: 'info'
  }

  const appOptions = ref<Api.Pay.AppOption[]>([])
  const channelOptions = ref<Api.Pay.ChannelOption[]>([])

  const searchForm = ref<Api.Pay.ChannelSearchParams>({
    appId: '',
    code: '',
    status: undefined
  })

  // 渠道编码下拉依赖后端枚举，初始化后再渲染搜索项
  const searchItems = computed(() => [
    {
      label: '所属应用',
      key: 'appId',
      type: 'select',
      span: 6,
      props: {
        placeholder: '全部',
        clearable: true,
        options: appOptions.value.map((item) => ({ label: item.name, value: item.id }))
      }
    },
    {
      label: '渠道编码',
      key: 'code',
      type: 'select',
      span: 6,
      props: {
        placeholder: '全部',
        clearable: true,
        options: channelOptions.value.map((item) => ({ label: item.name, value: item.code }))
      }
    },
    {
      label: '状态',
      key: 'status',
      type: 'select',
      span: 5,
      props: {
        placeholder: '全部',
        clearable: true,
        options: statusOptions
      }
    }
  ])

  const dialogVisible = ref(false)
  const dialogType = ref<DialogType>('add')
  const currentChannel = ref<ChannelItem | null>(null)

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
      apiFn: fetchGetPayChannelList,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        {
          prop: 'appId',
          label: '所属应用',
          minWidth: 150,
          formatter: (row: ChannelItem) => appName(row.appId)
        },
        {
          prop: 'channelName',
          label: '渠道',
          minWidth: 140
        },
        {
          prop: 'code',
          label: '渠道编码',
          minWidth: 130
        },
        {
          prop: 'feeRate',
          label: '费率',
          width: 100,
          formatter: (row: ChannelItem) => `${row.feeRate ?? 0}%`
        },
        {
          prop: 'status',
          label: '状态',
          width: 90,
          formatter: (row: ChannelItem) =>
            h(
              ElTag,
              { type: statusTagTypes[row.status] ?? 'info', effect: 'plain' },
              () => row.statusName || (row.status === 0 ? '开启' : '关闭')
            )
        },
        {
          prop: 'remark',
          label: '备注',
          minWidth: 150,
          showOverflowTooltip: true
        },
        { prop: 'createdAt', label: '创建时间', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          width: 170,
          fixed: 'right',
          formatter: (row: ChannelItem) =>
            h('div', { class: 'flex items-center' }, [
              ...(hasAuth('pay:channel:update')
                ? [
                    h(ArtButtonTable, {
                      type: 'edit',
                      text: '编辑',
                      onClick: () => showDialog('edit', row)
                    })
                  ]
                : []),
              ...(hasAuth('pay:channel:delete')
                ? [
                    h(ArtButtonTable, {
                      type: 'delete',
                      text: '删除',
                      onClick: () => deleteChannel(row)
                    })
                  ]
                : [])
            ])
        }
      ]
    }
  })

  // 按应用 ID 取应用名，未命中时回退展示 ID
  const appName = (appId: string): string => {
    return appOptions.value.find((item) => item.id === appId)?.name || appId
  }

  // 加载应用与渠道编码下拉，供筛选与弹窗共用
  const loadOptions = async (): Promise<void> => {
    try {
      const [apps, codes] = await Promise.all([
        fetchGetPayAppOptions(),
        fetchGetPayChannelOptions()
      ])
      appOptions.value = apps ?? []
      channelOptions.value = codes ?? []
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }

  onMounted(() => {
    loadOptions()
  })

  // 搜索条件变化后重新加载渠道列表
  const handleSearch = (params: Api.Pay.ChannelSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }

  // 打开新增 / 编辑弹窗，编辑时携带当前行用于回填
  const showDialog = (type: DialogType, row?: ChannelItem): void => {
    dialogType.value = type
    currentChannel.value = type === 'edit' ? (row ?? null) : null
    nextTick(() => {
      dialogVisible.value = true
    })
  }

  // 弹窗提交成功后按新增 / 编辑分别刷新列表
  const handleDialogSuccess = (): void => {
    if (dialogType.value === 'add') {
      refreshCreate()
      return
    }
    refreshUpdate()
  }

  // 删除支付渠道，确认后调用接口并刷新列表
  const deleteChannel = async (row: ChannelItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(
        `确定要删除渠道「${row.channelName}」吗？历史支付单会保留渠道快照。`,
        '删除支付渠道',
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
      await fetchDeletePayChannel(row.id)
      ElMessage.success('删除成功')
      refreshRemove()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }
</script>
