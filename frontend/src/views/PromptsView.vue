<template>
  <div class="prompts-container">
    <a-page-header title="Prompt 管理" sub-title="创建和管理 Prompt 模板">
      <template #extra>
        <a-button type="primary">
          <plus-outlined /> 新建 Prompt
        </a-button>
      </template>
    </a-page-header>

    <a-empty description="功能开发中，敬请期待">
      <template #image>
        <file-text-outlined style="font-size: 64px; color: #d9d9d9" />
      </template>
    </a-empty>
  </div>
</template>

<script setup lang="ts">
import { PlusOutlined, FileTextOutlined } from '@ant-design/icons-vue'
</script>

<style scoped>
.prompts-container {
  padding: 0;
}
</style>
