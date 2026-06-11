# 节点管理(dev_node表)

## 列表数据

请求路径：/api/dev/nodes

请求方法：GET

要求：

1. 按sort 排序，升序
2. 无需分页，全量查询

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
  ],
  "error": null,
  "message": "ok"
}
```

## 创建节点

请求路径：/api/dev/nodes

请求方法：POST

要求：

1. 操作时需校验businessId必填

2. 操作时需根据sort值 把 >=sort的其它同一个businessId下的数据的sort值，基于本次入参值依次+1。保证添加节点后顺序不乱。
   举例：在businessId=A的数据有5条：
   [
   {label:"标题1",sort:1},
   {label:"标题2",sort:2},
   {label:"标题3",sort:3},
   {label:"标题4",sort:4},
   {label:"标题5",sort:5},
   ]

   此时我创建节点入参 {label:"标题A",sort:2}

   那么创建后的顺序为：
   [
   {label:"标题1",sort:1},
   {label:"标题A",sort:2},
   {label:"标题2",sort:3},
   {label:"标题3",sort:4},
   {label:"标题4",sort:5},
   {label:"标题5",sort:6},
   ]

请求参数示例(body)：

```json
{
        "label": "需求提交",
        "value": "0",
        "sort": 1,
        "userId": "20a61684-41d0-4e03-b055-72a2844c86f5",
        "nodeType": 0,
        "remark": "流程开始节点",
        "businessType": "10",
        "businessId": "6e83cba6-66c2-4db1-947e-939b40acd8d2"
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

## 删除节点

请求路径：/api/dev/nodes

请求方法：DELETE

要求：

1. nodeType=0 或者 nodeType=3的数据不可删除
2. current=true的节点不可删除
3. sort 小于 current=true的节点不可删除

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

## 节点审批

请求路径：/api/dev/nodes/{nodeId}/approve

请求方法：PUT

要求：

1. 仅nodeType=2（审批类型）的节点可以审批
2. 审批后设置result值（0=待审批 1=通过 2=拒绝）
3. 审批后设置resultRichText（审核内容）
4. 审批完成后根据result判断是否流转到下一个节点
5. 把当前节点的 endDate 设置为当前时间

请求参数示例(path)：

```url
nodeId
```

请求参数示例(body)：

```json
{
  "result": 1,
  "resultRichText": "审批意见内容"
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

## 节点流转

请求路径：/api/dev/nodes/{nodeId}/next

请求方法：PUT

要求：

1. 把当前节点设置为非当前节点（current=0）
2. 根据sort值找到下一个节点，设置其为当前节点（current=1）
3. 当有上一个节点时，把上个节点的 endDate 设置为现在时间
4. 当有下一个节点时，把下个节点的 startDate 设置为现在时间
5. 当下一个节点node_type=3时。把下个节点的 startDate 、endDate设置为现在时间

请求参数示例(path)：

```url
nodeId
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