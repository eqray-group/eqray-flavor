<!-- 会员新增 / 编辑弹窗（字段对齐 yudao member_user） -->
<template>
  <ElDialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑会员' : '新增会员'"
    width="640px"
    align-center
    :close-on-click-modal="false"
  >
    <ElForm ref="formRef" :model="form" :rules="rules" label-width="90px">
      <ElFormItem label="头像" prop="avatar">
        <ElUpload
          class="member-avatar-upload"
          :show-file-list="false"
          :before-upload="beforeAvatarUpload"
          :http-request="handleAvatarUpload"
          accept=".jpg,.jpeg,.png,.gif,.webp,.bmp"
        >
          <div
            v-loading="avatarUploading"
            class="group relative w-20 h-20 overflow-hidden border-2 border-white rounded-full cursor-pointer"
          >
            <img class="w-full h-full object-cover" :src="avatarPreview" alt="avatar" />
            <div
              class="absolute inset-0 flex items-center justify-center text-xs text-white transition-opacity bg-black/40 opacity-0 group-hover:opacity-100"
            >
              更换
            </div>
          </div>
        </ElUpload>
        <span class="ml-3 text-xs text-g-600">支持 jpg / png / gif / webp / bmp，不超过 2MB</span>
      </ElFormItem>

      <div class="grid grid-cols-2 gap-x-4">
        <ElFormItem label="昵称" prop="nickname">
          <ElInput v-model="form.nickname" placeholder="请输入昵称" maxlength="64" />
        </ElFormItem>
        <ElFormItem label="手机号" prop="mobile">
          <ElInput v-model="form.mobile" placeholder="请输入手机号" maxlength="20" />
        </ElFormItem>
        <ElFormItem label="真实姓名" prop="name">
          <ElInput v-model="form.name" placeholder="选填" maxlength="30" />
        </ElFormItem>
        <ElFormItem label="性别" prop="sex">
          <ElSelect v-model="form.sex">
            <ElOption label="未知" :value="0" />
            <ElOption label="男" :value="1" />
            <ElOption label="女" :value="2" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="生日" prop="birthday">
          <ElDatePicker
            v-model="form.birthday"
            type="date"
            value-format="YYYY-MM-DD"
            placeholder="选择日期"
            class="w-full"
          />
        </ElFormItem>
        <ElFormItem label="会员等级" prop="levelId">
          <ElSelect v-model="form.levelId" placeholder="请选择等级">
            <ElOption
              v-for="item in levelOptions"
              :key="item.id"
              :label="`Lv.${item.level} ${item.name}`"
              :value="item.id"
            />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="积分" prop="point">
          <ElInputNumber v-model="form.point" :min="0" :max="99999999" controls-position="right" />
        </ElFormItem>
        <ElFormItem label="经验值" prop="experience">
          <ElInputNumber
            v-model="form.experience"
            :min="0"
            :max="99999999"
            controls-position="right"
          />
        </ElFormItem>
        <ElFormItem label="会员分组" prop="groupId">
          <ElSelect v-model="form.groupId" clearable placeholder="请选择分组">
            <ElOption
              v-for="item in groupOptions"
              :key="item.id"
              :label="item.name"
              :value="item.id"
            />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="状态" prop="status">
          <ElSwitch
            v-model="form.status"
            :active-value="1"
            :inactive-value="0"
            active-text="启用"
            inactive-text="禁用"
          />
        </ElFormItem>
      </div>

      <ElFormItem label="会员标签" prop="tagIds">
        <ElSelect v-model="form.tagIds" multiple clearable placeholder="请选择标签" class="w-full">
          <ElOption v-for="item in tagOptions" :key="item.id" :label="item.name" :value="item.id" />
        </ElSelect>
      </ElFormItem>

      <ElFormItem label="用户标记" prop="mark">
        <ElInput v-model="form.mark" placeholder="如：高价值客户" maxlength="255" />
      </ElFormItem>
    </ElForm>

    <template #footer>
      <ElButton @click="dialogVisible = false">取消</ElButton>
      <ElButton type="primary" :loading="submitting" @click="handleSubmit">提交</ElButton>
    </template>
  </ElDialog>
</template>

