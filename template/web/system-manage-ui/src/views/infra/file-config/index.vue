<!-- 文件配置页面：本地存储与主配置维护 -->
<template>
  <div class="file-config-page art-full-height">
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
            <ElButton v-auth="'infra:file-config:create'" @click="showDialog('add')" v-ripple>
              新增配置
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
    </ElCard>

    <FileConfigDialog
      v-model:visible="dialogVisible"
      :type="dialogType"
      :config-data="currentConfig"
      @success="handleDialogSuccess"
    />
  </div>
</template>

<script setup lang="ts">
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import {
    fetchDeleteFileConfig,
    fetchGetFileConfigList,
    fetchSetFileConfigMaster
  } from '@/api/infra/file-config'
  import { ElButton, ElCard, ElMessage, ElMessageBox, ElSpace, ElTag } from 'element-plus'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import FileConfigDialog from './modules/file-config-dialog.vue'

  defineOptions({ name: 'InfraFileConfig' })

  type FileConfigItem = Api.Infra.FileConfigItem
  type DialogType = 'add' | 'edit'

  // 存储类型展示文案
  const STORAGE_LABELS: Record<string, string> = {
    local: '本地磁盘',
    s3: 'S3 / MinIO'
  }

  // 把存储配置 JSON 转成表格可读摘要
  const formatStorageConfig = (row: FileConfigItem): string => {
    try {
      const conf = JSON.parse(row.config || '{}') as Record<string, any>
      if (row.storage === 's3') {
        const endpoint = conf.endpoint || '-'
        return conf.bucket ? `${conf.bucket} @ ${endpoint}` : endpoint
      }

      return conf.dir ? `目录：${conf.dir}` : '-'
    } catch {
      return row.config || '-'
    }
  }

  const { hasAuth } = useAuth()

  const searchForm = ref<Api.Infra.FileConfigSearchParams>({
    name: undefined,
    storage: undefined
  })

  const searchItems = [
    {
      label: '配置名称',
      key: 'name',
      type: 'input',
      placeholder: '请输入配置名称',
      clearable: true
    },
    {
      label: '存储类型',
      key: 'storage',
      type: 'select',
      props: {
        placeholder: '全部',
        clearable: true,
        options: [
          { label: '本地磁盘', value: 'local' },
          { label: 'S3 / MinIO', value: 's3' }
        ]
      }
    }
  ]

  const dialogVisible = ref(false)
  const dialogType = ref<DialogType>('add')
  const currentConfig = ref<FileConfigItem | null>(null)

  // 文件配置为低频数据，后端一次性返回全量；筛选与分页字段在前端处理
  const fetchFileConfigPage = async (
    params: Api.Infra.FileConfigSearchParams
  ): Promise<Api.Common.PaginatedResponse<FileConfigItem>> => {
    const res = await fetchGetFileConfigList()
    const keyword = (params.name || '').trim().toLowerCase()
    const storage = params.storage || ''

    const records = (res.list ?? []).filter((row) => {
      const nameMatched = !keyword || (row.name || '').toLowerCase().includes(keyword)
      const storageMatched = !storage || row.storage === storage
      return nameMatched && storageMatched
    })

    return {
      records,
      current: params.pageNo ?? 1,
      size: params.pageSize ?? 20,
      total: records.length
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
      apiFn: fetchFileConfigPage,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        { prop: 'name', label: '配置名称', minWidth: 160, showOverflowTooltip: true },
        {
          prop: 'storage',
          label: '存储类型',
          width: 130,
          formatter: (row: FileConfigItem) =>
            h(
              ElTag,
              { type: row.storage === 's3' ? 'warning' : 'info' },
              () => STORAGE_LABELS[row.storage] || row.storage || '-'
            )
        },
        {
          prop: 'master',
          label: '主配置',
          width: 100,
          formatter: (row: FileConfigItem) =>
            h(ElTag, { type: row.master === 1 ? 'success' : 'info' }, () =>
              row.master === 1 ? '是' : '否'
            )
        },
        {
          prop: 'config',
          label: '存储配置',
          minWidth: 240,
          showOverflowTooltip: true,
          formatter: (row: FileConfigItem) => formatStorageConfig(row)
        },
        { prop: 'createdAt', label: '创建时间', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          width: 260,
          fixed: 'right',
          formatter: (row: FileConfigItem) =>
            h('div', { class: 'flex items-center' }, [
              ...(hasAuth('infra:file-config:update') && row.master !== 1
                ? [
                    h(ArtButtonTable, {
                      type: 'action',
                      icon: 'ri:star-line',
                      text: '主配置',
                      onClick: () => handleSetMaster(row)
                    })
                  ]
                : []),
              ...(hasAuth('infra:file-config:update')
                ? [
                    h(ArtButtonTable, {
                      type: 'edit',
                      text: '编辑',
                      onClick: () => showDialog('edit', row)
                    })
                  ]
                : []),
              ...(hasAuth('infra:file-config:delete')
                ? [
                    h(ArtButtonTable, {
                      type: 'delete',
                      text: '删除',
                      disabled: row.master === 1,
                      title: row.master === 1 ? '主配置不允许删除' : '删除',
                      onClick: () => handleDelete(row)
                    })
                  ]
                : [])
            ])
        }
      ]
    }
  })

  const handleSearch = (params: Api.Infra.FileConfigSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }

  const showDialog = (type: DialogType, row?: FileConfigItem): void => {
    dialogType.value = type
    currentConfig.value = type === 'edit' ? (row ?? null) : null
    dialogVisible.value = true
  }

  const handleDialogSuccess = (): void => {
    if (dialogType.value === 'edit') {
      refreshUpdate()
    } else {
      refreshCreate()
    }
  }

  const handleSetMaster = async (row: FileConfigItem): Promise<void> => {
    if (row.master === 1) return

    try {
      await ElMessageBox.confirm(
        `确定把配置「${row.name}」设为主配置吗？新上传的文件会写入该配置的目录。`,
        '设为主配置',
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
      await fetchSetFileConfigMaster(row.id)
      ElMessage.success('已设为主配置')
      refreshUpdate()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }

  const handleDelete = async (row: FileConfigItem): Promise<void> => {
    if (row.master === 1) return

    try {
      await ElMessageBox.confirm(`确定要删除配置「${row.name}」吗？`, '删除文件配置', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch {
      return
    }

    try {
      await fetchDeleteFileConfig(row.id)
      ElMessage.success('删除成功')
      refreshRemove()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }
</script>
