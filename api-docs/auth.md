# 鉴权中心

## 登录接口

请求路径：/api/auth/login

请求方法：POST

要求：

1. 根据 status 值来判断。如果账号已禁用，报错提示 “该账号已被禁用”
1. password 不匹配时。报错“账号密码有误”
1. username 不匹配时。报错“账号密码有误”


请求参数示例：

```json
{"username":"vben","password":"123456"}
```

响应数据示例：

```json
{
  "code": 0,
  "data": {
    "accessToken": "xxxxxx" //登录的Token
  },
  "error": null,
  "message": "ok"
}
```

## 当前登录用户

请求路径：/api/auth/profile

请求方法：GET

要求：

1. 在根据当前请求头里的 authorization 所携带的 accessToken值，找到登录用户。

响应数据示例：

```json
{
  "code": 0,
  "data": {
    "userId": "fd8b5f2c-77c6-4e59-b81c-306c2fb85d44",
    "avatar": "https://picsum.photos/100/100",
    "username": "vben",
    "realName": "Vben",
    "roleTitles": ["super"], //登录用户所拥有的角色名称数组
    "roleIds": ["458e8285-cd9e-48ca-ac78-d2178a0e8c4f"], //登录用户所拥有的角色Id数组
    "desc": "超管",
    "email": "vben@example.com",
    "homePath": null,
    "deptTitles": ["技术部"],
    "deptIds": ["9de2fb68-7ba1-49cc-bfd4-e946a394f880"], //登录用户所有的部门Id列表
    "status": 1
  },
  "error": null,
  "message": "ok"
}
```

## 获取当前用户的菜单

请求路径：/api/auth/menus

请求方法：GET

要求：

1. 在根据当前请求头里的 authorization 所携带的 accessToken值，找到登录用户。
2. 然后找到关联此用户的角色(会关联多个角色)。然后再找到关联角色的所有菜单
3. 把路由相关的字段在 meta 里体现。
4. 按pid字段组为树结构。子级用 children表示 。再把组装好的树结构数据响应给前端
5. 关联角色表查询时，只查询 status = 1 的数据
6. 关联菜单表查询时，只查询 status = 1 AND type != "button" 的数据 
7. 当前登录用户是内置账户时 (用户表的is_sys字段) 响应所有的菜单。

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
            "activeIcon": null,
            "activePath": null,
            "affixTab": true,
            "affixTabOrder": 0,
            "badge": null,
            "badgeType": null,
            "badgeVariants": null,
            "hideChildrenInMenu": false,
            "hideInBreadcrumb": false,
            "hideInMenu": false,
            "hideInTab": false,
            "icon": "lucide:area-chart",
            "iframeSrc": null,
            "keepAlive": false,
            "link": null,
            "maxNumOfOpenTab": -1,
            "noBasicLayout": false,
            "openInNewWindow": false,
            "order": null,
            "query": null,
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
        "activeIcon": null,
        "activePath": null,
        "affixTab": false,
        "affixTabOrder": 0,
        "badge": null,
        "badgeType": null,
        "badgeVariants": null,
        "hideChildrenInMenu": false,
        "hideInBreadcrumb": false,
        "hideInMenu": false,
        "hideInTab": false,
        "icon": "carbon:workspace",
        "iframeSrc": null,
        "keepAlive": false,
        "link": null,
        "maxNumOfOpenTab": -1,
        "noBasicLayout": false,
        "openInNewWindow": false,
        "order": -1,
        "query": null,
        "title": "page.dashboard.title",
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

## 获取权限码列表

请求路径：/api/auth/codes
请求方法：GET

要求：

1. 在根据当前请求头里的 authorization 所携带的 accessToken值，找到登录用户。
1. 然后找到关联此用户的角色(会关联多个角色)。然后再找到关联角色的所有菜单。
1. 再把所有关联菜单的权限码 authCode Set去重后，按字母顺序 响应给前端
1. 关联角色表查询时，只查询 status = 1 的数据
1. 关联菜单表查询时，只查询 status = 1 的数据
1. 当前登录用户是内置账户时 (用户表的is_sys字段) 响应所有菜单的authCode。

响应数据示例：

```json
{
  "code": 0,
  "data": ["sys:workspace","sys:analytics"], //所有菜单数据的 authCode
  "error": null,
  "message": "ok"
}
```



## 退出登录

请求路径：/api/auth/logout

请求方法：POST

作用：

1. 在根据当前请求头里的 authorization 所携带的 accessToken值，找到登录用户。清掉登录状态
1. 退出登录后，标记token黑名单。旧token不能再被使用

响应数据示例：

```json
{
  "code": 0,
  "data": "",
  "error": null,
  "message": "ok"
}
```
