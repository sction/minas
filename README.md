# Minas 项目说明文档
[简体中文](README.md) | [English](README_EN.md) 

## 项目简介

Minas 是一个多功能的系统管理工具，专注于提供强大的数据管理与自动化运维能力。核心功能包括文件备份系统、日志与备份文件清理、定时任务调度以及 WebDAV 服务和超级终端管理等功能。它采用前后端分离的架构，后端使用 Go 语言开发，前端使用 Vue 3 + TypeScript 开发。

## 功能特性

- **数据文件备份**：支持多种备份模式（增量备份、单向镜像、完整备份），可设置定时执行，保护重要数据安全
- **日志与数据清理**：智能管理系统日志和备份数据，根据时间、数量等规则自动清理过期文件，释放存储空间
- **定时脚本任务**：强大的任务调度系统，支持cron表达式，可执行自定义脚本和系统命令
- **WebDAV 服务**：提供 WebDAV 协议支持，方便文件访问和管理
- **超级终端管理**：提供 Web 终端功能，可远程执行命令
- **多平台支持**：支持 Linux、Windows 等多种操作系统
- **Docker 支持**：提供 Docker 镜像，方便部署

## 开发者指南

### 环境要求

#### 后端开发环境

- Go 1.19 或更高版本
- 支持的数据库：SQLite、MySQL、PostgreSQL
- rclone 工具（用于文件备份功能）

#### 前端开发环境

- Node.js 14 或更高版本
- npm 或 yarn 包管理器

### 项目结构

```
minas/
├── build.sh                # 构建脚本
├── build-docker.sh         # Docker 构建脚本
├── Dockerfile              # Docker 配置文件
├── server/                 # 后端代码
│   ├── app/                # 应用层代码
│   ├── cmd/                # 命令行工具
│   ├── core/               # 核心代码
│   ├── data/               # 数据和配置
│   ├── middleware/         # 中间件
│   ├── route/              # 路由
│   ├── service/            # 服务层
│   ├── utils/              # 工具类
│   ├── www/                # 前端编译后的文件
│   ├── go.mod              # Go 模块依赖
│   └── main.go             # 主入口
├── web/                    # 前端代码
│   ├── public/             # 静态资源
│   ├── src/                # 源代码
│   ├── index.html          # HTML 入口
│   ├── package.json        # 依赖配置
│   └── vite.config.ts      # Vite 配置
└── airgo.sh                # 开发模式启动脚本
```

### 开发环境搭建

#### 克隆代码

```bash
git clone git@github.com:sorchk/minas.git
cd minas
```

