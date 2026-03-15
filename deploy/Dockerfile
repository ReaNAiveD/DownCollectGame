# Stage 1: Build frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app/src/frontend
COPY src/frontend/package.json src/frontend/package-lock.json ./
RUN npm ci
COPY src/frontend/ ./
RUN node node_modules/vite/bin/vite.js build

# Stage 2: Build Go server
FROM golang:1.25-alpine AS backend-builder
WORKDIR /app/src/backend
COPY src/backend/go.mod src/backend/go.sum ./
RUN go mod download
COPY src/backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o dist/server ./cmd/server/

# Stage 3: Final image
FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=backend-builder /app/src/backend/dist/server .
COPY --from=frontend-builder /app/src/frontend/dist ./public
ENV PORT=8080
EXPOSE 8080
CMD ["./server"]
