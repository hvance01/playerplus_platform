<template>
  <div class="faceswap-container">
    <a-row :gutter="24">
      <!-- Left: Upload list -->
      <a-col :span="6">
        <div class="panel">
          <div class="panel-header">
            <a-tabs v-model:activeKey="activeTab" size="small">
              <a-tab-pane key="uploads" tab="我的上传" />
              <a-tab-pane key="creations" tab="我的作品" />
            </a-tabs>
          </div>
          <div class="media-list">
            <div
              v-for="item in mediaList"
              :key="item.id"
              :class="['media-item', { active: selectedMedia?.id === item.id }]"
              @click="selectMedia(item)"
            >
              <img :src="item.thumbnail" :alt="item.name" />
              <span class="duration">{{ item.duration }}</span>
            </div>
          </div>
          <a-upload
            :show-upload-list="false"
            :before-upload="handleUpload"
            accept="video/*,image/*"
          >
            <a-button type="dashed" block>
              <upload-outlined /> 上传文件
            </a-button>
          </a-upload>
        </div>
      </a-col>

      <!-- Center: Preview -->
      <a-col :span="12">
        <div class="panel preview-panel">
          <div v-if="selectedMedia" class="preview-content">
            <video
              v-if="selectedMedia.type === 'video'"
              :src="selectedMedia.url"
              controls
              class="preview-video"
            />
            <img v-else :src="selectedMedia.url" class="preview-image" />
          </div>
          <a-empty v-else description="请选择或上传媒体文件" />
        </div>
      </a-col>

      <!-- Right: Face selection & Model -->
      <a-col :span="6">
        <div class="panel">
          <div class="panel-header">
            <h3>添加人脸开始换脸</h3>
          </div>

          <!-- Face slots -->
          <div class="face-grid">
            <div
              v-for="(slot, index) in faceSlots"
              :key="index"
              class="face-slot"
              @click="openFaceSelector(index)"
            >
              <template v-if="slot.face">
                <img :src="slot.face.thumbnail" />
                <a-button
                  type="text"
                  size="small"
                  class="remove-btn"
                  @click.stop="removeFace(index)"
                >
                  <close-outlined />
                </a-button>
              </template>
              <template v-else>
                <plus-outlined class="add-icon" />
              </template>
            </div>
          </div>

          <!-- Model selection -->
          <div class="model-section">
            <div class="section-header">
              <span>Model</span>
              <a-tooltip title="选择换脸算法模型">
                <question-circle-outlined />
              </a-tooltip>
            </div>
            <a-radio-group v-model:value="selectedModel" class="model-list">
              <a-radio-button
                v-for="model in models"
                :key="model.value"
                :value="model.value"
                class="model-item"
              >
                {{ model.label }}
                <a-tag v-if="model.tag" :color="model.tagColor" size="small">
                  {{ model.tag }}
                </a-tag>
              </a-radio-button>
            </a-radio-group>
          </div>

          <!-- HD Face toggle -->
          <div class="hd-section">
            <span>HD Face</span>
            <a-tag color="purple">PRO</a-tag>
            <a-switch v-model:checked="hdFace" :disabled="true" />
          </div>

          <!-- Create button -->
          <a-button
            type="primary"
            size="large"
            block
            :loading="processing"
            :disabled="!canCreate"
            @click="handleCreate"
          >
            <template #icon><thunderbolt-outlined /></template>
            Create
          </a-button>
        </div>
      </a-col>
    </a-row>

    <!-- Task status modal -->
    <a-modal
      v-model:open="taskModalVisible"
      title="处理进度"
      :footer="null"
      :closable="!processing"
      :maskClosable="!processing"
    >
      <div class="task-status">
        <a-spin v-if="currentTask?.status === 'processing'" />
        <check-circle-outlined v-else-if="currentTask?.status === 'completed'" class="success-icon" />
        <close-circle-outlined v-else-if="currentTask?.status === 'failed'" class="error-icon" />
        <p>{{ taskStatusText }}</p>
        <a-button
          v-if="currentTask?.status === 'completed'"
          type="primary"
          :href="currentTask?.result_url"
          target="_blank"
        >
          下载结果
        </a-button>
      </div>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { message } from 'ant-design-vue'
import {
  UploadOutlined,
  PlusOutlined,
  CloseOutlined,
  QuestionCircleOutlined,
  ThunderboltOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined
} from '@ant-design/icons-vue'
import { faceswapApi } from '@/api'

interface MediaItem {
  id: string
  name: string
  type: 'video' | 'image'
  url: string
  thumbnail: string
  duration?: string
}

interface FaceSlot {
  face: { id: string; thumbnail: string } | null
}

interface SwapTask {
  id: string
  status: string
  result_url?: string
  error?: string
}

const activeTab = ref('uploads')
const mediaList = ref<MediaItem[]>([])
const selectedMedia = ref<MediaItem | null>(null)
const faceSlots = ref<FaceSlot[]>([
  { face: null },
  { face: null },
  { face: null },
  { face: null }
])
const selectedModel = ref('shapefusion')
const hdFace = ref(false)
const processing = ref(false)
const taskModalVisible = ref(false)
const currentTask = ref<SwapTask | null>(null)

