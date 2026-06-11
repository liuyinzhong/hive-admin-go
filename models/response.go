package models

import (
	"time"
)

type Response struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Error   interface{} `json:"error"`
	Message string      `json:"message"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken string `json:"accessToken"`
}

type ProfileResponse struct {
	UserId     string   `json:"userId"`
	Avatar     *string  `json:"avatar"`
	Username   string   `json:"username"`
	RealName   string   `json:"realName"`
	RoleTitles []string `json:"roleTitles"`
	RoleIds    []string `json:"roleIds"`
	Desc       *string  `json:"desc"`
	Email      *string  `json:"email"`
	HomePath   *string  `json:"homePath"`
	DeptTitles []string `json:"deptTitles"`
	DeptIds    []string `json:"deptIds"`
	Status     int      `json:"status"`
	CreateDate *string  `json:"createDate"`
	UpdateDate *string  `json:"updateDate"`
}

type MenuMeta struct {
	ActiveIcon               *string `json:"activeIcon"`
	ActivePath               *string `json:"activePath"`
	AffixTab                 bool    `json:"affixTab"`
	AffixTabOrder            int     `json:"affixTabOrder"`
	Badge                    *string `json:"badge"`
	BadgeType                *string `json:"badgeType"`
	BadgeVariants            *string `json:"badgeVariants"`
	HideChildrenInMenu       bool    `json:"hideChildrenInMenu"`
	HideInBreadcrumb         bool    `json:"hideInBreadcrumb"`
	HideInMenu               bool    `json:"hideInMenu"`
	HideInTab                bool    `json:"hideInTab"`
	Icon                     *string `json:"icon"`
	IframeSrc                *string `json:"iframeSrc"`
	KeepAlive                bool    `json:"keepAlive"`
	Link                     *string `json:"link"`
	MaxNumOfOpenTab          int     `json:"maxNumOfOpenTab"`
	NoBasicLayout            bool    `json:"noBasicLayout"`
	OpenInNewWindow          bool    `json:"openInNewWindow"`
	Order                    *int    `json:"order"`
	Query                    *string `json:"query"`
	Title                    string  `json:"title"`
	DomCached                bool    `json:"domCached"`
	MenuVisibleWithForbidden bool    `json:"menuVisibleWithForbidden"`
}

type MenuTreeResponse struct {
	ID          string              `json:"id"`
	Pid         *string             `json:"pid"`
	Type        string              `json:"type"`
	AuthCode    *string             `json:"authCode"`
	Children    []*MenuTreeResponse `json:"children"`
	Component   *string             `json:"component"`
	Meta        MenuMeta            `json:"meta"`
	Name        string              `json:"name"`
	Path        *string             `json:"path"`
	CreatorId   *string             `json:"creatorId"`
	CreatorName *string             `json:"creatorName"`
	CreateDate  *string             `json:"createDate"`
	UpdateDate  *string             `json:"updateDate"`
	Status      int                 `json:"status"`
}

type UserListRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
	Username string `form:"username"`
	RealName string `form:"realName"`
	Status   *int   `form:"status"`
	Sorts    string `form:"sorts"`
}

type FileListRequest struct {
	Page         int    `form:"page"`
	PageSize     int    `form:"pageSize"`
	OriginalName string `form:"originalName"`
	Type         string `form:"type"`
	FileExt      string `form:"fileExt"`
	Sorts        string `form:"sorts"`
}

type CreateUserRequest struct {
	Username string   `json:"username" binding:"required"`
	RealName string   `json:"realName" binding:"required"`
	Password string   `json:"password" binding:"required"`
	Desc     *string  `json:"desc"`
	DeptIds  []string `json:"deptIds"`
	RoleIds  []string `json:"roleIds"`
}

type UpdateUserRequest struct {
	Username string   `json:"username" binding:"required"`
	RealName string   `json:"realName" binding:"required"`
	Desc     *string  `json:"desc"`
	DeptIds  []string `json:"deptIds"`
	RoleIds  []string `json:"roleIds"`
}

type UpdateUserStatusRequest struct {
	Status int `json:"status"`
}

type MenuListRequest struct {
	Name   string `form:"name"`
	Path   string `form:"path"`
	Type   string `form:"type"`
	Status *int   `form:"status"`
}

type CreateMenuRequest struct {
	Pid       *string  `json:"pid"`
	Type      string   `json:"type" binding:"required"`
	AuthCode  *string  `json:"authCode"`
	Component *string  `json:"component"`
	Name      string   `json:"name" binding:"required"`
	Path      *string  `json:"path"`
	Meta      MenuMeta `json:"meta" binding:"required"`
	Status    int      `json:"status"`
}

type UpdateMenuRequest struct {
	Pid       *string  `json:"pid"`
	Type      string   `json:"type" binding:"required"`
	AuthCode  *string  `json:"authCode"`
	Component *string  `json:"component"`
	Name      string   `json:"name" binding:"required"`
	Path      *string  `json:"path"`
	Meta      MenuMeta `json:"meta" binding:"required"`
	Status    int      `json:"status"`
}

type RoleListRequest struct {
	Page      int    `form:"page"`
	PageSize  int    `form:"pageSize"`
	RoleTitle string `form:"roleTitle"`
	Status    *int   `form:"status"`
	Remark    string `form:"remark"`
	StartDate string `form:"startDate"`
	EndDate   string `form:"endDate"`
	Sorts     string `form:"sorts"`
}

type CreateRoleRequest struct {
	RoleTitle   string   `json:"roleTitle" binding:"required"`
	Status      int      `json:"status"`
	Remark      *string  `json:"remark"`
	Permissions []string `json:"permissions"`
}

type UpdateRoleRequest struct {
	RoleTitle   string   `json:"roleTitle" binding:"required"`
	Status      int      `json:"status"`
	Remark      *string  `json:"remark"`
	Permissions []string `json:"permissions"`
}

type RoleDetailResponse struct {
	RoleId      string   `json:"roleId"`
	RoleTitle   string   `json:"roleTitle"`
	Status      int      `json:"status"`
	CreateDate  *string  `json:"createDate"`
	Remark      *string  `json:"remark"`
	Permissions []string `json:"permissions"`
}

type RoleSimpleResponse struct {
	RoleId    string `json:"roleId"`
	RoleTitle string `json:"roleTitle"`
	Status    int    `json:"status"`
}

type UpdateStatusRequest struct {
	Status int `json:"status"`
}

type DeptListRequest struct {
	DeptTitle string `form:"deptTitle"`
}

type DeptTreeResponse struct {
	DeptId     string              `json:"deptId"`
	Pid        *string             `json:"pid"`
	DeptTitle  string              `json:"deptTitle"`
	Status     int                 `json:"status"`
	CreateDate *string             `json:"createDate"`
	Remark     *string             `json:"remark"`
	Children   []*DeptTreeResponse `json:"children"`
}

type CreateDeptRequest struct {
	Pid       *string `json:"pid"`
	DeptTitle string  `json:"deptTitle" binding:"required"`
	Status    int     `json:"status"`
	Remark    *string `json:"remark"`
}

type UpdateDeptRequest struct {
	Pid       *string `json:"pid"`
	DeptTitle string  `json:"deptTitle" binding:"required"`
	Status    int     `json:"status"`
	Remark    *string `json:"remark"`
}

type DictListRequest struct {
	Label string `form:"label"`
	Value string `form:"value"`
	Type  string `form:"type"`
	Sorts string `form:"sorts"`
}

type DictTreeResponse struct {
	ID         string              `json:"id"`
	Pid        *string             `json:"pid"`
	Label      string              `json:"label"`
	Value      *string             `json:"value"`
	Type       string              `json:"type"`
	Remark     *string             `json:"remark"`
	Color      *string             `json:"color"`
	Status     int                 `json:"status"`
	CreateDate *string             `json:"createDate"`
	UpdateDate *string             `json:"updateDate"`
	Children   []*DictTreeResponse `json:"children"`
}

type CreateDictRequest struct {
	Pid    *string `json:"pid"`
	Type   string  `json:"type" binding:"required"`
	Label  string  `json:"label" binding:"required"`
	Value  *string `json:"value"`
	Color  *string `json:"color"`
	Status int     `json:"status"`
	Remark *string `json:"remark"`
}

type UpdateDictRequest struct {
	Pid    *string `json:"pid"`
	Type   string  `json:"type" binding:"required"`
	Label  string  `json:"label" binding:"required"`
	Value  *string `json:"value"`
	Color  *string `json:"color"`
	Status int     `json:"status"`
	Remark *string `json:"remark"`
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
		UserId:     user.UserID,
		Avatar:     user.Avatar,
		Username:   username,
		RealName:   realName,
		RoleTitles: roleTitles,
		RoleIds:    roleIds,
		Desc:       user.Desc,
		Email:      user.Email,
		HomePath:   user.HomePath,
		DeptTitles: deptTitles,
		DeptIds:    deptIds,
		Status:     user.Status,
		CreateDate: TimeToStringPtr(user.CreateDate),
		UpdateDate: TimeToStringPtr(user.UpdateDate),
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
	ProjectID    *string `json:"projectId"`
	ProjectTitle *string `json:"projectTitle"`
	ProjectLogo  *string `json:"projectLogo"`
	Description  *string `json:"description"`
	CreateDate   *string `json:"createDate"`
}

type ModuleResponse struct {
	ModuleID     *string `json:"moduleId"`
	ModuleTitle  *string `json:"moduleTitle"`
	ProjectID    *string `json:"projectId"`
	ProjectTitle *string `json:"projectTitle"`
	Sort         int     `json:"sort"`
	UpdateDate   *string `json:"updateDate"`
	CreateDate   *string `json:"createDate"`
}

type VersionResponse struct {
	VersionID         *string `json:"versionId"`
	Version           *string `json:"version"`
	VersionType       string  `json:"versionType"`
	Remark            *string `json:"remark"`
	CreatorID         *string `json:"creatorId"`
	CreatorName       *string `json:"creatorName"`
	CreateDate        *string `json:"createDate"`
	EndDate           *string `json:"endDate"`
	StartDate         *string `json:"startDate"`
	ProjectID         *string `json:"projectId"`
	ProjectTitle      *string `json:"projectTitle"`
	ReleaseStatus     string  `json:"releaseStatus"`
	ReleaseDate       *string `json:"releaseDate"`
	ChangeLogRichText *string `json:"changeLogRichText"`
	ChangeLog         *string `json:"changeLog"`
}

type StoryUserItem struct {
	UserID   *string `json:"userId"`
	Avatar   *string `json:"avatar"`
	RealName *string `json:"realName"`
}

type StoryResponse struct {
	StoryID       *string         `json:"storyId"`
	StoryTitle    *string         `json:"storyTitle"`
	StoryNum      int             `json:"storyNum"`
	CreatorName   *string         `json:"creatorName"`
	CreatorID     *string         `json:"creatorId"`
	StoryType     string          `json:"storyType"`
	StoryStatus   string          `json:"storyStatus"`
	StoryLevel    string          `json:"storyLevel"`
	VersionID     *string         `json:"versionId"`
	Version       *string         `json:"version"`
	ProjectID     *string         `json:"projectId"`
	ProjectTitle  *string         `json:"projectTitle"`
	ModuleID      *string         `json:"moduleId"`
	ModuleTitle   *string         `json:"moduleTitle"`
	Source        string          `json:"source"`
	UpdateDate    *string         `json:"updateDate"`
	CreateDate    *string         `json:"createDate"`
	UserList      []StoryUserItem `json:"userList"`
	UserIDs       []string        `json:"userIds"`
	StoryRichText *string         `json:"storyRichText"`
	FileIDs       []string        `json:"fileIds"`
	FileList      []FileResponse  `json:"fileList"`
	TaskList      []TaskResponse  `json:"taskList"`
	BugList       []BugResponse   `json:"bugList"`
	Nodes         []NodeResponse  `json:"nodes"`
}

type FileResponse struct {
	FileID        *string `json:"fileId"`
	URL           *string `json:"url"`
	Name          *string `json:"name"`
	Type          *string `json:"type"`
	Size          int64   `json:"size"`
	FileExt       *string `json:"fileExt"`
	OriginalName  *string `json:"originalName"`
	Path          *string `json:"path"`
	FullPath      *string `json:"fullPath"`
	ThumbnailPath *string `json:"thumbnailPath"`
	ThumbnailURL  *string `json:"thumbnailUrl"`
	CreatorID     *string `json:"creatorId"`
	CreatorName   *string `json:"creatorName"`
	CreateDate    *string `json:"createDate"`
}

type TaskResponse struct {
	TaskID       *string `json:"taskId"`
	StoryID      *string `json:"storyId"`
	StoryTitle   *string `json:"storyTitle"`
	ModuleID     *string `json:"moduleId"`
	ModuleTitle  *string `json:"moduleTitle"`
	VersionID    *string `json:"versionId"`
	Version      *string `json:"version"`
	ProjectID    *string `json:"projectId"`
	ProjectTitle *string `json:"projectTitle"`
	TaskTitle    *string `json:"taskTitle"`
	TaskNum      int     `json:"taskNum"`
	TaskStatus   string  `json:"taskStatus"`
	TaskType     string  `json:"taskType"`
	PlanHours    float64 `json:"planHours"`
	ActualHours  float64 `json:"actualHours"`
	EndDate      *string `json:"endDate"`
	StartDate    *string `json:"startDate"`
	CreateDate   *string `json:"createDate"`
	CreatorID    *string `json:"creatorId"`
	CreatorName  *string `json:"creatorName"`
	UserID       *string `json:"userId"`
	RealName     *string `json:"realName"`
	Avatar       *string `json:"avatar"`
	Percent      int     `json:"percent"`
	TaskRichText *string `json:"taskRichText"`
}

type BugResponse struct {
	BugID            *string `json:"bugId"`
	BugTitle         *string `json:"bugTitle"`
	BugNum           int     `json:"bugNum"`
	BugStatus        string  `json:"bugStatus"`
	BugConfirmStatus string  `json:"bugConfirmStatus"`
	BugLevel         string  `json:"bugLevel"`
	BugSource        string  `json:"bugSource"`
	BugType          string  `json:"bugType"`
	BugEnv           string  `json:"bugEnv"`
	BugUa            *string `json:"bugUa"`
	UserID           *string `json:"userId"`
	Avatar           *string `json:"avatar"`
	RealName         *string `json:"realName"`
	CreatorName      *string `json:"creatorName"`
	CreatorID        *string `json:"creatorId"`
	VersionID        *string `json:"versionId"`
	Version          *string `json:"version"`
	ModuleID         *string `json:"moduleId"`
	ModuleTitle      *string `json:"moduleTitle"`
	ProjectID        *string `json:"projectId"`
	ProjectTitle     *string `json:"projectTitle"`
	StoryID          *string `json:"storyId"`
	StoryTitle       *string `json:"storyTitle"`
	UpdateDate       *string `json:"updateDate"`
	CreateDate       *string `json:"createDate"`
	BugRichText      *string `json:"bugRichText"`
}

type ChangeHistoryResponse struct {
	ChangeID       *string `json:"changeId"`
	ChangeBehavior string  `json:"changeBehavior"`
	ChangeRichText *string `json:"changeRichText"`
	CreatorID      *string `json:"creatorId"`
	CreatorName    *string `json:"creatorName"`
	BusinessID     *string `json:"businessId"`
	BusinessType   string  `json:"businessType"`
	ExtendJson     *string `json:"extendJson"`
	CreateDate     *string `json:"createDate"`
	UpdateDate     *string `json:"updateDate"`
}

type CreateChangeHistoryRequest struct {
	BusinessID     string `json:"businessId" binding:"required"`
	BusinessType   string `json:"businessType" binding:"required"`
	ChangeBehavior string `json:"changeBehavior" binding:"required"`
	ChangeRichText string `json:"changeRichText"`
}

type CreateProjectRequest struct {
	ProjectTitle *string `json:"projectTitle" binding:"required"`
	Description  *string `json:"description"`
	ProjectLogo  *string `json:"projectLogo"`
}

type UpdateProjectRequest struct {
	ProjectTitle *string `json:"projectTitle" binding:"required"`
	Description  *string `json:"description"`
	ProjectLogo  *string `json:"projectLogo"`
}

type CreateModuleRequest struct {
	Sort        int     `json:"sort"`
	ModuleTitle *string `json:"moduleTitle" binding:"required"`
	ProjectID   string  `json:"projectId" binding:"required"`
}

type UpdateModuleRequest struct {
	Sort        int     `json:"sort"`
	ModuleTitle *string `json:"moduleTitle" binding:"required"`
}

type CreateVersionRequest struct {
	Version       *string `json:"version"`
	VersionType   string  `json:"versionType"`
	ReleaseStatus string  `json:"releaseStatus"`
	ProjectID     string  `json:"projectId"`
	Remark        *string `json:"remark"`
	EndDate       *string `json:"endDate"`
	StartDate     *string `json:"startDate"`
}

type UpdateVersionRequest struct {
	Version       *string `json:"version"`
	VersionType   string  `json:"versionType"`
	ReleaseStatus string  `json:"releaseStatus"`
	ProjectID     string  `json:"projectId"`
	Remark        *string `json:"remark"`
	EndDate       *string `json:"endDate"`
	StartDate     *string `json:"startDate"`
}

type UpdateVersionNextRequest struct {
	ReleaseStatus  string `json:"releaseStatus"`
	ChangeRichText string `json:"changeRichText"`
}

type CreateNodeItemRequest struct {
	Label    string  `json:"label" example:"需求提交"`                 // 节点名称
	Value    string  `json:"value" example:"0"`                    // 节点值
	Sort     int     `json:"sort" example:"1"`                     // 节点顺序
	UserID   string  `json:"userId" example:"UUID"`                // 负责人id
	NodeType int     `json:"nodeType" enums:"0,1,2,3" example:"0"` // 节点类型 0=开始 1=办理 2=审批 3=结束
	Remark   *string `json:"remark" example:"流程开始节点"`              // 备注
}

type CreateStoryRequest struct {
	StoryStatus   string                  `json:"storyStatus" example:"0"`                      // 需求状态,字典STORY_STATUS值
	StoryType     string                  `json:"storyType" binding:"required" example:"0"`     // 需求类型,字典STORY_TYPE值
	StoryLevel    string                  `json:"storyLevel" example:"0"`                       // 需求优先级,字典STORY_LEVEL值
	Source        string                  `json:"source" example:"0"`                           // 需求来源,字典STORY_SOURCE值
	StoryTitle    *string                 `json:"storyTitle" binding:"required" example:"需求标题"` // 需求标题
	StoryRichText *string                 `json:"storyRichText" example:"需求描述"`                 // 需求描述,富文本格式
	UserIDs       []string                `json:"userIds" example:"[\"UUID\"]"`                 // 参与人员id数组
	ProjectID     string                  `json:"projectId" binding:"required" example:"UUID"`  // 关联项目id
	VersionID     *string                 `json:"versionId" binding:"required" example:"UUID"`  // 关联版本id
	ModuleID      *string                 `json:"moduleId" binding:"required" example:"UUID"`   // 关联模块id
	FileIDs       []string                `json:"fileIds" example:"[\"UUID\"]"`                 // 附件id数组
	BusinessType  string                  `json:"businessType" example:"0"`                     // 业务类型,关联dev_node表business_type
	Nodes         []CreateNodeItemRequest `json:"nodes"`                                        // 节点信息,新增需求时创建对应节点
}

type UpdateStoryRequest struct {
	StoryStatus   string   `json:"storyStatus"`
	StoryType     string   `json:"storyType" binding:"required"`
	StoryLevel    string   `json:"storyLevel"`
	Source        string   `json:"source"`
	StoryTitle    *string  `json:"storyTitle" binding:"required"`
	StoryRichText *string  `json:"storyRichText"`
	UserIDs       []string `json:"userIds"`
	ProjectID     string   `json:"projectId" binding:"required"`
	VersionID     *string  `json:"versionId" binding:"required"`
	ModuleID      *string  `json:"moduleId" binding:"required"`
	FileIDs       []string `json:"fileIds"`
}

type UpdateStoryFieldRequest struct {
	Key   string      `json:"key" binding:"required"`
	Value interface{} `json:"value"`
}

type UpdateStoryNextRequest struct {
	StoryStatus    string `json:"storyStatus" binding:"required"`
	ChangeRichText string `json:"changeRichText"`
}

type CreateTaskRequest struct {
	PlanHours    float64 `json:"planHours" binding:"required"`
	TaskStatus   string  `json:"taskStatus"`
	TaskType     string  `json:"taskType"`
	TaskTitle    *string `json:"taskTitle" binding:"required"`
	ProjectID    string  `json:"projectId" binding:"required"`
	TaskRichText *string `json:"taskRichText"`
	VersionID    *string `json:"versionId"`
	ModuleID     *string `json:"moduleId"`
	StoryID      *string `json:"storyId"`
	UserID       *string `json:"userId" binding:"required"`
	EndDate      *string `json:"endDate" binding:"required"`
	StartDate    *string `json:"startDate" binding:"required"`
}

type UpdateTaskRequest struct {
	PlanHours    float64 `json:"planHours" binding:"required"`
	TaskStatus   string  `json:"taskStatus"`
	TaskType     string  `json:"taskType"`
	TaskTitle    *string `json:"taskTitle" binding:"required"`
	ProjectID    string  `json:"projectId" binding:"required"`
	TaskRichText *string `json:"taskRichText"`
	VersionID    *string `json:"versionId"`
	ModuleID     *string `json:"moduleId"`
	StoryID      *string `json:"storyId"`
	UserID       *string `json:"userId" binding:"required"`
	EndDate      *string `json:"endDate" binding:"required"`
	StartDate    *string `json:"startDate" binding:"required"`
}

type UpdateTaskFieldRequest struct {
	Key   string      `json:"key" binding:"required"`
	Value interface{} `json:"value"`
}

type UpdateTaskNextRequest struct {
	TaskStatus     string `json:"taskStatus" binding:"required"`
	ChangeRichText string `json:"changeRichText"`
}

type CreateBugRequest struct {
	BugLevel    string  `json:"bugLevel"`
	BugEnv      string  `json:"bugEnv"`
	BugStatus   string  `json:"bugStatus"`
	BugSource   string  `json:"bugSource"`
	BugType     string  `json:"bugType"`
	BugUa       *string `json:"bugUa"`
	BugTitle    *string `json:"bugTitle" binding:"required"`
	ProjectID   string  `json:"projectId" binding:"required"`
	BugRichText *string `json:"bugRichText"`
	VersionID   *string `json:"versionId"`
	ModuleID    *string `json:"moduleId"`
	StoryID     *string `json:"storyId"`
	UserID      *string `json:"userId" binding:"required"`
}

type UpdateBugRequest struct {
	BugLevel    string  `json:"bugLevel"`
	BugEnv      string  `json:"bugEnv"`
	BugStatus   string  `json:"bugStatus"`
	BugSource   string  `json:"bugSource"`
	BugType     string  `json:"bugType"`
	BugUa       *string `json:"bugUa"`
	BugTitle    *string `json:"bugTitle" binding:"required"`
	ProjectID   string  `json:"projectId" binding:"required"`
	BugRichText *string `json:"bugRichText"`
	VersionID   *string `json:"versionId"`
	ModuleID    *string `json:"moduleId"`
	StoryID     *string `json:"storyId"`
	UserID      *string `json:"userId" binding:"required"`
}

type UpdateBugFieldRequest struct {
	Key   string      `json:"key" binding:"required"`
	Value interface{} `json:"value"`
}

type UpdateBugNextRequest struct {
	BugStatus      string `json:"bugStatus" binding:"required"`
	ChangeRichText string `json:"changeRichText"`
}

type ConfirmBugRequest struct {
	BugConfirmStatus string `json:"bugConfirmStatus" binding:"required"`
	ChangeRichText   string `json:"changeRichText"`
}

type NodeListRequest struct {
	BusinessID string `form:"businessId"`
}

type NodeResponse struct {
	NodeID         string  `json:"nodeId"`
	Label          string  `json:"label"`
	Value          string  `json:"value"`
	Sort           int     `json:"sort"`
	UserID         string  `json:"userId"`
	RealName       string  `json:"realName"`
	Current        bool    `json:"current"`
	NodeType       int     `json:"nodeType"`
	Result         int     `json:"result"`
	Remark         *string `json:"remark"`
	ResultRichText *string `json:"resultRichText"`
	BusinessType   *string `json:"businessType"`
	BusinessID     string  `json:"businessId"`
	StartDate      *string `json:"startDate"`
	EndDate        *string `json:"endDate"`
	CreateDate     *string `json:"createDate"`
}

type CreateNodeRequest struct {
	Label        string  `json:"label" binding:"required"`
	Value        string  `json:"value" binding:"required"`
	Sort         int     `json:"sort"`
	UserID       string  `json:"userId" binding:"required"`
	NodeType     int     `json:"nodeType"`
	Remark       *string `json:"remark"`
	BusinessType string  `json:"businessType"`
	BusinessID   string  `json:"businessId" binding:"required"`
}

type NodeApproveRequest struct {
	Result         int    `json:"result" binding:"required"`
	ResultRichText string `json:"resultRichText"`
}

func DevNodeToNodeResponse(node DevNode, realName string) *NodeResponse {
	return &NodeResponse{
		NodeID:         node.NodeID,
		Label:          node.Label,
		Value:          node.Value,
		Sort:           node.Sort,
		UserID:         node.UserID,
		RealName:       realName,
		Current:        node.Current == 1,
		NodeType:       node.NodeType,
		Result:         node.Result,
		Remark:         node.Remark,
		ResultRichText: node.ResultRichText,
		BusinessType:   node.BusinessType,
		BusinessID:     node.BusinessID,
		StartDate:      TimeToStringPtr(node.StartDate),
		EndDate:        TimeToStringPtr(node.EndDate),
		CreateDate:     TimeToStringPtr(node.CreateDate),
	}
}
