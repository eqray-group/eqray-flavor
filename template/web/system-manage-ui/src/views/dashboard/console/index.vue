<!-- 控制台：项目概览、访问趋势、最近登录与快捷入口 -->
<template>
  <div v-loading="loading" class="dashboard-console">
    <div
      class="art-card mb-5 flex items-center justify-between p-5 max-sm:mb-4 max-sm:flex-col max-sm:items-start max-sm:gap-4"
    >
      <div class="flex items-center gap-4">
        <ElAvatar :size="52" :src="avatar">{{ displayName.slice(0, 1) }}</ElAvatar>
        <div>
          <h2 class="text-xl font-medium">你好，{{ displayName }}</h2>
          <p class="mt-1 text-sm text-g-600">{{ today }}，{{ greeting }}</p>
        </div>
      </div>

      <ElSpace wrap>
        <ElButton
          v-for="item in quickEntries.slice(0, 4)"
          :key="item.path"
          plain
          @click="go(item.path)"
        >
          <ArtSvgIcon :icon="item.icon" class="mr-1" />{{ item.label }}
        </ElButton>
      </ElSpace>
    </div>

    <ElRow :gutter="20">
      <ElCol v-for="card in cards" :key="card.label" :xs="24" :sm="12" :lg="6">
        <div class="art-card relative mb-5 flex h-35 flex-col justify-center px-5 max-sm:mb-4">
          <span class="text-sm text-g-700">{{ card.label }}</span>
          <ArtCountTo class="mt-2 text-[26px] font-medium" :target="card.value" :duration="1000" />
          <span class="mt-1 text-xs text-g-600">{{ card.tip }}</span>
          <div
            class="absolute top-0 right-5 bottom-0 m-auto flex-cc size-12.5 rounded-xl bg-theme/10"
          >
            <ArtSvgIcon :icon="card.icon" class="text-xl text-theme" />
          </div>
        </div>
      </ElCol>
    </ElRow>

    <ElRow :gutter="20">
      <ElCol :xs="24" :lg="16">
        <div class="art-card mb-5 p-5 max-sm:mb-4">
          <div class="art-card-header">
            <div class="title">
              <h4>访问趋势</h4>
              <p>近 7 天登录次数与 API 请求量</p>
            </div>
          </div>
          <ArtLineChart
            height="320px"
            :data="trendData"
            :xAxisData="trendLabels"
            :showAreaColor="true"
            :showAxisLine="false"
            :showLegend="true"
          />
        </div>
      </ElCol>

      <ElCol :xs="24" :lg="8">
        <div class="art-card mb-5 h-full p-5 max-sm:mb-4">
          <div class="art-card-header">
            <div class="title">
              <h4>最近登录</h4>
              <p>最新的登录记录</p>
            </div>
          </div>

          <div
            v-if="overview.recentLogins.length === 0"
            class="py-12 text-center text-sm text-g-500"
          >
            暂无登录记录
          </div>

          <div
            v-for="(item, index) in overview.recentLogins"
            :key="index"
            class="flex items-center justify-between border-b border-[var(--el-border-color-lighter)] py-3 last:border-0"
          >
            <div class="flex items-center gap-3">
              <div class="flex-cc size-8 rounded-full bg-theme/10 text-xs text-theme">
                {{ item.username.slice(0, 1).toUpperCase() }}
              </div>
              <div>
                <p class="text-sm">{{ item.username || '-' }}</p>
                <p class="text-xs text-g-600">{{ item.userIp || '-' }}</p>
              </div>
            </div>
            <div class="text-right">
              <ElTag size="small" :type="item.result === 1 ? 'success' : 'danger'">
                {{ item.result === 1 ? '成功' : '失败' }}
              </ElTag>
              <p class="mt-1 text-xs text-g-600">{{ item.createdAt }}</p>
            </div>
          </div>
        </div>
      </ElCol>
    </ElRow>

    <ElRow :gutter="20">
      <ElCol :xs="24" :lg="16">
        <div class="art-card mb-5 p-5 max-sm:mb-4">
          <div class="art-card-header">
            <div class="title">
              <h4>快捷入口</h4>
              <p>按当前账号权限展示</p>
            </div>
          </div>

          <div class="grid grid-cols-2 gap-3 sm:grid-cols-4">
            <div
              v-for="item in quickEntries"
              :key="item.path"
              class="flex cursor-pointer items-center gap-2 rounded-lg border border-[var(--el-border-color-lighter)] px-3 py-3 transition-colors hover:border-theme hover:text-theme"
              @click="go(item.path)"
            >
              <ArtSvgIcon :icon="item.icon" class="text-lg" />
              <span class="text-sm">{{ item.label }}</span>
            </div>
          </div>

          <div v-if="quickEntries.length === 0" class="py-8 text-center text-sm text-g-500">
            当前账号暂无可访问的功能
          </div>
        </div>
      </ElCol>

      <ElCol :xs="24" :lg="8">
        <div class="art-card mb-5 p-5 max-sm:mb-4">
          <div class="art-card-header">
            <div class="title">
              <h4>系统信息</h4>
            </div>
          </div>

          <div class="space-y-3 text-sm">
            <div class="flex items-center justify-between">
              <span class="text-g-600">登录账号</span>
              <span>{{ userName || '-' }}</span>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-g-600">Go 版本</span>
              <span>{{ overview.system.goVersion || '-' }}</span>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-g-600">服务器时间</span>
              <span>{{ overview.system.serverTime || '-' }}</span>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-g-600">文件占用</span>
              <span>{{ formatSize(overview.counts.fileSize) }}</span>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-g-600">今日错误</span>
              <ElTag size="small" :type="overview.counts.todayError > 0 ? 'danger' : 'success'">
                {{ overview.counts.todayError }}
              </ElTag>
            </div>
          </div>
        </div>
      </ElCol>
    </ElRow>
  </div>
