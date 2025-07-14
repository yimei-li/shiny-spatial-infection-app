# 云部署优化版Dockerfile - 减少构建时间
FROM rocker/shiny:latest

# 安装系统依赖（合并RUN命令减少层数）
RUN apt-get update && apt-get install -y \
    wget \
    curl \
    imagemagick \
    libmagick++-dev \
    libcurl4-openssl-dev \
    && rm -rf /var/lib/apt/lists/* \
    && wget -O go.tar.gz https://go.dev/dl/go1.21.0.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go.tar.gz \
    && rm go.tar.gz

ENV PATH=$PATH:/usr/local/go/bin

# 安装R包（使用二进制包减少编译时间）
RUN R -e "install.packages(c('shiny', 'shinyjs', 'magick', 'curl'), repos='https://cloud.r-project.org/', type='binary', dependencies=TRUE)"

# 创建应用目录并复制文件（合并操作）
WORKDIR /app
COPY main_app.R gif_generator.R mdbk_small_vero_0713.go run_app.R ./
RUN mkdir -p www && chmod -R 755 /app

# 暴露端口
EXPOSE 3838

# 启动命令
CMD ["R", "-e", "shiny::runApp('main_app.R', port=as.numeric(Sys.getenv('PORT', '3838')), host='0.0.0.0')"] 