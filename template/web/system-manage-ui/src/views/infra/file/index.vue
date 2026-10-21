<!-- 文件管理页面：上传记录、预览与删除 -->
<template>
  <div class="file-page art-full-height">
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
            <ElUpload
              :show-file-list="false"
              :http-request="handleUpload"
              accept=".jpg,.jpeg,.png,.gif,.webp,.bmp"
            >
              <ElButton type="primary" :loading="uploading" v-ripple>上传文件</ElButton>
            </ElUpload>
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
  </div>
</template>

<script setup lang="ts">
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import { fetchDeleteFile, fetchGetFileList, fetchUploadFile } from '@/api/infra/file'
  import type { UploadRequestOptions } from 'element-plus'
  import { ElButton, ElCard, ElMessage, ElMessageBox, ElSpace, ElUpload } from 'element-plus'
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import ArtSvgIcon from '@/components/core/base/art-svg-icon/index.vue'

  defineOptions({ name: 'InfraFile' })

  type FileItem = Api.Infra.FileItem

  const { hasAuth } = useAuth()

  const searchForm = ref<Api.Infra.FileSearchParams>({
    keyword: undefined
  })

  const searchItems = [
    {
      label: '文件名',
      key: 'keyword',
      type: 'input',
      placeholder: '请输入文件名',
      clearable: true
    }
  ]

  // 上传中状态：上传接口不要求权限点，任何登录用户都能传头像
  const uploading = ref(false)

  const isImage = (row: FileItem): boolean => (row.type || '').startsWith('image/')

  const formatSize = (size: number): string => {
    if (!size) return '0 B'

    const units = ['B', 'KB', 'MB', 'GB']
    let value = size
    let index = 0
    while (value >= 1024 && index < units.length - 1) {
      value /= 1024
      index += 1
    }

    return `${value.toFixed(index === 0 ? 0 : 1)} ${units[index]}`
  }

  const copyUrl = async (row: FileItem): Promise<void> => {
    const url = row.url.startsWith('http') ? row.url : `${window.location.origin}${row.url}`

    try {
      await navigator.clipboard.writeText(url)
      ElMessage.success('链接已复制')
    } catch {
      ElMessage.warning(`复制失败，请手动复制：${url}`)
    }
  }

  const handleUpload = async (options: UploadRequestOptions): Promise<void> => {
    uploading.value = true
    try {
      await fetchUploadFile(options.file)
      ElMessage.success('上传成功')
      refreshCreate()
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      uploading.value = false
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
    refreshRemove
  } = useTable({
    core: {
      apiFn: fetchGetFileList,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        {
          label: '预览',
          width: 90,
          formatter: (row: FileItem) =>
            isImage(row)
              ? h('img', {
                  src: row.url,
                  alt: row.name,
                  class: 'w-10 h-10 rounded object-cover border border-g-300'
                })
              : h(ArtSvgIcon, { icon: 'ri:file-3-line', class: 'text-2xl text-g-600' })
        },
        { prop: 'name', label: '文件名', minWidth: 220, showOverflowTooltip: true },
        {
          prop: 'type',
          label: '类型',
          width: 140,
          formatter: (row: FileItem) => row.type || '-'
        },
        {
          prop: 'size',
          label: '大小',
          width: 110,
          formatter: (row: FileItem) => formatSize(row.size)
        },
        { prop: 'createdAt', label: '上传时间', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          width: 220,
          fixed: 'right',
          formatter: (row: FileItem) =>
            h('div', { class: 'flex items-center' }, [
              h(ArtButtonTable, {
                type: 'action',
                icon: 'ri:file-copy-line',
                text: '复制链接',
                onClick: () => copyUrl(row)
              }),
              ...(hasAuth('infra:file:delete')
                ? [
                    h(ArtButtonTable, {
                      type: 'delete',
                      text: '删除',
                      onClick: () => handleDelete(row)
                    })
                  ]
                : [])
            ])
        }
      ]
    }
  })

  const handleSearch = (params: Api.Infra.FileSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }

  const handleDelete = async (row: FileItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(`确定要删除文件「${row.name}」吗？`, '删除文件', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch {
      return
    }

    try {
      await fetchDeleteFile(row.id)
      ElMessage.success('删除成功')
      refreshRemove()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }
</script>
