<!-- 个人中心页面 -->
<template>
  <div class="user-center w-full">
    <div class="flex items-start gap-5 max-lg:flex-col">
      <!-- 左侧：个人信息 -->
      <div class="w-96 shrink-0 max-lg:w-full">
        <div class="art-card-sm relative p-9 pb-8 overflow-hidden text-center">
          <img class="absolute top-0 left-0 w-full h-50 object-cover" src="@imgs/user/bg.webp" />
          <ElUpload
            class="relative z-10 mt-30 mx-auto"
            :show-file-list="false"
            :before-upload="beforeAvatarUpload"
            :http-request="handleAvatarUpload"
            accept=".jpg,.jpeg,.png,.gif,.webp,.bmp"
          >
            <div
              v-loading="avatarUploading"
              class="group relative w-20 h-20 overflow-hidden border-2 border-white rounded-full cursor-pointer"
            >
              <img class="w-full h-full object-cover" :src="avatarUrl" alt="avatar" />
              <div
                class="absolute inset-0 flex items-center justify-center text-xs text-white transition-opacity bg-black/40 opacity-0 group-hover:opacity-100"
              >
                更换头像
              </div>
            </div>
          </ElUpload>
          <h2 class="mt-5 text-xl font-normal">{{ displayName }}</h2>
          <p class="mt-2 text-sm text-g-600">@{{ userInfo.userName || '-' }}</p>

          <div class="w-75 mx-auto mt-7.5 text-left">
            <div class="mt-2.5 flex items-center">
              <ArtSvgIcon icon="ri:mail-line" class="shrink-0 text-g-700" />
              <span class="ml-2 text-sm truncate">{{ userInfo.email || '未设置邮箱' }}</span>
            </div>
            <div class="mt-2.5 flex items-center">
              <ArtSvgIcon icon="ri:smartphone-line" class="shrink-0 text-g-700" />
              <span class="ml-2 text-sm">{{ userInfo.mobile || '未设置手机号' }}</span>
            </div>
            <div class="mt-2.5 flex items-center">
              <ArtSvgIcon icon="ri:user-3-line" class="shrink-0 text-g-700" />
              <span class="ml-2 text-sm truncate">{{ roleText }}</span>
            </div>
            <div class="mt-2.5 flex items-center">
              <ArtSvgIcon icon="ri:shield-check-line" class="shrink-0 text-g-700" />
              <span class="ml-2 text-sm">{{ permissionCount }} 项操作权限</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 右侧：资料与密码 -->
      <div class="flex-1 min-w-0 max-lg:w-full">
        <div class="art-card-sm">
          <h1 class="p-4 text-xl font-normal border-b border-g-300">基本资料</h1>

          <ElForm
            ref="profileFormRef"
            :model="profileForm"
            :rules="profileRules"
            class="box-border p-5"
            label-position="top"
          >
            <ElRow :gutter="20">
              <ElCol :xs="24" :sm="12">
                <ElFormItem label="账号">
                  <ElInput :model-value="userInfo.userName" disabled />
                </ElFormItem>
              </ElCol>
              <ElCol :xs="24" :sm="12">
                <ElFormItem label="昵称" prop="nickname">
                  <ElInput v-model="profileForm.nickname" placeholder="请输入昵称" maxlength="64" />
                </ElFormItem>
              </ElCol>
              <ElCol :xs="24" :sm="12">
                <ElFormItem label="邮箱" prop="email">
                  <ElInput v-model="profileForm.email" placeholder="请输入邮箱" maxlength="128" />
                </ElFormItem>
              </ElCol>
              <ElCol :xs="24" :sm="12">
                <ElFormItem label="手机号" prop="mobile">
                  <ElInput v-model="profileForm.mobile" placeholder="请输入手机号" maxlength="20" />
                </ElFormItem>
              </ElCol>
              <ElCol :xs="24" :sm="12">
                <ElFormItem label="性别" prop="sex">
                  <ElSelect v-model="profileForm.sex" placeholder="请选择性别">
                    <ElOption
                      v-for="item in sexOptions"
                      :key="item.value"
                      :label="item.label"
                      :value="item.value"
                    />
                  </ElSelect>
                </ElFormItem>
              </ElCol>
            </ElRow>

            <div class="flex justify-end">
              <ElButton type="primary" :loading="saving" v-ripple @click="handleSaveProfile">
                保存修改
              </ElButton>
            </div>
          </ElForm>
        </div>

        <div class="art-card-sm mt-5">
          <h1 class="p-4 text-xl font-normal border-b border-g-300">修改密码</h1>

          <ElForm
            ref="passwordFormRef"
            :model="passwordForm"
            :rules="passwordRules"
            class="box-border p-5"
            label-position="top"
          >
            <ElRow :gutter="20">
              <ElCol :xs="24" :sm="12">
                <ElFormItem label="当前密码" prop="oldPassword">
                  <ElInput
                    v-model="passwordForm.oldPassword"
                    type="password"
                    placeholder="请输入当前密码"
                    show-password
                  />
                </ElFormItem>
              </ElCol>
              <ElCol :xs="24" :sm="12">
                <ElFormItem label="新密码" prop="newPassword">
                  <ElInput
                    v-model="passwordForm.newPassword"
                    type="password"
                    placeholder="6 ~ 72 个字符"
                    show-password
                  />
                </ElFormItem>
              </ElCol>
              <ElCol :xs="24" :sm="12">
                <ElFormItem label="确认新密码" prop="confirmPassword">
                  <ElInput
                    v-model="passwordForm.confirmPassword"
                    type="password"
                    placeholder="请再次输入新密码"
                    show-password
                  />
                </ElFormItem>
              </ElCol>
            </ElRow>

            <div class="flex justify-end">
              <ElButton
                type="primary"
                :loading="passwordSaving"
                v-ripple
                @click="handleChangePassword"
              >
                保存修改
              </ElButton>
            </div>
          </ElForm>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ElMessage } from 'element-plus'
  import type { FormInstance, FormRules, UploadRequestOptions } from 'element-plus'
  import defaultAvatar from '@imgs/user/avatar.webp'
  import { useUserStore } from '@/store/modules/user'
  import {
    fetchGetUserInfo,
    fetchUpdateProfile,
    fetchUpdateProfilePassword
  } from '@/api/system/auth'
  import { fetchUploadFile } from '@/api/infra/file'

  defineOptions({ name: 'SystemUserCenter' })

  const userStore = useUserStore()
  const userInfo = computed(() => userStore.getUserInfo)

  const displayName = computed(() => userInfo.value.nickname || userInfo.value.userName || '-')
  const avatarUrl = computed(() => userInfo.value.avatar || defaultAvatar)
  const roleText = computed(() => {
    const roles = userInfo.value.roles ?? []
    return roles.length ? roles.join('、') : '未分配角色'
  })
  const permissionCount = computed(() => userInfo.value.buttons?.length ?? 0)

  const sexOptions = [
    { value: 1, label: '男' },
    { value: 2, label: '女' },
    { value: 0, label: '未知' }
  ]

  // ---------- 基本资料 ----------
  const profileFormRef = ref<FormInstance>()
  const saving = ref(false)
  const profileForm = reactive({
    nickname: '',
    email: '',
    mobile: '',
    sex: 0
  })

  const profileRules: FormRules = {
    nickname: [
      { required: true, message: '请输入昵称', trigger: 'blur' },
      { max: 64, message: '昵称不能超过 64 个字符', trigger: 'blur' }
    ],
    email: [{ max: 128, message: '邮箱不能超过 128 个字符', trigger: 'blur' }],
    mobile: [{ max: 20, message: '手机号不能超过 20 个字符', trigger: 'blur' }]
  }

  // 用户信息变化后同步表单，避免刷新后仍显示旧值
  watch(
    userInfo,
    (info) => {
      profileForm.nickname = info.nickname ?? ''
      profileForm.email = info.email ?? ''
      profileForm.mobile = info.mobile ?? ''
      profileForm.sex = info.sex ?? 0
    },
    { immediate: true, deep: true }
  )

  const handleSaveProfile = async (): Promise<void> => {
    if (!profileFormRef.value) return

    const valid = await profileFormRef.value.validate().catch(() => false)
    if (!valid) return

    saving.value = true
    try {
      await fetchUpdateProfile({
        nickname: profileForm.nickname.trim(),
        email: profileForm.email.trim(),
        mobile: profileForm.mobile.trim(),
        sex: profileForm.sex
      })
      // 重新拉取信息，保证顶部头像 / 昵称等展示同步
      const info = await fetchGetUserInfo()
      userStore.setUserInfo(info)
      ElMessage.success('资料已更新')
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      saving.value = false
    }
  }

  // ---------- 头像上传 ----------
  const avatarUploading = ref(false)

  const beforeAvatarUpload = (file: File): boolean => {
    const allowed = ['image/jpeg', 'image/png', 'image/gif', 'image/webp', 'image/bmp']
    if (!allowed.includes(file.type)) {
      ElMessage.error('仅支持 jpg / png / gif / webp / bmp 图片')
      return false
    }
    if (file.size > 5 * 1024 * 1024) {
      ElMessage.error('图片不能超过 5MB')
      return false
    }
    return true
  }

  const handleAvatarUpload = async (options: UploadRequestOptions): Promise<void> => {
    avatarUploading.value = true
    try {
      const uploaded = await fetchUploadFile(options.file)
      const info = userInfo.value

      // 只提交头像，其余字段沿用当前值，避免把未保存的表单内容一起写进去
      await fetchUpdateProfile({
        nickname: info.nickname || info.userName || '',
        email: info.email ?? '',
        mobile: info.mobile ?? '',
        sex: info.sex ?? 0,
        avatar: uploaded.url
      })

      const latest = await fetchGetUserInfo()
      userStore.setUserInfo(latest)
      ElMessage.success('头像已更新')
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      avatarUploading.value = false
    }
  }

  // ---------- 修改密码 ----------
  const passwordFormRef = ref<FormInstance>()
  const passwordSaving = ref(false)
  const passwordForm = reactive({
    oldPassword: '',
    newPassword: '',
    confirmPassword: ''
  })

  const passwordRules: FormRules = {
    oldPassword: [
      { required: true, message: '请输入当前密码', trigger: 'blur' },
      { min: 6, max: 72, message: '密码长度需在 6 ~ 72 个字符', trigger: 'blur' }
    ],
    newPassword: [
      { required: true, message: '请输入新密码', trigger: 'blur' },
      { min: 6, max: 72, message: '密码长度需在 6 ~ 72 个字符', trigger: 'blur' },
      {
        validator: (_rule, value: string, callback) => {
          if (value && value === passwordForm.oldPassword) {
            callback(new Error('新密码不能与当前密码相同'))
            return
          }
          callback()
        },
        trigger: 'blur'
      }
    ],
    confirmPassword: [
      { required: true, message: '请再次输入新密码', trigger: 'blur' },
      {
        validator: (_rule, value: string, callback) => {
          if (value !== passwordForm.newPassword) {
            callback(new Error('两次输入的密码不一致'))
            return
          }
          callback()
        },
        trigger: 'blur'
      }
    ]
  }

  const handleChangePassword = async (): Promise<void> => {
    if (!passwordFormRef.value) return

    const valid = await passwordFormRef.value.validate().catch(() => false)
    if (!valid) return

    passwordSaving.value = true
    try {
      await fetchUpdateProfilePassword({
        oldPassword: passwordForm.oldPassword,
        newPassword: passwordForm.newPassword
      })
      ElMessage.success('密码已修改')
      passwordFormRef.value.resetFields()
    } catch {
      // 错误提示已由请求层统一弹出
    } finally {
      passwordSaving.value = false
    }
  }
</script>
