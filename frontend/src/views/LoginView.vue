<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <h1>PlayerPlus Platform</h1>
        <p>AI 工具平台</p>
      </div>

      <a-form
        :model="formState"
        @finish="handleLogin"
        layout="vertical"
        class="login-form"
      >
        <!-- Username with @playerplus.cn suffix -->
        <a-form-item
          label="用户名"
          name="username"
          :rules="[
            { required: true, message: '请输入用户名' },
            { pattern: /^[a-z0-9._-]+$/i, message: '用户名只能包含字母、数字、点、下划线和横线' }
          ]"
        >
          <a-input
            v-model:value="formState.username"
            placeholder="请输入用户名"
            size="large"
            :disabled="codeSent"
            @input="formState.username = formState.username.trim().toLowerCase()"
          >
            <template #suffix>
              <span class="email-suffix">@playerplus.cn</span>
            </template>
          </a-input>
        </a-form-item>

        <!-- Send Code Button (shown before code is sent) -->
        <a-form-item v-if="!codeSent">
          <a-button
            type="primary"
            size="large"
            block
            :loading="sendingCode"
            :disabled="!formState.username || sendingCode"
            @click="handleSendCode"
          >
            发送验证码
          </a-button>
        </a-form-item>

        <!-- Verification Code Input (shown after code is sent) -->
        <a-form-item
          v-if="codeSent"
          label="验证码"
          name="code"
          :rules="[
            { required: true, message: '请输入验证码' },
            { len: 6, message: '验证码为6位数字' }
          ]"
        >
          <a-input
            v-model:value="formState.code"
            placeholder="请输入6位验证码"
            size="large"
            maxlength="6"
            inputmode="numeric"
            @input="formState.code = formState.code.replace(/\D/g, '')"
          />
        </a-form-item>

        <!-- Login Button and Actions (shown after code is sent) -->
        <a-form-item v-if="codeSent">
          <a-button
            type="primary"
            html-type="submit"
            size="large"
            block
            :loading="loading"
            :disabled="formState.code.length !== 6"
          >
            登录
          </a-button>
          <div class="action-links">
            <a-button
              type="link"
              size="small"
              @click="handleResendCode"
              :disabled="countdown > 0 || sendingCode"
            >
              {{ countdown > 0 ? `${Math.floor(countdown / 60)}:${String(countdown % 60).padStart(2, '0')}后可重新发送` : '重新发送验证码' }}
            </a-button>
            <a-button type="link" size="small" @click="handleReset">
              返回修改用户名
            </a-button>
          </div>
        </a-form-item>
      </a-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, computed, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { authApi } from '@/api'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const formState = reactive({
  username: '',
  code: ''
})

const codeSent = ref(false)
const sendingCode = ref(false)
const loading = ref(false)
const countdown = ref(0)
const countdownTimer = ref<number | null>(null)

// Computed email with domain suffix
const email = computed(() => `${formState.username.trim().toLowerCase()}@playerplus.cn`)

// Clear countdown timer on unmount
onUnmounted(() => {
  if (countdownTimer.value) {
    clearInterval(countdownTimer.value)
  }
})

// Start countdown timer
const startCountdown = () => {
  // Clear existing timer if any
  if (countdownTimer.value) {
    clearInterval(countdownTimer.value)
  }

  countdown.value = 300
  countdownTimer.value = window.setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) {
      if (countdownTimer.value) {
        clearInterval(countdownTimer.value)
        countdownTimer.value = null
      }
    }
  }, 1000)
}

// Send verification code
const handleSendCode = async () => {
  if (!formState.username || sendingCode.value) {
    return
  }

  // Validate username format
  if (!/^[a-z0-9._-]+$/i.test(formState.username)) {
    message.error('用户名格式不正确')
    return
  }

  sendingCode.value = true
  try {
    await authApi.sendCode(email.value)
    codeSent.value = true
    startCountdown()
    message.success('验证码已发送到您的邮箱')
  } catch (error: any) {
    const errorMsg = error.response?.data?.error || '发送失败，请稍后重试'
    message.error(errorMsg)
  } finally {
    sendingCode.value = false
  }
}

// Resend verification code
const handleResendCode = async () => {
  if (countdown.value > 0 || sendingCode.value) {
    return
  }

  sendingCode.value = true
  try {
    await authApi.sendCode(email.value)
    startCountdown()
    message.success('验证码已重新发送')
  } catch (error: any) {
    const errorMsg = error.response?.data?.error || '发送失败，请稍后重试'
    message.error(errorMsg)
  } finally {
    sendingCode.value = false
  }
}

// Reset to username input state
const handleReset = () => {
  codeSent.value = false
  formState.code = ''
  countdown.value = 0
  if (countdownTimer.value) {
    clearInterval(countdownTimer.value)
    countdownTimer.value = null
  }
}

// Verify code and login
const handleLogin = async () => {
  if (formState.code.length !== 6) {
    message.error('请输入6位验证码')
    return
  }

  loading.value = true
  try {
    const { data } = await authApi.verify(email.value, formState.code)
    // Use email as user identifier (backend returns token only for verify)
    authStore.setAuth(data.token, formState.username)
    message.success('登录成功')
    router.push('/')
  } catch (error: any) {
    const errorMsg = error.response?.data?.error || '验证失败，请检查验证码'
    message.error(errorMsg)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.login-card {
  width: 400px;
  padding: 40px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.login-header h1 {
  margin: 0;
  font-size: 28px;
  color: #1a1a1a;
}

.login-header p {
  margin: 8px 0 0;
  color: #666;
}

.email-suffix {
  color: #999;
  font-size: 14px;
}

.action-links {
  display: flex;
  justify-content: space-between;
  margin-top: 12px;
}

.action-links .ant-btn-link {
  padding: 0;
  height: auto;
}
</style>
