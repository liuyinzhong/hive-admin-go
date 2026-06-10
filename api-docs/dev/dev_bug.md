# 缺陷管理(dev_bug表)

## 列表数据

请求路径：/api/dev/bugs

请求方法：GET

要求：

1. 过滤掉已删除的数据
2. 按创建时间倒序 - 最新的数据排在前面
3. 分页获取表数据。调用统一分页方法
4. 排序时仅支持：bugTitle、bugStatus、bugConfirmStatus、bugLevel 的排序

请求参数示例(query)：

```tex
//需分页,查询参数，排序参数
page=1&pageSize=20&bugNum=&bugTitle=&projectId=&versionId=&moduleId=&bugStatus=&sorts=
```

响应数据示例：

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "bugId": "ff4975a1-2e57-4b1a-89ba-ee8e87f0ecf2",
        "bugTitle": "缺陷标题",
        "bugNum": 1000, //缺陷编号
        "bugStatus": "99",
        "bugConfirmStatus": "0",
        "bugLevel": "10",
        "bugSource": "120",
        "bugType": "50",
        "bugEnv": "20",
        "bugUa": "Mozilla/5.0 (Windows NT 6.1; Win64; x64)",
        "userId": "fd8b5f2c-77c6-4e59-b81c-306c2fb85d44",
        "avatar": "https://picsum.photos/100/100", //根据 userId 连sys_user表得到
        "realName": "Vben", //根据 userId 连sys_user表得到
        "creatorName": "Vben", //根据 creatorId 连sys_user表得到
        "creatorId": "fd8b5f2c-77c6-4e59-b81c-306c2fb85d44",
        "versionId": "ca697069-5ce0-425e-a2fb-a25e4250ec20",
        "version": "1.7.4", //连dev_version表得到
        "moduleId": "dff65d3b-3a45-4ec1-a9ec-0397a810424c",
        "moduleTitle": "平台管理端", //连dev_module表得到
        "projectId": "0a0de794-26d6-47fb-bc20-5e3e2b1a1582",
        "projectTitle": "crudelis", //连dev_project表得到
        "storyId": "b4a0b759-ca17-420c-b006-bb4da72edd49",
        "storyTitle": "Umeru", //连dev_story表得到
        "updateDate": "2022-07-06 04:23:40",
        "createDate": "2022-07-06 04:23:40"
      },
    ],
    "total": 1
  },
  "error": null,
  "message": "ok"
}
```

## 列表数据(全量查询)

请求路径：/api/dev/bugs/all

请求方法：GET

要求：

1. 过滤掉已删除的数据
2. 按创建时间倒序 - 最新的数据排在前面

请求参数示例(query)：

```tex
storyId=&bugNum=&bugTitle=&projectId=&versionId=&moduleId=&bugStatus=
```

响应数据示例：

```json
{
  "code": 0,
  "data":[
      {
        "bugId": "ff4975a1-2e57-4b1a-89ba-ee8e87f0ecf2",
        "bugTitle": "缺陷标题",
        "bugNum": 1000, //缺陷编号
        "bugStatus": "99",
        "bugConfirmStatus": "0",
        "bugLevel": "10",
        "bugSource": "120",
        "bugType": "50",
        "bugEnv": "20",
        "bugUa": "Mozilla/5.0 (Windows NT 6.1; Win64; x64)",
        "userId": "fd8b5f2c-77c6-4e59-b81c-306c2fb85d44",
        "avatar": "https://picsum.photos/100/100", //根据 userId 连sys_user表得到
        "realName": "Vben", //根据 userId 连sys_user表得到
        "creatorName": "Vben", //根据 creatorId 连sys_user表得到
        "creatorId": "fd8b5f2c-77c6-4e59-b81c-306c2fb85d44",
        "versionId": "ca697069-5ce0-425e-a2fb-a25e4250ec20",
        "version": "1.7.4", //连dev_version表得到
        "moduleId": "dff65d3b-3a45-4ec1-a9ec-0397a810424c",
        "moduleTitle": "平台管理端", //连dev_module表得到
        "projectId": "0a0de794-26d6-47fb-bc20-5e3e2b1a1582",
        "projectTitle": "crudelis", //连dev_project表得到
        "storyId": "b4a0b759-ca17-420c-b006-bb4da72edd49",
        "storyTitle": "Umeru", //连dev_story表得到
        "updateDate": "2022-07-06 04:23:40",
        "createDate": "2022-07-06 04:23:40"
      },
    ],
  "error": null,
  "message": "ok"
}
```

## 创建缺陷

请求路径：/api/dev/bugs

请求方法：POST

要求：

1. 操作时需向 dev_change_history 表写入关联数据。change_behavior=0、business_type=20、business_id=当前创建的这条数据bugId
2. 操作时需校验：缺陷标题、关联项目、修复人 必填
3. bug_num 数据库字段要自增长

```json
{
  "bugLevel": "0",  // 字典BUG_LEVEL值
  "bugEnv": "0", // 字典BUG_ENV值
  "bugStatus": "0",  // 字典BUG_STATUS值
  "bugSource": "0",  // 字典BUG_SOURCE值
  "bugType": "0",  // 字典BUG_TYPE值
  "bugUa": "",
  "bugTitle": "缺陷标题",
  "projectId": "d8fd7341-b5ca-4548-88c5-9b6e67143e50",
  "bugRichText": "缺陷内容",
  "versionId": "038116db-a388-4255-94da-548d706718d0",
  "moduleId": "a15bff70-3e48-4619-8153-8c69edcf8f85",
  "storyId": "449de555-158a-495f-b097-7653a3385d69",
  "userId": "fd8b5f2c-77c6-4e59-b81c-306c2fb85d44" //修复人id
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

## 创建缺陷(批量)

请求路径：/api/dev/bugs/batch

请求方法：POST

要求：

1. 操作时需向 dev_change_history 表写入关联数据。change_behavior=0、business_type=20、business_id=当前创建的这条数据bugId
2. 操作时需校验：缺陷标题、关联项目、修复人 必填
3. bug_num 数据库字段要自增长

请求参数示例(body)：

```json
[
    {
      "bugLevel": "0",  // 字典BUG_LEVEL值
      "bugEnv": "0", // 字典BUG_ENV值
      "bugStatus": "0",  // 字典BUG_STATUS值
      "bugSource": "0",  // 字典BUG_SOURCE值
      "bugType": "0",  // 字典BUG_TYPE值
      "bugUa": "",
      "bugTitle": "缺陷标题",
      "projectId": "d8fd7341-b5ca-4548-88c5-9b6e67143e50",
      "bugRichText": "缺陷内容",
      "versionId": "038116db-a388-4255-94da-548d706718d0",
      "moduleId": "a15bff70-3e48-4619-8153-8c69edcf8f85",
      "storyId": "449de555-158a-495f-b097-7653a3385d69",
      "userId": "fd8b5f2c-77c6-4e59-b81c-306c2fb85d44" //修复人id
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

## 修改缺陷

请求路径：/api/dev/bugs/{bugId}

请求方法：PUT

要求：

1. 操作时需向 dev_change_history 表写入关联数据。change_behavior=10、business_type=20、business_id=当前创建的这条数据bugId
2. 操作时需校验：缺陷标题、关联项目、修复人 必填

请求参数示例(path)：

```url
bugId
```

请求参数示例(body)：

```json
{
  "bugLevel": "0",  // 字典BUG_LEVEL值
  "bugEnv": "0", // 字典BUG_ENV值
  "bugStatus": "0",  // 字典BUG_STATUS值
  "bugSource": "0",  // 字典BUG_SOURCE值
  "bugType": "0",  // 字典BUG_TYPE值
  "bugUa": "",
  "bugTitle": "缺陷标题",
  "projectId": "d8fd7341-b5ca-4548-88c5-9b6e67143e50",
  "bugRichText": "缺陷内容",
  "versionId": "038116db-a388-4255-94da-548d706718d0",
  "moduleId": "a15bff70-3e48-4619-8153-8c69edcf8f85",
  "storyId": "449de555-158a-495f-b097-7653a3385d69",
  "userId": "fd8b5f2c-77c6-4e59-b81c-306c2fb85d44" //修复人id
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

## 修改缺陷字段

请求路径：/api/dev/bugs/{bugId}/field

请求方法：PUT

要求：

1. 操作时需向 dev_change_history 表写入关联数据。change_behavior=10、business_type=20、business_id=bugId
2. 仅可修改：userId、bugLevel、bugEnv、bugType、bugSource

请求参数示例(path)：

```url
bugId
```

请求参数示例(body)：

```json
{
  "key": "bugType",
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

## 缺陷流转状态

请求路径：/api/dev/bugs/{bugId}/next

请求方法：PUT

要求：

1. 操作时需向 dev_change_history 表写入关联数据。change_behavior=40、business_type=20、business_id=bugId、changeRichText=参数值的changeRichText

请求参数示例(path)：

```url
bugId
```

请求参数示例(body)：

```json
{
  "bugStatus": "10", //用于修改缺陷表的状态
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

## 确认缺陷

请求路径：/api/dev/bugs/{bugId}/confirm

请求方法：PUT

要求：

1. 操作时需向 dev_change_history 表写入关联数据。change_behavior=50、business_type=20、business_id=bugId

请求参数示例(path)：

```url
bugId
```

请求参数示例(body)：

```json
{
  "bugConfirmStatus":"1",  //=1时把当前这条缺陷数据的bugStatus设置为10
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

## 缺陷详情

请求路径：/api/dev/bugs/{bugNum}

请求方法：GET

请求参数示例(path)：

```url
bugNum
```

响应数据示例：

```json
{
  "code": 0,
  "data": {
        "bugId": "ff4975a1-2e57-4b1a-89ba-ee8e87f0ecf2",
        "bugTitle": "缺陷标题",
        "bugNum": 1000, //缺陷编号
        "bugRichText": "内容",
        "bugStatus": "99",
        "bugConfirmStatus": "0",
        "bugLevel": "10",
        "bugSource": "120",
        "bugType": "50",
        "bugEnv": "20",
        "bugUa": "Mozilla/5.0 (Windows NT 6.1; Win64; x64)",
        "userId": "fd8b5f2c-77c6-4e59-b81c-306c2fb85d44",
        "avatar": "https://picsum.photos/100/100", //根据 userId 连sys_user表得到
        "realName": "Vben", //根据 userId 连sys_user表得到
        "creatorName": "Vben", //根据 creatorId 连sys_user表得到
        "creatorId": "fd8b5f2c-77c6-4e59-b81c-306c2fb85d44",
        "versionId": "ca697069-5ce0-425e-a2fb-a25e4250ec20",
        "version": "1.7.4", //连dev_version表得到
        "moduleId": "dff65d3b-3a45-4ec1-a9ec-0397a810424c",
        "moduleTitle": "平台管理端", //连dev_module表得到
        "projectId": "0a0de794-26d6-47fb-bc20-5e3e2b1a1582",
        "projectTitle": "crudelis", //连dev_project表得到
        "storyId": "b4a0b759-ca17-420c-b006-bb4da72edd49",
        "storyTitle": "Umeru", //连dev_story表得到
        "updateDate": "2022-07-06 04:23:40",
        "createDate": "2022-07-06 04:23:40"
      },
  "error": null,
  "message": "ok"
}
```

## 删除缺陷

请求路径：/api/dev/bugs

请求方法：DELETE

要求：

1. 逻辑删除 del_flag=1
2. bugStatus=0时可以删除。其它状态不许删除
3. 操作时需向 dev_change_history 表写入关联数据。change_behavior=20、business_type=20、business_id=当前创建的这条数据bugId

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