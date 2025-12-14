# VModel Video Face Swap - Development Plan

## API Research Results

### 1. video-face-detect API
- **Version ID**: `fa9317a2ad086f7633f4f9b38f35c82495b6c5f38fa2afbe32d9d9df8620b389`
- **Input**:
  ```json
  {
    "version": "fa9317a2...",
    "input": {
      "source": "https://example.com/video.mp4"
    }
  }
  ```
- **Output** (after task completion):
  ```json
  {
    "output": [{
      "id": "ODQVkD9DNWn",      // detect_id for swap
      "status": "succeed",
      "type": "face_detect",
      "faces": [
        {"id": 0, "link": "https://cdn.vmimgs.com/.../face0.jpg"},
        {"id": 1, "link": "https://cdn.vmimgs.com/.../face1.jpg"}
      ]
    }]
  }
  ```

### 2. video-multi-face-swap API
- **Version ID**: `8e960283784c5b58e5f67236757c40bb6796c85e3c733d060342bdf62f9f0c64`
- **Input**:
  ```json
  {
    "version": "8e960283...",
    "input": {
      "detect_id": "ODQVkD9DNWn",
      "face_map": "[{\"face_id\": 0, \"target\": \"https://example.com/new_face.jpg\"}]"
    }
  }
  ```
  Note: `face_map` is a **stringified JSON array**, not raw JSON
- **Output** (after task completion):
  ```json
  {
    "output": ["https://cdn.vmimgs.com/.../result_video.mp4"],
    "status": "succeeded"
  }
  ```

## Current Code Issues

### Issue 1: Status string mismatch
- Detection output uses `"succeed"`
- Task status uses `"succeeded"`
- Current code checks for both correctly ✓

### Issue 2: Output parsing
- Current code parses output correctly for both APIs ✓

### Issue 3: The code appears correct but needs verification
- Need to verify the actual API flow works end-to-end

## Development Plan

### Step 1: Update vmodel.go
No major changes needed - current implementation is correct. Minor improvements:
- Add better logging for debugging
- Ensure error messages are clear

### Step 2: Update vmodel_test.go
Add comprehensive tests:
- TestVModelDetectFaces - verify detect returns faces with thumbnails
- TestVModelCreateSwapTask - verify swap creates task successfully
- TestVModelGetTaskStatus - verify status polling works
- TestVModelFullSwapFlow - end-to-end test

### Step 3: Update handlers
Ensure handlers correctly:
- Parse detect results and return faces to frontend
- Accept face swap requests with detect_id and face mappings
- Return video URL when swap completes

### Step 4: Verify Frontend Integration
- Frontend should display detected face thumbnails
- User selects faces and uploads replacement images
- Frontend submits swap request with correct format

## API Flow Diagram

```
┌─────────────┐     ┌──────────────────┐     ┌─────────────────────┐
│   Frontend  │────▶│ POST /detect     │────▶│ video-face-detect   │
│             │     │ {video_url}      │     │ Returns: detect_id  │
│             │◀────│                  │◀────│ + faces[{id, link}] │
└─────────────┘     └──────────────────┘     └─────────────────────┘
      │
      │ User selects faces, uploads new face images
      ▼
┌─────────────┐     ┌──────────────────┐     ┌─────────────────────┐
│   Frontend  │────▶│ POST /swap       │────▶│ video-multi-face    │
│             │     │ {detect_id,      │     │ -swap               │
│             │     │  face_map}       │     │ Returns: task_id    │
└─────────────┘     └──────────────────┘     └─────────────────────┘
      │
      │ Poll for completion
      ▼
┌─────────────┐     ┌──────────────────┐     ┌─────────────────────┐
│   Frontend  │────▶│ GET /task/:id    │────▶│ Get task status     │
│   Shows     │◀────│                  │◀────│ Returns: video_url  │
│   Video     │     │                  │     │ when completed      │
└─────────────┘     └──────────────────┘     └─────────────────────┘
```

## Test Commands

```bash
# Test detect
curl -X POST https://api.vmodel.ai/api/tasks/v1/create \
  -H "Authorization: Bearer $VMODEL_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"version": "fa9317a2ad086f7633f4f9b38f35c82495b6c5f38fa2afbe32d9d9df8620b389", "input": {"source": "VIDEO_URL"}}'

# Test swap (after getting detect_id)
curl -X POST https://api.vmodel.ai/api/tasks/v1/create \
  -H "Authorization: Bearer $VMODEL_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"version": "8e960283784c5b58e5f67236757c40bb6796c85e3c733d060342bdf62f9f0c64", "input": {"detect_id": "DETECT_ID", "face_map": "[{\"face_id\": 0, \"target\": \"NEW_FACE_URL\"}]"}}'

# Check task status
curl -X GET https://api.vmodel.ai/api/tasks/v1/get/TASK_ID \
  -H "Authorization: Bearer $VMODEL_API_TOKEN"
```
