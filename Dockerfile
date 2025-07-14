# 简化版Dockerfile - 兼容Railway
FROM rocker/shiny:latest

# 安装系统依赖
RUN apt-get update && apt-get install -y \
    wget \
    curl \
    imagemagick \
    libmagick++-dev \
    && rm -rf /var/lib/apt/lists/*

# 安装Go
RUN wget -O go.tar.gz https://go.dev/dl/go1.21.0.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go.tar.gz \
    && rm go.tar.gz

ENV PATH=$PATH:/usr/local/go/bin

# 安装R包
RUN R -e "install.packages(c('shiny', 'shinyjs', 'magick'), repos='https://cloud.r-project.org/')"

# 设置工作目录
WORKDIR /app

# 复制文件
COPY main_app.R .
COPY gif_generator.R .
COPY mdbk_small_vero_0713.go .
COPY run_app.R .

# 创建www目录
RUN mkdir -p www

# 设置权限
RUN chmod -R 755 /app

# 暴露端口
EXPOSE 3838

# 启动命令
CMD ["R", "-e", "shiny::runApp('main_app.R', port=as.numeric(Sys.getenv('PORT', '3838')), host='0.0.0.0')"] 