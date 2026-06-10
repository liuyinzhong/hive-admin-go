# 项目管理(dev_project表)

## 列表数据

请求路径：/api/dev/projects

请求方法：GET

要求：

1. 按创建时间倒序 - 最新的数据排在前面
2. 无需分页，全量查询

响应数据示例：

```json
{
  "code": 0,
  "data": [
    {
      "projectId": "980162da-b28e-4bad-ae2b-b3ee4da7b213",
      "projectTitle": "标题",
      "projectLogo": "https://picsum.photos/100/100",
      "description": "描述",
      "createDate": "2024-05-15 08:57:47"
    }
  ],
  "error": null,
  "message": "ok"
}
```

## 创建项目

请求路径：/api/dev/projects

请求方法：POST

要求：

1. 操作时需校验标题是否已存在

请求参数示例(body)：

```json
{
    "projectTitle":"项目标题",
 	"description":"项目描述",
 	"projectLogo":"https://unpkg.com/@vbenjs/static-source@0.1.7/source/logo-v1.webp"
}
```

响应数据示例：

```json
{
  "code": 0,
  "data":null,
  "error": null,
  "message": "ok"
}
```

## 修改项目

请求路径：/api/dev/projects/{projectId}

请求方法：PUT

要求：

1. 操作时需校验标题是否已存在

请求参数示例(path)：

```url
projectId
```

请求参数示例(body)：

```json
{
    "projectTitle":"项目标题",
 	"description":"项目描述",
 	"projectLogo":"https://unpkg.com/@vbenjs/static-source@0.1.7/source/logo-v1.webp"
}
```

响应数据示例：

```json
{
  "code": 0,
  "data":null,
  "error": null,
  "message": "ok"
}
```

## 项目详情

请求路径：/api/dev/projects/{projectId}

请求方法：GET

请求参数示例(path)：

```url
projectId
```

响应数据示例：

```json
{
  "code": 0,
  "data": {
      "projectId": "d8fd7341-b5ca-4548-88c5-9b6e67143e50",
      "projectTitle": "voluptate",
      "projectLogo": "https://picsum.photos/100/100",
      "description": "vesco",
      "createDate": "2022-12-07 02:05:39"
   },
  "error": null,
  "message": "ok"
}
```