const models = [
  { value: 'shapefusion', label: 'ShapeFusion', tag: 'NEW', tagColor: 'green' },
  { value: 'shapetransformer_v3', label: 'ShapeTransformer V3.1' },
  { value: 'shapetransformer_v2', label: 'ShapeTransformer V2.0' },
  { value: 'shapekeeper_hd', label: 'ShapeKeeper HD' },
  { value: 'shapekeeper', label: 'ShapeKeeper' }
]

const activeFaces = computed(() => faceSlots.value.filter(s => s.face !== null))
const canCreate = computed(() => selectedMedia.value && activeFaces.value.length > 0)

const taskStatusText = computed(() => {
  if (!currentTask.value) return ''
  switch (currentTask.value.status) {
    case 'processing':
      return '正在处理中，请稍候...'
    case 'completed':
      return '处理完成！'
    case 'failed':
      return `处理失败: ${currentTask.value.error || '未知错误'}`
    default:
      return ''
  }
})

const selectMedia = (item: MediaItem) => {
  selectedMedia.value = item
}

const handleUpload = async (file: File) => {
  try {
    const { data } = await faceswapApi.upload(file)
    const newItem: MediaItem = {
      id: data.media_id,
      name: file.name,
      type: file.type.startsWith('video/') ? 'video' : 'image',
      url: URL.createObjectURL(file),
      thumbnail: URL.createObjectURL(file),
      duration: file.type.startsWith('video/') ? '--:--' : undefined
    }
    mediaList.value.unshift(newItem)
    selectedMedia.value = newItem
    message.success('上传成功')
  } catch (error) {
    message.error('上传失败')
  }
  return false
}

const openFaceSelector = (index: number) => {
  // TODO: Open face selector modal
  // For now, use a mock face
  const mockFace = {
    id: `face_${Date.now()}`,
    thumbnail: 'https://via.placeholder.com/80'
  }
  faceSlots.value[index].face = mockFace
}

const removeFace = (index: number) => {
  faceSlots.value[index].face = null
}

const handleCreate = async () => {
  if (!selectedMedia.value || activeFaces.value.length === 0) return

  processing.value = true
  taskModalVisible.value = true

  try {
    const faceIds = activeFaces.value.map(s => s.face!.id)
    const { data } = await faceswapApi.swap(selectedMedia.value.id, faceIds, selectedModel.value)
    currentTask.value = data

    // Poll for task status
    const pollInterval = setInterval(async () => {
      try {
        const { data: taskData } = await faceswapApi.getTask(currentTask.value!.id)
        currentTask.value = taskData

        if (taskData.status === 'completed' || taskData.status === 'failed') {
          clearInterval(pollInterval)
          processing.value = false
        }
      } catch (error) {
        clearInterval(pollInterval)
        processing.value = false
        currentTask.value = { ...currentTask.value!, status: 'failed', error: '获取状态失败' }
      }
    }, 2000)
  } catch (error) {
    processing.value = false
    currentTask.value = { id: '', status: 'failed', error: '创建任务失败' }
  }
}
</script>

<style scoped>
.faceswap-container {
  height: calc(100vh - 64px - 48px - 48px);
}

.panel {
  background: #f5f5f5;
  border-radius: 8px;
  padding: 16px;
  height: 100%;
}

.panel-header {
  margin-bottom: 16px;
}

.panel-header h3 {
  margin: 0;
  font-size: 14px;
  color: #666;
}

.media-list {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
  max-height: 400px;
  overflow-y: auto;
  margin-bottom: 16px;
}

.media-item {
  position: relative;
  aspect-ratio: 1;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  border: 2px solid transparent;
}

.media-item.active {
  border-color: #1890ff;
}

.media-item img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.media-item .duration {
  position: absolute;
  bottom: 4px;
  left: 4px;
  background: rgba(0, 0, 0, 0.6);
  color: white;
  font-size: 12px;
  padding: 2px 6px;
  border-radius: 4px;
}

.preview-panel {
  display: flex;
  align-items: center;
  justify-content: center;
  background: #000;
}

.preview-content {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.preview-video,
.preview-image {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}

.face-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  margin-bottom: 24px;
}

.face-slot {
  aspect-ratio: 1;
  border: 2px dashed #d9d9d9;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  position: relative;
  overflow: hidden;
  transition: all 0.3s;
}

.face-slot:hover {
  border-color: #1890ff;
}

.face-slot img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.face-slot .add-icon {
  font-size: 24px;
  color: #999;
}

.face-slot .remove-btn {
  position: absolute;
  top: 0;
  right: 0;
  background: rgba(0, 0, 0, 0.5);
  color: white;
  border-radius: 50%;
}

.model-section {
  margin-bottom: 16px;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  font-weight: 500;
}

.model-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.model-item {
  text-align: left;
  border-radius: 4px;
}

.hd-section {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 24px;
  padding: 12px;
  background: #fff;
  border-radius: 8px;
}

.task-status {
  text-align: center;
  padding: 24px;
}

.task-status .success-icon {
  font-size: 48px;
  color: #52c41a;
}

.task-status .error-icon {
  font-size: 48px;
  color: #ff4d4f;
}
</style>