</template>

<script setup lang="ts">
  import { useRouter } from 'vue-router'
  import { useUserStore } from '@/store/modules/user'
  import { useAuth } from '@/hooks/core/useAuth'
  import { fetchGetDashboardOverview } from '@/api/system/dashboard'

  defineOptions({ name: 'DashboardConsole' })

  interface QuickEntry {
    label: string
    icon: string
    path: string
    auth: string
  }

  const router = useRouter()
  const userStore = useUserStore()
  const { hasAuth } = useAuth()

  const loading = ref(true)

  const overview = ref<Api.Dashboard.Overview>({
    counts: {
      user: 0,
      role: 0,
      menu: 0,
      file: 0,
      fileSize: 0,
      todayLogin: 0,
      todayRequest: 0,
      todayError: 0
    },
    loginTrend: [],
    requestTrend: [],
    recentLogins: [],
    system: {
      goVersion: '',
      serverTime: ''
    }
  })

  const userName = computed(() => userStore.info?.userName || '')
  const displayName = computed(
    () => userStore.info?.nickname || userStore.info?.userName || '管理员'
  )
  const avatar = computed(() => userStore.info?.avatar || '')

  const today = computed(() =>
    new Date().toLocaleDateString('zh-CN', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      weekday: 'long'
    })
  )

  const greeting = computed(() => {
    const hour = new Date().getHours()
    if (hour < 6) return '注意休息，夜深了'
    if (hour < 12) return '上午好，开始今天的工作吧'
    if (hour < 18) return '下午好，喝杯水休息一下'
    return '晚上好，辛苦了'
  })

  const cards = computed(() => {
    const counts = overview.value.counts
    return [
      {
        label: '用户总数',
        value: counts.user,
        tip: `${counts.role} 个角色 · ${counts.menu} 项菜单`,
        icon: 'ri:user-line'
      },
      {
        label: '文件数量',
        value: counts.file,
        tip: `占用 ${formatSize(counts.fileSize)}`,
        icon: 'ri:folder-3-line'
      },
      {
        label: '今日请求',
        value: counts.todayRequest,
        tip: `今日登录 ${counts.todayLogin} 次`,
        icon: 'ri:route-line'
      },
      {
        label: '今日错误',
        value: counts.todayError,
        tip: counts.todayError > 0 ? '请及时处理' : '运行正常',
        icon: 'ri:error-warning-line'
      }
    ]
  })

  const trendLabels = computed(() => overview.value.loginTrend.map((item) => item.date))

  const trendData = computed(() => [
    {
      name: '登录次数',
      data: overview.value.loginTrend.map((item) => item.count),
      showAreaColor: true
    },
    {
      name: 'API 请求',
      data: overview.value.requestTrend.map((item) => item.count)
    }
  ])

  const entries: QuickEntry[] = [
    {
      label: '用户管理',
      icon: 'ri:user-settings-line',
      path: '/system/user',
      auth: 'system:user:list'
    },
    {
      label: '角色管理',
      icon: 'ri:shield-user-line',
      path: '/system/role',
      auth: 'system:role:list'
    },
    { label: '菜单管理', icon: 'ri:menu-line', path: '/system/menu', auth: 'system:menu:list' },
    { label: '字典管理', icon: 'ri:book-2-line', path: '/system/dict', auth: 'system:dict:list' },
    { label: '文件管理', icon: 'ri:folder-line', path: '/infra/file', auth: 'infra:file:list' },
    {
      label: '文件配置',
      icon: 'ri:database-2-line',
      path: '/infra/file-config',
      auth: 'infra:file-config:list'
    },
    {
      label: '参数配置',
      icon: 'ri:settings-3-line',
      path: '/infra/config',
      auth: 'infra:config:list'
    },
    {
      label: '访问日志',
      icon: 'ri:route-line',
      path: '/infra/api-access-log',
      auth: 'infra:api-log:list'
    }
  ]

  const quickEntries = computed(() => entries.filter((item) => hasAuth(item.auth)))

  const formatSize = (size: number): string => {
    if (!size) return '0 B'
    if (size < 1024) return `${size} B`
    if (size < 1024 * 1024) return `${(size / 1024).toFixed(1)} KB`
    if (size < 1024 * 1024 * 1024) return `${(size / 1024 / 1024).toFixed(1)} MB`
    return `${(size / 1024 / 1024 / 1024).toFixed(1)} GB`
  }

  const go = (path: string): void => {
    router.push(path)
  }

  const loadData = async (): Promise<void> => {
    loading.value = true
    try {
      overview.value = await fetchGetDashboardOverview()
    } catch {
      // 错误提示已由请求层统一弹出，控制台保留空态
    } finally {
      loading.value = false
    }
  }

  onMounted(() => {
    void loadData()
  })
</script>
