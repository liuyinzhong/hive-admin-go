# 部门管理 （sys_dept表）

## 列表数据

请求路径：/api/system/depts

请求方法：GET

要求：

1. 按pid字段组为树结构。子级用 children表示 。再把组装好的树结构数据响应给前端

请求参数示例(query)：

```tex
//无需分页
name=
```

响应数据示例：

```json
{
  "code": 0,
  "data": [
    {
      "deptId": "c7894d44-b9da-4e28-b590-cbf0b12b27ee",
      "pid": null,
      "deptTitle": "Clothing",
      "status": 1,
      "createDate": "2021/05/26 18:33:18",
      "remark": "Vomica tergo carbo demulceo auditor conturbo.",
      "children": [
        {
          "deptId": "663e1675-0cfd-4686-9cfa-324407c333bf",
          "pid": "c7894d44-b9da-4e28-b590-cbf0b12b27ee",
          "deptTitle": "Grocery",
          "status": 0,
          "createDate": "2023/04/21 04:45:27",
          "remark": "Deleo assumenda capio."
        }
      ]
    }
  ],
  "error": null,
  "message": "ok"
}
```

## 创建部门

请求路径：/api/system/depts

请求方法：POST

要求：

1. 操作时需校验名称是否已存在

请求参数示例(body)：

```json
{"deptTitle":"部门名称","status":1,"pid":"d522195c-b0d9-4538-b64a-f3530eb14612","remark":"备注"}
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

## 部门详情

请求路径：/api/system/depts/{deptId}

请求方法：GET

请求参数示例(path)：

```url
deptId
```

响应数据示例：

```json
{
  "code": 0,
  "data":{
    "deptId": "c7894d44-b9da-4e28-b590-cbf0b12b27ee",
    "pid": null,
    "deptTitle": "Clothing",
    "status": 1,
    "createDate": "2021/05/26 18:33:18",
    "remark": "Vomica tergo carbo demulceo auditor conturbo.",
    "children": [], //固定为空数组
  },
  "error": null,
  "message": "ok"
}
```

## 修改部门

请求路径：/api/system/depts/{deptId}

请求方法：PUT

要求：

1. 操作时需校验名称是否已存在

请求参数示例(path)：

```url
deptId
```

请求参数示例(body)：

```json
{"deptTitle":"部门名称","status":1,"pid":"d522195c-b0d9-4538-b64a-f3530eb14612","remark":"备注"}
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

## 删除部门

请求路径：/api/system/depts

请求方法：DELETE

要求：

1. 禁止删除有绑定用户的部门
2. 禁止删除有子项的部门
3. 逻辑删除 del_flag=1
4. 清理 sys_user_dept 关联表数据

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