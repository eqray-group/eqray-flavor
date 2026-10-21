<!-- 会员详情抽屉：资料 / 积分记录 / 等级记录 / 签到记录 / 收货地址 / 第三方绑定 -->
<template>
  <ElDrawer v-model="drawerVisible" title="会员详情" size="880px" :close-on-click-modal="true">
    <div v-loading="loading" class="member-detail">
      <div class="flex-c flex-col pb-4 mb-2 border-b border-g-300/80">
        <ElAvatar :size="72" :src="member?.avatar || defaultAvatar" />
        <p class="mt-3 mb-0 text-base font-medium">{{ member?.nickname || '-' }}</p>
        <p class="mt-1 mb-0 text-sm text-g-600">{{ member?.mobile || '-' }}</p>
      </div>

      <ElTabs v-model="activeTab">
        <ElTabPane label="基本信息" name="basic">
          <ElDescriptions v-if="member" :column="2" border>
            <ElDescriptionsItem label="会员 ID">{{ member.id }}</ElDescriptionsItem>
            <ElDescriptionsItem label="真实姓名">{{ member.name || '-' }}</ElDescriptionsItem>
            <ElDescriptionsItem label="性别">{{
              SEX_TEXT[member.sex] ?? '未知'
            }}</ElDescriptionsItem>
            <ElDescriptionsItem label="生日">
              {{ member.birthday ? member.birthday.slice(0, 10) : '-' }}
            </ElDescriptionsItem>
            <ElDescriptionsItem label="等级">{{ levelName }}</ElDescriptionsItem>
            <ElDescriptionsItem label="分组">{{ groupName }}</ElDescriptionsItem>
            <ElDescriptionsItem label="积分">{{ member.point }}</ElDescriptionsItem>
            <ElDescriptionsItem label="经验值">{{ member.experience }}</ElDescriptionsItem>
            <ElDescriptionsItem label="状态">
              <ElTag :type="member.status === 1 ? 'success' : 'info'">
                {{ member.status === 1 ? '启用' : '禁用' }}
              </ElTag>
            </ElDescriptionsItem>
            <ElDescriptionsItem label="用户标记">{{ member.mark || '-' }}</ElDescriptionsItem>
            <ElDescriptionsItem label="标签" :span="2">{{ tagNames }}</ElDescriptionsItem>
            <ElDescriptionsItem label="注册 IP">{{ member.registerIp || '-' }}</ElDescriptionsItem>
            <ElDescriptionsItem label="最后登录">{{ member.loginDate || '-' }}</ElDescriptionsItem>
            <ElDescriptionsItem label="注册时间">{{ member.createdAt || '-' }}</ElDescriptionsItem>
            <ElDescriptionsItem label="更新时间">{{ member.updatedAt || '-' }}</ElDescriptionsItem>
          </ElDescriptions>
        </ElTabPane>

        <ElTabPane label="积分记录" name="points" lazy>
          <div class="flex items-center justify-between mb-3">
            <span class="text-sm text-g-600">当前积分：{{ member?.point ?? 0 }}</span>
            <ElButton v-auth="'member:user:points'" type="primary" size="small" @click="openPoint">
              调整积分
            </ElButton>
          </div>

          <ArtTable
            :loading="pointLoading"
            :data="pointRecords"
            :columns="pointColumns"
            :pagination="pointPagination"
            :show-table-header="false"
            height="420"
            empty-height="420"
            @pagination:current-change="handlePointPageChange"
          />
        </ElTabPane>

        <ElTabPane label="等级记录" name="levels" lazy>
          <div class="flex justify-end mb-3">
            <ElButton v-auth="'member:user:level'" type="primary" size="small" @click="openLevel">
              调整等级
            </ElButton>
          </div>

          <ArtTable
            :loading="levelLoading"
            :data="levelRecords"
            :columns="levelColumns"
            :pagination="levelPagination"
            :show-table-header="false"
            height="420"
            empty-height="420"
            @pagination:current-change="handleLevelPageChange"
          />
        </ElTabPane>

        <ElTabPane label="签到记录" name="signs" lazy>
          <ArtTable
            :loading="signLoading"
            :data="signRecords"
            :columns="signColumns"
            :pagination="signPagination"
            :show-table-header="false"
            height="420"
            empty-height="420"
            @pagination:current-change="handleSignPageChange"
          />
        </ElTabPane>

        <ElTabPane label="收货地址" name="addresses" lazy>
          <ArtTable
            :loading="addressLoading"
            :data="addresses"
            :columns="addressColumns"
            :show-table-header="false"
            height="420"
            empty-height="420"
          />
        </ElTabPane>

        <ElTabPane label="第三方绑定" name="socials" lazy>
          <ArtTable
            :loading="socialLoading"
            :data="socialBindings"
            :columns="socialColumns"
            :show-table-header="false"
            height="420"
            empty-height="420"
          />
        </ElTabPane>
      </ElTabs>
    </div>

    <!-- 调整积分 -->
    <ElDialog v-model="pointVisible" title="调整积分" width="420px" append-to-body>
      <ElForm ref="pointFormRef" :model="pointForm" :rules="pointRules" label-width="90px">
        <ElFormItem label="积分变动" prop="point">
          <ElInputNumber
            v-model="pointForm.point"
            :min="-1000000"
            :max="1000000"
            controls-position="right"
          />
          <span class="ml-3 text-xs text-g-600">正数增加，负数扣减</span>
        </ElFormItem>
        <ElFormItem label="标题" prop="title">
          <ElInput v-model="pointForm.title" maxlength="64" placeholder="默认「后台调整」" />
        </ElFormItem>
        <ElFormItem label="描述" prop="description">
          <ElInput
            v-model="pointForm.description"
            type="textarea"
            :rows="2"
            maxlength="255"
            show-word-limit
            placeholder="选填"
          />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="pointVisible = false">取消</ElButton>
        <ElButton type="primary" :loading="pointSubmitting" @click="submitPoint">确定</ElButton>
      </template>
    </ElDialog>

    <!-- 调整等级 -->
    <ElDialog v-model="levelVisible" title="调整等级" width="440px" append-to-body>
      <ElForm ref="levelFormRef" :model="levelForm" :rules="levelRules" label-width="90px">
        <ElFormItem label="会员等级" prop="levelId">
          <ElSelect v-model="levelForm.levelId" placeholder="请选择等级">
            <ElOption
              v-for="item in levelOptions"
              :key="item.id"
              :label="`Lv.${item.level} ${item.name}`"
              :value="item.id"
            />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="经验值" prop="experience">
          <ElInputNumber
            v-model="levelForm.experience"
            :min="0"
            :max="99999999"
            controls-position="right"
          />
        </ElFormItem>
        <ElFormItem label="标题" prop="title">
          <ElInput v-model="levelForm.title" maxlength="64" placeholder="默认「后台调整等级」" />
        </ElFormItem>
        <ElFormItem label="描述" prop="description">
          <ElInput
            v-model="levelForm.description"
            type="textarea"
            :rows="2"
            maxlength="255"
            show-word-limit
            placeholder="选填"
          />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="levelVisible = false">取消</ElButton>
        <ElButton type="primary" :loading="levelSubmitting" @click="submitLevel">确定</ElButton>
      </template>
    </ElDialog>
  </ElDrawer>
