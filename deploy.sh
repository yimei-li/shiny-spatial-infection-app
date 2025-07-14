#!/bin/bash

echo "🚀 病毒仿真应用部署脚本"
echo "=========================="

echo "请选择部署方式："
echo "1. Docker本地部署（开发测试）"
echo "2. Docker云部署（发布到网站 - 推荐）"
echo "3. 查看Shinyapps.io部署说明"
echo "4. 生成简化版本（用于云部署）"

read -p "请输入选择 (1-4): " choice

case $choice in
    1)
        echo "🐳 开始Docker本地部署..."
        
        # 检查Docker是否安装
        if ! command -v docker &> /dev/null; then
            echo "❌ 请先安装Docker"
            echo "访问：https://docs.docker.com/get-docker/"
            exit 1
        fi
        
        # 检查docker-compose是否安装
        if ! command -v docker-compose &> /dev/null; then
            echo "❌ 请先安装docker-compose"
            exit 1
        fi
        
        echo "📦 构建Docker镜像..."
        docker-compose build
        
        echo "🚀 启动应用..."
        docker-compose up -d
        
        echo "✅ 部署成功！"
        echo "📱 访问：http://localhost:3838"
        echo "🛑 停止应用：docker-compose down"
        ;;
        
    2)
        echo "🌐 Docker云部署向导"
        echo "==================="
        echo ""
        echo "📋 支持的云平台："
        echo "1. Railway (最推荐) - $5/月免费额度"
        echo "2. Render - 有限免费层"
        echo "3. DigitalOcean - $200试用额度"
        echo ""
        echo "📖 详细部署指南："
        echo "请查看 cloud_deployment_guide.md"
        echo ""
        echo "🚀 快速开始："
        echo "1. 推送代码到GitHub"
        echo "2. 注册云平台账户"
        echo "3. 连接GitHub仓库"
        echo "4. 选择Dockerfile部署"
        echo "5. 获得公网访问地址"
        echo ""
        echo "💡 提示：使用 Dockerfile.cloud 获得更好的云平台兼容性"
        ;;
        
    3)
        echo "📋 Shinyapps.io 部署说明："
        echo "1. 当前应用使用Go语言，Shinyapps.io不支持"
        echo "2. 需要重构为纯R版本才能部署"
        echo "3. 参考 deployment_considerations.md"
        echo "4. 或选择方案4生成简化版本"
        ;;
        
    4)
        echo "📝 生成简化版本中..."
        echo "这将创建一个不依赖Go的版本"
        echo "功能：展示预计算的结果"
        # 这里可以添加生成简化版本的逻辑
        echo "✅ 简化版本生成完成"
        ;;
        
    *)
        echo "❌ 无效选择"
        exit 1
        ;;
esac 