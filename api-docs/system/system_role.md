# 角色管理 （sys_role表）

## 列表数据

请求路径：/api/system/roles

请求方法：GET

要求：

1. 按创建时间倒序 - 最新的数据排在前面
2. 分页获取表数据。调用统一分页方法

请求参数示例(query)：

```tex
//需分页,查询参数，排序参数
page=1&pageSize=20&name=&id=&status=1&remark=&startDate=2026-05-15&endDate=2026-05-17&sorts=createDate,desc;
```

响应数据示例：

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "roleId": "458e8285-cd9e-48ca-ac78-d2178a0e8c4f",
        "roleTitle": "SuperAdmin",
        "status": 1,
        "createDate": null,
        "remark": "超级管理员，普通人不要给这个"
      }
    ],
    "total": 1
  },
  "error": null,
  "message": "ok"
}
```

## 创建角色

请求路径：/api/system/roles

请求方法：POST

要求：

1. 操作时需校验名称是否已存在

请求参数示例(body)：

```json
{
  "status": 1,
  "roleTitle": "名称",
  "remark": "备注",
  "permissions": ["45843bfc-4e97-4061-87cb-7e4835100d43"] //菜单id
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

## 角色详情

请求路径：/api/system/roles/{roleId}

请求方法：GET

请求参数示例(path)：

```url
roleId
```

响应数据示例：

```json
{
  "code": 0,
  "data": {
    "roleId": "458e8285-cd9e-48ca-ac78-d2178a0e8c4f",
    "roleTitle": "SuperAdmin",
    "status": 1,
    "createDate": null,
    "permissions": ["205ce73c-baa0-4df9-b853-f6ae810d38ef"],//菜单id
    "remark": "超级管理员，普通人不要给这个"
  },
  "error": null,
  "message": "ok"
}
```

## 修改角色

请求路径：/api/system/roles/{roleId}

请求方法：PUT

要求：

1. 操作时需校验名称是否已存在

请求参数示例(path)：

```url
roleId
```

请求参数示例(body)：

```json
{
  "status": 1,
  "roleTitle": "名称",
  "remark": "备注",
  "permissions": ["45843bfc-4e97-4061-87cb-7e4835100d43"] //菜单id
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

## 修改角色状态

请求路径：/api/system/roles/{roleId}/status

请求方法：PUT

请求参数示例(path)：

```url
roleId
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

## 删除角色

请求路径：/api/system/roles

请求方法：DELETE

要求：

1. 禁止删除有绑定用户的角色
2. 逻辑删除 del_flag=1
3. 清理 sys_user_role 关联表数据
4. 清理 sys_role_menu 关联表数据

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