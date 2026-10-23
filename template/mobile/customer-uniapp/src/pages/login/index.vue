<script setup lang="ts">
/**
 * 登录：手机号 + 验证码。后端在手机号未注册时自动建会员，所以没有单独的注册入口。
 */
import { sendSmsCode } from '@/api/member'
import { useSessionStore } from '@/store/session'

definePage({
  name: 'login',
  style: {
    navigationStyle: 'custom',
  },
})

const session = useSessionStore()
const router = useRouter()

const mobile = ref('')
const code = ref('')
const sending = ref(false)
const submitting = ref(false)
const countdown = ref(0)

const canSubmit = computed(() => /^1\d{10}$/.test(mobile.value) && code.value.length >= 4)
let timer: ReturnType<typeof setInterval> | null = null

async function sendCode(): Promise<void> {
  if (sending.value || countdown.value > 0) return
  if (!/^1\d{10}$/.test(mobile.value)) {
    useGlobalToast().warning({ msg: '请输入 11 位手机号' })
    return
  }

  sending.value = true
  try {
    await sendSmsCode(mobile.value)
    useGlobalToast().success({ msg: '验证码已发送' })
    countdown.value = 60
    timer = setInterval(() => {
      countdown.value -= 1
      if (countdown.value <= 0 && timer) {
        clearInterval(timer)
        timer = null
      }
    }, 1000)
  } catch {
    // 提示已由请求层统一处理
  } finally {
    sending.value = false
  }
}

async function submit(): Promise<void> {
  if (!canSubmit.value || submitting.value) return

  submitting.value = true
  try {
    await session.login(mobile.value, code.value)
    router.replaceAll({ name: 'menu' })
  } catch {
    code.value = ''
  } finally {
    submitting.value = false
  }
}

onUnload(() => {
  if (timer) clearInterval(timer)
})
</script>

<template>
  <view class="paper">
    <view class="masthead">
      <text class="kicker">WELCOME</text>
      <text class="title">点单</text>
      <text class="subtitle">登录后可用会员价、积分与优惠</text>
      <view class="rule">
        <view class="rule-thick" />
      </view>
    </view>

    <view class="form">
      <view class="field">
        <text class="label">手机号</text>
        <input
          v-model="mobile"
          class="input"
          type="number"
          :maxlength="11"
          placeholder="请输入手机号"
          placeholder-class="ph"
        />
      </view>

      <view class="field">
        <text class="label">验证码</text>
        <view class="code-row">
          <input
            v-model="code"
            class="input code-input"
            type="number"
            :maxlength="6"
            placeholder="6 位验证码"
            placeholder-class="ph"
          />
          <text class="send" :class="{ off: countdown > 0 || sending }" @click="sendCode">
            {{ countdown > 0 ? `${countdown}s` : sending ? '发送中' : '获取验证码' }}
          </text>
        </view>
      </view>

      <view class="submit" :class="{ off: !canSubmit || submitting }" @click="submit">
        {{ submitting ? '登录中…' : '登录并开始点餐' }}
      </view>

      <text class="fineprint">未注册的手机号将自动创建会员，继续即表示同意用户协议</text>
    </view>
  </view>
</template>

<style lang="scss" scoped>
$paper: #f7f2e7;
$paper-deep: #efe7d6;
$ink: #17130f;
$ink-soft: #5c5448;
$ink-faint: #9a9183;
$cinnabar: #b23a2f;
$serif: 'Songti SC', 'Noto Serif CJK SC', 'Source Han Serif SC', 'SimSun', serif;

.paper {
  min-height: 100vh;
  padding: 0 48rpx;
  background:
    radial-gradient(90% 45% at 50% 0%, #fbf7ee 0%, $paper 60%),
    $paper;
  color: $ink;
}

.masthead {
  padding-top: 160rpx;
}
.kicker {
  font-size: 20rpx;
  color: $cinnabar;
  letter-spacing: 8rpx;
}
.title {
  display: block;
  margin-top: 18rpx;
  font-family: $serif;
  font-size: 96rpx;
  font-weight: 600;
  letter-spacing: 12rpx;
}
.subtitle {
  display: block;
  margin-top: 16rpx;
  font-size: 24rpx;
  color: $ink-faint;
  letter-spacing: 1rpx;
}
.rule {
  margin-top: 34rpx;
}
.rule-thick {
  height: 6rpx;
  background: $ink;
}

.form {
  margin-top: 60rpx;
}
.field {
  margin-bottom: 44rpx;
}
.label {
  display: block;
  margin-bottom: 12rpx;
  font-size: 21rpx;
  color: $ink-soft;
  letter-spacing: 3rpx;
}
.input {
  padding: 16rpx 0;
  font-family: $serif;
  font-size: 40rpx;
  letter-spacing: 3rpx;
  color: $ink;
  border-bottom: 1px solid rgba(23, 19, 15, 0.3);
}
.ph {
  font-family: system-ui, -apple-system, 'PingFang SC', sans-serif;
  font-size: 26rpx;
  color: $ink-faint;
  letter-spacing: 0;
}
.code-row {
  display: flex;
  align-items: flex-end;
  gap: 20rpx;
}
.code-input {
  flex: 1;
}
.send {
  padding-bottom: 16rpx;
  font-size: 23rpx;
  color: $cinnabar;
  white-space: nowrap;
  letter-spacing: 1rpx;
}
.send.off {
  color: $ink-faint;
}

.submit {
  margin-top: 30rpx;
  padding: 28rpx;
  font-size: 28rpx;
  color: $paper;
  letter-spacing: 4rpx;
  text-align: center;
  background: $ink;
  border-radius: 999rpx;
}
.submit.off {
  opacity: 0.4;
}

.fineprint {
  display: block;
  margin-top: 28rpx;
  font-size: 20rpx;
  line-height: 1.7;
  color: $ink-faint;
  text-align: center;
}
</style>
