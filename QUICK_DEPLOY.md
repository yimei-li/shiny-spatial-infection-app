# 🚀 快速云部署指南

## ⚡ 5分钟发布到网站

### 🥇 Railway（最推荐）

1. **注册账户**：https://railway.app
2. **连接GitHub**：推送代码到GitHub仓库
3. **创建项目**：点击"Deploy from GitHub repo"
4. **选择仓库**：选择您的项目仓库
5. **自动部署**：Railway检测到Dockerfile后自动部署
6. **获得网址**：部署完成后获得公网访问地址

**成本**：每月$5免费额度

### 🥈 Render

1. **注册账户**：https://render.com
2. **创建Web Service**：连接GitHub仓库
3. **配置设置**：
   - Runtime: Docker
   - Dockerfile Path: `./Dockerfile.cloud`
4. **部署**：自动构建和部署
5. **访问**：通过Render提供的URL访问

**成本**：有限免费层

### 🥉 DigitalOcean

1. **注册账户**：https://cloud.digitalocean.com
2. **创建App**：选择GitHub仓库
3. **配置**：
   - Type: Web Service
   - Source: Dockerfile
   - HTTP Port: 3838
4. **部署**：自动构建和部署

**成本**：$200试用额度

## 📋 部署前准备

### 1. 推送到GitHub
```bash
git init
git add .
git commit -m "Initial commit"
git remote add origin https://github.com/你的用户名/viral-simulation-app.git
git push -u origin main
```

### 2. 选择Dockerfile
- **本地测试**：使用 `Dockerfile`
- **云部署**：使用 `Dockerfile.cloud`（支持动态端口）

### 3. 环境变量
大多数云平台会自动设置PORT环境变量，无需手动配置。

## 🎯 推荐流程

1. **本地测试**：
   ```bash
   ./deploy.sh  # 选择选项1
   ```

2. **云部署**：
   ```bash
   ./deploy.sh  # 选择选项2，查看详细指南
   ```

3. **访问应用**：通过云平台提供的URL访问

## 🔧 故障排除

### 构建失败
- 检查Dockerfile路径
- 确认所有文件已推送到GitHub
- 查看构建日志

### 端口问题
- 使用`Dockerfile.cloud`而不是`Dockerfile`
- 确保应用监听`0.0.0.0`而不是`127.0.0.1`

### 内存不足
- 选择更高配置的实例
- 考虑优化Docker镜像大小

## 💡 提示

- **Railway**：最简单，适合快速部署
- **Render**：免费层受限，但稳定
- **DigitalOcean**：最灵活，适合生产环境

现在您的病毒仿真应用可以通过以上任何平台发布到互联网！ 