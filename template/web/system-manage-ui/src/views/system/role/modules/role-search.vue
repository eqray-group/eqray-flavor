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
    modelValue: Api.SystemManage.RoleSearchParams
  }
  interface Emits {
    (e: 'update:modelValue', value: Api.SystemManage.RoleSearchParams): void
    (e: 'search', params: Api.SystemManage.RoleSearchParams): void
    (e: 'reset'): void
  }
  const props = defineProps<Props>()
  const emit = defineEmits<Emits>()

  const searchBarRef = ref()
  const formData = computed({
    get: () => props.modelValue,
    set: (val) => emit('update:modelValue', val)
  })

  const rules = {}

  // 与 system_role.status 一致（0 禁用 / 1 启用）
  const statusOptions = [
    { label: '启用', value: 1 },
    { label: '禁用', value: 0 }
  ]

  const formItems = computed(() => [
    {
      label: '角色名称',
      key: 'name',
      type: 'input',
      placeholder: '请输入角色名称',
      clearable: true
    },
    {
      label: '角色标识',
      key: 'code',
      type: 'input',
      placeholder: '请输入角色标识',
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

  function handleReset() {
    emit('reset')
  }

  async function handleSearch(params: Api.SystemManage.RoleSearchParams) {
    await searchBarRef.value.validate()
    emit('search', params)
  }
</script>
