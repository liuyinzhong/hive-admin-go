# 版本迭代表(dev_version表)

## 列表数据

请求路径：/api/dev/versions

请求方法：GET

要求：

1. 过滤掉已删除的数据
2. 按创建时间倒序 - 最新的数据排在前面
3. 分页获取表数据。调用统一分页方法
4. 排序时仅支持：version、startDate、endDate 的排序

请求参数示例(query)：

```tex
//需分页,查询参数，排序参数
page=1&pageSize=20&version=&projectId=&releaseStatus=&sorts=
```

响应数据示例：

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "versionId": "0886cc53-8186-4584-a33e-65a2dae9181a",
        "version": "1.7.9",
        "versionType": "0", //更新类型 字典VERSION_TYPE值
        "remark": "备注",
        "creatorId": "4951a5ed-d835-4256-98f0-f68e90247d44",
        "creatorName": "创建人",  //连sys_user表得到
        "createDate": "2024-05-15 08:57:47",
        "endDate": "",
        "startDate": "",
        "projectId": "a9c33c6b-b87d-42c6-822b-66034104baac",
        "projectTitle":"",//连dev_project表得到
        "releaseStatus": "0", // 发布状态 字典RELEASE_STATUS值
        "releaseDate": "",//发布时间
      }
    ],
    "total": 1
  },
  "error": null,
  "message": "ok"
}
```

## 列表数据(全量查询)

请求路径：/api/dev/versions/all

请求方法：GET

要求：

1. 过滤掉已删除的数据
2. 按创建时间倒序 - 最新的数据排在前面

请求参数示例(query)：

```tex
version=&projectId=&releaseStatus=
```

响应数据示例：

```json
{
  "code": 0,
  "data": [
      {
        "versionId": "0886cc53-8186-4584-a33e-65a2dae9181a",
        "version": "1.7.9",
        "versionType": "0", //更新类型 字典VERSION_TYPE值
        "remark": "备注",
        "creatorId": "4951a5ed-d835-4256-98f0-f68e90247d44",
        "creatorName": "创建人",  //连sys_user表得到
        "createDate": "2024-05-15 08:57:47",
        "endDate": "",
        "startDate": "",
        "projectId": "a9c33c6b-b87d-42c6-822b-66034104baac",
        "projectTitle":"",//连dev_project表得到
        "releaseStatus": "0", // 发布状态 字典RELEASE_STATUS值
        "releaseDate": "",//发布时间
      }
    ],
  "error": null,
  "message": "ok"
}
```

## 创建版本

请求路径：/api/dev/versions

请求方法：POST

要求：

1. 同一项目下需判断版本号是否已存在
2. 操作时需向 dev_change_history 表写入关联数据。change_behavior=0、business_type=30、business_id=当前创建的这条数据versionId
3. 操作时需校验：关联项目、更新类型、版本号、发布状态。必填

请求参数示例(body)：

```json
{
    "version": "2.0.0",
    "versionType": "20", // 字典VERSION_TYPE值
    "releaseStatus": "0", // 字典RELEASE_STATUS值
    "projectId": "e5a6a9f6-3699-4a29-9ebc-d9160d8b8f8d",
    "remark": "备注",
    "endDate": "2026-05-03 00:00:00",
    "startDate": "2026-05-03 00:00:00"
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

## 修改版本

请求路径：/api/dev/versions/{versionId}

请求方法：PUT

要求：

1. 操作时需判断版本号是否已存在
2. 操作时需向 dev_change_history 表写入关联数据。change_behavior=10、business_type=30、business_id=当前创建的这条数据versionId

请求参数示例(path)：

```url
versionId
```

请求参数示例(body)：

```json
{
    "version": "2.0.0",
    "versionType": "20",
    "releaseStatus": "0",
    "projectId": "e5a6a9f6-3699-4a29-9ebc-d9160d8b8f8d",
    "firstVersion": "1.2.6",
    "remark": "备注",
    "endDate": "2026-05-01 00:00:00",
    "startDate": "2026-05-03 00:00:00"
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

## 版本流转状态

请求路径：/api/dev/versions/{versionId}/next

请求方法：PUT

要求：

1. 操作时需向 dev_change_history 表写入关联数据。change_behavior=40、business_type=30、business_id=versionId、changeRichText=参数值的changeRichText

请求参数示例(path)：

```url
versionId
```

请求参数示例(body)：

```json
{
  "releaseStatus": "10", //用于修改版本表的状态
  "changeRichText": "内容" //用于填充 dev_change_history 表的 changeRichText
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

## 版本详情

请求路径：/api/dev/versions/{versionId}

请求方法：GET

请求参数示例(path)：

```url
versionId
```

响应数据示例：

```json
{
  "code": 0,
  "data": {
        "versionId": "1da045d0-be37-4b1e-8e05-f74cb28eb68c",
        "version": "1.9.6",
        "versionType": "20",
        "remark": "",
        "creatorId": "fec6d6b8-df5e-4fbd-8347-565f15c24b51",
        "creatorName": "Nathan Rolfson", //连sys_user表得到
        "createDate": "2024-04-21 21:38:01",
        "endDate": "",
        "startDate": "",
        "projectId": "bcff0ba9-0a65-4190-9bbc-16cf4962420e",
      	"projectTitle":"",//连dev_project表得到
        "releaseStatus": "99",
        "releaseDate": "",
        "changeLogRichText": "",
        "changeLog": ""
      },
  "error": null,
  "message": "ok"
}
```

## 获取最新版本号

请求路径：/api/dev/versions/getLastVersion

请求方法：GET

要求：

1. 获取指定项目下。按version排序后的最大版本号。version的格式为1.0.0

请求参数示例(query)：

```url
projectId
```

响应数据示例：

```json
{
  "code": 0,
  "data": {
        "versionId": "1da045d0-be37-4b1e-8e05-f74cb28eb68c",
        "version": "1.9.6",
        "versionType": "20",
        "remark": "",
        "creatorId": "fec6d6b8-df5e-4fbd-8347-565f15c24b51",
        "creatorName": "Nathan Rolfson", //连sys_user表得到
        "createDate": "2024-04-21 21:38:01",
        "endDate": "",
        "startDate": "",
        "projectId": "bcff0ba9-0a65-4190-9bbc-16cf4962420e",
      	"projectTitle":"",//连dev_project表得到
        "releaseStatus": "99",
        "releaseDate": "",
        "changeLogRichText": "",
        "changeLog": ""
      },
  "error": null,
  "message": "ok"
}
```

## 删除版本

请求路径：/api/dev/versions

请求方法：DELETE

要求：

1. 逻辑删除 del_flag=1
2. releaseStatus=0时可以删除，其它状态不许删除
3. 操作时需向 dev_change_history 表写入关联数据。change_behavior=20、business_type=30、business_id=当前创建的这条数据versionId

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