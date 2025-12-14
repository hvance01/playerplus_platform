<template>
  <a-layout class="layout">
    <a-layout-sider v-model:collapsed="collapsed" :trigger="null" collapsible class="sider">
      <div class="logo">
        <span v-if="!collapsed">PlayPlus</span>
        <span v-else>P</span>
      </div>
      <a-menu
        v-model:selectedKeys="selectedKeys"
        theme="dark"
        mode="inline"
        @click="handleMenuClick"
      >
        <a-menu-item key="faceswap">
          <template #icon><swap-outlined /></template>
          <span>视频换脸</span>
        </a-menu-item>
        <a-menu-item key="prompts">
          <template #icon><file-text-outlined /></template>
          <span>Prompt 管理</span>
        </a-menu-item>
      </a-menu>
    </a-layout-sider>

    <a-layout>
      <a-layout-header class="header">
        <menu-unfold-outlined
          v-if="collapsed"
          class="trigger"
          @click="collapsed = false"
        />
        <menu-fold-outlined v-else class="trigger" @click="collapsed = true" />

        <div class="header-right">
          <a-dropdown>
            <a-space>
              <a-avatar size="small">{{ userInitial }}</a-avatar>
              <span>{{ authStore.email }}</span>
            </a-space>
            <template #overlay>
              <a-menu>
                <a-menu-item key="logout" @click="handleLogout">
                  <logout-outlined />
                  退出登录
                </a-menu-item>
              </a-menu>
            </template>
          </a-dropdown>
        </div>
      </a-layout-header>

      <a-layout-content class="content">
        <router-view />
      </a-layout-content>
    </a-layout>
  </a-layout>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import {
  MenuUnfoldOutlined,
  MenuFoldOutlined,
  SwapOutlined,
  FileTextOutlined,
  LogoutOutlined
} from '@ant-design/icons-vue'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const collapsed = ref(false)
const selectedKeys = computed(() => [route.name as string])

const userInitial = computed(() => {
  return authStore.email?.charAt(0).toUpperCase() || 'U'
})

const handleMenuClick = ({ key }: { key: string }) => {
  router.push({ name: key })
}

const handleLogout = () => {
  authStore.logout()
  router.push('/login')
}
</script>

<style scoped>
.layout {
  min-height: 100vh;
}

.sider {
  background: #001529;
}

.logo {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 20px;
  font-weight: bold;
  background: rgba(255, 255, 255, 0.1);
}

.header {
  background: white;
  padding: 0 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);
}

.trigger {
  font-size: 18px;
  cursor: pointer;
  transition: color 0.3s;
}

.trigger:hover {
  color: #1890ff;
}

.header-right {
  cursor: pointer;
}

.content {
  margin: 24px;
  padding: 24px;
  background: white;
  border-radius: 8px;
  min-height: calc(100vh - 64px - 48px);
}
</style>
