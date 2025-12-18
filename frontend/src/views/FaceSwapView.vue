<template>
  <div class="faceswap-container">
    <a-tabs v-model:activeKey="activeTab" class="main-tabs">
      <a-tab-pane key="faceswap" tab="视频换脸">
    <a-row :gutter="24">
      <!-- Left: Video Upload & Preview -->
      <a-col :span="14">
        <div class="panel video-panel">
          <!-- Step 1: Upload Video -->
          <div v-if="!videoUrl" class="upload-area">
            <a-upload
              :show-upload-list="false"
              :before-upload="handleVideoUpload"
              accept="video/*"
              :disabled="uploading"
            >
              <div class="upload-placeholder">
                <template v-if="uploading">
                  <a-progress type="circle" :percent="uploadProgress" :width="80" />
                  <p>正在上传视频...</p>
                </template>
                <template v-else>
                  <cloud-upload-outlined class="upload-icon" />
                  <p>点击或拖拽上传视频</p>
                  <p class="hint">支持 MP4, WebM 格式</p>
                </template>
              </div>
            </a-upload>
          </div>

          <!-- Step 2: Video Player with Frame Capture -->
          <div v-else class="video-section">
            <div class="video-wrapper">
              <video
                ref="videoRef"
                :src="videoUrl"
                @loadedmetadata="onVideoLoaded"
                @timeupdate="onTimeUpdate"
                controls
                class="video-player"
              />
              <!-- Face boxes overlay -->
              <div v-if="detectedFaces.length" class="face-overlay">
                <div
                  v-for="face in detectedFaces"
                  :key="face.index"
                  :class="['face-box', { selected: selectedFaceIndices.includes(face.index) }]"
                  :style="getFaceBoxStyle(face)"
                  @click="toggleFaceSelection(face.index)"
                >
                  <span class="face-label">{{ face.index + 1 }}</span>
                </div>
              </div>
            </div>

            <!-- Video controls -->
            <div class="video-controls">
              <a-slider
                v-model:value="currentTime"
                :max="duration"
                :step="0.1"
                :tip-formatter="formatTime"
                @change="seekVideo"
              />

              <!-- Detection progress -->
              <div v-if="detecting" class="detect-progress">
                <a-spin size="small" />
                <span>正在分析视频中的人脸，请稍候...</span>
              </div>

              <div class="control-buttons">
                <a-button @click="clearVideo" danger :disabled="detecting">
                  <delete-outlined /> 重新上传
                </a-button>
                <a-button
                  type="primary"
                  @click="detectFacesFromVideo"
                  :loading="detecting"
                  :disabled="!videoPublicUrl || detecting || uploading"
                >
                  <scan-outlined /> {{ detecting ? '检测中...' : '检测人脸' }}
                </a-button>
              </div>
            </div>

            <!-- Detection result info -->
            <a-alert
              v-if="detectedFaces.length"
              type="success"
              :message="`检测到 ${detectedFaces.length} 张人脸，点击视频中的人脸框选择要替换的人脸`"
              show-icon
              class="detect-alert"
            />
          </div>
        </div>
      </a-col>

      <!-- Right: Face Selection & Replacement -->
      <a-col :span="10">
        <div class="panel face-panel">
          <h3 class="panel-title">人脸替换设置</h3>

          <!-- No faces detected yet -->
          <a-empty
            v-if="!detectedFaces.length"
            description="请先上传视频并检测人脸"
          />

          <!-- Face selection and replacement -->
          <div v-else class="face-config">
            <!-- Selected faces for replacement -->
            <div class="section">
              <h4>选择要替换的人脸 (已选 {{ selectedFaceIndices.length }})</h4>
              <div class="face-grid">
                <div
                  v-for="face in detectedFaces"
                  :key="face.index"
                  :class="['face-card', { selected: selectedFaceIndices.includes(face.index) }]"
                  @click="toggleFaceSelection(face.index)"
                >
                  <div class="face-thumbnail">
                    <img v-if="face.thumbnail" :src="face.thumbnail" />
                    <div v-else class="face-number">{{ face.index + 1 }}</div>
                  </div>
                  <check-circle-filled
                    v-if="selectedFaceIndices.includes(face.index)"
                    class="check-icon"
                  />
                </div>
              </div>
            </div>

            <!-- Upload replacement faces -->
            <div v-if="selectedFaceIndices.length" class="section">
              <h4>上传替换人脸照片</h4>
              <div class="replacement-list">
                <div
                  v-for="faceIndex in selectedFaceIndices"
                  :key="faceIndex"
                  class="replacement-row"
                >
                  <div class="original-face">
                    <span class="label">人脸 {{ faceIndex + 1 }}</span>
                  </div>
                  <arrow-right-outlined class="arrow" />
                  <a-upload
                    :show-upload-list="false"
                    :before-upload="(file: File) => handleReplacementUpload(faceIndex, file)"
                    accept="image/*"
                  >
                    <div
                      v-if="replacementFaces[faceIndex]"
                      class="uploaded-face"
                    >
                      <img :src="replacementFaces[faceIndex].preview" />
                      <close-circle-filled
                        class="remove-btn"
                        @click.stop="removeReplacementFace(faceIndex)"
                      />
                    </div>
                    <a-button v-else>
                      <upload-outlined /> 上传新脸
                    </a-button>
                  </a-upload>
                </div>
              </div>
            </div>

            <!-- Options -->
            <div class="section">
              <h4>选项</h4>
              <a-checkbox v-model:checked="faceEnhance">
                HD 人脸增强 (消耗更多积分)
              </a-checkbox>
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
              开始换脸
            </a-button>
          </div>
        </div>
      </a-col>
    </a-row>

      </a-tab-pane>
      <a-tab-pane key="guide" tab="使用说明">
        <MarkdownViewer :content="guideContent" />
      </a-tab-pane>
    </a-tabs>

    <!-- Task Progress Modal -->
    <a-modal
      v-model:open="taskModalVisible"
      title="换脸处理进度"
      :footer="null"
      :closable="!processing"
      :maskClosable="!processing"
      width="400px"
    >
      <div class="task-status">
        <a-spin v-if="currentTask?.status === 'queuing' || currentTask?.status === 'processing' || currentTask?.status === 'transferring'" size="large" />
        <check-circle-outlined v-else-if="currentTask?.status === 'completed'" class="success-icon" />
        <close-circle-outlined v-else-if="currentTask?.status === 'failed'" class="error-icon" />

        <p class="status-text">{{ taskStatusText }}</p>

        <div v-if="currentTask?.status === 'completed'" class="result-actions">
          <a-button
            type="primary"
            :href="currentTask.result_url"
            target="_blank"
            :disabled="currentTask.transfer_status === 'failed'"
          >
            <download-outlined /> 下载结果视频
          </a-button>
          <a-button @click="resetAll">处理新视频</a-button>
        </div>

        <a-button v-if="currentTask?.status === 'failed'" @click="taskModalVisible = false">
          关闭
        </a-button>
      </div>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { message } from 'ant-design-vue'
