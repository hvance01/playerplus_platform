# Multi-stage build for PlayerPlus Platform
# Stage 1: Build frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN npm install -g pnpm && pnpm install --frozen-lockfile
COPY frontend/ ./
RUN pnpm build

# Stage 2: Build backend
FROM golang:1.22-alpine AS backend-builder
WORKDIR /app
COPY backend/ ./backend/
COPY --from=frontend-builder /app/frontend/dist ./backend/internal/handler/dist/
WORKDIR /app/backend
RUN go build -o server ./cmd/server

# Stage 3: Final image
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=backend-builder /app/backend/server ./server
EXPOSE 8080
CMD ["./server"]
