# 需求管理(dev_story表)

## 列表数据

请求路径：/api/dev/storys

请求方法：GET

要求：

1. 过滤掉已删除的数据
2. 按创建时间倒序 - 最新的数据排在前面
3. 分页获取表数据。调用统一分页方法
4. 排序时仅支持：storyStatus、storyLevel、storyTitle 的排序

请求参数示例(query)：

```tex
//需分页,查询参数，排序参数
page=1&pageSize=20&storyNum=&storyTitle=&projectId=&versionId=&moduleId=&storyStatus=&sorts=
```

响应数据示例：

```json
{
    "code": 0,
    "data": {
        "items": [
            {
                "storyId": "7921d829-de0f-4856-8c6e-8293d4306bc2",
                "storyTitle": "需求标题",
                "storyNum": 1000, //需求编号
                "creatorName": "创建人名称", //连sys_user表得到
                "creatorId": "53ae7f3e-fd1f-48a3-9f84-2f1112380249",
                "storyType": "20",
                "storyStatus": "10",
                "storyLevel": "10",
                "versionId": "e9d67560-b967-481e-b8be-29b7db039ce3",
                "version": "1.3.6", //连dev_version表得到
                "projectId": "344c319c-c49a-4145-84ec-676af87825c4",
                "projectTitle": "项目标题", //连dev_project表得到
                "moduleId": "add845d9-76e3-42ba-86d3-ec6707e92d2c",
                "moduleTitle": "模块标题", //连dev_module表得到
                "source": "10",
                "updateDate": "2024-05-15 08:57:47",
                "createDate": "2024-05-15 08:57:47",
                "userList":[
                    //根据表字段 userIds(id逗号分隔) 关联查询用户表
                    {
                        "userId": "fd8b5f2c-77c6-4e59-b81c-306c2fb85d44",
                        "avatar": "https://picsum.photos/100/100",
                        "realName": "Vben"
                    }
                ],
                "userIds": ["fd8b5f2c-77c6-4e59-b81c-306c2fb85d44"] //都逗号分隔的格式为数组
            },
        ],
        "total": 1
    },
    "error": null,
    "message": "ok"
}
```

## 列表数据(全量查询)

请求路径：/api/dev/storys/all

请求方法：GET

要求：

1. 过滤掉已删除的数据
2. 按创建时间倒序 - 最新的数据排在前面

请求参数示例(query)：

```tex
storyNum=&storyTitle=&projectId=&versionId=&moduleId=&storyStatus=
```

响应数据示例：

```json
{
    "code": 0,
    "data":  [
            {
                "storyId": "7921d829-de0f-4856-8c6e-8293d4306bc2",
                "storyTitle": "需求标题",
                "storyNum": 1000, //需求编号
                "creatorName": "创建人名称", //连sys_user表得到
                "creatorId": "53ae7f3e-fd1f-48a3-9f84-2f1112380249",
                "storyType": "20",
                "storyStatus": "10",
                "storyLevel": "10",
                "versionId": "e9d67560-b967-481e-b8be-29b7db039ce3",
                "version": "1.3.6", //连dev_version表得到
                "projectId": "344c319c-c49a-4145-84ec-676af87825c4",
                "projectTitle": "项目标题", //连dev_project表得到
                "moduleId": "add845d9-76e3-42ba-86d3-ec6707e92d2c",
                "moduleTitle": "模块标题", //连dev_module表得到
                "source": "10",
                "updateDate": "2024-05-15 08:57:47",
                "createDate": "2024-05-15 08:57:47",
                "userList":[
                    //根据表字段 userIds(id逗号分隔) 关联查询用户表
                    {
                        "userId": "fd8b5f2c-77c6-4e59-b81c-306c2fb85d44",
                        "avatar": "https://picsum.photos/100/100",
                        "realName": "Vben"
                    }
                ],
                "userIds": ["fd8b5f2c-77c6-4e59-b81c-306c2fb85d44"] //都逗号分隔的格式为数组
            },
        ],
    "error": null,
    "message": "ok"
}
```

## 创建需求

请求路径：/api/dev/storys

请求方法：POST

要求：

1. 操作时需向 dev_change_history 表写入关联数据。change_behavior=0、business_type=0、business_id=当前创建的这条数据storyId
2. 操作时需校验 需求标题、关联项目、关联版本、关联模块、需求类型。必填
3. userIds 以逗号分隔的方式存入表字段
4. fileIds 以逗号分隔的方式存入表字段
5. story_num 数据库字段要自增长

请求参数示例(body)：

