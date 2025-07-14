# Shinyapps.io 部署脚本
# 请先访问 https://www.shinyapps.io/ 注册账户

library(rsconnect)

# 第一次部署需要配置账户信息
# 在 shinyapps.io 账户设置中获取这些信息：
# rsconnect::setAccountInfo(name='你的用户名',
#                          token='你的token',
#                          secret='你的secret')

# 部署应用
# rsconnect::deployApp(appDir = ".", 
#                     appName = "viral-simulation-app",
#                     account = "你的用户名")

print("请按以下步骤发布应用:")
print("1. 访问 https://www.shinyapps.io/ 注册账户")
print("2. 在账户设置中获取 token 和 secret")
print("3. 取消注释上面的 setAccountInfo 和 deployApp 命令")
print("4. 填入你的实际信息并运行脚本") 