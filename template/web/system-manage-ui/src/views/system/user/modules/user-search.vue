<template>
  <ArtSearchBar
    ref="searchBarRef"
    v-model="formData"
    :items="formItems"
    :rules="rules"
    @reset="handleReset"
    @search="handleSearch"
  >
  </ArtSearchBar>
</template>

<script setup lang="ts">
  interface Props {
    modelValue: Api.SystemManage.UserSearchParams
  }
  interface Emits {
    (e: 'update:modelValue', value: Api.SystemManage.UserSearchParams): void
    (e: 'search', params: Api.SystemManage.UserSearchParams): void
    (e: 'reset'): void
  }
  const props = defineProps<Props>()
  const emit = defineEmits<Emits>()

  // 表单数据双向绑定
  const searchBarRef = ref()
  const formData = computed({
    get: () => props.modelValue,
    set: (val) => emit('update:modelValue', val)
  })

  // 校验规则
  const rules = {}

  // 状态选项：与 system_users.status 一致（0 禁用 / 1 启用）
  const statusOptions = [
    { label: '启用', value: 1 },
    { label: '禁用', value: 0 }
  ]

  // 表单配置：key 必须与后端接受的查询参数同名
  const formItems = computed(() => [
    {
      label: '账号',
      key: 'username',
      type: 'input',
      placeholder: '请输入账号',
      clearable: true
    },
    {
      label: '昵称',
      key: 'nickname',
      type: 'input',
      placeholder: '请输入昵称',
      clearable: true
    },
    {
      label: '手机号',
      key: 'mobile',
      type: 'input',
      placeholder: '请输入手机号',
      clearable: true
    },
    {
      label: '状态',
      key: 'status',
      type: 'select',
      props: {
        placeholder: '请选择状态',
        options: statusOptions
      }
    }
  ])

  // 事件
  function handleReset() {
    emit('reset')
  }

  async function handleSearch(params: Api.SystemManage.UserSearchParams) {
    await searchBarRef.value.validate()
    emit('search', params)
  }
</script>