<script setup lang="ts">
  import { ElMessage } from 'element-plus'
  import type { FormInstance, FormRules, UploadRequestOptions } from 'element-plus'
  import defaultAvatar from '@imgs/user/avatar.webp'
  import { fetchUploadFile } from '@/api/infra/file'
  import { fetchCreateMember, fetchUpdateMember } from '@/api/member/user'
  import { fetchGetLevelOptions } from '@/api/member/level'
  import { fetchGetTagOptions } from '@/api/member/tag'
  import { fetchGetGroupOptions } from '@/api/member/group'

  defineOptions({ name: 'MemberDialog' })

  type MemberItem = Api.Member.MemberItem
  type DialogType = 'add' | 'edit'

  interface Props {
    visible: boolean
    type: DialogType
    memberData?: MemberItem | null
  }

  const props = withDefaults(defineProps<Props>(), {
    memberData: null
  })

  const emit = defineEmits<{
    (e: 'update:visible', value: boolean): void
    (e: 'success'): void
  }>()

  const dialogVisible = computed({
    get: () => props.visible,
    set: (value: boolean) => emit('update:visible', value)
  })

  const isEdit = computed(() => props.type === 'edit')

  const formRef = ref<FormInstance>()
  const submitting = ref(false)
  const avatarUploading = ref(false)

  // 下拉数据
  const levelOptions = ref<Api.Member.LevelItem[]>([])
  const tagOptions = ref<Api.Member.TagItem[]>([])
  const groupOptions = ref<Api.Member.GroupItem[]>([])
  const optionsLoaded = ref(false)

  const form = reactive({
    nickname: '',
    mobile: '',
    avatar: '',
    name: '',
    sex: 0,
    birthday: '',
    levelId: '',
    point: 0,
    experience: 0,
    tagIds: [] as string[],
    mark: '',
    groupId: '',
    status: 1
  })

  const avatarPreview = computed(() => form.avatar || defaultAvatar)

  const rules: FormRules = {
    nickname: [
      { required: true, message: '请输入昵称', trigger: 'blur' },
      { max: 64, message: '昵称不能超过 64 个字符', trigger: 'blur' }
    ],
    mobile: [
      { required: true, message: '请输入手机号', trigger: 'blur' },
      { max: 20, message: '手机号不能超过 20 个字符', trigger: 'blur' }
    ],
    levelId: [{ required: true, message: '请选择会员等级', trigger: 'change' }],
    name: [{ max: 30, message: '真实姓名不能超过 30 个字符', trigger: 'blur' }],
    mark: [{ max: 255, message: '用户标记不能超过 255 个字符', trigger: 'blur' }]
  }

  const loadOptions = async (): Promise<void> => {
    try {
      const [levels, tags, groups] = await Promise.all([
        fetchGetLevelOptions(),
        fetchGetTagOptions(),
        fetchGetGroupOptions()
      ])
      levelOptions.value = levels
      tagOptions.value = tags
      groupOptions.value = groups
      optionsLoaded.value = true

      // 新增时默认选中最低等级，避免等级必填报错
      if (!isEdit.value && !form.levelId && levels.length) {
        form.levelId = levels[0].id
      }
    } catch {
      // 错误提示已由请求层统一弹出
    }
  }

  watch(
    () => props.visible,
    (visible) => {
      if (!visible) return
      if (!optionsLoaded.value) {
        loadOptions()
      }
    }
  )

  // 每次打开弹窗时按当前数据重置表单
  watch(
    () => [props.visible, props.memberData] as const,
    ([visible]) => {
      if (!visible) return

      const data = props.memberData
      Object.assign(form, {
        nickname: data?.nickname ?? '',
        mobile: data?.mobile ?? '',
        avatar: data?.avatar ?? '',
        name: data?.name ?? '',
        sex: data?.sex ?? 0,
        birthday: data?.birthday ? data.birthday.slice(0, 10) : '',
        levelId: data?.levelId ?? '',
        point: data?.point ?? 0,
        experience: data?.experience ?? 0,
        tagIds: data?.tagIds ? data.tagIds.split(',').filter(Boolean) : [],
        mark: data?.mark ?? '',
        groupId: data?.groupId ?? '',
        status: data?.status ?? 1
      })

      nextTick(() => formRef.value?.clearValidate())
    },
    { immediate: true }
  )

  // ---------- 头像上传 ----------
  const beforeAvatarUpload = (file: File): boolean => {
    const allowed = ['image/jpeg', 'image/png', 'image/gif', 'image/webp', 'image/bmp']
    if (!allowed.includes(file.type)) {
      ElMessage.error('仅支持 jpg / png / gif / webp / bmp 图片')
      return false
    }
    if (file.size / 1024 / 1024 > 2) {
      ElMessage.error('图片不能超过 2MB')
      return false
    }
    return true
  }

  const handleAvatarUpload = async (options: UploadRequestOptions): Promise<void> => {
    avatarUploading.value = true
    try {
      const uploaded = await fetchUploadFile(options.file)
      form.avatar = uploaded.url
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      avatarUploading.value = false
    }
  }

  // ---------- 提交 ----------
  const handleSubmit = async (): Promise<void> => {
    if (!formRef.value) return

    const valid = await formRef.value.validate().catch(() => false)
    if (!valid) return

    const payload: Api.Member.MemberSaveParams = {
      nickname: form.nickname.trim(),
      mobile: form.mobile.trim(),
      avatar: form.avatar,
      name: form.name.trim(),
      sex: form.sex,
      birthday: form.birthday ? `${form.birthday} 00:00:00` : '',
      levelId: form.levelId,
      point: form.point,
      experience: form.experience,
      tagIds: form.tagIds.join(','),
      mark: form.mark.trim(),
      groupId: form.groupId || undefined,
      status: form.status
    }

    submitting.value = true
    try {
      if (isEdit.value && props.memberData) {
        await fetchUpdateMember(props.memberData.id, payload)
        ElMessage.success('保存成功')
      } else {
        await fetchCreateMember(payload)
        ElMessage.success('新增成功')
      }

      dialogVisible.value = false
      emit('success')
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      submitting.value = false
    }
  }
</script>
