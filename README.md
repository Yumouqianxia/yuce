# 预测系统 - Docker 版本

基于 Vue 3 + Go + MySQL + Redis 的现代化预测系统，支持 Docker 一键部署。

## ✨ 特性

- 🚀 **现代化技术栈**: Vue 3 + Go + MySQL + Redis
- 🐳 **Docker 部署**: 一键启动，开箱即用
- 🔒 **安全可靠**: JWT 认证、权限控制、数据加密
- 📱 **响应式设计**: 支持 PC 和移动端
- ⚡ **高性能**: Go 后端、Redis 缓存
- 🛡️ **健康检查**: 完整的监控和故障恢复
- � **排行载榜系统**: 实时排名、统计分析

## 🚀 快速开始

### Windows 一键启动（推荐）

```cmd
# 1. 克隆项目
git clone <your-repo-url>
cd yuce

# 2. 启动所有服务
start.bat

# 3. 停止服务
stop.bat
```

### 手动启动

```cmd
# 启动所有服务（包括数据库）
docker-compose -f docker-compose.hub.yml --profile local up -d

# 仅启动应用服务（不包括数据库）
docker-compose -f docker-compose.hub.yml up -d backend frontend

# 查看服务状态
docker ps

# 查看日志
docker-compose -f docker-compose.hub.yml logs -f

# 停止服务
docker-compose -f docker-compose.hub.yml down
```

### 配置文件说明

项目使用 `docker-compose.hub.yml` 作为生产环境配置文件：

- **mysql/redis**: 使用 `--profile local` 启动，适合本地开发
- **backend/frontend**: 默认启动，适合生产部署
- **adminer**: 数据库管理工具，使用 `--profile local` 启动

### 环境变量配置（可选）

创建 `.env` 文件自定义配置：

```env
# 数据库配置
MYSQL_ROOT_PASSWORD=your_root_password
MYSQL_PASSWORD=your_password
DB_HOST=mysql
DB_USER=prediction
DB_NAME=prediction_system

# Redis 配置
REDIS_PASSWORD=your_redis_password

# JWT 配置
JWT_SECRET=your_jwt_secret_key

# 日志级别
LOG_LEVEL=info
```

## 📋 系统要求

- Docker 20.10+
- Docker Compose 2.0+
- 2GB+ 可用内存
- 10GB+ 可用磁盘空间

## 🌐 访问地址

启动成功后，您可以通过以下地址访问：

- **前端应用**: http://localhost:5408
- **后端 API**: http://localhost:1874
- **API 文档**: http://localhost:1874/swagger/index.html
- **数据库管理**: http://localhost:8082 (Adminer)
- **健康检查**: http://localhost:1874/health

### 默认账号

- **管理员**: root / admin123
- **数据库**: prediction / prediction123

## 🔧 管理命令

### Windows 命令

```cmd
# 启动服务
start.bat

# 停止服务
stop.bat

# 查看服务状态
docker ps

# 查看日志
docker-compose -f docker-compose.hub.yml logs -f

# 重启服务
docker-compose -f docker-compose.hub.yml restart

# 重新构建
docker-compose -f docker-compose.hub.yml build

# 清理所有数据（危险操作）
docker-compose -f docker-compose.hub.yml down -v
```

## 📊 服务架构

```
┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Backend       │
│   (Vue 3)       │◄──►│   (Go)          │
│   Port:5408     │    │   Port:1874     │
└─────────────────┘    └─────────────────┘
                                │
                   ┌─────────────────────────┐
                   │                         │
          ┌─────────────────┐    ┌─────────────────┐
          │   MySQL:3306    │    │   Redis:6379    │
          │   (数据存储)     │    │   (缓存)        │
          └─────────────────┘    └─────────────────┘
                                │
                   ┌─────────────────────────┐
                   │   Adminer:8082          │
                   │   (数据库管理)           │
                   └─────────────────────────┘
```

### 容器列表

