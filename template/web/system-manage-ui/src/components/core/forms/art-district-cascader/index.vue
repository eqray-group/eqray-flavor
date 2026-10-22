<!-- 省 / 市 / 区县级联选择：数据取自 infra 行政区划，自动跳过国家级根节点 -->
<template>
  <ElCascader
    :model-value="innerValue"
    :options="options"
    :props="CASCADER_PROPS"
    :placeholder="placeholder"
    :disabled="disabled || loading"
    :clearable="clearable"
    filterable
    style="width: 100%"
    @update:model-value="handleChange"
    @visible-change="handleVisibleChange"
  />
</template>

<script setup lang="ts">
  import { ElCascader } from 'element-plus'
  import type { CascaderOption } from 'element-plus'
  import { fetchGetDistrictTree } from '@/api/infra/district'

  defineOptions({ name: 'ArtDistrictCascader' })

  type DistrictTreeNode = Api.Infra.DistrictTreeNode

  interface Props {
    /** 选中的区县 ID，雪花 ID 用字符串；空串表示未选择 */
    modelValue?: string
    placeholder?: string
    disabled?: boolean
    clearable?: boolean
  }

  const props = withDefaults(defineProps<Props>(), {
    modelValue: '',
    placeholder: '省 / 市 / 区县',
    disabled: false,
    clearable: true
  })

  const emit = defineEmits<{
    (e: 'update:modelValue', value: string): void
    /** 选择完成时返回区划路径名称，便于上层拼接地址 */
    (e: 'change', payload: { id: string; names: string[] }): void
  }>()

  // 悬停展开；选项统一为 value / label / children
  const CASCADER_PROPS = { expandTrigger: 'hover' } as const

  const loading = ref(false)
  const options = ref<CascaderOption[]>([])
  const innerValue = ref<string[]>([])
  // 去掉国家级根节点后的树，用于按 ID 反查名称
  const rootNodes = ref<DistrictTreeNode[]>([])

  // 区划树是静态数据，模块级缓存一次，多个页面复用同一份请求结果
  let treePromise: Promise<DistrictTreeNode[]> | null = null
  const loadTree = (): Promise<DistrictTreeNode[]> => {
    if (!treePromise) {
      treePromise = fetchGetDistrictTree().catch((error) => {
        // 失败后清空缓存，下次展开可重试
        treePromise = null
        throw error
      })
    }

    return treePromise
  }

  // 跳过国家级根节点（level 0，如「中国」），直接从省（level 1）开始
  const pickRoots = (nodes: DistrictTreeNode[]): DistrictTreeNode[] => {
    const roots: DistrictTreeNode[] = []
    const walk = (list: DistrictTreeNode[]) => {
      list.forEach((node) => {
        if (node.level >= 1) {
          roots.push(node)
          return
        }
        walk(node.children ?? [])
      })
    }
    walk(nodes)

    return roots
  }

  // 区划树转级联选项，叶子不带 children
  const toOptions = (nodes: DistrictTreeNode[]): CascaderOption[] =>
    nodes.map((node) => ({
      value: node.id,
      label: node.areaName,
      children: node.children?.length ? toOptions(node.children) : undefined
    }))

  // 按 ID 反查省 / 市 / 区县路径，用于回填
  const findPath = (nodes: DistrictTreeNode[], targetId: string): string[] => {
    for (const node of nodes) {
      if (node.id === targetId) return [node.id]

      const childPath = findPath(node.children ?? [], targetId)
      if (childPath.length > 0) return [node.id, ...childPath]
    }

    return []
  }

  // 按 ID 反查各级名称，供上层拼接完整地址
  const findNames = (nodes: DistrictTreeNode[], path: string[]): string[] => {
    const names: string[] = []
    let current = nodes
    for (const id of path) {
      const hit = current.find((node) => node.id === id)
      if (!hit) break
      names.push(hit.areaName)
      current = hit.children ?? []
    }

    return names
  }

  // 加载选项并回填当前值对应的路径
  const ensureOptions = async (): Promise<void> => {
    if (options.value.length > 0) return

    loading.value = true
    try {
      const tree = await loadTree()
      rootNodes.value = pickRoots(tree)
      options.value = toOptions(rootNodes.value)
      innerValue.value = props.modelValue ? findPath(rootNodes.value, props.modelValue) : []
    } catch {
      // 错误提示已由请求层统一弹出，保持空选项
    } finally {
      loading.value = false
    }
  }

  // 首次展开或外部直接设值时都要保证选项就绪
  const handleVisibleChange = (visible: boolean): void => {
    if (visible) ensureOptions()
  }

  // 级联选择变化：取最后一级作为区县 ID，清空时回传空串
  const handleChange = (value: unknown): void => {
    const path = Array.isArray(value) ? value.map((item) => String(item)) : []
    innerValue.value = path

    const id = path.length > 0 ? path[path.length - 1] : ''
    emit('update:modelValue', id)
    emit('change', { id, names: findNames(rootNodes.value, path) })
  }

  onMounted(() => {
    if (props.modelValue) ensureOptions()
  })

  // 外部值变化（如编辑回填）时重新解析路径
  watch(
    () => props.modelValue,
    async (value) => {
      if (!value) {
        innerValue.value = []
        return
      }
      await ensureOptions()
      const current = innerValue.value[innerValue.value.length - 1] ?? ''
      if (current !== value) {
        innerValue.value = findPath(rootNodes.value, value)
      }
    }
  )
</script>
