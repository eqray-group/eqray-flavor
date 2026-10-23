<!-- 后厨登录：账号密码 + 验证码，多门店时选一家进店 -->
<template>
  <div class="login-page">
    <div class="login-box">
      <h1 class="title">后厨看板</h1>
      <p class="subtitle">登录后进入本店制作队列</p>

      <template v-if="!stores.length">
        <label class="field">
          <span>账号</span>
          <input v-model="form.username" autocomplete="username" @keyup.enter="submit" />
        </label>
        <label class="field">
          <span>密码</span>
          <input v-model="form.password" type="password" autocomplete="current-password" @keyup.enter="submit" />
        </label>
        <label class="field">
          <span>验证码</span>
          <div class="captcha-row">
            <input v-model="form.captchaCode" maxlength="6" @keyup.enter="submit" />
            <img v-if="captchaImage" :src="captchaImage" alt="验证码" @click="loadCaptcha" />
            <button v-else class="link" type="button" @click="loadCaptcha">获取</button>
          </div>
        </label>

        <button class="primary" :disabled="loading" type="button" @click="submit">
          {{ loading ? '登录中…' : '登录' }}
        </button>
      </template>

      <template v-else>
        <p class="subtitle">该账号有多个门店，请选择后厨所在门店</p>
        <button
          v-for="store in stores"
          :key="store.id"
          class="store-item"
          type="button"
          @click="chooseStore(store)"
        >
          {{ store.name }}
        </button>
      </template>

      <p v-if="error" class="error">{{ error }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { onMounted, reactive, ref } from 'vue'
  import { useRouter } from 'vue-router'

  import { fetchCaptcha, fetchLogin, fetchVisibleStores, type VisibleStore } from '@/api/auth'
  import { REFRESH_KEY, STORE_KEY, TOKEN_KEY, notify } from '@/utils/http'

  const router = useRouter()

  const form = reactive({
    username: '',
    password: '',
    captchaId: '',
    captchaCode: ''
  })
  const captchaImage = ref('')
  const loading = ref(false)
  const error = ref('')
  const stores = ref<VisibleStore[]>([])

  const loadCaptcha = async (): Promise<void> => {
    try {
      const res = await fetchCaptcha()
      form.captchaId = res.captchaId
      captchaImage.value = res.image
    } catch {
      // 提示已由请求层弹出
    }
  }

  const submit = async (): Promise<void> => {
    if (!form.username || !form.password) {
      error.value = '请输入账号与密码'

      return
    }

    loading.value = true
    error.value = ''
    try {
      const res = await fetchLogin({ ...form })
      localStorage.setItem(TOKEN_KEY, res.accessToken)
      localStorage.setItem(REFRESH_KEY, res.refreshToken)

      // 门店决定看板数据来源，必须落定后才能进看板
      const visible = await fetchVisibleStores()
      if (!visible.length) {
        error.value = '该账号没有可管理的门店'

        return
      }
      if (visible.length === 1) {
        localStorage.setItem(STORE_KEY, visible[0].id)
        await router.replace('/')

        return
      }
      stores.value = visible
    } catch {
      // 验证码一次性，失败后换一张
      void loadCaptcha()
    } finally {
      loading.value = false
    }
  }

  const chooseStore = async (store: VisibleStore): Promise<void> => {
    localStorage.setItem(STORE_KEY, store.id)
    notify(`已进入 ${store.name}`, 'success')
    await router.replace('/')
  }

  onMounted(loadCaptcha)
</script>

<style scoped>
  .login-page {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
    background: #14161a;
  }

  .login-box {
    width: 340px;
    padding: 28px 24px;
    background: #1c1f26;
    border: 1px solid #2a2e37;
    border-radius: 10px;
  }

  .title {
    margin: 0;
    font-size: 22px;
    text-align: center;
  }

  .subtitle {
    margin: 6px 0 18px;
    font-size: 13px;
    color: #8b93a1;
    text-align: center;
  }

  .field {
    display: block;
    margin-bottom: 14px;
  }

  .field span {
    display: block;
    margin-bottom: 6px;
    font-size: 13px;
    color: #a9b1bf;
  }

  .field input {
    width: 100%;
    padding: 9px 10px;
    font-size: 14px;
    color: #e8eaed;
    background: #14161a;
    border: 1px solid #2f3541;
    border-radius: 6px;
    outline: none;
  }

  .field input:focus {
    border-color: #3f8cff;
  }

  .captcha-row {
    display: flex;
    gap: 8px;
    align-items: center;
  }

  .captcha-row img {
    width: 100px;
    height: 38px;
    cursor: pointer;
    border-radius: 6px;
  }

  .primary {
    width: 100%;
    padding: 11px;
    margin-top: 6px;
    font-size: 15px;
    color: #fff;
    background: #3f8cff;
    border: none;
    border-radius: 6px;
  }

  .primary:disabled {
    opacity: 0.6;
  }

  .store-item {
    display: block;
    width: 100%;
    padding: 12px;
    margin-bottom: 10px;
    font-size: 15px;
    color: #e8eaed;
    text-align: left;
    background: #14161a;
    border: 1px solid #2f3541;
    border-radius: 6px;
  }

  .store-item:hover {
    border-color: #3f8cff;
  }

  .link {
    padding: 8px 12px;
    color: #3f8cff;
    background: transparent;
    border: 1px solid #2f3541;
    border-radius: 6px;
  }

  .error {
    margin: 12px 0 0;
    font-size: 13px;
    color: #ff7875;
    text-align: center;
  }
</style>
