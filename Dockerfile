# 云部署优化版Dockerfile
FROM rocker/shiny:latest

# 安装系统依赖
RUN apt-get update && apt-get install -y \
    wget \
    curl \
    imagemagick \
    libmagick++-dev \
    libcurl4-openssl-dev \
    && rm -rf /var/lib/apt/lists/*

# 安装Go语言
RUN wget -O go.tar.gz https://go.dev/dl/go1.21.0.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go.tar.gz \
    && rm go.tar.gz

ENV PATH=$PATH:/usr/local/go/bin

# 安装R包（提前安装所有依赖，避免运行时安装）
RUN R -e "install.packages(c('shiny', 'shinyjs', 'magick', 'curl'), repos='https://cloud.r-project.org/', dependencies=TRUE)"

# 创建应用目录
WORKDIR /app

# 复制应用文件
COPY main_app.R .
COPY gif_generator.R .
COPY mdbk_small_vero_0713.go .
COPY run_app.R .

# 创建www目录
RUN mkdir -p www

# 设置权限
RUN chmod -R 755 /app

# 暴露端口（固定端口，Railway会自动映射）
EXPOSE 3838

# 启动命令（支持动态端口）
CMD ["R", "-e", "shiny::runApp('main_app.R', port=as.numeric(Sys.getenv('PORT', '3838')), host='0.0.0.0')"] 