import MarkdownViewer from '@/components/MarkdownViewer.vue'
import guideContent from '@/docs/faceswap.md?raw'
import {
  CloudUploadOutlined,
  ScanOutlined,
  DeleteOutlined,
  UploadOutlined,
  ArrowRightOutlined,
  CheckCircleFilled,
  CloseCircleFilled,
  ThunderboltOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  DownloadOutlined
} from '@ant-design/icons-vue'
import { faceswapApiV2, type DetectedFace, type TaskStatusResponse } from '@/api'

// --- State ---
const activeTab = ref('faceswap')
const videoRef = ref<HTMLVideoElement | null>(null)
const videoUrl = ref('')
const videoPublicUrl = ref('')  // URL for API access
const duration = ref(0)
const currentTime = ref(0)
const frameImageUrl = ref('')   // Frame used for detection
const detectId = ref('')        // VModel detect_id for swap

const uploading = ref(false)
const uploadProgress = ref(0)
const detecting = ref(false)
const processing = ref(false)

const detectedFaces = ref<DetectedFace[]>([])
const selectedFaceIndices = ref<number[]>([])
const replacementFaces = ref<Record<number, { url: string; preview: string }>>({})
const faceEnhance = ref(false)

const taskModalVisible = ref(false)
// Reuse TaskStatusResponse type from API for type safety
const currentTask = ref<TaskStatusResponse['data'] | null>(null)

