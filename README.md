# 007Password - 密码管理器

一个安全、简单的密码管理应用，使用Go后端和Vue.js前端开发。

## 功能特点

- SQLite数据库采用SQLCipher加密
- 加密所有存储的密码
- 支持多种授权登录方式记录：Google, Facebook, Twitter, Github, 微信, 微博, QQ等
- 直观的用户界面，方便管理所有密码
- JWT令牌认证保障安全性
- 支持Docker部署

## 安全说明

- 主密码用于生成SQLite数据库的加密密钥
- 所有密码均以加密形式存储
- 主密码不会在服务器端存储明文，将主密码用做SQLite数据库的密码
- 请务必记住您的主密码，如果忘记将无法恢复数据

## 技术栈

- **后端**: Go, Gin框架, SQLite, SQLCipher, JWT
- **前端**: Vue.js, Tailwind CSS, Axios
- **部署**: Docker, Nginx

## 本地开发

### 后端

```bash
cd backend
go mod download
go run main.go
```

### 前端

```bash
cd frontend
npm install
npm run serve
```

应用将在 http://localhost:8081 运行

## Docker 部署 (推荐)

项目支持使用 Docker Compose 进行一键部署。

### 前提条件

- 安装 [Docker](https://docs.docker.com/get-docker/)
- 安装 [Docker Compose](https://docs.docker.com/compose/install/)

### 部署步骤

1. 克隆仓库

```bash
git clone https://github.com/yourusername/007Password.git
cd 007Password
```

2. 构建并启动容器

```bash
docker-compose up -d
```

应用将在 http://localhost:8081 上运行。

### 查看日志

```bash
docker-compose logs -f
```

### 停止服务

```bash
docker-compose down
```

### 数据持久化

数据库文件存储在Docker卷中，确保数据不会丢失。如果需要备份数据，可以使用以下命令：

```bash
docker run --rm -v 007password-data:/data -v $(pwd):/backup alpine tar -czvf /backup/007password-backup.tar.gz /data
```

### 恢复数据

```bash
docker run --rm -v 007password-data:/data -v $(pwd):/backup alpine sh -c "rm -rf /data/* && tar -xzvf /backup/007password-backup.tar.gz -C /"
```

## 页面效果
![Image](https://github.com/007Secret/007Password/blob/main/image/list.png)
![Image](https://github.com/007Secret/007Password/blob/main/image/add.png)
![Image](https://github.com/007Secret/007Password/blob/main/image/import.png)

## 许可证

Apache Lincense 2.0 
