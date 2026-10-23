<!-- 会员收货地址弹窗：后台代客新增 / 编辑地址 -->
<template>
  <ElDialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑收货地址' : '新增收货地址'"
    width="560px"
    align-center
    :close-on-click-modal="false"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="96px">
      <ElFormItem label="收件人" prop="name">
        <ElInput v-model="form.name" maxlength="32" placeholder="请填写收件人姓名" />
      </ElFormItem>

      <ElFormItem label="手机号" prop="mobile">
        <ElInput v-model="form.mobile" maxlength="20" placeholder="请填写联系电话" />
      </ElFormItem>

      <ElFormItem label="所在地区" prop="areaId">
        <ArtDistrictCascader v-model="form.areaId" placeholder="省 / 市 / 区县（选填）" />
      </ElFormItem>

      <ElFormItem label="详细地址" prop="detailAddress">
        <ElInput
          v-model="form.detailAddress"
          type="textarea"
          :rows="2"
          maxlength="255"
          show-word-limit
          placeholder="街道、门牌号等"
        />
      </ElFormItem>

      <ElFormItem label="设为默认">
        <ElSwitch v-model="form.defaultStatus" :active-value="1" :inactive-value="0" />
        <span class="ml-2 text-xs text-g-600">设默认后该会员其它地址会自动取消默认</span>
      </ElFormItem>
    </ElForm>

    <template #footer>
      <ElButton @click="dialogVisible = false">取消</ElButton>
      <ElButton type="primary" :loading="submitting" @click="handleSubmit">保存</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { ElMessage } from 'element-plus'
  import type { FormInstance, FormRules } from 'element-plus'
  import ArtDistrictCascader from '@/components/core/forms/art-district-cascader/index.vue'
  import { fetchCreateMemberAddress, fetchUpdateMemberAddress } from '@/api/member/user'

  defineOptions({ name: 'MemberAddressDialog' })

  interface Props {
    visible: boolean
    memberId: string
    /** 传入地址表示编辑，为空表示新增 */
    address?: Api.Member.AddressItem | null
  }

  const props = withDefaults(defineProps<Props>(), {
    address: null
  })

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void
    (e: 'success'): void
  }>()

  const dialogVisible = computed({
    get: () => props.visible,
    set: (value: boolean) => emit('update:visible', value)
  })

  const isEdit = computed(() => !!props.address?.id)

  const formRef = ref<FormInstance>()
  const submitting = ref(false)

  const form = reactive({
    name: '',
    mobile: '',
    // 区划 ID 为雪花 ID，表单里用字符串，提交时转数字
    areaId: '',
    detailAddress: '',
    defaultStatus: 0
  })

  const rules: FormRules = {
    name: [{ required: true, message: '请填写收件人', trigger: 'blur' }],
    mobile: [{ required: true, message: '请填写手机号', trigger: 'blur' }],
    detailAddress: [{ required: true, message: '请填写详细地址', trigger: 'blur' }]
  }

  // 打开弹窗时按编辑 / 新增回填
  watch(
    () => props.visible,
    (visible) => {
      if (!visible) return

      form.name = props.address?.name ?? ''
      form.mobile = props.address?.mobile ?? ''
      form.areaId = props.address?.areaId ? String(props.address.areaId) : ''
      form.detailAddress = props.address?.detailAddress ?? ''
      form.defaultStatus = props.address?.defaultStatus ?? 0
      nextTick(() => formRef.value?.clearValidate())
    },
    { immediate: true }
  )

  // 提交地址：区划 ID 转数字后调用新增或编辑接口
  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value || !props.memberId) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    const payload: Api.Member.AddressSaveParams = {
      name: form.name,
      mobile: form.mobile,
      areaId: form.areaId ? Number(form.areaId) : 0,
      detailAddress: form.detailAddress,
      defaultStatus: form.defaultStatus
    }

    submitting.value = true
    try {
      if (isEdit.value && props.address) {
        await fetchUpdateMemberAddress(props.memberId, props.address.id, payload)
      } else {
        await fetchCreateMemberAddress(props.memberId, payload)
      }
      ElMessage.success('地址已保存')
      dialogVisible.value = false
      emit('success')
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      submitting.value = false
    }
  }
</script>
