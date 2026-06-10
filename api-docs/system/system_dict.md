## 字典管理 （sys_dict表）

### 列表数据

请求路径：/api/system/dicts

要求：

1. 按pid字段组为树结构。子级用 children表示 。再把组装好的树结构数据响应给前端
2. 排序时仅支持：label、type、value
3. 如果 value 可以转为 int类型。则转int后在参与排序。如果不可以转为int类型。着按照字母顺序排
4. 当不指定任何排序时。一级数据默认为：title desc。二级数据默认为：value desc。可根据pid来区分一级、二级数据

请求方法：GET

请求参数示例(query)：

```text
//无需分页
label=&value=&type=&sorts=
```

响应数据示例：

```json
{
  "code": 0,
  "data": [
    {
      "id": "a4fed3fb-137c-4e93-99f2-1595ca3e6209",
      "pid": null,
      "label": "变更行为",
      "value": null,
      "type": "CHANGE_BEHAVIOR",
      "remark": null,
      "color": null,
      "status": 1,
      "createDate": "2026/01/30 01:39:20",
      "updateDate": "2026/01/30 01:39:20",
      "children": [
        {
          "id": "14aa690b-ba87-46ee-89ac-a7158f964a42",
          "pid": "a4fed3fb-137c-4e93-99f2-1595ca3e6209",
          "label": "流转",
          "value": 30,
          "type": "CHANGE_BEHAVIOR",
          "remark": null,
          "color": null,
          "status": 1,
          "createDate": "2026/01/30 01:39:20",
          "updateDate": "2026/01/30 01:39:20"
        }
      ]
    }
  ],
  "error": null,
  "message": "ok"
}
```

### 创建字典

请求路径：/api/system/dicts

请求方法：POST

要求：

1. 操作时需校验。同一 type 下 label 唯一
2. 操作时需校验字典类型是否已存在

请求参数示例(body)：

```json
{"type":"字典类型","label":"字典标题","value":"","color":"default","status":1,"remark":"备注"}
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

### 字典详情

请求路径：/api/system/dicts/{dictId}

请求方法：GET

请求参数示例(path)：

```url
dictId
```

响应数据示例：

```json
{
  "code": 0,
  "data": {
    "id": "a4fed3fb-137c-4e93-99f2-1595ca3e6209",
    "pid": null,
    "label": "变更行为",
    "value": null,
    "type": "CHANGE_BEHAVIOR",
    "remark": null,
    "color": null,
    "status": 1,
    "createDate": "2026/01/30 01:39:20",
    "updateDate": "2026/01/30 01:39:20",
    "children": [], //固定为空数组
  },
  "error": null,
  "message": "ok"
}
```

### 修改字典

请求路径：/api/system/dicts/{id}

请求方法：PUT

要求：

1. 操作时需校验。同一 type 下 label 唯一
2. 操作时需校验字典类型是否已存在

请求参数示例(path)：

```url
id
```

请求参数示例(body)：

```json
{"type":"字典类型","label":"字典标题","value":"","color":"default","status":1,"remark":"备注"}
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

### 修改字典状态

请求路径：/api/system/dicts/{id}/status

请求方法：PUT

要求：

1. 当修改状态时，所有子数据（无论层级深度）都会同步更新为相同状态

请求参数示例(path)：

```url
id
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

### 删除字典

请求路径：/api/system/dicts

请求方法：DELETE

要求：

1. 禁止删除有子项的字典
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
