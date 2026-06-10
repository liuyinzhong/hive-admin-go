# Hive Admin Go

## 项目简介

这是一个基于Go语言开发的后台管理接口项目，使用Gin框架和GORM ORM。

swagger 地址：http://localhost:9191/swagger/index.html
swagger json地址: http://localhost:9191/swagger/doc.json

## 技术栈

- Go 1.21+
- Gin Web Framework
- GORM ORM
- MySQL 8.0+
- JWT Authentication

## 项目结构

```
hive-admin-go/
├── config/              # 配置加载模块
│   └── config.go       # 配置文件结构定义
├── controllers/        # 控制器层
│   └── auth_controller.go  # 认证接口控制器
├── database/           # 数据库连接模块
│   └── database.go    # 数据库初始化
├── middleware/         # 中间件
│   └── auth.go        # JWT认证中间件
├── models/             # 数据模型
│   ├── models.go      # 数据库模型定义
│   └── response.go    # 响应结构体
├── router/             # 路由配置
│   └── router.go      # 路由设置
├── services/          # 业务逻辑层
│   └── auth_service.go # 认证业务逻辑
├── utils/             # 工具函数
│   └── jwt.go        # JWT工具函数
├── api-docs/          # API文档
├── config.json        # 配置文件
├── go.mod            # Go模块文件
└── main.go           # 入口文件
```

## 配置说明

编辑 `config.json` 文件：

```json
{
  "server": {
    "port": 9191,        // 服务端口
    "mode": "debug"      // 运行模式
  },
  "database": {
    "host": "localhost",
    "port": 3306,
    "username": "root",
    "password": "123456",
    "dbname": "hive",
    "charset": "utf8mb4"
  },
  "jwt": {
    "secret": "your-secret-key",
    "expire": 720        // Token过期时间（小时）
  }
}
```

## 快速开始

1. 安装依赖：

```bash
go mod tidy
```

2. 配置数据库连接（编辑 `config.json`）

3. 运行项目：

```bash
go run main.go
```

## API接口列表

### 认证模块 `/api/auth`

| 接口名称 | 请求路径 | 请求方式 | 说明 | 认证 |
|---------|---------|---------|------|------|
| 登录 | `/api/auth/login` | POST | 用户登录 | 否 |
| 获取当前用户信息 | `/api/auth/profile` | GET | 获取登录用户详情 | 是 |
| 获取用户菜单 | `/api/auth/menus` | GET | 获取用户菜单权限 | 是 |
| 获取权限码列表 | `/api/auth/codes` | GET | 获取用户所有权限码 | 是 |
| 退出登录 | `/api/auth/logout` | POST | 用户退出登录 | 是 |

## 接口详细说明

### 1. 登录接口

**请求路径**: `/api/auth/login`

**请求方法**: POST

**请求参数**:
```json
{
  "username": "vben",
  "password": "123456"
}
```

**响应示例**:
```json
{
  "code": 0,
  "data": {
    "accessToken": "xxxxxx"
  },
  "error": null,
  "message": "ok"
}
```

### 2. 获取当前用户信息

**请求路径**: `/api/auth/profile`

**请求方法**: GET

**请求头**:
```
Authorization: Bearer {accessToken}
```

**响应示例**:
```json
{
  "code": 0,
  "data": {
    "userId": "fd8b5f2c-77c6-4e59-b81c-306c2fb85d44",
    "avatar": "https://picsum.photos/100/100",
    "username": "vben",
    "realName": "Vben",
    "roleTitles": ["super"],
    "roleIds": ["458e8285-cd9e-48ca-ac78-d2178a0e8c4f"],
    "desc": "超管",
    "email": "vben@example.com",
    "homePath": null,
    "deptTitles": ["技术部"],
    "deptIds": ["9de2fb68-7ba1-49cc-bfd4-e946a394f880"],
    "status": 1
  },
  "error": null,
  "message": "ok"
}
```

### 3. 获取用户菜单

**请求路径**: `/api/auth/menus`

**请求方法**: GET

**请求头**:
```
Authorization: Bearer {accessToken}
```

**响应示例**:
```json
{
  "code": 0,
  "data": [
    {
      "id": "205ce73c-baa0-4df9-b853-f6ae810d38ef",
      "pid": null,
      "type": "catalog",
      "authCode": "sys:workspace",
      "children": [...],
      "component": null,
      "meta": {
        "icon": "carbon:workspace",
        "title": "page.dashboard.title",
        ...
      },
      "name": "Dashboard",
      "path": "/dashboard",
      "status": 1
    }
  ],
  "error": null,
  "message": "ok"
}
```

### 4. 获取权限码列表

**请求路径**: `/api/auth/codes`

**请求方法**: GET

**请求头**:
```
Authorization: Bearer {accessToken}
```

**响应示例**:
```json
{
  "code": 0,
  "data": ["sys:workspace", "sys:analytics"],
  "error": null,
  "message": "ok"
}
```

### 5. 退出登录

**请求路径**: `/api/auth/logout`

**请求方法**: POST

**请求头**:
```
Authorization: Bearer {accessToken}
```

**响应示例**:
```json
{
  "code": 0,
  "data": "",
  "error": null,
  "message": "ok"
}
```

## 统一响应格式

```json
{
  "code": 0,           // 业务状态码（0=成功，-1=失败）
  "data": null,       // 成功时的业务数据
  "error": null,      // 失败时的错误详情
  "message": "ok"     // 提示文本
}
```

## 响应码说明

- `0`: 请求成功
- `-1`: 请求失败

## 数据库初始化

1. 创建数据库：
```sql
CREATE DATABASE hive CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

2. 导入SQL文件：
```bash
mysql -u root -p hive < api-docs/hive.sql
```

## 内置账户

- **用户名**: superAdmin
- **密码**: 123456
- **特性**: 系统内置超级管理员，无需分配角色，拥有所有权限

## 开发规范

1. 所有接口遵循RESTful设计规范
2. 使用UUID作为主键格式
3. 时间格式：YYYY-MM-DD HH:mm:ss
4. 支持逻辑删除（del_flag字段）
5. 所有接口添加Swagger注释

## 许可证

MIT License
