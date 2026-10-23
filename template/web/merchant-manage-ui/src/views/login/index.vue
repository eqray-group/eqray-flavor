<!-- 门店端登录：与后台共用账号体系，登录后由令牌权限决定能看哪些操作 -->
<template>
  <div class="login-page">
    <ElCard class="login-card" shadow="never">
      <h2 class="login-title">门店端登录</h2>
      <ElForm :model="form" label-position="top" @submit.prevent="submit">
        <ElFormItem label="账号">
          <ElInput v-model="form.username" placeholder="后台账号" clearable />
        </ElFormItem>
        <ElFormItem label="密码">
          <ElInput v-model="form.password" type="password" placeholder="登录密码" show-password />
        </ElFormItem>
        <ElFormItem label="验证码">
          <div class="captcha-row">
            <ElInput v-model="form.captchaCode" placeholder="右侧数字" @keyup.enter="submit" />
            <img
              v-if="captcha"
              class="captcha-img"
              :src="captchaUrl"
              alt="验证码"
              title="点击刷新"
              @click="loadCaptcha"
            />
          </div>
        </ElFormItem>
      </ElForm>

      <ElButton type="primary" class="submit" :loading="loading" @click="submit">登录</ElButton>
      <p class="tip">账号由总部在「系统管理 → 用户管理」中分配</p>
    </ElCard>
  </div>
</template>

<script setup lang="ts">
  import { ElMessage } from 'element-plus'
  import { computed, onMounted, reactive, ref } from 'vue'
  import { useRouter } from 'vue-router'

  import { fetchCaptcha } from '@/api/auth'
  import { useAuthStore } from '@/stores/auth'

  const router = useRouter()
  const auth = useAuthStore()

  const loading = ref(false)
  const captcha = ref<Api.Auth.Captcha | null>(null)

  const form = reactive({
    username: '',
    password: '',
    captchaCode: ''
  })

  // 后端返回的可能是纯 base64，补上 data URI 前缀才能直接给 img 用
  const captchaUrl = computed(() => {
    const image = captcha.value?.image || ''

    return image.startsWith('data:') ? image : `data:image/png;base64,${image}`
  })

  const loadCaptcha = async (): Promise<void> => {
    captcha.value = await fetchCaptcha()
    form.captchaCode = ''
  }

  const submit = async (): Promise<void> => {
    if (!form.username || !form.password) {
      ElMessage.warning('请输入账号与密码')
      return
    }
    if (!captcha.value) {
      ElMessage.warning('验证码未加载，请刷新')
      return
    }

    loading.value = true
    try {
      await auth.login({
        username: form.username,
        password: form.password,
        captchaId: captcha.value.captchaId,
        captchaCode: form.captchaCode
      })
      await auth.loadUser()
      ElMessage.success('登录成功')
      router.push('/')
    } catch {
      // 失败提示由请求层统一弹出，这里只刷新验证码
      await loadCaptcha()
    } finally {
      loading.value = false
    }
  }

  onMounted(loadCaptcha)
</script>

<style scoped>
  .login-page {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
    background: linear-gradient(135deg, #1f2329 0%, #2f3640 100%);
  }

  .login-card {
    width: 380px;
    padding: 8px 12px 20px;
    border-radius: 12px;
  }

  .login-title {
    margin: 12px 0 24px;
    font-size: 20px;
    text-align: center;
  }

  .captcha-row {
    display: flex;
    gap: 10px;
    align-items: center;
    width: 100%;
  }

  .captcha-img {
    height: 40px;
    cursor: pointer;
    border: 1px solid #e5e7eb;
    border-radius: 6px;
  }

  .submit {
    width: 100%;
    height: 42px;
    margin-top: 8px;
  }

  .tip {
    margin-top: 14px;
    font-size: 12px;
    color: #9ca3af;
    text-align: center;
  }
</style>
