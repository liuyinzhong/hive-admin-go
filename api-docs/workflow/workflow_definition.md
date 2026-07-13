# 流程定义接口

## 获取流程定义列表

请求路径：/api/workflow/definitions

请求方法：GET

请求参数示例(query)：

```text
page=1&pageSize=20&definitionName=&definitionKey=&category=&status=0,1&sorts=updateDate,desc
```

响应数据示例：

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "definitionId": "cc1a8564-37e7-47df-ad60-6c0a7f199d31",
        "definitionKey": "story_approval",
        "definitionName": "需求审批流程",
        "category": "dev",
        "status": "0",
        "version": 0,
        "flowData": "{\"nodes\":[],\"edges\":[]}",
        "remark": "用于需求审批",
        "creatorId": "cc1a8564-37e7-47df-ad60-6c0a7f199d31",
        "creatorName": "管理员",
        "createDate": "2026-05-18 15:30:26",
        "updateDate": "2026-05-18 15:30:26"
      }
    ],
    "total": 1
  },
  "error": null,
  "message": "ok"
}
```

## 获取所有流程定义

请求路径：/api/workflow/definitions/all

请求方法：GET

请求参数示例(query)：

```text
definitionName=&category=dev&status=1
```

## 创建流程定义

请求路径：/api/workflow/definitions

请求方法：POST

请求参数示例(body)：

```json
{
  "definitionKey": "story_approval",
  "definitionName": "需求审批流程",
  "category": "dev",
  "flowData": "{\"nodes\":[],\"edges\":[]}",
  "remark": "用于需求审批"
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

## 流程定义详情

请求路径：/api/workflow/definitions/{definitionId}

请求方法：GET

路径参数：

```text
definitionId
```

## 更新流程定义

请求路径：/api/workflow/definitions/{definitionId}

请求方法：PUT

请求参数示例(body)：

```json
{
  "definitionKey": "story_approval",
  "definitionName": "需求审批流程",
  "category": "dev",
  "flowData": "{\"nodes\":[],\"edges\":[]}",
  "remark": "用于需求审批"
}
```

## 保存流程画布

请求路径：/api/workflow/definitions/{definitionId}/canvas

请求方法：PUT

请求参数示例(body)：

```json
{
  "flowData": "{\"nodes\":[],\"edges\":[]}"
}
```

## 发布流程定义

请求路径：/api/workflow/definitions/{definitionId}/publish

请求方法：PUT

说明：发布会将状态更新为已发布，并递增 version。

## 更新流程定义状态

请求路径：/api/workflow/definitions/{definitionId}/status

请求方法：PUT

请求参数示例(body)：

```json
{
  "status": "2"
}
```

状态说明：

```text
0 草稿
1 已发布
2 已停用
```

## 删除流程定义

请求路径：/api/workflow/definitions

请求方法：DELETE

请求参数示例(body)：

```json
[
  "cc1a8564-37e7-47df-ad60-6c0a7f199d31"
]
```
