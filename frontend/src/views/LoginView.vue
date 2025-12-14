<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <h1>PlayerPlus Platform</h1>
        <p>AI 工具平台</p>
      </div>

      <a-form
        :model="formState"
        @finish="handleSubmit"
        layout="vertical"
        class="login-form"
      >
        <a-form-item
          label="用户名"
          name="username"
          :rules="[{ required: true, message: '请输入用户名' }]"
        >
          <a-input
            v-model:value="formState.username"
            placeholder="请输入用户名"
            size="large"
          />
        </a-form-item>

        <a-form-item
          label="密码"
          name="password"
          :rules="[{ required: true, message: '请输入密码' }]"
        >
          <a-input-password
            v-model:value="formState.password"
            placeholder="请输入密码"
            size="large"
          />
        </a-form-item>

        <a-form-item>
          <a-button
            type="primary"
            html-type="submit"
            size="large"
            block
            :loading="loading"
          >
            登录
          </a-button>
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
  username: '',
  password: ''
})

const loading = ref(false)

const handleSubmit = async () => {
  loading.value = true
  try {
    const { data } = await authApi.login(formState.username, formState.password)
    authStore.setAuth(data.token, data.user)
    message.success('登录成功')
    router.push('/')
  } catch (error: any) {
    message.error(error.response?.data?.error || '登录失败')
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
