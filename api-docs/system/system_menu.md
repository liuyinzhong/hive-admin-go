# 菜单管理 （sys_menu表）

## 列表数据

请求路径：/api/system/menus

请求方法：GET

要求：

1. 把路由相关的字段在 meta 里体现。
2. 按 order 排序。升序
3. 按pid字段组为树结构。子级用 children表示 。再把组装好的树结构数据响应给前端
4. 查询参数只查询一级数据

请求参数示例(query)：

```tex
name=&path=&type=&status=
```

响应数据示例：

```json
{
  "code": 0,
  "data": [
    {
      "authCode": "sys:workspace",
      "children": [
        {
          "authCode": "sys:analytics",
          "children": [],
          "component": "/dashboard/analytics/index",
          "id": "7c17031e-63fd-4f5d-8c00-17d85673457c",
          "meta": {
            "activeIcon": "string",
            "activePath": "string",
            "affixTab": true,
            "affixTabOrder": 0,
            "badge": "string",
            "badgeType": "string",
            "badgeVariants": "string",
            "hideChildrenInMenu": false,
            "hideInBreadcrumb": false,
            "hideInMenu": false,
            "hideInTab": false,
            "icon": "lucide:area-chart",
            "iframeSrc": "string",
            "keepAlive": false,
            "link": "string",
            "maxNumOfOpenTab": -1,
            "noBasicLayout": false,
            "openInNewWindow": false,
            "order": 0,
            "query": "string",
            "title": "page.dashboard.analytics",
            "domCached": false,
            "menuVisibleWithForbidden": false
          },
          "name": "Analytics",
          "path": "/analytics",
          "pid": "205ce73c-baa0-4df9-b853-f6ae810d38ef",
          "type": "menu",
          "creatorId": null,
          "creatorName": null,
          "createDate": null,
          "updateDate": null,
          "status": 1
        }
      ],
      "component": null,
      "id": "205ce73c-baa0-4df9-b853-f6ae810d38ef",
      "meta": {
        "activeIcon": "string",
        "activePath": "string",
        "affixTab": true,
        "affixTabOrder": 0,
        "badge": "string",
        "badgeType": "string",
        "badgeVariants": "string",
        "hideChildrenInMenu": false,
        "hideInBreadcrumb": false,
        "hideInMenu": false,
        "hideInTab": false,
        "icon": "lucide:area-chart",
        "iframeSrc": "string",
        "keepAlive": false,
        "link": "string",
        "maxNumOfOpenTab": -1,
        "noBasicLayout": false,
        "openInNewWindow": false,
        "order": 0,
        "query": "string",
        "title": "page.dashboard.analytics",
        "domCached": false,
        "menuVisibleWithForbidden": false
      },
      "name": "Dashboard",
      "path": "/dashboard",
      "pid": null,
      "type": "catalog",
      "creatorId": null,
      "creatorName": null,
      "createDate": null,
      "updateDate": null,
      "status": 1
    }
  ],
  "error": null,
  "message": "ok"
}
```

## 检查菜单名称是否存在

请求路径：/api/system/menus/name-exists

请求方法：GET

要求：

1. 查询菜单表中名称是否已存在
2. id（排除自身）

请求参数示例(query)：

```tex
name="主页"&id=
```

响应数据示例：

```json
{
  "code": 0,
  "data": false, //false=不存在 true=已存在
  "error": null,
  "message": "ok"
}
```

## 检查菜单路由路径是否存在

请求路径：/api/system/menus/path-exists

请求方法：GET

要求：

1. 查询菜单表中路由路径是否已存在
2. id（排除自身）

请求参数示例(query)：

```tex
path="/system"&id=
```

响应数据示例：

```json
{
  "code": 0,
  "data": false, //false=不存在 true=已存在
  "error": null,
  "message": "ok"
}
```

## 创建菜单

请求路径：/api/system/menus

请求方法：POST

请求参数示例(body)：