// --- Computed ---
const canCreate = computed(() => {
  if (selectedFaceIndices.value.length === 0) return false
  // Check all selected faces have replacement
  return selectedFaceIndices.value.every(idx => replacementFaces.value[idx])
})

const taskStatusText = computed(() => {
  if (!currentTask.value) return ''
  switch (currentTask.value.status) {
    case 'queuing':
      return '任务排队中...'
    case 'processing':
      return '正在处理视频，请稍候...'
    case 'transferring':
      return '正在转存视频到服务器...'
    case 'completed':
      // Check if transfer failed
      if (currentTask.value.transfer_status === 'failed') {
        return '处理完成，但视频转存失败，可能无法正常下载'
      }
      return '处理完成！'
    case 'failed':
      return `处理失败: ${currentTask.value.error || '未知错误'}`
    default:
      return ''
  }
})

// --- Video Handlers ---
const handleVideoUpload = async (file: File) => {
  uploading.value = true
  uploadProgress.value = 0

  // Store file for later use
  const localPreviewUrl = URL.createObjectURL(file)

  try {
    // Upload to server with progress tracking
    const { data } = await faceswapApiV2.uploadMedia(file, (progress) => {
      uploadProgress.value = progress
    })

    // Set video URL only after upload completes
    videoUrl.value = localPreviewUrl
    videoPublicUrl.value = data.url
    message.success('视频上传成功')
  } catch (error) {
    message.error('视频上传失败')
    URL.revokeObjectURL(localPreviewUrl) // Clean up
  } finally {
    uploading.value = false
    uploadProgress.value = 0
  }
  return false
}

const onVideoLoaded = () => {
  if (videoRef.value) {
    duration.value = videoRef.value.duration
    // Auto detect faces from video after upload completes
    if (videoPublicUrl.value) {
      detectFacesFromVideo()
    }
  }
}

const onTimeUpdate = () => {
  if (videoRef.value) {
    currentTime.value = videoRef.value.currentTime
  }
}

const seekVideo = (time: number) => {
  if (videoRef.value) {
    videoRef.value.currentTime = time
  }
}

