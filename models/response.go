package models

import (
	"time"
)

type Response struct {
	Code    int         `json:"code" example:"0"`     // 状态码 0=成功 -1=失败
	Data    interface{} `json:"data"`                 // 响应数据
	Error   interface{} `json:"error"`                // 错误信息
	Message string      `json:"message" example:"ok"` // 提示信息
}

type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"admin"`  // 登录用户名
	Password string `json:"password" binding:"required" example:"123456"` // 登录密码
}

type LoginResponse struct {
	AccessToken string `json:"accessToken" example:"eyJhbGciOiJIUzI1NiIs..."` // JWT访问令牌
}

type ProfileResponse struct {
	UserId         string   `json:"userId" example:"UUID"`                    // 用户ID
	Avatar         *string  `json:"avatar" example:"https://xxx/avatar.jpg"`  // 用户头像URL
	Username       string   `json:"username" example:"admin"`                 // 登录用户名
	RealName       string   `json:"realName" example:"管理员"`                   // 真实姓名
	Phone          *string  `json:"phone" example:"13800138000"`              // 手机号
	RoleTitles     []string `json:"roleTitles"`                               // 角色名称数组
	RoleIds        []string `json:"roleIds"`                                  // 角色id数组
	Desc           *string  `json:"desc" example:"超级管理员"`                     // 用户描述
	Email          *string  `json:"email" example:"admin@example.com"`        // 邮箱
	HomePath       *string  `json:"homePath" example:"/dashboard/analytics"`  // 首页路径
	LeaderUserId   *string  `json:"leaderUserId" example:"UUID"`              // 直属上级用户ID
	LeaderUserName *string  `json:"leaderUserName" example:"张三"`              // 直属上级用户姓名
	DeptTitles     []string `json:"deptTitles"`                               // 部门名称数组
	DeptIds        []string `json:"deptIds"`                                  // 部门id数组
	Status         int      `json:"status" example:"1"`                       // 用户状态 0=禁用 1=启用
	CreateDate     *string  `json:"createDate" example:"2024-01-01 12:00:00"` // 创建时间
	UpdateDate     *string  `json:"updateDate" example:"2024-01-01 12:00:00"` // 更新时间
}

type MenuMeta struct {
	ActiveIcon               *string `json:"activeIcon" example:"lucide:home"`         // 激活图标
	ActivePath               *string `json:"activePath" example:"/dashboard"`          // 激活路径
	AffixTab                 bool    `json:"affixTab" example:"false"`                 // 固定在标签页 0=否 1=是
	AffixTabOrder            int     `json:"affixTabOrder" example:"0"`                // 固定标签页的排序
	Badge                    *string `json:"badge" example:"new"`                      // 徽章内容
	BadgeType                *string `json:"badgeType" example:"dot"`                  // 徽标类型
	BadgeVariants            *string `json:"badgeVariants" example:"destructive"`      // 徽标样式
	HideChildrenInMenu       bool    `json:"hideChildrenInMenu" example:"false"`       // 隐藏子菜单 0=否 1=是
	HideInBreadcrumb         bool    `json:"hideInBreadcrumb" example:"false"`         // 在面包屑中隐藏 0=否 1=是
	HideInMenu               bool    `json:"hideInMenu" example:"false"`               // 在菜单中隐藏 0=否 1=是
	HideInTab                bool    `json:"hideInTab" example:"false"`                // 在标签栏中隐藏 0=否 1=是
	Icon                     *string `json:"icon" example:"lucide:home"`               // 图标
	IframeSrc                *string `json:"iframeSrc" example:"https://example.com"`  // 内嵌页面的iframe地址
	KeepAlive                bool    `json:"keepAlive" example:"false"`                // 缓存标签页 0=否 1=是
	Link                     *string `json:"link" example:"https://example.com"`       // 外链跳转路径
	MaxNumOfOpenTab          int     `json:"maxNumOfOpenTab" example:"-1"`             // 标签页最大打开数量
	NoBasicLayout            bool    `json:"noBasicLayout" example:"false"`            // 无基础布局 0=否 1=是
	OpenInNewWindow          bool    `json:"openInNewWindow" example:"false"`          // 在新窗口打开 0=否 1=是
	Order                    *int    `json:"order" example:"0"`                        // 排序
	Query                    *string `json:"query" example:"id=1"`                     // 额外的路由参数
	Title                    string  `json:"title" example:"page.dashboard.analytics"` // 页面标题
	DomCached                bool    `json:"domCached" example:"false"`                // 缓存DOM 0=否 1=是
	MenuVisibleWithForbidden bool    `json:"menuVisibleWithForbidden" example:"false"` // 菜单可见但访问受限 0=否 1=是
}

type MenuTreeResponse struct {
	ID          string              `json:"id" example:"UUID"`                              // 菜单ID
	Pid         *string             `json:"pid" example:"UUID"`                             // 上级菜单ID
	Type        string              `json:"type" example:"menu"`                            // 菜单类型 catalog=目录 menu=菜单
	AuthCode    *string             `json:"authCode" example:"sys:analytics"`               // 权限标识
	Children    []*MenuTreeResponse `json:"children"`                                       // 子菜单
	Component   *string             `json:"component" example:"/dashboard/analytics/index"` // 页面组件路径
	Meta        MenuMeta            `json:"meta"`                                           // 菜单元数据
	Name        string              `json:"name" example:"Analytics"`                       // 菜单名称
	Path        *string             `json:"path" example:"/analytics"`                      // 路由地址
	CreatorId   *string             `json:"creatorId" example:"UUID"`                       // 创建人id
	CreatorName *string             `json:"creatorName" example:"管理员"`                      // 创建人姓名
	CreateDate  *string             `json:"createDate" example:"2024-01-01 12:00:00"`       // 创建时间
	UpdateDate  *string             `json:"updateDate" example:"2024-01-01 12:00:00"`       // 修改时间
	Status      int                 `json:"status" example:"1"`                             // 状态 0=禁用 1=启用
}

type UserListRequest struct {
	Page     int    `form:"page" example:"1"`                // 页码
	PageSize int    `form:"pageSize" example:"20"`           // 每页大小
	Username string `form:"username" example:"admin"`        // 用户名，模糊搜索
	RealName string `form:"realName" example:"管理员"`          // 真实姓名，模糊搜索
	Phone    string `form:"phone" example:"13800138000"`     // 手机号，模糊搜索
	Status   *int   `form:"status" example:"1"`              // 状态 0=禁用 1=启用
	Sorts    string `form:"sorts" example:"createDate,desc"` // 排序参数
	DeptId   string `form:"deptId" example:"UUID"`           // 部门ID，查询该部门及子部门的用户
}

type FileListRequest struct {
	Page         int    `form:"page" example:"1"`                         // 页码
	PageSize     int    `form:"pageSize" example:"20"`                    // 每页大小
	OriginalName string `form:"originalName" example:"文件名"`               // 原始文件名，模糊搜索
	Type         string `form:"type" example:"image/jpeg"`                // MIME类型，精确匹配
	FileExt      string `form:"fileExt" example:".jpg"`                   // 文件扩展名，精确匹配
	Sorts        string `form:"sorts" example:"createDate,desc;size,asc"` // 排序参数
}

type CreateUserRequest struct {
	Username     string   `json:"username" binding:"required" example:"newuser"` // 登录用户名
	RealName     string   `json:"realName" binding:"required" example:"新用户"`     // 真实姓名
	Password     string   `json:"password" binding:"required" example:"123456"`  // 密码
	Phone        *string  `json:"phone" example:"13800138000"`                   // 手机号
	Desc         *string  `json:"desc" example:"普通用户"`                           // 描述
	DeptIds      []string `json:"deptIds" example:"[\"UUID\"]"`                  // 部门id数组
	RoleIds      []string `json:"roleIds" example:"[\"UUID\"]"`                  // 角色id数组
	LeaderUserId *string  `json:"leaderUserId" example:"UUID"`                   // 直属上级用户ID
}

type UpdateUserRequest struct {
	Username     string   `json:"username" binding:"required" example:"newuser"` // 登录用户名
	RealName     string   `json:"realName" binding:"required" example:"新用户"`     // 真实姓名
	Phone        *string  `json:"phone" example:"13800138000"`                   // 手机号
	Desc         *string  `json:"desc" example:"普通用户"`                           // 描述
	DeptIds      []string `json:"deptIds" example:"[\"UUID\"]"`                  // 部门id数组
	RoleIds      []string `json:"roleIds" example:"[\"UUID\"]"`                  // 角色id数组
	LeaderUserId *string  `json:"leaderUserId" example:"UUID"`                   // 直属上级用户ID
}

type UpdateUserStatusRequest struct {
	Status int `json:"status" example:"1"` // 状态 0=禁用 1=启用
}

type MenuListRequest struct {
	Name   string `form:"name" example:"Analytics"`  // 菜单名称，模糊搜索
	Path   string `form:"path" example:"/analytics"` // 路由路径，模糊搜索
	Type   string `form:"type" example:""`           // 菜单类型
	Status *int   `form:"status" example:"1"`        // 状态 0=禁用 1=启用
}

type CreateMenuRequest struct {
	Pid       *string  `json:"pid" example:"UUID"`                             // 上级菜单ID，空表示顶级菜单
	Type      string   `json:"type" binding:"required" example:"menu"`         // 菜单类型 catalog=目录 menu=菜单
	AuthCode  *string  `json:"authCode" example:"sys:analytics"`               // 权限标识
	Component *string  `json:"component" example:"/dashboard/analytics/index"` // 页面组件路径
	Name      string   `json:"name" binding:"required" example:"Analytics"`    // 菜单名称
	Path      *string  `json:"path" example:"/analytics"`                      // 路由地址
	Meta      MenuMeta `json:"meta" binding:"required"`                        // 菜单元数据
	Status    int      `json:"status" example:"1"`                             // 状态 0=禁用 1=启用
}

type UpdateMenuRequest struct {
	Pid       *string  `json:"pid" example:"UUID"`                             // 上级菜单ID，空表示顶级菜单
	Type      string   `json:"type" binding:"required" example:"menu"`         // 菜单类型 catalog=目录 menu=菜单
	AuthCode  *string  `json:"authCode" example:"sys:analytics"`               // 权限标识
	Component *string  `json:"component" example:"/dashboard/analytics/index"` // 页面组件路径
	Name      string   `json:"name" binding:"required" example:"Analytics"`    // 菜单名称
	Path      *string  `json:"path" example:"/analytics"`                      // 路由地址
	Meta      MenuMeta `json:"meta" binding:"required"`                        // 菜单元数据
	Status    int      `json:"status" example:"1"`                             // 状态 0=禁用 1=启用
}

type RoleListRequest struct {
	Page      int    `form:"page" example:"1"`                // 页码
	PageSize  int    `form:"pageSize" example:"20"`           // 每页大小
	RoleTitle string `form:"roleTitle" example:"管理员"`         // 角色名称，模糊搜索
	Status    *int   `form:"status" example:"1"`              // 状态 0=禁用 1=启用
	Remark    string `form:"remark" example:"备注"`             // 备注，模糊搜索
	StartDate string `form:"startDate" example:"2024-01-01"`  // 创建日期起始
	EndDate   string `form:"endDate" example:"2024-12-31"`    // 创建日期截止
	Sorts     string `form:"sorts" example:"createDate,desc"` // 排序参数
}

type CreateRoleRequest struct {
	RoleTitle   string   `json:"roleTitle" binding:"required" example:"编辑员"` // 角色名称
	Status      int      `json:"status" example:"1"`                         // 状态 0=禁用 1=启用
	Remark      *string  `json:"remark" example:"编辑角色"`                      // 备注
	Permissions []string `json:"permissions" example:"[\"UUID\"]"`           // 菜单id数组
}

type UpdateRoleRequest struct {
	RoleTitle   string   `json:"roleTitle" binding:"required" example:"编辑员"` // 角色名称
	Status      int      `json:"status" example:"1"`                         // 状态 0=禁用 1=启用
	Remark      *string  `json:"remark" example:"编辑角色"`                      // 备注
	Permissions []string `json:"permissions" example:"[\"UUID\"]"`           // 菜单id数组
}

type RoleDetailResponse struct {
	RoleId      string   `json:"roleId" example:"UUID"`                    // 角色ID
	RoleTitle   string   `json:"roleTitle" example:"SuperAdmin"`           // 角色名称
	Status      int      `json:"status" example:"1"`                       // 状态 0=禁用 1=启用
	CreateDate  *string  `json:"createDate" example:"2024-01-01 12:00:00"` // 创建时间
	Remark      *string  `json:"remark" example:"超级管理员"`                   // 备注
	Permissions []string `json:"permissions" example:"[\"UUID\"]"`         // 菜单id数组
}

type RoleSimpleResponse struct {
	RoleId    string `json:"roleId" example:"UUID"`          // 角色ID
	RoleTitle string `json:"roleTitle" example:"SuperAdmin"` // 角色名称
	Status    int    `json:"status" example:"1"`             // 状态 0=禁用 1=启用
}

type UpdateStatusRequest struct {
	Status int `json:"status" example:"1"` // 状态 0=禁用 1=启用
}

type DeptListRequest struct {
	DeptTitle string `form:"deptTitle" example:"技术部"` // 部门名称，模糊搜索
}

type DeptTreeResponse struct {
	DeptId     string              `json:"deptId" example:"UUID"`                    // 部门ID
	Pid        *string             `json:"pid" example:"UUID"`                       // 父级部门ID
	DeptTitle  string              `json:"deptTitle" example:"技术部"`                  // 部门名称
	Status     int                 `json:"status" example:"1"`                       // 状态 0=禁用 1=启用
	CreateDate *string             `json:"createDate" example:"2024-01-01 12:00:00"` // 创建时间
	Remark     *string             `json:"remark" example:"技术部门"`                    // 备注
	Children   []*DeptTreeResponse `json:"children"`                                 // 子部门
}

type CreateDeptRequest struct {
	Pid       *string `json:"pid" example:"UUID"`                         // 父级部门ID，空表示顶级部门
	DeptTitle string  `json:"deptTitle" binding:"required" example:"技术部"` // 部门名称
	Status    int     `json:"status" example:"1"`                         // 状态 0=禁用 1=启用
	Remark    *string `json:"remark" example:"技术部门"`                      // 备注
}

type UpdateDeptRequest struct {
	Pid       *string `json:"pid" example:"UUID"`                         // 父级部门ID
	DeptTitle string  `json:"deptTitle" binding:"required" example:"技术部"` // 部门名称
	Status    int     `json:"status" example:"1"`                         // 状态 0=禁用 1=启用
	Remark    *string `json:"remark" example:"技术部门"`                      // 备注
}

type DictListRequest struct {
	Label string `form:"label" example:"需求类型"`            // 字典标题，模糊搜索
	Value string `form:"value" example:"0"`               // 字典值
	Type  string `form:"type" example:"STORY_TYPE"`       // 字典类型
	Sorts string `form:"sorts" example:"createDate,desc"` // 排序参数
}

type DictTreeResponse struct {
	ID         string              `json:"id" example:"UUID"`                        // 字典ID
	Pid        *string             `json:"pid" example:"UUID"`                       // 父级字典ID
	Label      string              `json:"label" example:"功能需求"`                     // 字典标题
	Value      *string             `json:"value" example:"0"`                        // 字典值
	Type       string              `json:"type" example:"STORY_TYPE"`                // 字典类型
	Remark     *string             `json:"remark" example:"需求类型"`                    // 备注
	Color      *string             `json:"color" example:"#2db7f5"`                  // 主题色
	Status     int                 `json:"status" example:"1"`                       // 状态 0=禁用 1=启用
	CreateDate *string             `json:"createDate" example:"2024-01-01 12:00:00"` // 创建时间
	UpdateDate *string             `json:"updateDate" example:"2024-01-01 12:00:00"` // 更新时间
	Children   []*DictTreeResponse `json:"children"`                                 // 子字典
}

type CreateDictRequest struct {
	Pid    *string `json:"pid" example:"UUID"`                           // 父级字典ID
	Type   string  `json:"type" binding:"required" example:"STORY_TYPE"` // 字典类型
	Label  string  `json:"label" binding:"required" example:"功能需求"`      // 字典标题
	Value  *string `json:"value" example:"0"`                            // 字典值
	Color  *string `json:"color" example:"#2db7f5"`                      // 主题色
	Status int     `json:"status" example:"1"`                           // 状态 0=禁用 1=启用
	Remark *string `json:"remark" example:"需求类型"`                        // 备注
}

type UpdateDictRequest struct {
	Pid    *string `json:"pid" example:"UUID"`                           // 父级字典ID
	Type   string  `json:"type" binding:"required" example:"STORY_TYPE"` // 字典类型
	Label  string  `json:"label" binding:"required" example:"功能需求"`      // 字典标题
	Value  *string `json:"value" example:"0"`                            // 字典值
	Color  *string `json:"color" example:"#2db7f5"`                      // 主题色
	Status int     `json:"status" example:"1"`                           // 状态 0=禁用 1=启用
	Remark *string `json:"remark" example:"需求类型"`                        // 备注
}

func NewSuccessResponse(data interface{}) Response {
	return Response{
		Code:    0,
		Data:    data,
		Error:   nil,
		Message: "ok",
	}
}

func NewErrorResponse(err interface{}, message string) Response {
	return Response{
		Code:    -1,
		Data:    nil,
		Error:   err,
		Message: message,
	}
}

func SysUserToProfileResponse(user SysUser, roleTitles, roleIds, deptTitles, deptIds []string) *ProfileResponse {
	username := ""
	if user.Username != nil {
		username = *user.Username
	}
	realName := ""
	if user.RealName != nil {
		realName = *user.RealName
	}

	return &ProfileResponse{
		UserId:       user.UserID,
		Avatar:       user.Avatar,
		Username:     username,
		RealName:     realName,
		Phone:        user.Phone,
		RoleTitles:   roleTitles,
		RoleIds:      roleIds,
		Desc:         user.Desc,
		Email:        user.Email,
		HomePath:     user.HomePath,
		LeaderUserId: user.LeaderUserID,
		DeptTitles:   deptTitles,
		DeptIds:      deptIds,
		Status:       user.Status,
		CreateDate:   TimeToStringPtr(user.CreateDate),
		UpdateDate:   TimeToStringPtr(user.UpdateDate),
	}
}

func TimeToStringPtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("2006-01-02 15:04:05")
	return &s
}

type ProjectResponse struct {
	ProjectID    *string `json:"projectId" example:"UUID"`                   // 项目ID
	ProjectTitle *string `json:"projectTitle" example:"crudelis"`            // 项目标题
	ProjectLogo  *string `json:"projectLogo" example:"https://xxx/logo.png"` // 项目Logo
	Description  *string `json:"description" example:"项目描述"`                 // 项目描述
	CreateDate   *string `json:"createDate" example:"2024-01-01 12:00:00"`   // 创建时间
}

type ModuleResponse struct {
	ModuleID     *string `json:"moduleId" example:"UUID"`                  // 模块ID
	ModuleTitle  *string `json:"moduleTitle" example:"模块名称"`               // 模块标题
	ProjectID    *string `json:"projectId" example:"UUID"`                 // 关联项目ID
	ProjectTitle *string `json:"projectTitle" example:"crudelis"`          // 关联项目标题
	Sort         int     `json:"sort" example:"1"`                         // 排序
	UpdateDate   *string `json:"updateDate" example:"2024-01-01 12:00:00"` // 更新时间
	CreateDate   *string `json:"createDate" example:"2024-01-01 12:00:00"` // 创建时间
}

type VersionResponse struct {
	VersionID         *string `json:"versionId" example:"UUID"`                 // 版本ID
	Version           *string `json:"version" example:"v1.0.0"`                 // 版本号
	VersionType       string  `json:"versionType" example:"0"`                  // 版本类型
	Remark            *string `json:"remark" example:"版本备注"`                    // 备注
	CreatorID         *string `json:"creatorId" example:"UUID"`                 // 创建人ID
	CreatorName       *string `json:"creatorName" example:"管理员"`                // 创建人姓名
	CreateDate        *string `json:"createDate" example:"2024-01-01 12:00:00"` // 创建时间
	EndDate           *string `json:"endDate" example:"2024-12-31"`             // 结束日期
	StartDate         *string `json:"startDate" example:"2024-01-01"`           // 开始日期
	ProjectID         *string `json:"projectId" example:"UUID"`                 // 关联项目ID
	ProjectTitle      *string `json:"projectTitle" example:"crudelis"`          // 关联项目标题
	ReleaseStatus     string  `json:"releaseStatus" example:"0"`                // 发布状态
	ReleaseDate       *string `json:"releaseDate" example:"2024-06-30"`         // 发布日期
	ChangeLogRichText *string `json:"changeLogRichText" example:"<p>更新日志</p>"`  // 更新日志(富文本)
	ChangeLog         *string `json:"changeLog" example:"更新日志文本"`               // 更新日志
}

type StoryUserItem struct {
	UserID   *string `json:"userId" example:"UUID"`                   // 用户ID
	Avatar   *string `json:"avatar" example:"https://xxx/avatar.jpg"` // 用户头像
	RealName *string `json:"realName" example:"张三"`                   // 真实姓名
}

type StoryResponse struct {
	StoryID       *string         `json:"storyId" example:"UUID"`                   // 需求ID
	StoryTitle    *string         `json:"storyTitle" example:"需求标题"`                // 需求标题
	StoryNum      int             `json:"storyNum" example:"1"`                     // 需求编号
	CreatorName   *string         `json:"creatorName" example:"管理员"`                // 创建人姓名
	CreatorID     *string         `json:"creatorId" example:"UUID"`                 // 创建人ID
	StoryType     string          `json:"storyType" example:"0"`                    // 需求类型
	StoryStatus   string          `json:"storyStatus" example:"0"`                  // 需求状态
	StoryLevel    string          `json:"storyLevel" example:"0"`                   // 需求优先级
	VersionID     *string         `json:"versionId" example:"UUID"`                 // 关联版本ID
	Version       *string         `json:"version" example:"v1.0.0"`                 // 关联版本号
	ProjectID     *string         `json:"projectId" example:"UUID"`                 // 关联项目ID
	ProjectTitle  *string         `json:"projectTitle" example:"crudelis"`          // 关联项目标题
	ModuleID      *string         `json:"moduleId" example:"UUID"`                  // 关联模块ID
	ModuleTitle   *string         `json:"moduleTitle" example:"模块名称"`               // 关联模块标题
	Source        string          `json:"source" example:"0"`                       // 需求来源
	UpdateDate    *string         `json:"updateDate" example:"2024-01-01 12:00:00"` // 更新时间
	CreateDate    *string         `json:"createDate" example:"2024-01-01 12:00:00"` // 创建时间
	UserList      []StoryUserItem `json:"userList"`                                 // 参与人员列表
	UserIDs       []string        `json:"userIds"`                                  // 参与人员ID数组
	StoryRichText *string         `json:"storyRichText" example:"<p>需求描述</p>"`      // 需求描述(富文本)
	FileIDs       []string        `json:"fileIds"`                                  // 附件ID数组
	FileList      []FileResponse  `json:"fileList"`                                 // 附件列表
	TaskList      []TaskResponse  `json:"taskList"`                                 // 关联任务列表
	BugList       []BugResponse   `json:"bugList"`                                  // 关联缺陷列表
}

type FileResponse struct {
	FileID        *string `json:"fileId" example:"UUID"`                         // 文件ID
	URL           *string `json:"url" example:"/uploads/abc.jpg"`                // 文件访问URL
	Name          *string `json:"name" example:"abc.jpg"`                        // 存储文件名(UUID重命名)
	Type          *string `json:"type" example:"image/jpeg"`                     // MIME类型
	Size          int64   `json:"size" example:"102400"`                         // 文件大小(字节)
	FileExt       *string `json:"fileExt" example:".jpg"`                        // 文件扩展名
	OriginalName  *string `json:"originalName" example:"原始文件名.jpg"`              // 原始文件名
	Path          *string `json:"path" example:"/uploads/"`                      // 文件存储路径(不含文件名)
	FullPath      *string `json:"fullPath" example:"/uploads/abc.jpg"`           // 完整路径
	ThumbnailPath *string `json:"thumbnailPath" example:"/uploads/thumb"`        // 缩略图路径(图片专用)
	ThumbnailURL  *string `json:"thumbnailUrl" example:"/uploads/thumb/abc.jpg"` // 缩略图URL(图片专用)
	CreatorID     *string `json:"creatorId" example:"UUID"`                      // 创建人id
	CreatorName   *string `json:"creatorName" example:"创建人姓名"`                   // 创建人姓名
	CreateDate    *string `json:"createDate" example:"2024-01-01 12:00:00"`      // 创建日期
}

type TaskResponse struct {
	TaskID       *string `json:"taskId" example:"UUID"`                    // 任务ID
	StoryID      *string `json:"storyId" example:"UUID"`                   // 关联需求ID
	StoryTitle   *string `json:"storyTitle" example:"需求标题"`                // 关联需求标题
	ModuleID     *string `json:"moduleId" example:"UUID"`                  // 关联模块ID
	ModuleTitle  *string `json:"moduleTitle" example:"模块名称"`               // 关联模块标题
	VersionID    *string `json:"versionId" example:"UUID"`                 // 关联版本ID
	Version      *string `json:"version" example:"v1.0.0"`                 // 关联版本号
	ProjectID    *string `json:"projectId" example:"UUID"`                 // 关联项目ID
	ProjectTitle *string `json:"projectTitle" example:"crudelis"`          // 关联项目标题
	TaskTitle    *string `json:"taskTitle" example:"任务标题"`                 // 任务标题
	TaskNum      int     `json:"taskNum" example:"1"`                      // 任务编号
	TaskStatus   string  `json:"taskStatus" example:"0"`                   // 任务状态
	TaskType     string  `json:"taskType" example:"0"`                     // 任务类型
	PlanHours    float64 `json:"planHours" example:"8"`                    // 计划工时
	ActualHours  float64 `json:"actualHours" example:"6"`                  // 实际工时
	EndDate      *string `json:"endDate" example:"2024-12-31"`             // 结束日期
	StartDate    *string `json:"startDate" example:"2024-01-01"`           // 开始日期
	CreateDate   *string `json:"createDate" example:"2024-01-01 12:00:00"` // 创建时间
	CreatorID    *string `json:"creatorId" example:"UUID"`                 // 创建人ID
	CreatorName  *string `json:"creatorName" example:"管理员"`                // 创建人姓名
	UserID       *string `json:"userId" example:"UUID"`                    // 负责人ID
	RealName     *string `json:"realName" example:"张三"`                    // 负责人姓名
	Avatar       *string `json:"avatar" example:"https://xxx/avatar.jpg"`  // 负责人头像
	Percent      int     `json:"percent" example:"50"`                     // 任务进度百分比
	TaskRichText *string `json:"taskRichText" example:"<p>任务描述</p>"`       // 任务描述(富文本)
}

type BugResponse struct {
	BugID            *string `json:"bugId" example:"UUID"`                     // 缺陷ID
	BugTitle         *string `json:"bugTitle" example:"缺陷标题"`                  // 缺陷标题
	BugNum           int     `json:"bugNum" example:"1"`                       // 缺陷编号
	BugStatus        string  `json:"bugStatus" example:"0"`                    // 缺陷状态
	BugConfirmStatus string  `json:"bugConfirmStatus" example:"0"`             // 缺陷确认状态
	BugLevel         string  `json:"bugLevel" example:"0"`                     // 缺陷等级
	BugSource        string  `json:"bugSource" example:"0"`                    // 缺陷来源
	BugType          string  `json:"bugType" example:"0"`                      // 缺陷类型
	BugEnv           string  `json:"bugEnv" example:"0"`                       // 缺陷环境
	BugUa            *string `json:"bugUa" example:"Mozilla/5.0"`              // 用户代理
	UserID           *string `json:"userId" example:"UUID"`                    // 指派人ID
	Avatar           *string `json:"avatar" example:"https://xxx/avatar.jpg"`  // 指派人头像
	RealName         *string `json:"realName" example:"张三"`                    // 指派人姓名
	CreatorName      *string `json:"creatorName" example:"管理员"`                // 创建人姓名
	CreatorID        *string `json:"creatorId" example:"UUID"`                 // 创建人ID
	VersionID        *string `json:"versionId" example:"UUID"`                 // 关联版本ID
	Version          *string `json:"version" example:"v1.0.0"`                 // 关联版本号
	ModuleID         *string `json:"moduleId" example:"UUID"`                  // 关联模块ID
	ModuleTitle      *string `json:"moduleTitle" example:"模块名称"`               // 关联模块标题
	ProjectID        *string `json:"projectId" example:"UUID"`                 // 关联项目ID
	ProjectTitle     *string `json:"projectTitle" example:"crudelis"`          // 关联项目标题
	StoryID          *string `json:"storyId" example:"UUID"`                   // 关联需求ID
	StoryTitle       *string `json:"storyTitle" example:"需求标题"`                // 关联需求标题
	UpdateDate       *string `json:"updateDate" example:"2024-01-01 12:00:00"` // 更新时间
	CreateDate       *string `json:"createDate" example:"2024-01-01 12:00:00"` // 创建时间
	BugRichText      *string `json:"bugRichText" example:"<p>缺陷描述</p>"`        // 缺陷描述(富文本)
}

type ChangeHistoryResponse struct {
	ChangeID       *string `json:"changeId" example:"UUID"`                  // 变更记录ID
	ChangeBehavior string  `json:"changeBehavior" example:"0"`               // 变更行为
	ChangeRichText *string `json:"changeRichText" example:"<p>变更详情</p>"`     // 变更详情(富文本)
	CreatorID      *string `json:"creatorId" example:"UUID"`                 // 创建人ID
	CreatorName    *string `json:"creatorName" example:"管理员"`                // 创建人姓名
	BusinessID     *string `json:"businessId" example:"UUID"`                // 业务ID
	BusinessType   string  `json:"businessType" example:"0"`                 // 业务类型
	ExtendJson     *string `json:"extendJson" example:"{\"key\":\"value\"}"` // 扩展JSON
	CreateDate     *string `json:"createDate" example:"2024-01-01 12:00:00"` // 创建时间
	UpdateDate     *string `json:"updateDate" example:"2024-01-01 12:00:00"` // 更新时间
}

type CreateChangeHistoryRequest struct {
	BusinessID     string `json:"businessId" binding:"required" example:"UUID"`  // 业务ID
	BusinessType   string `json:"businessType" binding:"required" example:"0"`   // 业务类型
	ChangeBehavior string `json:"changeBehavior" binding:"required" example:"0"` // 变更行为
	ChangeRichText string `json:"changeRichText" example:"<p>变更详情</p>"`          // 变更详情(富文本)
}

type CreateProjectRequest struct {
	ProjectTitle *string `json:"projectTitle" binding:"required" example:"crudelis"` // 项目标题
	Description  *string `json:"description" example:"项目描述"`                         // 项目描述
	ProjectLogo  *string `json:"projectLogo" example:"https://xxx/logo.png"`         // 项目Logo
}

type UpdateProjectRequest struct {
	ProjectTitle *string `json:"projectTitle" binding:"required" example:"crudelis"` // 项目标题
	Description  *string `json:"description" example:"项目描述"`                         // 项目描述
	ProjectLogo  *string `json:"projectLogo" example:"https://xxx/logo.png"`         // 项目Logo
}

type CreateModuleRequest struct {
	Sort        int     `json:"sort" example:"1"`                              // 排序
	ModuleTitle *string `json:"moduleTitle" binding:"required" example:"模块名称"` // 模块标题
	ProjectID   string  `json:"projectId" binding:"required" example:"UUID"`   // 关联项目ID
}

type UpdateModuleRequest struct {
	Sort        int     `json:"sort" example:"1"`                              // 排序
	ModuleTitle *string `json:"moduleTitle" binding:"required" example:"模块名称"` // 模块标题
}

type CreateVersionRequest struct {
	Version       *string `json:"version" example:"v1.0.0"`       // 版本号
	VersionType   string  `json:"versionType" example:"0"`        // 版本类型
	ReleaseStatus string  `json:"releaseStatus" example:"0"`      // 发布状态
	ProjectID     string  `json:"projectId" example:"UUID"`       // 关联项目ID
	Remark        *string `json:"remark" example:"版本备注"`          // 备注
	EndDate       *string `json:"endDate" example:"2024-12-31"`   // 结束日期
	StartDate     *string `json:"startDate" example:"2024-01-01"` // 开始日期
}

type UpdateVersionRequest struct {
	Version       *string `json:"version" example:"v1.0.0"`       // 版本号
	VersionType   string  `json:"versionType" example:"0"`        // 版本类型
	ReleaseStatus string  `json:"releaseStatus" example:"0"`      // 发布状态
	ProjectID     string  `json:"projectId" example:"UUID"`       // 关联项目ID
	Remark        *string `json:"remark" example:"版本备注"`          // 备注
	EndDate       *string `json:"endDate" example:"2024-12-31"`   // 结束日期
	StartDate     *string `json:"startDate" example:"2024-01-01"` // 开始日期
}

type UpdateVersionNextRequest struct {
	ReleaseStatus  string `json:"releaseStatus" example:"0"`            // 发布状态
	ChangeRichText string `json:"changeRichText" example:"<p>变更详情</p>"` // 变更详情(富文本)
}

type CreateStoryRequest struct {
	StoryStatus   string   `json:"storyStatus" example:"0"`                      // 需求状态,字典STORY_STATUS值
	StoryType     string   `json:"storyType" binding:"required" example:"0"`     // 需求类型,字典STORY_TYPE值
	StoryLevel    string   `json:"storyLevel" example:"0"`                       // 需求优先级,字典STORY_LEVEL值
	Source        string   `json:"source" example:"0"`                           // 需求来源,字典STORY_SOURCE值
	StoryTitle    *string  `json:"storyTitle" binding:"required" example:"需求标题"` // 需求标题
	StoryRichText *string  `json:"storyRichText" example:"需求描述"`                 // 需求描述,富文本格式
	UserIDs       []string `json:"userIds" example:"[\"UUID\"]"`                 // 参与人员id数组
	ProjectID     string   `json:"projectId" binding:"required" example:"UUID"`  // 关联项目id
	VersionID     *string  `json:"versionId" binding:"required" example:"UUID"`  // 关联版本id
	ModuleID      *string  `json:"moduleId" binding:"required" example:"UUID"`   // 关联模块id
	FileIDs       []string `json:"fileIds" example:"[\"UUID\"]"`                 // 附件id数组
}

type UpdateStoryRequest struct {
	StoryStatus   string   `json:"storyStatus" example:"0"`                      // 需求状态
	StoryType     string   `json:"storyType" binding:"required" example:"0"`     // 需求类型
	StoryLevel    string   `json:"storyLevel" example:"0"`                       // 需求优先级
	Source        string   `json:"source" example:"0"`                           // 需求来源
	StoryTitle    *string  `json:"storyTitle" binding:"required" example:"需求标题"` // 需求标题
	StoryRichText *string  `json:"storyRichText" example:"<p>需求描述</p>"`          // 需求描述(富文本)
	UserIDs       []string `json:"userIds"`                                      // 参与人员ID数组
	ProjectID     string   `json:"projectId" binding:"required" example:"UUID"`  // 关联项目ID
	VersionID     *string  `json:"versionId" binding:"required" example:"UUID"`  // 关联版本ID
	ModuleID      *string  `json:"moduleId" binding:"required" example:"UUID"`   // 关联模块ID
	FileIDs       []string `json:"fileIds"`                                      // 附件ID数组
}

type UpdateStoryFieldRequest struct {
	Key   string      `json:"key" binding:"required" example:"storyStatus"` // 要更新的字段名
	Value interface{} `json:"value" example:"0"`                            // 要更新的字段值
}

type UpdateStoryNextRequest struct {
	StoryStatus    string `json:"storyStatus" binding:"required" example:"0"` // 需求状态
	ChangeRichText string `json:"changeRichText" example:"<p>流转说明</p>"`       // 流转说明(富文本)
}

type CreateTaskRequest struct {
	PlanHours    float64 `json:"planHours" binding:"required" example:"8"`          // 计划工时
	TaskStatus   string  `json:"taskStatus" example:"0"`                            // 任务状态
	TaskType     string  `json:"taskType" example:"0"`                              // 任务类型
	TaskTitle    *string `json:"taskTitle" binding:"required" example:"任务标题"`       // 任务标题
	ProjectID    string  `json:"projectId" binding:"required" example:"UUID"`       // 关联项目ID
	TaskRichText *string `json:"taskRichText" example:"<p>任务描述</p>"`                // 任务描述(富文本)
	VersionID    *string `json:"versionId" example:"UUID"`                          // 关联版本ID
	ModuleID     *string `json:"moduleId" example:"UUID"`                           // 关联模块ID
	StoryID      *string `json:"storyId" example:"UUID"`                            // 关联需求ID
	UserID       *string `json:"userId" binding:"required" example:"UUID"`          // 负责人ID
	EndDate      *string `json:"endDate" binding:"required" example:"2024-12-31"`   // 结束日期
	StartDate    *string `json:"startDate" binding:"required" example:"2024-01-01"` // 开始日期
}

type UpdateTaskRequest struct {
	PlanHours    float64 `json:"planHours" binding:"required" example:"8"`          // 计划工时
	TaskStatus   string  `json:"taskStatus" example:"0"`                            // 任务状态
	TaskType     string  `json:"taskType" example:"0"`                              // 任务类型
	TaskTitle    *string `json:"taskTitle" binding:"required" example:"任务标题"`       // 任务标题
	ProjectID    string  `json:"projectId" binding:"required" example:"UUID"`       // 关联项目ID
	TaskRichText *string `json:"taskRichText" example:"<p>任务描述</p>"`                // 任务描述(富文本)
	VersionID    *string `json:"versionId" example:"UUID"`                          // 关联版本ID
	ModuleID     *string `json:"moduleId" example:"UUID"`                           // 关联模块ID
	StoryID      *string `json:"storyId" example:"UUID"`                            // 关联需求ID
	UserID       *string `json:"userId" binding:"required" example:"UUID"`          // 负责人ID
	EndDate      *string `json:"endDate" binding:"required" example:"2024-12-31"`   // 结束日期
	StartDate    *string `json:"startDate" binding:"required" example:"2024-01-01"` // 开始日期
}

type UpdateTaskFieldRequest struct {
	Key   string      `json:"key" binding:"required" example:"taskStatus"` // 要更新的字段名
	Value interface{} `json:"value" example:"0"`                           // 要更新的字段值
}

type UpdateTaskNextRequest struct {
	TaskStatus     string `json:"taskStatus" binding:"required" example:"0"` // 任务状态
	ChangeRichText string `json:"changeRichText" example:"<p>流转说明</p>"`      // 流转说明(富文本)
}

type CreateBugRequest struct {
	BugLevel    string  `json:"bugLevel" example:"0"`                        // 缺陷等级
	BugEnv      string  `json:"bugEnv" example:"0"`                          // 缺陷环境
	BugStatus   string  `json:"bugStatus" example:"0"`                       // 缺陷状态
	BugSource   string  `json:"bugSource" example:"0"`                       // 缺陷来源
	BugType     string  `json:"bugType" example:"0"`                         // 缺陷类型
	BugUa       *string `json:"bugUa" example:"Mozilla/5.0"`                 // 用户代理
	BugTitle    *string `json:"bugTitle" binding:"required" example:"缺陷标题"`  // 缺陷标题
	ProjectID   string  `json:"projectId" binding:"required" example:"UUID"` // 关联项目ID
	BugRichText *string `json:"bugRichText" example:"<p>缺陷描述</p>"`           // 缺陷描述(富文本)
	VersionID   *string `json:"versionId" example:"UUID"`                    // 关联版本ID
	ModuleID    *string `json:"moduleId" example:"UUID"`                     // 关联模块ID
	StoryID     *string `json:"storyId" example:"UUID"`                      // 关联需求ID
	UserID      *string `json:"userId" binding:"required" example:"UUID"`    // 指派人ID
}

type UpdateBugRequest struct {
	BugLevel    string  `json:"bugLevel" example:"0"`                        // 缺陷等级
	BugEnv      string  `json:"bugEnv" example:"0"`                          // 缺陷环境
	BugStatus   string  `json:"bugStatus" example:"0"`                       // 缺陷状态
	BugSource   string  `json:"bugSource" example:"0"`                       // 缺陷来源
	BugType     string  `json:"bugType" example:"0"`                         // 缺陷类型
	BugUa       *string `json:"bugUa" example:"Mozilla/5.0"`                 // 用户代理
	BugTitle    *string `json:"bugTitle" binding:"required" example:"缺陷标题"`  // 缺陷标题
	ProjectID   string  `json:"projectId" binding:"required" example:"UUID"` // 关联项目ID
	BugRichText *string `json:"bugRichText" example:"<p>缺陷描述</p>"`           // 缺陷描述(富文本)
	VersionID   *string `json:"versionId" example:"UUID"`                    // 关联版本ID
	ModuleID    *string `json:"moduleId" example:"UUID"`                     // 关联模块ID
	StoryID     *string `json:"storyId" example:"UUID"`                      // 关联需求ID
	UserID      *string `json:"userId" binding:"required" example:"UUID"`    // 指派人ID
}

type UpdateBugFieldRequest struct {
	Key   string      `json:"key" binding:"required" example:"bugStatus"` // 要更新的字段名
	Value interface{} `json:"value" example:"0"`                          // 要更新的字段值
}

type UpdateBugNextRequest struct {
	BugStatus      string `json:"bugStatus" binding:"required" example:"0"` // 缺陷状态
	ChangeRichText string `json:"changeRichText" example:"<p>流转说明</p>"`     // 流转说明(富文本)
}

type ConfirmBugRequest struct {
	BugConfirmStatus string `json:"bugConfirmStatus" binding:"required" example:"0"` // 缺陷确认状态
	ChangeRichText   string `json:"changeRichText" example:"<p>确认说明</p>"`            // 确认说明(富文本)
}
