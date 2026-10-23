<!-- 门店端布局：顶部选店 + 左侧导航，选店结果由请求层统一注入 X-Store-Id -->
<template>
  <ElContainer class="layout">
    <ElAside width="180px" class="aside">
      <div class="brand">门店端</div>
      <ElMenu
        :default-active="route.path"
        router
        class="menu"
        background-color="#1f2329"
        text-color="#c9cdd4"
        active-text-color="#ffffff"
      >
        <ElMenuItem index="/orders">订单看板</ElMenuItem>
        <ElMenuItem index="/tables">桌台</ElMenuItem>
        <ElMenuItem index="/pos">点单</ElMenuItem>
        <ElMenuItem index="/sold-out">沽清</ElMenuItem>
        <ElMenuItem index="/reports">报表</ElMenuItem>
        <ElMenuItem index="/print-tasks">打印任务</ElMenuItem>
      </ElMenu>
    </ElAside>

    <ElContainer>
      <ElHeader class="header">
        <span class="page-title">{{ route.meta.title || '' }}</span>
        <div class="header-right">
          <ElSelect
            :model-value="storeContext.currentID"
            placeholder="选择门店"
            class="w-52"
            @update:model-value="handleStoreChange"
          >
            <ElOption
              v-for="item in storeContext.stores"
              :key="item.id"
              :label="item.name"
              :value="item.id"
            />
          </ElSelect>
          <span class="user">{{ auth.displayName }}</span>
          <ElButton link type="primary" @click="handleLogout">退出</ElButton>
        </div>
      </ElHeader>

      <ElMain class="main">
        <router-view />
      </ElMain>
    </ElContainer>
  </ElContainer>
</template>

<script setup lang="ts">
  import { ElMessage } from 'element-plus'
  import { onMounted } from 'vue'
  import { useRoute, useRouter } from 'vue-router'

  import { useAuthStore } from '@/stores/auth'
  import { useStoreContext } from '@/stores/store-context'

  const route = useRoute()
  const router = useRouter()
  const auth = useAuthStore()
  const storeContext = useStoreContext()

  // 切换门店后刷新整页，确保列表按新门店重新拉取
  const handleStoreChange = (id: string): void => {
    storeContext.select(id)
    location.reload()
  }

  const handleLogout = async (): Promise<void> => {
    await auth.logout()
    ElMessage.success('已退出')
    router.push('/login')
  }

  onMounted(async () => {
    if (!auth.user) {
      await auth.loadUser()
    }
    await storeContext.load()
  })
</script>

<style scoped>
  .layout {
    height: 100vh;
  }

  .aside {
    background: #1f2329;
  }

  .brand {
    padding: 18px 16px;
    font-size: 16px;
    font-weight: 600;
    color: #fff;
  }

  .menu {
    border-right: none;
    background: transparent;
  }

  .header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    background: #fff;
    border-bottom: 1px solid #e5e7eb;
  }

  .header-right {
    display: flex;
    gap: 12px;
    align-items: center;
  }

  .user {
    font-size: 13px;
    color: #6b7280;
  }

  .main {
    padding: 0;
    overflow: auto;
  }

  .w-52 {
    width: 200px;
  }
</style>