```json
{
  "storyStatus": "0", // 字典STORY_STATUS值
  "storyType": "0", // 字典STORY_TYPE值
  "storyLevel": "0", // 字典STORY_LEVEL值
  "source": "0", // 字典STORY_SOURCE值
  "storyTitle": "需求标题",
  "storyRichText":"需求内容",
  "userIds": [
    "fd8b5f2c-77c6-4e59-b81c-306c2fb85d44", //用户id数组
    "68fd6081-f907-44a2-a309-4d856a3276e9"
  ],
  "projectId": "bfeaf474-55e7-409c-9dfe-cbef7d7510f3",
  "versionId": "2c24b19a-7356-4d7a-ad89-07ef4fda269f",
  "moduleId": "535efe60-1899-4070-8484-b400bfbec788",
  "fileIds": ["535efe60-1899-4070-8484-b400bfbec789"], //文件id数组
   "nodes":[
         {
        "nodeId": "7a94958e-c5fe-4d44-aa8d-0fd4001a72e0",
        "label": "需求提交",
        "value": "0",
        "sort": 1,
        "userId": "20a61684-41d0-4e03-b055-72a2844c86f5",
        "current": true,
        "nodeType": 0,
        "result": 0,
        "remark": "流程开始节点",
        "resultRichText": "",
        "businessType": "10",
        "businessId": "6e83cba6-66c2-4db1-947e-939b40acd8d2",
        "startDate": "2026-05-23 14:11:50", //或者null
        "endDate": "2026-05-24 14:11:50", //或者null
        "createDate": "2026-05-20 14:11:50"
    	},
        {
        "nodeId": "7a94958e-c5fe-4d44-aa8d-0fd4001a72e0",
        "label": "需求提交",
        "value": "0",
        "sort": 1,
        "userId": "20a61684-41d0-4e03-b055-72a2844c86f5",
        "current": true,
        "nodeType": 0,
        "result": 0,
        "remark": "流程开始节点",
        "resultRichText": "",
        "businessType": "10",
        "businessId": "6e83cba6-66c2-4db1-947e-939b40acd8d2",
        "startDate": "2026-05-23 14:11:50", //或者null
        "endDate": "2026-05-24 14:11:50", //或者null
        "createDate": "2026-05-20 14:11:50"
    	},
    ]
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

## 创建需求(批量)

请求路径：/api/dev/storys/batch

请求方法：POST

要求：

1. 操作时需向 dev_change_history 表写入关联数据。change_behavior=0、business_type=0、business_id=当前创建的这条数据storyId
2. 操作时需校验 需求标题、关联项目、关联版本、关联模块、需求类型。必填
3. userIds 以逗号分隔的方式存入表字段
4. fileIds 以逗号分隔的方式存入表字段
5. story_num 数据库字段要自增长

请求参数示例(body)：

```json
[
  {
      "storyStatus": "0", // 字典STORY_STATUS值
      "storyType": "0", // 字典STORY_TYPE值
      "storyLevel": "0", // 字典STORY_LEVEL值
      "source": "0", // 字典STORY_SOURCE值
      "storyTitle": "需求标题",
      "storyRichText":"需求内容",
      "userIds": [
        "fd8b5f2c-77c6-4e59-b81c-306c2fb85d44", //用户id数组
        "68fd6081-f907-44a2-a309-4d856a3276e9"
      ],
      "projectId": "bfeaf474-55e7-409c-9dfe-cbef7d7510f3",
      "versionId": "2c24b19a-7356-4d7a-ad89-07ef4fda269f",
      "moduleId": "535efe60-1899-4070-8484-b400bfbec788",
      "fileIds": ["535efe60-1899-4070-8484-b400bfbec789"] //文件id数组
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

## 修改需求

请求路径：/api/dev/storys/{storyId}

请求方法：PUT

要求：

1. 操作时需向 dev_change_history 表写入关联数据。change_behavior=10、business_type=0、business_id=当前创建的这条数据storyId
2. 操作时需校验 需求标题、关联项目、关联版本、关联模块、需求类型。必填
3. userIds 以逗号分隔的方式存入表字段
4. fileIds 以逗号分隔的方式存入表字段

请求参数示例(path)：

```url
storyId
```

请求参数示例(body)：

```json
{
  "storyStatus": "0", // 字典STORY_STATUS值
  "storyType": "0", // 字典STORY_TYPE值
  "storyLevel": "0", // 字典STORY_LEVEL值
  "source": "0", // 字典STORY_SOURCE值
  "storyTitle": "需求标题",
  "storyRichText":"需求内容",
  "userIds": [
    "fd8b5f2c-77c6-4e59-b81c-306c2fb85d44", //用户id数组
    "68fd6081-f907-44a2-a309-4d856a3276e9"
  ],
  "projectId": "bfeaf474-55e7-409c-9dfe-cbef7d7510f3",
  "versionId": "2c24b19a-7356-4d7a-ad89-07ef4fda269f",
  "moduleId": "535efe60-1899-4070-8484-b400bfbec788",
  "fileIds": ["535efe60-1899-4070-8484-b400bfbec789"] //文件id数组
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

## 修改需求字段

请求路径：/api/dev/storys/{storyId}/field

请求方法：PUT

要求：

1. 操作时需向 dev_change_history 表写入关联数据。change_behavior=10、business_type=0、business_id=storyId
2. 仅可修改：userIds、storyType、storyLevel、source

请求参数示例(path)：

```url
storyId
```

请求参数示例(body)：

```json
{
  "key": "storyType",
  "value": "10"
}
或者
{
  "key": "userIds",
  "value": ["fd8b5f2c-77c6-4e59-b81c-306c2fb85d44","68fd6081-f907-44a2-a309-4d856a3276e9"]
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

## 需求流转状态

请求路径：/api/dev/storys/{storyId}/next

请求方法：PUT

要求：

1. 操作时需向 dev_change_history 表写入关联数据。change_behavior=40、business_type=0、business_id=storyId、changeRichText=参数值的changeRichText

请求参数示例(path)：

```url
storyId
```

请求参数示例(body)：

```json
{
  "storyStatus": "10", //用于修改需求表的状态
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

## 需求详情

请求路径：/api/dev/storys/{storyNum}

请求方法：GET

请求参数示例(path)：

```url
storyNum
```

响应数据示例：

```json
{
  "code": 0,
  "data": {
    "storyId": "7921d829-de0f-4856-8c6e-8293d4306bc2",
    "storyTitle": "需求标题",
    "storyNum": 1000, //需求编号
    "creatorName": "创建人名称", //连sys_user表得到
    "creatorId": "53ae7f3e-fd1f-48a3-9f84-2f1112380249",
    "storyType": "20",
    "storyStatus": "10",
    "storyLevel": "10",
    "storyRichText": "",
    "versionId": "e9d67560-b967-481e-b8be-29b7db039ce3",
    "version": "1.3.6", //连dev_version表得到
    "projectId": "344c319c-c49a-4145-84ec-676af87825c4",
    "projectTitle": "项目标题", //连dev_project表得到
    "moduleId": "add845d9-76e3-42ba-86d3-ec6707e92d2c",
    "moduleTitle": "模块标题", //连dev_module表得到
    "source": "10",
    "updateDate": "2024-05-15 08:57:47",
    "createDate": "2024-05-15 08:57:47",
    "fileIds": ["文件id"], //都逗号分隔的格式为数组
    "taskList": [ //在任务表里查询出关联的storyId的数据。其数据模型和 任务详情接口一致
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
        "avatar": "", //根据 userId 连sys_user表得到
        "percent": 60 //进度字段。60=60% 是一个逻辑字段，计算方式：actualHours/planHours
      }
    ],
    "bugList": [ //在缺陷表里查询出关联的storyId的数据。其数据模型和 缺陷详情接口一致
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
        "avatar": "", //根据 userId 连sys_user表得到
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
      }
    ],
    "fileList": [
      //根据表字段 fileIds(id逗号分隔) 关联查询sys_file表
      {
        "fileId": "文件id",
        "url": "文件访问URL",
        "name": "eddd69de2d534933a8ce285a9273579c.jpeg", //存储文件名(UUID去横线重命名后的),
        "type": "image/jpeg",
        "size": 4800000,
        "fileExt": ".jpeg",
        "originalName": "测试.jpeg",
        "path": "/uploads/",
        "fullPath": "/uploads/eddd69de2d534933a8ce285a9273579c.jpeg",
        "thumbnailPath": "/uploads/thumb", //图片专用。其它格式为null
        "thumbnailUrl": "/uploads/thumb/eddd69de2d534933a8ce285a9273539c.jpeg", //图片专用。其它格式为null
        "creatorId": "",
        "creatorName": "", //连sys_user表得到
        "createDate": "2024-05-15 08:57:47"
      }
    ],
    "userIds": ["fd8b5f2c-77c6-4e59-b81c-306c2fb85d44"], //都逗号分隔的格式为数组
    "userList": [
      //根据表字段 userIds(id逗号分隔) 关联查询用户表
      {
        "userId": "fd8b5f2c-77c6-4e59-b81c-306c2fb85d44",
        "avatar": "",
        "realName": "Vben"
      }
    ]
  },
  "error": null,
  "message": "ok"
}
```

## 删除需求

请求路径：/api/dev/storys

请求方法：DELETE

要求：

1. 逻辑删除 del_flag=1
2. storyStatus=0时可以删除。其它状态不许删除
3. 操作时需向 dev_change_history 表写入关联数据。change_behavior=20、business_type=0、business_id=当前创建的这条数据storyId

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