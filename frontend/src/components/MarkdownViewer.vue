<template>
  <div class="markdown-viewer" v-html="renderedContent"></div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { marked } from 'marked'
import DOMPurify from 'dompurify'

const props = defineProps<{
  content: string
}>()

const renderedContent = computed(() => {
  return DOMPurify.sanitize(marked(props.content) as string)
})
</script>

<style scoped>
.markdown-viewer {
  padding: 24px;
  line-height: 1.8;
  color: #333;
}

.markdown-viewer :deep(h1) {
  font-size: 28px;
  font-weight: 600;
  margin: 0 0 24px 0;
  padding-bottom: 12px;
  border-bottom: 1px solid #e8e8e8;
}

.markdown-viewer :deep(h2) {
  font-size: 22px;
  font-weight: 600;
  margin: 32px 0 16px 0;
  color: #1890ff;
}

.markdown-viewer :deep(h3) {
  font-size: 18px;
  font-weight: 600;
  margin: 24px 0 12px 0;
}

.markdown-viewer :deep(h4) {
  font-size: 16px;
  font-weight: 600;
  margin: 20px 0 10px 0;
}

.markdown-viewer :deep(p) {
  margin: 12px 0;
}

.markdown-viewer :deep(ul),
.markdown-viewer :deep(ol) {
  padding-left: 24px;
  margin: 12px 0;
}

.markdown-viewer :deep(li) {
  margin: 8px 0;
}

.markdown-viewer :deep(code) {
  background: #f5f5f5;
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 14px;
}

.markdown-viewer :deep(pre) {
  background: #f5f5f5;
  padding: 16px;
  border-radius: 8px;
  overflow-x: auto;
}

.markdown-viewer :deep(pre code) {
  background: none;
  padding: 0;
}

.markdown-viewer :deep(blockquote) {
  border-left: 4px solid #1890ff;
  padding-left: 16px;
  margin: 16px 0;
  color: #666;
}

.markdown-viewer :deep(table) {
  width: 100%;
  border-collapse: collapse;
  margin: 16px 0;
}

.markdown-viewer :deep(th),
.markdown-viewer :deep(td) {
  border: 1px solid #e8e8e8;
  padding: 12px;
  text-align: left;
}

.markdown-viewer :deep(th) {
  background: #fafafa;
  font-weight: 600;
}

.markdown-viewer :deep(hr) {
  border: none;
  border-top: 1px solid #e8e8e8;
  margin: 24px 0;
}

.markdown-viewer :deep(a) {
  color: #1890ff;
  text-decoration: none;
}

.markdown-viewer :deep(a:hover) {
  text-decoration: underline;
}
</style>
