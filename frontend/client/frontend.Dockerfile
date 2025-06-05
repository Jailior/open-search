# Build
FROM node:20-alpine AS builder

WORKDIR /app

COPY ./frontend/client/package.json ./frontend/client/package-lock.json ./

RUN npm install

COPY frontend/client/ ./

RUN npm run build

# Serve
FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html

# Maybe add nginx configuaration here

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]