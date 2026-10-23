<!-- 通知派发日志抽屉：按次数展示每次派发结果与业务方响应，便于排查回调未生效 -->
<template>
  <ElDrawer v-model="drawerVisible" title="派发日志" size="620px">
    <div v-if="task" class="mb-4 text-sm text-g-600">
      商户订单号：{{ task.merchantOrderId }} · 回调地址：{{ task.notifyUrl }}
    </div>

    <ElTable v-loading="loading" :data="logs" size="small" border max-height="560">
      <ElTableColumn prop="notifyTimes" label="第几次" width="90" />
      <ElTableColumn label="结果" width="100">
        <template #default="{ row }">
          <ElTag :type="statusTagType(row.status)" effect="plain">
            {{ row.status === 2 ? '成功' : '失败' }}
          </ElTag>
        </template>
      </ElTableColumn>
      <ElTableColumn prop="createdAt" label="执行时间" width="170" />
      <ElTableColumn label="业务方响应" min-width="220">
        <template #default="{ row }">
          <span class="text-xs">{{ row.response || '-' }}</span>
        </template>
      </ElTableColumn>
    </ElTable>

    <div v-if="!loading && logs.length === 0" class="mt-4 text-sm text-g-600">
      暂无派发记录，任务会在下一分钟的定时调度里执行。
    </div>
  </ElDrawer>
</template>

<script setup lang="ts">
  import { fetchGetPayNotifyTaskLogs } from '@/api/pay/notify'

  defineOptions({ name: 'PayNotifyLogDrawer' })

  interface Props {
    visible: boolean
    task?: Api.Pay.NotifyTaskItem | null
  }

  const props = withDefaults(defineProps<Props>(), {
    task: null
  })

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void
  }>()

  const drawerVisible = computed({
    get: () => props.visible,
    set: (value: boolean) => emit('update:visible', value)
  })

  const loading = ref(false)
  const logs = ref<Api.Pay.NotifyLogItem[]>([])

  // 日志状态对应的标签类型
  const statusTagType = (status: number): 'success' | 'danger' => {
    return status === 2 ? 'success' : 'danger'
  }

  // 打开抽屉时拉该任务的派发日志，最多 50 条
  watch(
    () => props.visible,
    async (visible) => {
      if (!visible || !props.task) return

      loading.value = true
      logs.value = []
      try {
        const res = await fetchGetPayNotifyTaskLogs(props.task.id, { pageNo: 1, pageSize: 50 })
        logs.value = res?.list ?? []
      } catch {
        // 错误提示已由请求层统一弹出
      } finally {
        loading.value = false
      }
    },
    { immediate: true }
  )
</script>