```json
{
  "authCode": "sys:analytics",
  "component": "/dashboard/analytics/index",
  "meta": {
    "activeIcon": "string",
    "activePath": "string",
    "affixTab": true,
    "affixTabOrder": 0,
    "badge": "string",
    "badgeType": "string",
    "badgeVariants": "string",
    "hideChildrenInMenu": false,
    "hideInBreadcrumb": false,
    "hideInMenu": false,
    "hideInTab": false,
    "icon": "lucide:area-chart",
    "iframeSrc": "string",
    "keepAlive": false,
    "link": "string",
    "maxNumOfOpenTab": -1,
    "noBasicLayout": false,
    "openInNewWindow": false,
    "order": 0,
    "query": "string",
    "title": "page.dashboard.analytics",
    "domCached": false,
    "menuVisibleWithForbidden": false
  },
  "name": "Analytics",
  "path": "/analytics",
  "pid": "205ce73c-baa0-4df9-b853-f6ae810d38ef", //pid为空表示一级菜单
  "type": "menu",
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

## 菜单详情

请求路径：/api/system/menus/{id}

请求方法：GET

请求参数示例(path)：

```url
id
```

响应数据示例：

```json
{
  "code": 0,
  "data": {
    "authCode": "sys:analytics",
    "children": [],
    "component": "/dashboard/analytics/index",
    "id": "7c17031e-63fd-4f5d-8c00-17d85673457c",
    "meta": {
      "activeIcon": "string",
      "activePath": "string",
      "affixTab": true,
      "affixTabOrder": 0,
      "badge": "string",
      "badgeType": "string",
      "badgeVariants": "string",
      "hideChildrenInMenu": false,
      "hideInBreadcrumb": false,
      "hideInMenu": false,
      "hideInTab": false,
      "icon": "lucide:area-chart",
      "iframeSrc": "string",
      "keepAlive": false,
      "link": "string",
      "maxNumOfOpenTab": -1,
      "noBasicLayout": false,
      "openInNewWindow": false,
      "order": 0,
      "query": "string",
      "title": "page.dashboard.analytics",
      "domCached": false,
      "menuVisibleWithForbidden": false
    },
    "name": "Analytics",
    "path": "/analytics",
    "pid": "205ce73c-baa0-4df9-b853-f6ae810d38ef",
    "type": "menu",
    "creatorId": null,
    "creatorName": null,
    "createDate": null,
    "updateDate": null,
    "status": 1
  },
  "error": null,
  "message": "ok"
}
```

## 修改菜单

请求路径：/api/system/menus/{id}

请求方法：PUT

请求参数示例(path)：

```url
id
```

请求参数示例(body)：

```json
{
  "authCode": "sys:analytics",
  "component": "/dashboard/analytics/index",
  "meta": {
    "activeIcon": "string",
    "activePath": "string",
    "affixTab": true,
    "affixTabOrder": 0,
    "badge": "string",
    "badgeType": "string",
    "badgeVariants": "string",
    "hideChildrenInMenu": false,
    "hideInBreadcrumb": false,
    "hideInMenu": false,
    "hideInTab": false,
    "icon": "lucide:area-chart",
    "iframeSrc": "string",
    "keepAlive": false,
    "link": "string",
    "maxNumOfOpenTab": -1,
    "noBasicLayout": false,
    "openInNewWindow": false,
    "order": 0,
    "query": "string",
    "title": "page.dashboard.analytics",
    "domCached": false,
    "menuVisibleWithForbidden": false
  },
  "name": "Analytics",
  "path": "/analytics",
  "pid": "205ce73c-baa0-4df9-b853-f6ae810d38ef", //pid为空表示一级菜单
  "type": "menu",
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

## 删除菜单

请求路径：/api/system/menus

请求方法：DELETE

要求：

1. 禁止删除有子项的菜单
2. 禁止删除有角色关联的菜单
3. 逻辑删除 del_flag=1
4. 删除后要清理 sys_role_menu 关联表数据

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