# 任务管理(dev_task表)

## 列表数据

请求路径：/api/dev/tasks

请求方法：GET

要求：

1. 过滤掉已删除的数据
2. 按创建时间倒序 - 最新的数据排在前面
3. 分页获取表数据。调用统一分页方法
4. 排序时仅支持：taskTitle、taskStatus、startDate、endDate 的排序

请求参数示例(query)：

```tex
//需分页,查询参数，排序参数
page=1&pageSize=20&taskNum=&taskTitle=&projectId=&versionId=&taskStatus=&sorts=
```

响应数据示例：

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "taskId": "015ea079-a0cc-409d-ae1a-bf6a562538a2",
        "storyId": "516a9893-9144-4835-ade9-4e18b2d4e2ae",
        "storyTitle": "", //连dev_story表得到
        "moduleId": "a15bff70-3e48-4619-8153-8c69edcf8f85",
        "moduleTitle": "EXE端", //连dev_module表得到
        "versionId": "038116db-a388-4255-94da-548d706718d0",
        "version": "1.3.4", //连dev_version表得到
        "projectId": "d8fd7341-b5ca-4548-88c5-9b6e67143e50",
        "projectTitle": "voluptate", //连dev_project表得到
        "taskTitle": "任务标题",
        "taskNum": 1000, //任务编号
        "taskStatus": "0",
        "taskType": "110",
        "planHours": 4,
        "actualHours": 10,
        "endDate": "2022-07-06 04:23:40",
        "startDate": "2022-07-06 04:23:40",
        "createDate": "2022-07-06 04:23:40",
        "creatorId": "68fd6081-f907-44a2-a309-4d856a3276e9",
        "creatorName": "Admin", //根据 creatorId 连sys_user表得到
        "userId": "68fd6081-f907-44a2-a309-4d856a3276e9", //执行人id
        "realName": "Admin", //根据 userId 连sys_user表得到
        "avatar": "https://picsum.photos/100/100", //根据 userId 连sys_user表得到
        "percent": 60 //进度字段。60=60% 是一个逻辑字段，计算方式：actualHours/planHours
      }
    ],
    "total": 1
  },
  "error": null,
  "message": "ok"
}
```

## 列表数据(全量查询)

请求路径：/api/dev/tasks/all

请求方法：GET

要求：

1. 过滤掉已删除的数据
2. 按创建时间倒序 - 最新的数据排在前面

请求参数示例(query)：

```tex
storyId=&taskNum=&taskTitle=&projectId=&versionId=&taskStatus=
```

响应数据示例：

```json
{
  "code": 0,
  "data": [
      {
        "taskId": "015ea079-a0cc-409d-ae1a-bf6a562538a2",
        "storyId": "516a9893-9144-4835-ade9-4e18b2d4e2ae",
        "storyTitle": "", //连dev_story表得到
        "moduleId": "a15bff70-3e48-4619-8153-8c69edcf8f85",
        "moduleTitle": "EXE端", //连dev_module表得到
        "versionId": "038116db-a388-4255-94da-548d706718d0",
        "version": "1.3.4", //连dev_version表得到
        "projectId": "d8fd7341-b5ca-4548-88c5-9b6e67143e50",
        "projectTitle": "voluptate", //连dev_project表得到
        "taskTitle": "任务标题",
        "taskNum": 1000, //任务编号
        "taskStatus": "0",
        "taskType": "110",
        "planHours": 4,
        "actualHours": 10,
        "endDate": "2022-07-06 04:23:40",
        "startDate": "2022-07-06 04:23:40",
        "createDate": "2022-07-06 04:23:40",
        "creatorId": "68fd6081-f907-44a2-a309-4d856a3276e9",
        "creatorName": "Admin", //根据 creatorId 连sys_user表得到
        "userId": "68fd6081-f907-44a2-a309-4d856a3276e9", //执行人id
        "realName": "Admin", //根据 userId 连sys_user表得到
        "avatar": "https://picsum.photos/100/100", //根据 userId 连sys_user表得到
        "percent": 60 //进度字段。60=60% 是一个逻辑字段，计算方式：actualHours/planHours
      }
    ],
  "error": null,
  "message": "ok"
}
```

## 创建任务

请求路径：/api/dev/tasks

请求方法：POST

要求：

1. 操作时需向 dev_change_history 表写入关联数据。change_behavior=0、business_type=10、business_id=当前创建的这条数据taskId
2. 操作时需校验：任务标题、开始时间、结束时间、计划工时、关联项目、执行人。必填
3. **结束时间必须大于或等于开始时间**
4. task_num 数据库字段要自增长

请求参数示例(body)：

```json
{
  "planHours": 2,
  "taskStatus": "0", // 字典TASK_STATUS值
  "taskType": "0", // 字典TASK_TYPE值
  "taskTitle": "任务标题",
  "projectId": "d8fd7341-b5ca-4548-88c5-9b6e67143e50",
  "taskRichText": "任务内容：1231",
  "versionId": "038116db-a388-4255-94da-548d706718d0",
  "moduleId": "a15bff70-3e48-4619-8153-8c69edcf8f85",
  "storyId": "449de555-158a-495f-b097-7653a3385d69",
  "userId": "fd8b5f2c-77c6-4e59-b81c-306c2fb85d44",
  "endDate": "2026-05-01 17:16:14",
  "startDate": "2026-05-31 17:16:14"
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

## 创建任务(批量)

请求路径：/api/dev/tasks/batch

请求方法：POST

要求：

1. 操作时需向 dev_change_history 表写入关联数据。change_behavior=0、business_type=10、business_id=当前创建的这条数据taskId
2. 操作时需校验：任务标题、开始时间、结束时间、计划工时、关联项目、执行人。必填
3. **结束时间必须大于或等于开始时间**
4. task_num 数据库字段要自增长

请求参数示例(body)：

```json
[
  {
      "planHours": 2,
      "taskStatus": "0", // 字典TASK_STATUS值
      "taskType": "0", // 字典TASK_TYPE值
      "taskTitle": "任务标题",
      "projectId": "d8fd7341-b5ca-4548-88c5-9b6e67143e50",
      "taskRichText": "任务内容：1231",
      "versionId": "038116db-a388-4255-94da-548d706718d0",
      "moduleId": "a15bff70-3e48-4619-8153-8c69edcf8f85",
      "storyId": "449de555-158a-495f-b097-7653a3385d69",
      "userId": "fd8b5f2c-77c6-4e59-b81c-306c2fb85d44",
      "endDate": "2026-05-01 17:16:14",
      "startDate": "2026-05-31 17:16:14"
 }
]
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

## 修改任务

请求路径：/api/dev/tasks/{taskId}

请求方法：PUT

要求：

1. 操作时需向 dev_change_history 表写入关联数据。change_behavior=10、business_type=10、business_id=当前创建的这条数据taskId
2. 操作时需校验：任务标题、开始时间、结束时间、计划工时、关联项目、执行人。必填
3. **结束时间必须大于或等于开始时间**

请求参数示例(path)：

```url
taskId
```

请求参数示例(body)：

```json
{
  "planHours": 2,
  "taskStatus": "0", // 字典TASK_STATUS值
  "taskType": "0", // 字典TASK_TYPE值
  "taskTitle": "任务标题",
  "projectId": "d8fd7341-b5ca-4548-88c5-9b6e67143e50",
  "taskRichText": "任务内容：1231",
  "versionId": "038116db-a388-4255-94da-548d706718d0",
  "moduleId": "a15bff70-3e48-4619-8153-8c69edcf8f85",
  "storyId": "449de555-158a-495f-b097-7653a3385d69",
  "userId": "fd8b5f2c-77c6-4e59-b81c-306c2fb85d44",
  "endDate": "2026-05-01 17:16:14",
  "startDate": "2026-05-31 17:16:14"
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

## 修改任务字段

请求路径：/api/dev/tasks/{taskId}/field

请求方法：PUT

要求：

1. 操作时需向 dev_change_history 表写入关联数据。change_behavior=10、business_type=10、business_id=taskId
2. 仅可修改：userId、taskType、startDate、endDate
3. 修改 startDate、endDate 时，需校验 结束时间不能小于开始时间、开始时间不能大于结束时间

请求参数示例(path)：

```url
taskId
```

请求参数示例(body)：

```json
{
  "key": "taskType",
  "value": "10"
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

## 任务流转状态

请求路径：/api/dev/tasks/{taskId}/next

请求方法：PUT

要求：

1. 操作时需向 dev_change_history 表写入关联数据。change_behavior=40、business_type=10、business_id=taskId、changeRichText=参数值的changeRichText

请求参数示例(path)：

```url
taskId
```

请求参数示例(body)：

```json
{
  "taskStatus": "10", //用于修改任务表的状态
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

## 任务详情

请求路径：/api/dev/tasks/{taskNum}

请求方法：GET

请求参数示例(path)：

```url
taskNum
```

响应数据示例：

```json
{
  "code": 0,
  "data": {
        "taskId": "015ea079-a0cc-409d-ae1a-bf6a562538a2",
        "storyId": "516a9893-9144-4835-ade9-4e18b2d4e2ae",
        "storyTitle": "", //连dev_story表得到
        "moduleId": "a15bff70-3e48-4619-8153-8c69edcf8f85",
        "moduleTitle": "EXE端", //连dev_module表得到
        "versionId": "038116db-a388-4255-94da-548d706718d0",
        "version": "1.3.4", //连dev_version表得到
        "projectId": "d8fd7341-b5ca-4548-88c5-9b6e67143e50",
        "projectTitle": "voluptate", //连dev_project表得到
        "taskTitle": "任务标题",
        "taskNum": 1000, //任务编号
        "taskRichText": "任务内容",
        "taskStatus": "0",
        "taskType": "110",
        "planHours": 4,
        "actualHours": 10,
        "endDate": "2022-07-06 04:23:40",
        "startDate": "2022-07-06 04:23:40",
        "createDate": "2022-07-06 04:23:40",
        "creatorId": "68fd6081-f907-44a2-a309-4d856a3276e9",
        "creatorName": "Admin", //根据 creatorId 连sys_user表得到
        "userId": "68fd6081-f907-44a2-a309-4d856a3276e9", //执行人id
        "realName": "Admin", //根据 userId 连sys_user表得到
        "avatar": "https://picsum.photos/100/100",
        "percent": 60 //进度字段。60=60% 是一个逻辑字段，计算方式：actualHours/planHours
      },
  "error": null,
  "message": "ok"
}
```

## 删除任务

请求路径：/api/dev/tasks

请求方法：DELETE

要求：

1. 逻辑删除 del_flag=1
2. taskStatus=0时可以删除。其它状态不许删除
3. 操作时需向 dev_change_history 表写入关联数据。change_behavior=20、business_type=10、business_id=当前创建的这条数据taskId

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
