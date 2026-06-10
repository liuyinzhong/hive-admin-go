# 项目模块管理(dev_module表)

## 列表数据

请求路径：/api/dev/modules

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
      "moduleId": "cd774399-15c6-44cb-9f47-6520d418bc5a",
      "moduleTitle": "商户管理端",
      "projectId": "62b48f5b-29da-41fe-aa87-0c0739cc8bb0",
      "sort": 9,
      "updateDate": "2024-05-15 08:57:47",
      "createDate": "2024-05-15 08:57:47"
    }
  ],
  "error": null,
  "message": "ok"
}
```

## 创建模块

请求路径：/api/dev/modules

请求方法：POST

要求：

1. 同一个项目下判断标题是否已存在

请求参数示例(body)：

```json
{"sort":0,"moduleTitle":"模块标题","projectId":""}
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

## 修改模块

请求路径：/api/dev/modules/{moduleId}

请求方法：PUT

要求：

1. 操作时需判断标题是否已存在

请求参数示例(path)：

```url
moduleId
```

请求参数示例(body)：

```json
{"sort":0,"moduleTitle":"模块标题"}
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

## 模块详情

请求路径：/api/dev/modules/{moduleId}

请求方法：GET

请求参数示例(path)：

```url
moduleId
```

响应数据示例：

```json
{
  "code": 0,
  "data": {
      "moduleId": "a94556c4-30e9-4371-8602-7a1a6dc7f7f0",
      "moduleTitle": "药店端",
      "projectTitle":"", //连dev_project表得到
      "projectId": "1296c41a-47bf-429e-9703-411b23c1f61c",
      "sort": 5,
      "updateDate": "2022-07-06 04:23:40",
      "createDate": "2022-07-06 04:23:40"
    },
  "error": null,
  "message": "ok"
}
```

## 删除模块

请求路径：/api/dev/modules

请求方法：DELETE

要求：

1. 逻辑删除 del_flag=1

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