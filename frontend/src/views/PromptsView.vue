<template>
  <div class="prompts-container">
    <a-tabs v-model:activeKey="activeTab" class="main-tabs">
      <a-tab-pane key="prompts" tab="Prompt 管理">
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
      </a-tab-pane>
      <a-tab-pane key="guide" tab="使用说明">
        <MarkdownViewer :content="guideContent" />
      </a-tab-pane>
    </a-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { PlusOutlined, FileTextOutlined } from '@ant-design/icons-vue'
import MarkdownViewer from '@/components/MarkdownViewer.vue'
import guideContent from '@/docs/prompts.md?raw'

const activeTab = ref('prompts')
</script>

<style scoped>
.prompts-container {
  padding: 24px;
}

.main-tabs {
  background: #fff;
  border-radius: 8px;
  padding: 0 16px;
}

.main-tabs :deep(.ant-tabs-nav) {
  margin-bottom: 0;
}

.main-tabs :deep(.ant-tabs-content-holder) {
  padding: 16px 0;
}
</style>
