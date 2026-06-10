# 文件管理(sys_file表)

## 上传文件

请求路径：/api/system/upload

请求方法：POST

要求：

1. name 格式为 UUID去横线重命名后的
2. 把文件暂存在本项目根目录下：static/uploads 下
3. 未登录用户不可上传文件
4. 限制文件大小 < 200M。

请求参数示例(body)：

```tex
file 二进制数据
```

响应数据示例：

```json
{
  "code": 0,
  "data": {
                "fileId": "文件id",
                "url": "文件访问URL",
                "name": "eddd69de2d534933a8ce285a9273579c.jpeg" //存储文件名(UUID去横线重命名后的),
                "type": "image/jpeg",
                "size": 4800000,
                "fileExt": ".jpeg",
                "originalName": "测试.jpeg",
                "path": "/uploads/",
                "fullPath": "/uploads/eddd69de2d534933a8ce285a9273579c.jpeg",
                "thumbnailPath": "/uploads/thumb", //图片专用。其它格式为null
                "thumbnailUrl": "/uploads/thumb/eddd69de2d534933a8ce285a9273539c.jpeg", //图片专用。其它格式为null
                "createDate": "2024-05-15 08:57:47"
            },
  "error": null,
  "message": "ok"
}
```

## 列表数据

请求路径：/api/system/files

请求方法：GET

要求：

1. 按创建时间倒序 - 最新的数据排在前面
2. 分页获取表数据。调用统一分页方法
3. 排序时仅支持：originalName、size、createDate 的排序

请求参数示例(query)：

```tex
//需分页,查询参数，排序参数
page=1&pageSize=20&originalName=&type=&fileExt=&sorts=
```

响应数据示例：

```json
{
  "code": 0,
  "data": {
    "items": [
      {
                "fileId": "文件id",
                "url": "文件访问URL",
                "name": "eddd69de2d534933a8ce285a9273579c.jpeg" //存储文件名(UUID去横线重命名后的),
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
    "total": 1
  },
  "error": null,
  "message": "ok"
}
```