#### 后端开发
1. 安装 Go 语言环境
   - 请参考 [Go 官方文档](https://golang.org/doc/install) 安装 Go 语言环境
```bash
export GOPATH=$HOME/go
export GOROOT=/home/sorc/tools/go
export PATH=$PATH:$GOROOT/bin
export PATH=$PATH:$GOPATH/bin
```

1. 安装 Go 依赖

```bash
cd server
go mod download
```

2. 安装 rclone 工具（用于文件备份功能）

```bash
# Linux
curl https://rclone.org/install.sh | sudo bash

# 或者手动下载并安装
# 下载地址：https://rclone.org/downloads/
```

3. 安装 air 工具（用于热重载）

air 是一个热重载工具，可以在代码变更时自动重新编译和运行程序，非常适合开发环境。

```bash
# 使用 go install 安装
go install github.com/air-verse/air@latest

# 确保 $GOPATH/bin 在您的 PATH 环境变量中
# 对于 Linux/macOS，可以添加到 ~/.bashrc 或 ~/.zshrc
# export PATH=$PATH:$GOPATH/bin

# 验证安装
air -v
```

或者使用其他方式安装：

```bash
# 使用 curl 安装
# Linux/macOS
curl -sSfL https://raw.githubusercontent.com/air-verse/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

```

4. 开发模式启动

```bash
# 使用 air 工具进行热重载
cd ..
chmod +x *.sh
./airgo.sh
```

#### 前端开发

1. 安装依赖

```bash
cd web
npm install
# 或者
yarn
```

2. 开发模式启动

```bash
npm run dev
# 或者
yarn dev
```

前端开发服务器将在 http://localhost:3002 启动，并自动代理 API 请求到后端服务。

### 构建项目

使用提供的构建脚本可以一键构建整个项目：

```bash
./build.sh
```

这将：
1. 构建前端代码并将其放入 `server/www/dist` 目录
2. 构建后端代码，生成多平台的可执行文件到 `dist` 目录

### Docker 构建

```bash
./build-docker.sh [版本号]
```

如果不指定版本号，将使用默认版本 1.3.3。

### 自定义开发

#### 修改配置

主要配置文件位于 `server/data/config.yaml`，可以根据需要修改。

#### 添加新功能

1. 在 `server/app` 目录下创建新的功能模块
2. 在 `server/route/Index.go` 中注册新的路由
3. 在 `web/src` 中添加对应的前端代码

## 用户使用指南

### 安装方式

#### 二进制安装

1. 从 [发布页面](https://github.com/sorchk/minas/releases) 下载适合您系统的二进制文件
2. 解压文件
3. 运行可执行文件

```bash
# Linux
chmod +x minas_amd64  # 或 minas_arm64
./minas_amd64 server

# Windows
minas.exe server
```

#### Docker 安装

```bash
docker run -d --name minas \
  -p 8002:8002 \
  -p 8003:8003 \
  -v $(pwd)/data:/app/data \
  -v $(pwd)/backup:/backup \
  sorc/minas:latest
```

### 配置说明

首次运行时，系统会在 `data` 目录下生成默认配置文件 `config.yaml`。主要配置项说明：

```yaml
app:
  host: 0.0.0.0            # 监听地址
  port: 8002               # HTTP 端口
  context-path: "/minas"   # 上下文路径
  ssl-port: 8003           # HTTPS 端口
  enable-ui: true          # 是否启用 UI
  type: "all"              # 服务类型：all 或 webdav
term:
  command: "bash"          # 终端命令
db:
  type: sqlite             # 数据库类型：sqlite、mysql、postgres
  # 其他数据库配置...
rclone:
  bin-path: "rclone"       # rclone 可执行文件路径
  port: 5572               # rclone 服务端口
log:
  level: "debug"           # 日志级别
  log-path: "logs"         # 日志路径
  # 其他日志配置...
```

### 服务管理

Minas 提供了多种服务管理命令：

```bash
# 启动服务
minas server

# 安装为系统服务
minas install

# 启动系统服务
minas start

# 停止系统服务
minas stop

# 卸载系统服务
minas uninstall

# 查看版本
minas -v
```

### 功能使用

#### 初始化

首次访问系统时，需要进行初始化设置，创建管理员账号。

#### 数据备份与同步

1. 进入 "计划任务" 页面，选择"文件同步"功能
2. 创建新的备份/同步任务
3. 选择同步类型：
   - **备份模式**：单向复制文件，适合数据备份
   - **镜像模式**：使目标目录与源目录完全一致，会删除目标中源中不存在的文件
   - **双向同步**：保持两个目录内容一致，任一方的更改都会同步到另一方
4. 设置源路径和目标路径（支持本地路径、远程FTP/SFTP、云存储等）
5. 配置高级选项（如文件过滤规则、带宽限制等）
6. 设置执行计划（每天、每周、每月或自定义cron表达式）
7. 保存并激活任务

#### 日志与数据清理

1. 进入 "计划任务" 页面，选择"文件清理"功能
2. 创建新的清理任务
3. 设置清理目标目录
4. 配置清理规则：
   - 按时间清理（如删除30天前的文件）
   - 按数量清理（如只保留最近10个备份文件）
   - 按文件大小清理（如当目录超过10GB时清理最旧文件）
   - 按文件名模式清理（支持通配符和正则表达式）
5. 设置执行计划
6. 可选择启用日志记录和清理报告
7. 保存并激活任务

#### 定时脚本任务

1. 进入 "计划任务" 页面，选择"脚本任务"功能
2. 创建新的脚本任务
3. 输入任务名称和描述
4. 编写或上传要执行的脚本内容（支持shell、python等）
5. 设置执行参数和环境变量
6. 配置执行计划（支持cron表达式）
7. 设置失败重试策略和超时时间
8. 配置任务完成通知（可选）
9. 保存并激活任务

#### WebDAV 服务

1. 登录系统后，进入 "WebDAV 管理" 页面
2. 点击 "新建" 创建 WebDAV 账号
3. 设置账号名称、主目录和权限
4. 保存后，可以使用 WebDAV 客户端连接：
   - 地址：`http(s)://服务器地址:端口/minas/dav`
   - 用户名：创建的账号
   - 密码：系统生成的令牌


#### 超级终端

1. 进入 "终端" 页面
2. 系统会自动打开一个终端会话
3. 可以执行命令行操作

### 常见问题

1. **无法启动服务**
   - 检查端口是否被占用
   - 检查配置文件是否正确
   - 查看日志文件获取详细错误信息

2. **WebDAV 连接失败**
   - 确认 WebDAV 账号状态是否启用
   - 检查用户名和令牌是否正确
   - 确认网络连接和防火墙设置

3. **文件备份/同步失败**
   - 检查 rclone 是否正确安装
   - 确认源路径和目标路径是否有访问权限
   - 验证网络连接和远程存储配置
   - 查看任务日志获取详细错误信息

4. **定时任务未执行**
   - 检查系统时间是否正确
   - 确认cron表达式语法无误
   - 检查任务状态是否为"启用"
   - 查看任务执行日志

5. **日志清理不生效**
   - 确认清理规则配置正确
   - 检查目标目录权限
   - 验证文件匹配规则是否正确匹配目标文件

## 许可证

本项目使用 MIT 许可证。详见 LICENSE 文件。
