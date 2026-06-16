# ── Stage 1: build ────────────────────────────────────────────────────────────
FROM node:20-alpine AS builder
WORKDIR /app

# Install dependencies first (cached layer).
COPY package.json package-lock.json ./
RUN npm ci

# Copy source and build.
COPY . .
RUN npm run build
# Output is in /app/dist

# ── Stage 2: serve ────────────────────────────────────────────────────────────
FROM nginx:1.27-alpine

# Copy the built assets.
COPY --from=builder /app/dist /usr/share/nginx/html

# Custom nginx config: serve SPA and proxy /ws to the backend service.
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