const formatTime = (seconds: number) => {
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

const clearVideo = () => {
  videoUrl.value = ''
  videoPublicUrl.value = ''
  detectedFaces.value = []
  selectedFaceIndices.value = []
  replacementFaces.value = {}
  frameImageUrl.value = ''
  detectId.value = ''
}

// --- Face Detection from Video URL ---
const detectFacesFromVideo = async () => {
  if (!videoPublicUrl.value || detecting.value) return

  detecting.value = true
  try {
    // Call API with video URL to detect ALL faces in the entire video
    const { data } = await faceswapApiV2.detectFaces(videoPublicUrl.value)

    if (data.code !== 0) {
      throw new Error(data.msg || 'Detection failed')
    }

    if (!data.data?.faces.length) {
      message.warning('视频中未检测到人脸，请确保视频中有清晰的人脸画面')
      return
    }

    detectedFaces.value = data.data.faces
    frameImageUrl.value = data.data.frame_image
    detectId.value = data.data.detect_id || ''
    selectedFaceIndices.value = []
    replacementFaces.value = {}

    message.success(`检测到 ${data.data.faces.length} 张人脸，请选择要替换的人脸`)
  } catch (error: any) {
    message.error('人脸检测失败: ' + (error.message || '未知错误'))
  } finally {
    detecting.value = false
  }
}

const getFaceBoxStyle = (face: DetectedFace) => {
  if (!face.bbox || !videoRef.value) return {}

  const video = videoRef.value
  const [x, y, w, h] = face.bbox

  // Convert to percentage for responsive positioning
  const scaleX = video.clientWidth / video.videoWidth
  const scaleY = video.clientHeight / video.videoHeight

  return {
    left: `${x * scaleX}px`,
    top: `${y * scaleY}px`,
    width: `${w * scaleX}px`,
    height: `${h * scaleY}px`
  }
}

// --- Face Selection ---
const toggleFaceSelection = (index: number) => {
  const idx = selectedFaceIndices.value.indexOf(index)
  if (idx === -1) {
    selectedFaceIndices.value.push(index)
  } else {
    selectedFaceIndices.value.splice(idx, 1)
    delete replacementFaces.value[index]
  }
}

const handleReplacementUpload = async (faceIndex: number, file: File) => {
  try {
    const preview = URL.createObjectURL(file)
    const { data } = await faceswapApiV2.uploadFace(file)

    replacementFaces.value[faceIndex] = {
      url: data.url,
      preview
    }
    message.success('人脸照片上传成功')
  } catch (error) {
    message.error('上传失败')
  }
  return false
}

const removeReplacementFace = (faceIndex: number) => {
  delete replacementFaces.value[faceIndex]
}

// --- Create Task ---
const handleCreate = async () => {
  if (!canCreate.value) return

  processing.value = true
  taskModalVisible.value = true

  try {
    // Build face swaps array
    const faceSwaps = selectedFaceIndices.value.map(idx => {
      const face = detectedFaces.value[idx]
      return {
        source_image_url: replacementFaces.value[idx].url,
        face_id: face.face_id,                    // VModel: face ID
        landmarks_str: face.landmarks_str || ''   // Legacy (optional)
      }
    })

    // Create task
    const { data } = await faceswapApiV2.createSwapTask({
      target_video_url: videoPublicUrl.value,
      detect_id: detectId.value,           // VModel: detection ID
      frame_image_url: frameImageUrl.value, // Legacy (optional)
      face_swaps: faceSwaps,
      face_enhance: faceEnhance.value
    })

    if (data.code !== 0) {
      throw new Error(data.msg || 'Failed to create task')
    }

    currentTask.value = {
      task_id: data.data!.task_id,
      status: data.data!.status as any
    }

    // Start polling
    pollTaskStatus(data.data!.task_id)
  } catch (error: any) {
    processing.value = false
    currentTask.value = {
      task_id: '',
      status: 'failed',
      error: error.message || '创建任务失败'
    }
  }
}

const pollTaskStatus = async (taskId: string) => {
  const poll = async () => {
    try {
      const { data } = await faceswapApiV2.getTaskStatus(taskId)

      if (data.code !== 0) {
        throw new Error(data.msg)
      }

      currentTask.value = data.data!

      const taskStatus = data.data!.status

      // Stop polling when task is failed or completed
      if (taskStatus === 'failed') {
        processing.value = false
        return
      }

      if (taskStatus === 'completed') {
        // Task fully completed (including transfer)
        processing.value = false
        return
      }

      // Continue polling for queuing, processing, or transferring
      // Use faster polling for transferring status
      const pollInterval = taskStatus === 'transferring' ? 2000 : 3000
      setTimeout(poll, pollInterval)
    } catch (error: any) {
      processing.value = false
      currentTask.value = {
        task_id: taskId,
        status: 'failed',
        error: error.message || '获取状态失败'
      }
    }
  }

  poll()
}

const resetAll = () => {
  clearVideo()
  taskModalVisible.value = false
  currentTask.value = null
}
</script>

<style scoped>
.faceswap-container {
  padding: 24px;
  min-height: calc(100vh - 64px);
}

.panel {
  background: #fff;
  border-radius: 8px;
  padding: 20px;
  height: 100%;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.panel-title {
  margin: 0 0 20px 0;
  font-size: 16px;
  font-weight: 600;
}

/* Upload Area */
.upload-area {
  height: 400px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.upload-placeholder {
  text-align: center;
  padding: 60px;
  border: 2px dashed #d9d9d9;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s;
  width: 100%;
}

.upload-placeholder:hover {
  border-color: #1890ff;
}

.upload-icon {
  font-size: 48px;
  color: #999;
  margin-bottom: 16px;
}

.upload-placeholder .hint {
  color: #999;
  font-size: 12px;
}

/* Video Section */
.video-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.video-wrapper {
  position: relative;
  background: #000;
  border-radius: 8px;
  overflow: hidden;
}

.video-player {
  width: 100%;
  max-height: 400px;
  display: block;
}

.face-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  pointer-events: none;
}

.face-box {
  position: absolute;
  border: 2px solid #52c41a;
  border-radius: 4px;
  cursor: pointer;
  pointer-events: auto;
  transition: all 0.2s;
}

.face-box:hover {
  border-color: #1890ff;
  background: rgba(24, 144, 255, 0.1);
}

.face-box.selected {
  border-color: #1890ff;
  border-width: 3px;
  background: rgba(24, 144, 255, 0.2);
}

.face-label {
  position: absolute;
  top: -20px;
  left: 50%;
  transform: translateX(-50%);
  background: #52c41a;
  color: white;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
}

.face-box.selected .face-label {
  background: #1890ff;
}

.video-controls {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.control-buttons {
  display: flex;
  gap: 12px;
  justify-content: flex-end;
}

.detect-progress {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 12px;
  background: #e6f7ff;
  border-radius: 4px;
  color: #1890ff;
}

.detect-alert {
  margin-top: 8px;
}

/* Face Config Panel */
.face-config {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.section h4 {
  margin: 0 0 12px 0;
  font-size: 14px;
  color: #666;
}

.face-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
}

.face-card {
  position: relative;
  aspect-ratio: 1;
  border: 2px solid #d9d9d9;
  border-radius: 8px;
  cursor: pointer;
  overflow: hidden;
  transition: all 0.2s;
}

.face-card:hover {
  border-color: #1890ff;
}

.face-card.selected {
  border-color: #1890ff;
  border-width: 3px;
}

.face-thumbnail {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f5f5f5;
}

.face-thumbnail img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.face-number {
  font-size: 24px;
  font-weight: bold;
  color: #999;
}

.face-card .check-icon {
  position: absolute;
  top: 4px;
  right: 4px;
  font-size: 20px;
  color: #1890ff;
}

/* Replacement Section */
.replacement-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.replacement-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.original-face {
  width: 60px;
  height: 60px;
  border-radius: 8px;
  background: #f0f0f0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.original-face .label {
  font-size: 12px;
  color: #666;
}

.arrow {
  color: #999;
}

.uploaded-face {
  position: relative;
  width: 60px;
  height: 60px;
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
}

.uploaded-face img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.uploaded-face .remove-btn {
  position: absolute;
  top: -6px;
  right: -6px;
  font-size: 18px;
  color: #ff4d4f;
  background: white;
  border-radius: 50%;
}

/* Task Modal */
.task-status {
  text-align: center;
  padding: 24px;
}

.task-status .success-icon {
  font-size: 64px;
  color: #52c41a;
}

.task-status .error-icon {
  font-size: 64px;
  color: #ff4d4f;
}

.status-text {
  margin: 16px 0;
  font-size: 16px;
}

.result-actions {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

/* Tabs */
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
