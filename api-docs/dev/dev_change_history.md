# 变更记录管理（dev_change_history表）

## 列表数据

请求路径：/api/dev/changeHistory

请求方法：GET

要求：

1. 按创建时间倒序 - 最新的数据排在前面

请求参数示例(query)：

```tex
businessId=
```

响应数据示例：

```json
{
  "code": 0,
  "data": [
    {
      "changeId": "031d3137-82ec-4321-b477-5a8e19019612",
      "changeBehavior": "30",
      "changeRichText": "内容",
      "creatorId": "2d0058c0-6347-4b73-92fa-c1b12f5e6454",
      "creatorName": "Jack", //连sys_user表得到
      "businessId": "ba02d29d-65b0-4cf4-a35c-75ce29865f90",
      "businessType": "10",
      "extendJson": "",
      "createDate": "2022-05-26 01:23:36",
      "updateDate": "2022-05-26 01:23:36"
    }
  ],
  "error": null,
  "message": "ok"
}
```

## 插入数据(评论)

请求路径：/api/dev/changeHistory

请求方法：POST

```json
{
  "businessId": "0",
  "businessType": "0",
  "changeBehavior": "0",
  "changeRichText": "0",
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