</template>

<script setup lang="ts">
  import { ElMessage, ElTag } from 'element-plus'
  import type { FormInstance, FormRules } from 'element-plus'
  import defaultAvatar from '@imgs/user/avatar.webp'
  import {
    fetchAdjustMemberLevel,
    fetchAdjustMemberPoint,
    fetchGetMemberAddresses,
    fetchGetMemberDetail,
    fetchGetMemberLevelRecords,
    fetchGetMemberPointRecords,
    fetchGetMemberSignRecords,
    fetchGetMemberSocialBindings
  } from '@/api/member/user'
  import { fetchGetLevelOptions } from '@/api/member/level'
  import { fetchGetTagOptions } from '@/api/member/tag'
  import { fetchGetGroupOptions } from '@/api/member/group'

  defineOptions({ name: 'MemberDetail' })

  interface Props {
    visible: boolean
    memberId?: string
  }

  const props = withDefaults(defineProps<Props>(), {
    memberId: ''
  })

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void
    /** 资料 / 积分 / 等级变更后通知父页面刷新列表 */
    (e: 'updated'): void
  }>()

  const SEX_TEXT: Record<number, string> = { 0: '未知', 1: '男', 2: '女' }

  const BIZ_TYPE_TEXT: Record<number, string> = {
    0: '其它',
    10: '后台调整',
    20: '签到',
    30: '消费'
  }

  const drawerVisible = computed({
    get: () => props.visible,
    set: (value: boolean) => emit('update:visible', value)
  })

  const activeTab = ref('basic')
  const loading = ref(false)
  const member = ref<Api.Member.MemberItem | null>(null)

  const levelOptions = ref<Api.Member.LevelItem[]>([])
  const tagMap = ref<Record<string, string>>({})
  const groupMap = ref<Record<string, string>>({})

  const levelName = computed(() => {
    const level = levelOptions.value.find((item) => item.id === member.value?.levelId)
    return level ? `Lv.${level.level} ${level.name}` : '-'
  })

  const groupName = computed(() => groupMap.value[member.value?.groupId ?? ''] || '-')

  const tagNames = computed(() => {
    const ids = member.value?.tagIds
    if (!ids) return '-'
    return (
      ids
        .split(',')
        .map((id) => tagMap.value[id])
        .filter(Boolean)
        .join('、') || '-'
    )
  })

  const loadOptions = async (): Promise<void> => {
    try {
      const [levels, tags, groups] = await Promise.all([
        fetchGetLevelOptions(),
        fetchGetTagOptions(),
        fetchGetGroupOptions()
      ])
      levelOptions.value = levels
      tagMap.value = Object.fromEntries(tags.map((item) => [item.id, item.name]))
      groupMap.value = Object.fromEntries(groups.map((item) => [item.id, item.name]))
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }

  const loadDetail = async (): Promise<void> => {
    if (!props.memberId) return

    loading.value = true
    try {
      member.value = await fetchGetMemberDetail(props.memberId)
    } catch {
      member.value = null
    } finally {
      loading.value = false
    }
  }

  watch(
    () => [props.visible, props.memberId] as const,
    ([visible]) => {
      if (!visible) return

      activeTab.value = 'basic'
      pointPage.value = 1
      levelPage.value = 1
      signPage.value = 1
      loadOptions()
      loadDetail()
    },
    { immediate: true }
  )

  watch(activeTab, (tab) => {
    if (tab === 'points') loadPointRecords()
    if (tab === 'levels') loadLevelRecords()
    if (tab === 'signs') loadSignRecords()
    if (tab === 'addresses') loadAddresses()
    if (tab === 'socials') loadSocialBindings()
  })

  // ---------- 积分记录 ----------
  const pointRecords = ref<Api.Member.PointRecordItem[]>([])
  const pointLoading = ref(false)
  const pointPage = ref(1)
  const pointTotal = ref(0)

  const pointPagination = computed(() => ({
    current: pointPage.value,
    size: 10,
    total: pointTotal.value
  }))

  const pointColumns = [
    { prop: 'createdAt', label: '时间', width: 160 },
    {
      prop: 'bizType',
      label: '类型',
      width: 100,
      formatter: (row: Api.Member.PointRecordItem) => BIZ_TYPE_TEXT[row.bizType] ?? '其它'
    },
    {
      prop: 'point',
      label: '变动',
      width: 90,
      formatter: (row: Api.Member.PointRecordItem) =>
        h(
          'span',
          {
            style:
              row.point >= 0 ? 'color: var(--el-color-success)' : 'color: var(--el-color-danger)'
          },
          row.point >= 0 ? `+${row.point}` : `${row.point}`
        )
    },
    { prop: 'totalPoint', label: '余额', width: 80 },
    {
      prop: 'title',
      label: '标题',
      minWidth: 110,
      showOverflowTooltip: true
    },
    {
      prop: 'description',
      label: '描述',
      minWidth: 120,
      showOverflowTooltip: true,
      formatter: (row: Api.Member.PointRecordItem) => row.description || '-'
    }
  ]

  const loadPointRecords = async (): Promise<void> => {
    if (!props.memberId) return

    pointLoading.value = true
    try {
      const res = await fetchGetMemberPointRecords(props.memberId, {
        pageNo: pointPage.value,
        pageSize: 10
      })
      pointRecords.value = res.list
      pointTotal.value = res.total
    } catch {
      pointRecords.value = []
      pointTotal.value = 0
    } finally {
      pointLoading.value = false
    }
  }

  const handlePointPageChange = (page: number): void => {
    pointPage.value = page
    loadPointRecords()
  }

  // ---------- 等级记录 ----------
  const levelRecords = ref<Api.Member.LevelRecordItem[]>([])
  const levelLoading = ref(false)
  const levelPage = ref(1)
  const levelTotal = ref(0)

  const levelPagination = computed(() => ({
    current: levelPage.value,
    size: 10,
    total: levelTotal.value
  }))

  const levelColumns = [
    { prop: 'createdAt', label: '时间', width: 160 },
    {
      prop: 'levelId',
      label: '等级',
      width: 120,
      formatter: (row: Api.Member.LevelRecordItem) => {
        const level = levelOptions.value.find((item) => item.id === row.levelId)
        return level ? `Lv.${level.level} ${level.name}` : '-'
      }
    },
    { prop: 'experience', label: '经验值', width: 90 },
    { prop: 'title', label: '标题', minWidth: 120, showOverflowTooltip: true },
    {
      prop: 'description',
      label: '描述',
      minWidth: 120,
      showOverflowTooltip: true,
      formatter: (row: Api.Member.LevelRecordItem) => row.description || '-'
    }
  ]

  const loadLevelRecords = async (): Promise<void> => {
    if (!props.memberId) return

    levelLoading.value = true
    try {
      const res = await fetchGetMemberLevelRecords(props.memberId, {
        pageNo: levelPage.value,
        pageSize: 10
      })
      levelRecords.value = res.list
      levelTotal.value = res.total
    } catch {
      levelRecords.value = []
      levelTotal.value = 0
    } finally {
      levelLoading.value = false
    }
  }

  const handleLevelPageChange = (page: number): void => {
    levelPage.value = page
    loadLevelRecords()
  }

  // ---------- 签到记录 ----------
  const signRecords = ref<Api.Member.SignRecordItem[]>([])
  const signLoading = ref(false)
  const signPage = ref(1)
  const signTotal = ref(0)

  const signPagination = computed(() => ({
    current: signPage.value,
    size: 10,
    total: signTotal.value
  }))

  const signColumns = [
    {
      prop: 'signDate',
      label: '签到日期',
      width: 140,
      formatter: (row: Api.Member.SignRecordItem) => (row.signDate || '').slice(0, 10)
    },
    { prop: 'point', label: '获得积分', width: 100 },
    { prop: 'createdAt', label: '创建时间', minWidth: 160 }
  ]

  const loadSignRecords = async (): Promise<void> => {
    if (!props.memberId) return

    signLoading.value = true
    try {
      const res = await fetchGetMemberSignRecords(props.memberId, {
        pageNo: signPage.value,
        pageSize: 10
      })
      signRecords.value = res.list
      signTotal.value = res.total
    } catch {
      signRecords.value = []
      signTotal.value = 0
    } finally {
      signLoading.value = false
    }
  }

  const handleSignPageChange = (page: number): void => {
    signPage.value = page
    loadSignRecords()
  }

  // ---------- 收货地址 ----------
  const addresses = ref<Api.Member.AddressItem[]>([])
  const addressLoading = ref(false)

  const addressColumns = [
    { prop: 'name', label: '收件人', width: 100 },
    { prop: 'mobile', label: '手机号', width: 130 },
    { prop: 'detailAddress', label: '详细地址', minWidth: 200, showOverflowTooltip: true },
    {
      prop: 'defaultStatus',
      label: '默认',
      width: 80,
      formatter: (row: Api.Member.AddressItem) =>
        h(ElTag, { type: row.defaultStatus === 1 ? 'success' : 'info', effect: 'plain' }, () =>
          row.defaultStatus === 1 ? '是' : '否'
        )
    }
  ]

  const loadAddresses = async (): Promise<void> => {
    if (!props.memberId) return

    addressLoading.value = true
    try {
      addresses.value = await fetchGetMemberAddresses(props.memberId)
    } catch {
      addresses.value = []
    } finally {
      addressLoading.value = false
    }
  }

  // ---------- 第三方绑定 ----------
  const SOCIAL_TYPE_TEXT: Record<string, string> = {
    wechat_mini: '微信小程序',
    wechat_mp: '微信公众号',
    wechat_open: '微信开放平台'
  }

  const socialBindings = ref<Api.Member.SocialBindingItem[]>([])
  const socialLoading = ref(false)

  const socialColumns = [
    {
      prop: 'socialType',
      label: '渠道',
      width: 130,
      formatter: (row: Api.Member.SocialBindingItem) =>
        h(
          ElTag,
          { type: 'success', effect: 'plain' },
          () => SOCIAL_TYPE_TEXT[row.socialType] ?? row.socialType
        )
    },
    { prop: 'nickname', label: '第三方昵称', minWidth: 140, showOverflowTooltip: true },
    { prop: 'createdAt', label: '绑定时间', width: 160 }
  ]

  const loadSocialBindings = async (): Promise<void> => {
    if (!props.memberId) return

    socialLoading.value = true
    try {
      socialBindings.value = await fetchGetMemberSocialBindings(props.memberId)
    } catch {
      socialBindings.value = []
    } finally {
      socialLoading.value = false
    }
  }

  // ---------- 调整积分 ----------
  const pointVisible = ref(false)
  const pointSubmitting = ref(false)
  const pointFormRef = ref<FormInstance>()
  const pointForm = reactive({
    point: 100,
    title: '',
    description: ''
  })

  const pointRules: FormRules = {
    point: [{ required: true, message: '请输入积分变动数量', trigger: 'blur' }]
  }

  const openPoint = (): void => {
    pointForm.point = 100
    pointForm.title = ''
    pointForm.description = ''
    pointVisible.value = true
    nextTick(() => pointFormRef.value?.clearValidate())
  }

  const submitPoint = async (): Promise<void> => {
    if (!pointFormRef.value || !props.memberId) return

    const valid = await pointFormRef.value.validate().catch(() => false)
    if (!valid) return
    if (pointForm.point === 0) {
      ElMessage.warning('积分变动数量不能为 0')
      return
    }

    pointSubmitting.value = true
    try {
      await fetchAdjustMemberPoint(props.memberId, {
        point: pointForm.point,
        title: pointForm.title.trim(),
        description: pointForm.description.trim()
      })
      ElMessage.success('积分已更新')

      pointVisible.value = false
      await Promise.all([loadDetail(), loadPointRecords()])
      emit('updated')
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      pointSubmitting.value = false
    }
  }

  // ---------- 调整等级 ----------
  const levelVisible = ref(false)
  const levelSubmitting = ref(false)
  const levelFormRef = ref<FormInstance>()
  const levelForm = reactive({
    levelId: '',
    experience: 0,
    title: '',
    description: ''
  })

  const levelRules: FormRules = {
    levelId: [{ required: true, message: '请选择会员等级', trigger: 'change' }]
  }

  const openLevel = (): void => {
    levelForm.levelId = member.value?.levelId ?? ''
    levelForm.experience = member.value?.experience ?? 0
    levelForm.title = ''
    levelForm.description = ''
    levelVisible.value = true
    nextTick(() => levelFormRef.value?.clearValidate())
  }

  const submitLevel = async (): Promise<void> => {
    if (!levelFormRef.value || !props.memberId) return

    const valid = await levelFormRef.value.validate().catch(() => false)
    if (!valid) return

    levelSubmitting.value = true
    try {
      await fetchAdjustMemberLevel(props.memberId, {
        levelId: levelForm.levelId,
        experience: levelForm.experience,
        title: levelForm.title.trim(),
        description: levelForm.description.trim()
      })
      ElMessage.success('等级已更新')

      levelVisible.value = false
      await Promise.all([loadDetail(), loadLevelRecords()])
      emit('updated')
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      levelSubmitting.value = false
    }
  }
</script>
