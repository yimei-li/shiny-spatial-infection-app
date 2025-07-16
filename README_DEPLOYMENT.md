# 🚀 病毒仿真应用发布指南

## 📋 发布选项对比

| 方案 | 难度 | 支持所有功能 | 成本 | 推荐度 |
|------|------|-------------|------|--------|
| Docker本地 | ⭐⭐ | ✅ | 免费 | ⭐⭐⭐⭐⭐ |
| 自建服务器 | ⭐⭐⭐ | ✅ | 低-中 | ⭐⭐⭐⭐ |
| Shinyapps.io | ⭐ | ❌ | 免费/付费 | ⭐⭐ |
| 云平台Docker | ⭐⭐⭐⭐ | ✅ | 中-高 | ⭐⭐⭐ |

## 🐳 方案1：Docker部署（推荐）

### 优点
- ✅ 支持所有功能（Go + ImageMagick）
- ✅ 环境一致性
- ✅ 易于部署和管理

### 步骤
```bash
# 1. 运行部署脚本
./deploy.sh

# 2. 选择选项1 - Docker部署

# 3. 访问应用
open http://localhost:3838
```

### 手动部署
```bash
# 构建镜像
docker-compose build

# 启动应用
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止应用
docker-compose down
```

## 🌐 方案2：Shinyapps.io（需要重构）

### 限制
- ❌ 不支持Go语言
- ❌ 不支持系统调用
- ❌ 文件写入限制

### 解决方案
1. 将Go逻辑重写为R代码
2. 使用预计算结果
3. 简化为静态展示

### 部署步骤（重构后）
```r
# 安装rsconnect包
install.packages("rsconnect")

# 配置账户（从shinyapps.io获取）
rsconnect::setAccountInfo(
  name='你的用户名',
  token='你的token', 
  secret='你的secret'
)

# 部署应用
rsconnect::deployApp(
  appDir = ".", 
  appName = "viral-simulation-app"
)
```

## 🔧 方案3：自建服务器

### 系统要求
- Ubuntu/CentOS Linux
- R >= 4.0
- Go >= 1.18
- ImageMagick

### 安装步骤
```bash
# 1. 安装R和Shiny Server
sudo apt-get update
sudo apt-get install r-base
# ... (安装Shiny Server)

# 2. 安装Go
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz

# 3. 安装ImageMagick
sudo apt-get install imagemagick

# 4. 部署应用文件
sudo cp -r * /srv/shiny-server/viral-sim/
```

## 🚀 快速开始

```bash
# 1. 克隆或下载项目文件
# 2. 进入项目目录
cd 36_Shinny_add_video

# 3. 运行部署脚本
./deploy.sh

# 4. 按提示选择部署方式
```

## 📁 项目文件说明

- `main_app.R` - 主应用文件
- `gif_generator.R` - GIF生成功能
- `mdbk_small_vero_0716.go` - Go仿真核心
- `Dockerfile` - Docker镜像配置
- `docker-compose.yml` - Docker编排文件
- `deploy.sh` - 一键部署脚本

## 🆘 故障排除

### Docker相关
```bash
# 查看容器状态
docker-compose ps

# 查看日志
docker-compose logs viral-simulation

# 重启容器
docker-compose restart
```

### 端口冲突
```bash
# 修改docker-compose.yml中的端口映射
ports:
  - "3839:3838"  # 改为其他端口
```

## 📞 技术支持

如果遇到部署问题：
1. 查看 `deployment_considerations.md`
2. 检查Docker日志
3. 确认所有依赖已安装 