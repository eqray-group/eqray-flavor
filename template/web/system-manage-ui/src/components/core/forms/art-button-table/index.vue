<!-- 表格操作按钮：图标 + 文字 -->
<template>
  <ElButton
    :size="size"
    :type="buttonType"
    :plain="plain"
    :disabled="disabled"
    :class="{ '!px-2': !text }"
    @click="handleClick"
  >
    <ArtSvgIcon v-if="iconContent" :icon="iconContent" class="mr-1" />
    <span v-if="text">{{ text }}</span>
  </ElButton>
</template>

<script setup lang="ts">
  defineOptions({ name: 'ArtButtonTable' })

  interface Props {
    /** 操作语义，决定默认图标与按钮颜色 */
    type?: 'add' | 'edit' | 'delete' | 'view' | 'action' | 'more'
    /** 按钮文字，为空时只显示图标 */
    text?: string
    /** 自定义图标，优先级高于 type 的默认图标 */
    icon?: string
    /** 按钮尺寸 */
    size?: 'small' | 'default' | 'large'
    /** 是否朴素按钮（浅色底 + 边框） */
    plain?: boolean
    /** 是否禁用 */
    disabled?: boolean
  }

  const props = withDefaults(defineProps<Props>(), {
    size: 'small',
    plain: true,
    disabled: false
  })

  const emit = defineEmits<{
    (e: 'click'): void
  }>()

  // 语义 -> 默认图标 + Element Plus 按钮类型
  const defaultButtons = {
    add: { icon: 'ri:add-line', type: 'primary' },
    edit: { icon: 'ri:pencil-line', type: 'primary' },
    delete: { icon: 'ri:delete-bin-5-line', type: 'danger' },
    view: { icon: 'ri:eye-line', type: 'info' },
    action: { icon: '', type: 'primary' },
    more: { icon: 'ri:more-2-fill', type: 'primary' }
  } as const

  // 获取图标内容
  const iconContent = computed(() => {
    return props.icon || (props.type ? defaultButtons[props.type].icon : '')
  })

  // 获取按钮类型
  const buttonType = computed(() => (props.type ? defaultButtons[props.type].type : 'primary'))

  const handleClick = () => {
    emit('click')
  }
</script>
