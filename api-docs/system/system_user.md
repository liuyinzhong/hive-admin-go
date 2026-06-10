# 用户管理 （sys_user表）

## 列表数据

请求路径：/api/system/users

请求方法：GET

要求：

1. 过滤掉内置账户
2. 过滤掉已删除的数据
3. 按创建时间倒序 - 最新的数据排在前面
4. 分页获取表数据。调用统一分页方法

请求参数示例(query)：

```tex
//需分页,查询参数，排序参数
page=1&pageSize=20&username=&realName=&status=1&sorts=avatar,desc;username,desc
```

响应数据示例：

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "userId": "fd8b5f2c-77c6-4e59-b81c-306c2fb85d44", //用户ID
        "avatar": "https://picsum.photos/100/100", //头像URL
        "username": "vben", //登录用户名
        "realName": "Vben", //真实姓名
        "roleTitles": ["super"], //角色名称数组，根据sys_role表关联查询得到
        "roleIds": ["458e8285-cd9e-48ca-ac78-d2178a0e8c4f"], //角色id数组，根据sys_user_role表关联查询得到
        "desc": "超管", //用户描述
        "email": "vben@example.com", //邮箱
        "homePath": null, //首页路径
        "deptTitles": ["技术部"], //部门名称数组，根据sys_user_dept和sys_dept表关联查询得到
        "deptIds": ["6a379b2d-2cb9-46f7-8703-2523317a2473"], //部门id数组，根据sys_user_dept表关联查询得到
        "status": 1, //用户状态：0=禁用，1=启用
        "createDate": null, //创建时间
        "updateDate": null //更新时间
      }
    ],  //items：分页时的数据数组，不分页时包含所有数据
    "total": 1 //总记录数
  },
  "error": null,
  "message": "ok"
}
```

## 获取有效的用户（全量）

请求路径：/api/system/users/all

请求方法：GET

要求：

1. 过滤掉内置账户
2. 过滤已删除的用户
3. 过滤已禁用的用户

请求参数示例(query)：

```tex
realName=
```

响应数据示例：

```json
{
  "code": 0,
  "data": [
      {
        "userId": "fd8b5f2c-77c6-4e59-b81c-306c2fb85d44", //用户ID
        "avatar": "https://picsum.photos/100/100", //头像URL
        "username": "vben", //登录用户名
        "realName": "Vben", //真实姓名
        "roleTitles": ["super"], //角色名称数组，根据sys_role表关联查询得到
        "roleIds": ["458e8285-cd9e-48ca-ac78-d2178a0e8c4f"], //角色id数组，根据sys_user_role表关联查询得到
        "desc": "超管", //用户描述
        "email": "vben@example.com", //邮箱
        "homePath": null, //首页路径
        "deptTitles": ["技术部"], //部门名称数组，根据sys_user_dept和sys_dept表关联查询得到
        "deptIds": ["6a379b2d-2cb9-46f7-8703-2523317a2473"], //部门id数组，根据sys_user_dept表关联查询得到
        "status": 1, //用户状态：0=禁用，1=启用
        "createDate": null, //创建时间
        "updateDate": null //更新时间
      }
    ],
  "error": null,
  "message": "ok"
}
```

## 创建用户

请求路径：/api/system/users

请求方法：POST

要求：

1. 操作时需判断登录名是否已存在

请求参数示例(body)：

```json
{
  "username": "登录名",
  "realName": "真实姓名",
  "password": "密码",
  "desc": "描述",
  "deptIds": ["c7894d44-b9da-4e28-b590-cbf0b12b27ee", "49306a50-cfa9-4ff0-88e3-95eda0ae73e0"],
  "roleIds": ["458e8285-cd9e-48ca-ac78-d2178a0e8c4f", "6b81f1cf-301a-444f-a5b4-2ffa333de39f"]
}
```

响应数据示例：

```json
{
  "code": 0,
  "data": null,
  "error": null,
  "message": "ok"
}
```

## 用户详情

请求路径：/api/system/users/{userId}

请求方法：GET

请求参数示例(path)：

```url
userId
```

响应数据示例：

```json
{
  "code": 0,
  "data": {
        "userId": "fd8b5f2c-77c6-4e59-b81c-306c2fb85d44",
        "avatar": "https://picsum.photos/100/100",
        "username": "vben",
        "realName": "Vben",
        "roleTitles": ["super"], //角色名称数组
        "roleIds": ["458e8285-cd9e-48ca-ac78-d2178a0e8c4f"], //角色id数组
        "desc": "超管",
        "email": "vben@example.com",
        "homePath": null,
        "deptTitles": ["技术部"],
        "deptIds": ["6a379b2d-2cb9-46f7-8703-2523317a2473"],
        "status": 1,
        "createDate": null,
        "updateDate": null
      },
  "error": null,
  "message": "ok"
}
```

## 修改用户信息

请求路径：/api/system/users/{userId}

请求方法：PUT

要求：

1. 操作时需判断登录名是否已存在

请求参数示例(path)：

```url
userId
```

请求参数示例(body)：

```json
{
  "username": "登录名",
  "realName": "真实姓名",
  "desc": "描述",
  "deptIds": ["c7894d44-b9da-4e28-b590-cbf0b12b27ee", "49306a50-cfa9-4ff0-88e3-95eda0ae73e0"],
  "roleIds": ["458e8285-cd9e-48ca-ac78-d2178a0e8c4f", "6b81f1cf-301a-444f-a5b4-2ffa333de39f"]
}
```

响应数据示例：

```json
{
  "code": 0,
  "data": null,
  "error": null,
  "message": "ok"
}
```

## 修改用户状态

请求路径：/api/system/users/{userId}/status

请求方法：PUT

请求参数示例(path)：

```url
userId
```

请求参数示例(body)：

```json
{
  "status": 1
}
```

响应数据示例：

```json
{
  "code": 0,
  "data": null,
  "error": null,
  "message": "ok"
}
```

## 删除用户

请求路径：/api/system/users

请求方法：DELETE

要求：

1. 禁止删除当前登录用户
2. 禁止删除有管理员角色的用户
3. 禁止删除内置账户，用户表的 is_sys字段
4. 逻辑删除 del_flag=1
5. 删除后要清理 sys_user_role 关联表数据
6. 删除后要清理 sys_user_dept 关联表数据

请求参数示例(body)：

```json
["UUID 字符串数组"]
```

响应数据示例：

```json
{
  "code": 0,
  "data": null,
  "error": null,
  "message": "ok"
}
```