| 容器名 | 服务 | 端口 | 说明 |
|--------|------|------|------|
| yuce-frontend | Vue 3 前端 | 5408 | Web 应用界面 |
| yuce-backend | Go 后端 | 1874 | RESTful API |
| yuce-mysql | MySQL 8.0 | 3306 | 数据库 |
| yuce-redis | Redis 7 | 6379 | 缓存服务 |
| yuce-adminer | Adminer | 8082 | 数据库管理工具 |

## 🔒 安全配置

### 生产环境部署前必须修改：

1. **数据库密码**: 修改 `.env` 中的 `MYSQL_ROOT_PASSWORD` 和 `MYSQL_PASSWORD`
2. **Redis 密码**: 修改 `REDIS_PASSWORD`
3. **JWT 密钥**: 修改 `JWT_SECRET`
4. **防火墙**: 配置防火墙规则，只开放必要端口
5. **SSL 证书**: 配置 HTTPS 证书

### 推荐安全措施：

```bash
# 生成安全的随机密码
openssl rand -base64 32

# 配置防火墙（Ubuntu）
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable

# 安装SSL证书
sudo certbot --nginx -d your-domain.com
```

## 📈 性能优化

### 系统优化

- 增加文件描述符限制
- 优化内核网络参数
- 配置 Docker 日志轮转
- 启用 Gzip 压缩

### 应用优化

- Redis 缓存热点数据
- 数据库连接池
- 静态资源 CDN
- 图片压缩优化

## 🔄 备份策略

### 自动备份

系统会自动创建备份脚本，每天凌晨 2 点备份：

- 数据库数据
- 用户上传文件
- 配置文件

### 手动备份

```bash
# 备份数据库
docker exec prediction-mysql mysqldump -u root -p prediction_system > backup.sql

# 备份上传文件
tar -czf uploads_backup.tar.gz backend-go/uploads/
```

## 🚨 故障排除

### 常见问题

1. **端口被占用**

   ```cmd
   # 查看端口占用
   netstat -ano | findstr :5408
   netstat -ano | findstr :1874
   
   # 停止占用端口的进程
   taskkill /PID <进程ID> /F
   ```

2. **Docker 未启动**

   ```cmd
   # 检查 Docker 状态
   docker info
   
   # 如果失败，请启动 Docker Desktop
   ```

3. **服务启动失败**

   ```cmd
   # 查看详细日志
   docker-compose -f docker-compose.hub.yml logs backend
   
   # 重新构建
   docker-compose -f docker-compose.hub.yml build --no-cache backend
   
   # 完全重启
   docker-compose -f docker-compose.hub.yml down
   docker-compose -f docker-compose.hub.yml --profile local up -d
   ```

4. **数据库连接失败**

   ```cmd
   # 检查数据库是否运行
   docker ps | findstr mysql
   
   # 查看数据库日志
   docker logs yuce-mysql
   
   # 重启数据库
   docker restart yuce-mysql
   ```

5. **清理并重新开始**

   ```cmd
   # 停止所有服务
   docker-compose -f docker-compose.hub.yml down
   
   # 清理所有数据（警告：会删除数据库数据）
   docker-compose -f docker-compose.hub.yml down -v
   
   # 重新启动
   start.bat
   ```

## 📚 开发指南

### 进入容器

```cmd
# 进入后端容器
docker exec -it yuce-backend sh

# 进入前端容器
docker exec -it yuce-frontend sh

# 进入数据库
docker exec -it yuce-mysql mysql -u prediction -p
```

### 查看日志

```cmd
# 查看所有服务日志
docker-compose -f docker-compose.hub.yml logs -f

# 查看特定服务日志
docker-compose -f docker-compose.hub.yml logs -f backend
docker-compose -f docker-compose.hub.yml logs -f frontend

# 查看最近100行日志
docker logs yuce-backend --tail 100
```

### 数据库管理

访问 http://localhost:8082 使用 Adminer 管理数据库：

- **系统**: MySQL
- **服务器**: mysql
- **用户名**: prediction
- **密码**: prediction123
- **数据库**: prediction_system

## 🤝 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 📞 技术支持

- 📧 邮箱: support@example.com
- 💬 QQ 群: 123456789
- 📱 微信群: 扫码加入

---

**⚠️ 重要提示**: 生产环境部署前请务必修改所有默认密码和配置！
