<!-- 充值套餐页面：维护钱包充值档位与赠送金额，到账金额为实付加赠送 -->
<template>
  <div class="pay-wallet-package-page art-full-height">
    <ElCard class="art-table-card">
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElSpace wrap>
            <ElButton v-auth="'pay:wallet-package:create'" @click="showDialog('add')" v-ripple>
              新增充值套餐
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

      <PackageDialog
        v-model:visible="dialogVisible"
        :type="dialogType"
        :package-data="currentPackage"
        @success="handleDialogSuccess"
      />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import { fetchDeletePayWalletPackage, fetchGetPayWalletPackageList } from '@/api/pay/wallet'
  import { ElButton, ElMessage, ElMessageBox, ElSpace, ElTag } from 'element-plus'
  import PackageDialog from './modules/package-dialog.vue'

  defineOptions({ name: 'PayWalletPackage' })

  type PackageItem = Api.Pay.WalletPackageItem
  type DialogType = 'add' | 'edit'

  const { hasAuth } = useAuth()

  const statusTagTypes: Record<number, 'success' | 'info'> = {
    0: 'success',
    1: 'info'
  }

  const dialogVisible = ref(false)
  const dialogType = ref<DialogType>('add')
  const currentPackage = ref<PackageItem | null>(null)

  // 分转元展示，保留两位小数
  const formatYuan = (fen: number): string => ((fen || 0) / 100).toFixed(2)

  const {
    columns,
    columnChecks,
    data,
    loading,
    pagination,
    handleSizeChange,
    handleCurrentChange,
    refreshData,
    refreshCreate,
    refreshUpdate,
    refreshRemove
  } = useTable({
    core: {
      apiFn: fetchGetPayWalletPackageList,
      apiParams: { pageNo: 1, pageSize: 20 },
      columnsFactory: () => [
        { prop: 'name', label: '套餐名', minWidth: 160 },
        {
          prop: 'payPrice',
          label: '实付（元）',
          width: 120,
          formatter: (row: PackageItem) => formatYuan(row.payPrice)
        },
        {
          prop: 'bonusPrice',
          label: '赠送（元）',
          width: 120,
          formatter: (row: PackageItem) => formatYuan(row.bonusPrice)
        },
        {
          prop: 'totalPrice',
          label: '到账（元）',
          width: 120,
          formatter: (row: PackageItem) => formatYuan(row.payPrice + row.bonusPrice)
        },
        {
          prop: 'status',
          label: '状态',
          width: 90,
          formatter: (row: PackageItem) =>
            h(ElTag, { type: statusTagTypes[row.status] ?? 'info', effect: 'plain' }, () =>
              row.status === 0 ? '开启' : '关闭'
            )
        },
        {
          prop: 'description',
          label: '描述',
          minWidth: 200,
          showOverflowTooltip: true
        },
        { prop: 'createdAt', label: '创建时间', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          width: 170,
          fixed: 'right',
          formatter: (row: PackageItem) =>
            h('div', { class: 'flex items-center' }, [
              ...(hasAuth('pay:wallet-package:update')
                ? [
                    h(ArtButtonTable, {
                      type: 'edit',
                      text: '编辑',
                      onClick: () => showDialog('edit', row)
                    })
                  ]
                : []),
              ...(hasAuth('pay:wallet-package:delete')
                ? [
                    h(ArtButtonTable, {
                      type: 'delete',
                      text: '删除',
                      onClick: () => deletePackage(row)
                    })
                  ]
                : [])
            ])
        }
      ]
    }
  })

  // 打开新增 / 编辑弹窗，编辑时携带当前行用于回填
  const showDialog = (type: DialogType, row?: PackageItem): void => {
    dialogType.value = type
    currentPackage.value = type === 'edit' ? (row ?? null) : null
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

  // 删除充值套餐，确认后调用接口并刷新列表
  const deletePackage = async (row: PackageItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(
        `确定要删除充值套餐「${row.name}」吗？已产生的充值记录会保留。`,
        '删除充值套餐',
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
      await fetchDeletePayWalletPackage(row.id)
      ElMessage.success('删除成功')
      refreshRemove()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }
</script>
