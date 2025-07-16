# 🌐 Docker云部署指南

## 📋 云平台对比

| 平台 | 难度 | 成本 | 免费额度 | 推荐度 |
|------|------|------|----------|--------|
| **Railway** | ⭐ | 低 | $5/月免费 | ⭐⭐⭐⭐⭐ |
| **Render** | ⭐⭐ | 低 | 有限免费 | ⭐⭐⭐⭐ |
| **DigitalOcean** | ⭐⭐⭐ | 中 | $200试用 | ⭐⭐⭐⭐ |
| **AWS/Azure/GCP** | ⭐⭐⭐⭐ | 高 | 有免费层 | ⭐⭐⭐ |
| **Heroku** | ⭐⭐ | 中-高 | 无免费 | ⭐⭐ |

## 🚀 方案1：Railway（最推荐）

### 优点
- ✅ 支持Docker
- ✅ 自动HTTPS
- ✅ 简单部署
- ✅ $5/月免费额度

### 部署步骤
1. **注册Railway**：访问 https://railway.app
2. **连接GitHub**：上传项目到GitHub
3. **创建项目**：选择"Deploy from GitHub repo"
4. **自动部署**：Railway自动检测Dockerfile并部署

### 配置文件
```bash
# railway.json (可选)
{
  "build": {
    "builder": "DOCKERFILE"
  },
  "deploy": {
    "startCommand": "docker-compose up",
    "healthcheckPath": "/",
    "healthcheckTimeout": 100
  }
}
```

## 🔧 方案2：Render

### 优点
- ✅ 免费层可用
- ✅ 简单配置
- ✅ 自动SSL

### 部署步骤
1. **注册Render**：访问 https://render.com
2. **连接GitHub**：链接你的仓库
3. **创建Web Service**：
   - Runtime: Docker
   - Build Command: `docker build -t app .`
   - Start Command: `docker run -p 10000:3838 app`

## 💻 方案3：DigitalOcean App Platform

### 优点
- ✅ 强大的基础设施
- ✅ 容易扩展
- ✅ $200免费试用

### 部署步骤
1. **创建DigitalOcean账户**
2. **创建App**：选择GitHub仓库
3. **配置**：
   - Type: Web Service
   - Source: Dockerfile
   - HTTP Port: 3838

## 🏗️ 准备云部署

### 1. 修改Dockerfile（云优化版）
```dockerfile
# 云部署优化版Dockerfile
FROM rocker/shiny:latest

# 安装系统依赖
RUN apt-get update && apt-get install -y \
    wget \
    curl \
    imagemagick \
    libmagick++-dev \
    && rm -rf /var/lib/apt/lists/*

# 安装Go语言
RUN wget -O go.tar.gz https://go.dev/dl/go1.21.0.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go.tar.gz \
    && rm go.tar.gz

ENV PATH=$PATH:/usr/local/go/bin

# 安装R包
RUN R -e "install.packages(c('shiny', 'shinyjs'), repos='https://cloud.r-project.org/')"

# 创建应用目录
WORKDIR /app

# 复制应用文件
COPY . .

# 创建www目录
RUN mkdir -p www

# 设置权限
RUN chmod -R 755 /app

# 暴露端口（云平台通常使用环境变量）
EXPOSE $PORT

# 启动命令（支持动态端口）
CMD ["sh", "-c", "R -e \"library(shiny); shiny::runApp('main_app.R', port=as.numeric(Sys.getenv('PORT', '3838')), host='0.0.0.0')\""]
```

### 2. 创建启动脚本
```bash
#!/bin/bash
# start.sh
export PORT=${PORT:-3838}
R -e "library(shiny); shiny::runApp('main_app.R', port=as.numeric(Sys.getenv('PORT')), host='0.0.0.0')"
```

### 3. 环境变量配置
```bash
# .env (本地测试用)
PORT=3838
SHINY_LOG_STDERR=1
```

## 📤 部署流程

### GitHub准备
```bash
# 1. 初始化Git（如果还没有）
git init
git add .
git commit -m "Initial commit"

# 2. 创建GitHub仓库并推送
git remote add origin https://github.com/你的用户名/viral-simulation-app.git
git push -u origin main
```

### Railway部署
```bash
# 1. 安装Railway CLI（可选）
npm install -g @railway/cli

# 2. 登录并部署
railway login
railway link
railway up
```

## 🔧 云部署配置文件

### docker-compose.cloud.yml
```yaml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "${PORT:-3838}:${PORT:-3838}"
    environment:
      - PORT=${PORT:-3838}
      - SHINY_LOG_STDERR=1
    restart: unless-stopped
```

### .dockerignore
```
.git
.gitignore
README*.md
Dockerfile*
docker-compose*
*.log
output/
deployment_considerations.md
deploy_setup.R
```

## 🌍 域名和HTTPS

### 自定义域名（Railway）
1. 在Railway项目设置中添加自定义域名
2. 配置DNS指向Railway提供的地址
3. 自动获得SSL证书

### 自定义域名（Render）
1. 在Render设置中添加自定义域名
2. 更新DNS记录
3. 自动SSL配置

## 📊 性能优化

### 1. 多阶段构建（减小镜像大小）
```dockerfile
# 构建阶段
FROM golang:1.21 AS go-builder
WORKDIR /build
COPY mdbk_small_vero_0716.go .
RUN go build -o simulation mdbk_small_vero_0716.go

# 运行阶段
FROM rocker/shiny:latest
COPY --from=go-builder /build/simulation /usr/local/bin/
# ... 其他配置
```

### 2. 缓存优化
```dockerfile
# 先复制依赖文件，利用Docker缓存
COPY DESCRIPTION .
RUN R -e "install.packages(...)"

# 再复制应用代码
COPY . .
```

## 🚀 一键云部署脚本

```bash
#!/bin/bash
# deploy_cloud.sh

echo "🌐 云部署向导"
echo "=============="

echo "选择云平台："
echo "1. Railway (推荐)"
echo "2. Render" 
echo "3. DigitalOcean"

read -p "请选择 (1-3): " platform

case $platform in
    1)
        echo "🚂 Railway部署"
        echo "1. 访问 https://railway.app"
        echo "2. 连接GitHub仓库"
        echo "3. 选择项目并部署"
        echo "4. 访问提供的URL"
        ;;
    2)
        echo "🎨 Render部署"
        echo "1. 访问 https://render.com"
        echo "2. 创建Web Service"
        echo "3. 连接GitHub仓库"
        echo "4. 设置Docker配置"
        ;;
    3)
        echo "🌊 DigitalOcean部署"
        echo "1. 访问 https://cloud.digitalocean.com"
        echo "2. 创建App Platform应用"
        echo "3. 连接GitHub仓库"
        echo "4. 配置Docker设置"
        ;;
esac

echo ""
echo "📋 部署前检查清单："
echo "✅ 代码已推送到GitHub"
echo "✅ Dockerfile已优化"
echo "✅ 端口配置正确"
echo "✅ 环境变量设置"
```

## 💰 成本估算

### Railway
- 免费：$5/月额度
- 付费：$0.000463/GB-hour RAM + $0.000231/vCPU-hour

### Render
- 免费：512MB RAM，有睡眠限制
- 付费：$7/月起（512MB RAM）

### DigitalOcean
- 基础：$5/月（512MB RAM）
- 标准：$12/月（1GB RAM）

现在您的应用可以通过任何云平台发布到互联网了！推荐从Railway开始，因为它最简单且有免费额度。 