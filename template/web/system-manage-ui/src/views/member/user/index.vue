<!-- 会员列表页面（表结构对齐 yudao member_user） -->
<template>
  <div class="member-user-page art-full-height">
    <!-- 搜索栏 -->
    <MemberSearch v-model="searchForm" @search="handleSearch" @reset="resetSearchParams" />

    <ElCard class="art-table-card">
      <!-- 表格头部 -->
      <ArtTableHeader v-model:columns="columnChecks" :loading="loading" @refresh="refreshData">
        <template #left>
          <ElSpace wrap>
            <ElButton v-auth="'member:user:create'" @click="showDialog('add')" v-ripple>
              新增会员
            </ElButton>
          </ElSpace>
        </template>
      </ArtTableHeader>

      <!-- 表格 -->
      <ArtTable
        :loading="loading"
        :data="data"
        :columns="columns"
        :pagination="pagination"
        @pagination:size-change="handleSizeChange"
        @pagination:current-change="handleCurrentChange"
      >
      </ArtTable>

      <!-- 新增 / 编辑弹窗 -->
      <MemberDialog
        v-model:visible="dialogVisible"
        :type="dialogType"
        :member-data="currentMember"
        @success="handleDialogSuccess"
      />

      <!-- 详情抽屉 -->
      <MemberDetail v-model:visible="detailVisible" :member-id="detailId" @updated="refreshData" />
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import ArtButtonTable from '@/components/core/forms/art-button-table/index.vue'
  import { useTable } from '@/hooks/core/useTable'
  import { useAuth } from '@/hooks/core/useAuth'
  import { fetchDeleteMember, fetchGetMemberList, fetchUpdateMemberStatus } from '@/api/member/user'
  import { fetchGetLevelOptions } from '@/api/member/level'
  import { fetchGetTagOptions } from '@/api/member/tag'
  import { fetchGetGroupOptions } from '@/api/member/group'
  import defaultAvatar from '@imgs/user/avatar.webp'
  import {
    ElAvatar,
    ElButton,
    ElMessage,
    ElMessageBox,
    ElSpace,
    ElSwitch,
    ElTag
  } from 'element-plus'
  import MemberSearch from './modules/member-search.vue'
  import MemberDialog from './modules/member-dialog.vue'
  import MemberDetail from './modules/member-detail.vue'

  defineOptions({ name: 'MemberUser' })

  type MemberItem = Api.Member.MemberItem
  type DialogType = 'add' | 'edit'

  const { hasAuth } = useAuth()

  // 性别展示配置，取值与 member_user 表一致
  const SEX_TEXT: Record<number, string> = { 0: '未知', 1: '男', 2: '女' }

  // 等级 / 标签 / 分组映射，列表只返回 ID，这里换成名称展示
  const levelMap = ref<Record<string, string>>({})
  const tagMap = ref<Record<string, string>>({})
  const groupMap = ref<Record<string, string>>({})

  const loadOptions = async (): Promise<void> => {
    try {
      const [levels, tags, groups] = await Promise.all([
        fetchGetLevelOptions(),
        fetchGetTagOptions(),
        fetchGetGroupOptions()
      ])

      levelMap.value = Object.fromEntries(levels.map((item) => [item.id, item.name]))
      tagMap.value = Object.fromEntries(tags.map((item) => [item.id, item.name]))
      groupMap.value = Object.fromEntries(groups.map((item) => [item.id, item.name]))
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }

  onMounted(loadOptions)

  // 搜索表单（key 与后端查询参数一致）
  const searchForm = ref<Api.Member.MemberSearchParams>({
    nickname: undefined,
    mobile: undefined,
    status: undefined
  })

  // 弹窗相关
  const dialogVisible = ref(false)
  const dialogType = ref<DialogType>('add')
  const currentMember = ref<MemberItem | null>(null)

  // 详情抽屉
  const detailVisible = ref(false)
  const detailId = ref('')

  const tagNames = (row: MemberItem): string => {
    if (!row.tagIds) return '-'
    return row.tagIds
      .split(',')
      .map((id) => tagMap.value[id])
      .filter(Boolean)
      .join('、')
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
      apiFn: fetchGetMemberList,
      apiParams: {
        pageNo: 1,
        pageSize: 20,
        ...searchForm.value
      },
      columnsFactory: () => [
        {
          prop: 'avatar',
          label: '头像',
          width: 80,
          align: 'center',
          formatter: (row: MemberItem) =>
            h(ElAvatar, {
              size: 36,
              shape: 'circle',
              fit: 'cover',
              src: row.avatar || defaultAvatar
            })
        },
        {
          prop: 'nickname',
          label: '昵称',
          minWidth: 140,
          formatter: (row: MemberItem) =>
            h('div', [
              h('p', row.nickname || '-'),
              row.name
                ? h(
                    'p',
                    { style: 'color: var(--el-text-color-secondary); font-size: 12px' },
                    row.name
                  )
                : null
            ])
        },
        { prop: 'mobile', label: '手机号', width: 130 },
        {
          prop: 'sex',
          label: '性别',
          width: 70,
          formatter: (row: MemberItem) => SEX_TEXT[row.sex] ?? '未知'
        },
        {
          prop: 'levelId',
          label: '等级',
          width: 110,
          formatter: (row: MemberItem) =>
            levelMap.value[row.levelId]
              ? h(ElTag, { type: 'warning', effect: 'plain' }, () => levelMap.value[row.levelId])
              : '-'
        },
        {
          prop: 'point',
          label: '积分',
          width: 90,
          formatter: (row: MemberItem) => row.point ?? 0
        },
        {
          prop: 'tagIds',
          label: '标签',
          minWidth: 140,
          showOverflowTooltip: true,
          formatter: (row: MemberItem) => tagNames(row)
        },
        {
          prop: 'groupId',
          label: '分组',
          width: 110,
          formatter: (row: MemberItem) => groupMap.value[row.groupId] || '-'
        },
        {
          prop: 'status',
          label: '状态',
          width: 90,
          formatter: (row: MemberItem) =>
            h(ElSwitch, {
              modelValue: row.status,
              activeValue: 1,
              inactiveValue: 0,
              disabled: !hasAuth('member:user:update'),
              'onUpdate:modelValue': (value: string | number | boolean) =>
                handleStatusChange(row, Number(value))
            })
        },
        { prop: 'createdAt', label: '注册时间', width: 170 },
        {
          prop: 'operation',
          label: '操作',
          // 详情 / 编辑 / 删除 三个带图标文字的按钮，220 会把「删除」挤掉
          width: 280,
          fixed: 'right',
          formatter: (row: MemberItem) =>
            h('div', { class: 'flex items-center' }, [
              h(ArtButtonTable, {
                type: 'view',
                text: '详情',
                onClick: () => showDetail(row)
              }),
              ...(hasAuth('member:user:update')
                ? [
                    h(ArtButtonTable, {
                      type: 'edit',
                      text: '编辑',
                      onClick: () => showDialog('edit', row)
                    })
                  ]
                : []),
              ...(hasAuth('member:user:delete')
                ? [
                    h(ArtButtonTable, {
                      type: 'delete',
                      text: '删除',
                      onClick: () => deleteMember(row)
                    })
                  ]
                : [])
            ])
        }
      ]
    }
  })

  /**
   * 搜索处理
   */
  const handleSearch = (params: Api.Member.MemberSearchParams) => {
    Object.assign(searchParams, params)
    getData()
  }

  /**
   * 显示新增 / 编辑弹窗
   */
  const showDialog = (type: DialogType, row?: MemberItem): void => {
    dialogType.value = type
    currentMember.value = type === 'edit' ? (row ?? null) : null
    nextTick(() => {
      dialogVisible.value = true
    })
  }

  /**
   * 打开详情抽屉
   */
  const showDetail = (row: MemberItem): void => {
    detailId.value = row.id
    detailVisible.value = true
  }

  /**
   * 弹窗提交成功后刷新列表
   */
  const handleDialogSuccess = (): void => {
    if (dialogType.value === 'add') {
      refreshCreate()
      return
    }
    refreshUpdate()
  }

  /**
   * 启用 / 禁用会员：先乐观更新开关，失败回滚
   */
  const handleStatusChange = async (row: MemberItem, value: number): Promise<void> => {
    const previous = row.status
    row.status = value

    try {
      await fetchUpdateMemberStatus(row.id, value)
      ElMessage.success(value === 1 ? '已启用' : '已禁用')
    } catch {
      row.status = previous
    }
  }

  /**
   * 删除会员（后端为软删除）
   */
  const deleteMember = async (row: MemberItem): Promise<void> => {
    try {
      await ElMessageBox.confirm(`确定要删除会员「${row.nickname}」吗？`, '删除会员', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      })
    } catch {
      return
    }

    try {
      await fetchDeleteMember(row.id)
      ElMessage.success('删除成功')
      refreshRemove()
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }
</script>
