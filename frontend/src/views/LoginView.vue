<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <h1>PlayPlus Platform</h1>
        <p>AI 工具平台</p>
      </div>

      <a-form
        :model="formState"
        @finish="handleSubmit"
        layout="vertical"
        class="login-form"
      >
        <a-form-item
          label="邮箱"
          name="email"
          :rules="[
            { required: true, message: '请输入邮箱' },
            { type: 'email', message: '请输入有效的邮箱地址' },
            { validator: validateDomain }
          ]"
        >
          <a-input
            v-model:value="formState.email"
            placeholder="请输入 @playerplus.cn 邮箱"
            size="large"
          />
        </a-form-item>

        <a-form-item v-if="codeSent" label="验证码" name="code" :rules="[{ required: true, message: '请输入验证码' }]">
          <a-input
            v-model:value="formState.code"
            placeholder="请输入6位验证码"
            size="large"
            maxlength="6"
          />
        </a-form-item>

        <a-form-item>
          <a-button
            v-if="!codeSent"
            type="primary"
            html-type="submit"
            size="large"
            block
            :loading="loading"
          >
            发送验证码
          </a-button>
          <a-space v-else direction="vertical" style="width: 100%">
            <a-button type="primary" html-type="submit" size="large" block :loading="loading">
              登录
            </a-button>
            <a-button type="link" block :disabled="countdown > 0" @click="resendCode">
              {{ countdown > 0 ? `${countdown}秒后可重新发送` : '重新发送验证码' }}
            </a-button>
          </a-space>
        </a-form-item>
      </a-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { message } from 'ant-design-vue'
import { authApi } from '@/api'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const formState = reactive({
  email: '',
  code: ''
})

const loading = ref(false)
const codeSent = ref(false)
const countdown = ref(0)

const validateDomain = (_rule: any, value: string) => {
  if (value && !value.toLowerCase().endsWith('@playerplus.cn')) {
    return Promise.reject('只允许 @playerplus.cn 域名邮箱')
  }
  return Promise.resolve()
}

const startCountdown = () => {
  countdown.value = 60
  const timer = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) {
      clearInterval(timer)
    }
  }, 1000)
}

const handleSubmit = async () => {
  loading.value = true
  try {
    if (!codeSent.value) {
      await authApi.sendCode(formState.email)
      codeSent.value = true
      startCountdown()
      message.success('验证码已发送到您的邮箱')
    } else {
      const { data } = await authApi.verify(formState.email, formState.code)
      authStore.setAuth(data.token, formState.email)
      message.success('登录成功')
      router.push('/')
    }
  } catch (error: any) {
    message.error(error.response?.data?.error || '操作失败')
  } finally {
    loading.value = false
  }
}

const resendCode = async () => {
  loading.value = true
  try {
    await authApi.sendCode(formState.email)
    startCountdown()
    message.success('验证码已重新发送')
  } catch (error: any) {
    message.error(error.response?.data?.error || '发送失败')
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
</style>
