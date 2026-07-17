package models

import (
	"time"
)

type SysUser struct {
	UserID       string     `gorm:"column:user_id;type:char(36);primaryKey" json:"userId"`
	Avatar       *string    `gorm:"column:avatar;type:varchar(256)" json:"avatar"`
	Username     *string    `gorm:"column:username;type:varchar(36)" json:"username"`
	RealName     *string    `gorm:"column:real_name;type:varchar(12)" json:"realName"`
	Desc         *string    `gorm:"column:desc;type:varchar(128)" json:"desc"`
	Email        *string    `gorm:"column:email;type:varchar(128)" json:"email"`
	Phone        *string    `gorm:"column:phone;type:varchar(20)" json:"phone"`
	Password     *string    `gorm:"column:password;type:varchar(512)" json:"-"`
	HomePath     *string    `gorm:"column:home_path;type:varchar(128)" json:"homePath"`
	LeaderUserID *string    `gorm:"column:leader_user_id;type:char(36)" json:"leaderUserId"`
	Status       int        `gorm:"column:status;type:tinyint;default:1" json:"status"`
	CreateDate   *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate   *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag      int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
	IsSys        int        `gorm:"column:is_sys;type:tinyint;default:0" json:"isSys"`
}

func (SysUser) TableName() string {
	return "sys_user"
}

type SysDept struct {
	DeptID     string     `gorm:"column:dept_id;type:char(36);primaryKey" json:"deptId"`
	DeptTitle  *string    `gorm:"column:dept_title;type:varchar(36)" json:"deptTitle"`
	Pid        *string    `gorm:"column:pid;type:varchar(36)" json:"pid"`
	Remark     *string    `gorm:"column:remark;type:varchar(128)" json:"remark"`
	CreateDate *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag    int        `gorm:"column:del_flag;type:int;default:0" json:"delFlag"`
	Status     int        `gorm:"column:status;type:tinyint;default:1" json:"status"`
}

func (SysDept) TableName() string {
	return "sys_dept"
}

type SysRole struct {
	RoleID     string     `gorm:"column:role_id;type:char(36);primaryKey" json:"roleId"`
	RoleTitle  *string    `gorm:"column:role_title;type:varchar(36)" json:"roleTitle"`
	Remark     *string    `gorm:"column:remark;type:varchar(128)" json:"remark"`
	Status     int        `gorm:"column:status;type:tinyint;default:1" json:"status"`
	CreateDate *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag    int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

func (SysRole) TableName() string {
	return "sys_role"
}

type SysMenu struct {
	ID                       string     `gorm:"column:id;type:char(36);primaryKey" json:"id"`
	Pid                      *string    `gorm:"column:pid;type:varchar(36)" json:"pid"`
	Type                     string     `gorm:"column:type;type:varchar(36)" json:"type"`
	Icon                     *string    `gorm:"column:icon;type:varchar(128)" json:"icon"`
	ActiveIcon               *string    `gorm:"column:active_icon;type:varchar(128)" json:"activeIcon"`
	KeepAlive                int        `gorm:"column:keep_alive;type:tinyint;default:0" json:"keepAlive"`
	HideInMenu               int        `gorm:"column:hide_in_menu;type:tinyint;default:0" json:"hideInMenu"`
	HideInTab                int        `gorm:"column:hide_in_tab;type:tinyint;default:0" json:"hideInTab"`
	HideInBreadcrumb         int        `gorm:"column:hide_in_breadcrumb;type:tinyint;default:0" json:"hideInBreadcrumb"`
	HideChildrenInMenu       int        `gorm:"column:hide_children_in_menu;type:tinyint;default:0" json:"hideChildrenInMenu"`
	Badge                    *string    `gorm:"column:badge;type:varchar(4)" json:"badge"`
	BadgeType                *string    `gorm:"column:badge_type;type:varchar(36)" json:"badgeType"`
	BadgeVariants            *string    `gorm:"column:badge_variants;type:varchar(36)" json:"badgeVariants"`
	ActivePath               *string    `gorm:"column:active_path;type:varchar(128)" json:"activePath"`
	AuthCode                 *string    `gorm:"column:auth_code;type:varchar(512)" json:"authCode"`
	AffixTab                 int        `gorm:"column:affix_tab;type:tinyint;default:0" json:"affixTab"`
	Component                *string    `gorm:"column:component;type:varchar(128)" json:"component"`
	Title                    string     `gorm:"column:title;type:varchar(128)" json:"title"`
	Name                     *string    `gorm:"column:name;type:varchar(128)" json:"name"`
	Path                     *string    `gorm:"column:path;type:varchar(128)" json:"path"`
	Status                   int        `gorm:"column:status;type:tinyint;default:1" json:"status"`
	Link                     *string    `gorm:"column:link;type:varchar(1024)" json:"link"`
	IframeSrc                *string    `gorm:"column:iframe_src;type:varchar(1024)" json:"iframeSrc"`
	Order                    *int       `gorm:"column:order;type:tinyint" json:"order"`
	MaxNumOfOpenTab          int        `gorm:"column:max_num_of_open_tab;type:tinyint;default:-1" json:"maxNumOfOpenTab"`
	AffixTabOrder            int        `gorm:"column:affix_tab_order;type:tinyint;default:0" json:"affixTabOrder"`
	NoBasicLayout            int        `gorm:"column:no_basic_layout;type:tinyint;default:0" json:"noBasicLayout"`
	OpenInNewWindow          int        `gorm:"column:open_in_new_window;type:tinyint;default:0" json:"openInNewWindow"`
	DomCached                int        `gorm:"column:dom_cached;type:tinyint;default:0" json:"domCached"`
	Query                    *string    `gorm:"column:query;type:varchar(512)" json:"query"`
	MenuVisibleWithForbidden int        `gorm:"column:menu_visible_with_forbidden;type:tinyint;default:0" json:"menuVisibleWithForbidden"`
	CreatorID                *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	CreatorName              *string    `gorm:"column:creator_name;type:varchar(12)" json:"creatorName"`
	CreateDate               *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate               *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag                  int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

func (SysMenu) TableName() string {
	return "sys_menu"
}

type SysUserRole struct {
	ID         string     `gorm:"column:id;type:char(36);primaryKey" json:"id"`
	UserID     string     `gorm:"column:user_id;type:char(36)" json:"userId"`
	RoleID     string     `gorm:"column:role_id;type:char(36)" json:"roleId"`
	CreateDate *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag    int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

func (SysUserRole) TableName() string {
	return "sys_user_role"
}

type SysUserDept struct {
	ID         string     `gorm:"column:id;type:char(36);primaryKey" json:"id"`
	UserID     string     `gorm:"column:user_id;type:char(36)" json:"userId"`
	DeptID     string     `gorm:"column:dept_id;type:char(36)" json:"deptId"`
	CreateDate *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag    int        `gorm:"column:del_flag;type:int;default:0" json:"delFlag"`
}

func (SysUserDept) TableName() string {
	return "sys_user_dept"
}

type SysRoleMenu struct {
	ID         string     `gorm:"column:id;type:char(36);primaryKey" json:"id"`
	RoleID     string     `gorm:"column:role_id;type:char(36)" json:"roleId"`
	MenuID     string     `gorm:"column:menu_id;type:char(36)" json:"menuId"`
	CreateDate *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag    int        `gorm:"column:del_flag;type:int;default:0" json:"delFlag"`
}

func (SysRoleMenu) TableName() string {
	return "sys_role_menu"
}

type SysDict struct {
	ID         string     `gorm:"column:id;type:char(36);primaryKey" json:"id"`
	Pid        *string    `gorm:"column:pid;type:varchar(36)" json:"pid"`
	Label      *string    `gorm:"column:label;type:varchar(128)" json:"label"`
	Value      *string    `gorm:"column:value;type:varchar(36)" json:"value"`
	Type       string     `gorm:"column:type;type:varchar(36)" json:"type"`
	Remark     *string    `gorm:"column:remark;type:varchar(128)" json:"remark"`
	Color      *string    `gorm:"column:color;type:varchar(7)" json:"color"`
	CreateDate *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag    int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
	Status     int        `gorm:"column:status;type:int;default:1" json:"status"`
}

func (SysDict) TableName() string {
	return "sys_dict"
}

type DevProject struct {
	ProjectID    string     `gorm:"column:project_id;type:char(36);primaryKey" json:"projectId"`
	ProjectTitle *string    `gorm:"column:project_title;type:varchar(16)" json:"projectTitle"`
	ProjectLogo  *string    `gorm:"column:project_logo;type:varchar(256)" json:"projectLogo"`
	Description  *string    `gorm:"column:description;type:varchar(128)" json:"description"`
	CreateDate   *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate   *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag      int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

func (DevProject) TableName() string {
	return "dev_project"
}

type DevModule struct {
	ModuleID    string     `gorm:"column:module_id;type:char(36);primaryKey" json:"moduleId"`
	ProjectID   string     `gorm:"column:project_id;type:char(36)" json:"projectId"`
	ModuleTitle *string    `gorm:"column:module_title;type:varchar(128)" json:"moduleTitle"`
	Sort        int        `gorm:"column:sort;type:int;default:0" json:"sort"`
	CreateDate  *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate  *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag     int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

func (DevModule) TableName() string {
	return "dev_module"
}

type DevVersion struct {
	VersionID         string     `gorm:"column:version_id;type:char(36);primaryKey" json:"versionId"`
	Version           *string    `gorm:"column:version;type:varchar(36)" json:"version"`
	Remark            *string    `gorm:"column:remark;type:text" json:"remark"`
	VersionType       int        `gorm:"column:version_type;type:tinyint;default:0" json:"versionType"`
	CreatorID         *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	EndDate           *time.Time `gorm:"column:end_date" json:"endDate"`
	StartDate         *time.Time `gorm:"column:start_date" json:"startDate"`
	ProjectID         string     `gorm:"column:project_id;type:char(36)" json:"projectId"`
	ReleaseStatus     int        `gorm:"column:release_status;type:tinyint;default:0" json:"releaseStatus"`
	ReleaseDate       *time.Time `gorm:"column:release_date" json:"releaseDate"`
	ChangeLogRichText *string    `gorm:"column:change_log_rich_text;type:longtext" json:"changeLogRichText"`
	ChangeLog         *string    `gorm:"column:change_log;type:longtext" json:"changeLog"`
	CreateDate        *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate        *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag           int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

func (DevVersion) TableName() string {
	return "dev_version"
}

type DevStory struct {
	StoryID       string     `gorm:"column:story_id;type:char(36);primaryKey" json:"storyId"`
	StoryTitle    *string    `gorm:"column:story_title;type:varchar(128)" json:"storyTitle"`
	StoryNum      int        `gorm:"column:story_num;type:int;autoIncrement" json:"storyNum"`
	CreatorID     *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	StoryRichText *string    `gorm:"column:story_rich_text;type:longtext" json:"storyRichText"`
	StoryType     int        `gorm:"column:story_type;type:tinyint;default:0" json:"storyType"`
	StoryStatus   int        `gorm:"column:story_status;type:tinyint;default:0" json:"storyStatus"`
	StoryLevel    int        `gorm:"column:story_level;type:tinyint;default:0" json:"storyLevel"`
	VersionID     *string    `gorm:"column:version_id;type:char(36)" json:"versionId"`
	ProjectID     string     `gorm:"column:project_id;type:char(36)" json:"projectId"`
	ModuleID      *string    `gorm:"column:module_id;type:char(36)" json:"moduleId"`
	Source        int        `gorm:"column:source;type:tinyint;default:0" json:"source"`
	FileIDs       *string    `gorm:"column:file_ids;type:text" json:"fileIds"`
	UserIDs       *string    `gorm:"column:user_ids;type:varchar(128)" json:"userIds"`
	CreateDate    *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate    *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag       int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

func (DevStory) TableName() string {
	return "dev_story"
}

type DevTask struct {
	TaskID       string     `gorm:"column:task_id;type:char(36);primaryKey" json:"taskId"`
	TaskTitle    *string    `gorm:"column:task_title;type:varchar(128)" json:"taskTitle"`
	TaskNum      int        `gorm:"column:task_num;type:int;autoIncrement" json:"taskNum"`
	TaskRichText *string    `gorm:"column:task_rich_text;type:longtext" json:"taskRichText"`
	TaskStatus   int        `gorm:"column:task_status;type:tinyint;default:0" json:"taskStatus"`
	CreatorID    *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	UserID       *string    `gorm:"column:user_id;type:char(36)" json:"userId"`
	TaskType     int        `gorm:"column:task_type;type:tinyint;default:0" json:"taskType"`
	PlanHours    float64    `gorm:"column:plan_hours;type:float;default:0" json:"planHours"`
	ActualHours  float64    `gorm:"column:actual_hours;type:float;default:0" json:"actualHours"`
	StoryID      *string    `gorm:"column:story_id;type:char(36)" json:"storyId"`
	ModuleID     *string    `gorm:"column:module_id;type:char(36)" json:"moduleId"`
	VersionID    *string    `gorm:"column:version_id;type:char(36)" json:"versionId"`
	ProjectID    string     `gorm:"column:project_id;type:char(36)" json:"projectId"`
	EndDate      *time.Time `gorm:"column:end_date" json:"endDate"`
	StartDate    *time.Time `gorm:"column:start_date" json:"startDate"`
	CreateDate   *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate   *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag      int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

func (DevTask) TableName() string {
	return "dev_task"
}

type DevBug struct {
	BugID            string     `gorm:"column:bug_id;type:char(36);primaryKey" json:"bugId"`
	BugTitle         *string    `gorm:"column:bug_title;type:varchar(128)" json:"bugTitle"`
	BugNum           int        `gorm:"column:bug_num;type:int;autoIncrement" json:"bugNum"`
	BugRichText      *string    `gorm:"column:bug_rich_text;type:longtext" json:"bugRichText"`
	BugStatus        int        `gorm:"column:bug_status;type:tinyint;default:0" json:"bugStatus"`
	BugConfirmStatus int        `gorm:"column:bug_confirm_status;type:tinyint;default:0" json:"bugConfirmStatus"`
	BugLevel         int        `gorm:"column:bug_level;type:tinyint;default:0" json:"bugLevel"`
	BugEnv           int        `gorm:"column:bug_env;type:tinyint;default:0" json:"bugEnv"`
	BugSource        int        `gorm:"column:bug_source;type:tinyint;default:0" json:"bugSource"`
	BugType          int        `gorm:"column:bug_type;type:tinyint;default:0" json:"bugType"`
	BugUa            *string    `gorm:"column:bug_ua;type:varchar(256)" json:"bugUa"`
	UserID           *string    `gorm:"column:user_id;type:char(36)" json:"userId"`
	CreatorID        *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	VersionID        *string    `gorm:"column:version_id;type:char(36)" json:"versionId"`
	ModuleID         *string    `gorm:"column:module_id;type:char(36)" json:"moduleId"`
	ProjectID        string     `gorm:"column:project_id;type:char(36)" json:"projectId"`
	StoryID          *string    `gorm:"column:story_id;type:char(36)" json:"storyId"`
	CreateDate       *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate       *time.Time `gorm:"column:update_date" json:"updateDate"`
	DelFlag          int        `gorm:"column:del_flag;type:tinyint;default:0" json:"delFlag"`
}

func (DevBug) TableName() string {
	return "dev_bug"
}

type DevChangeHistory struct {
	ChangeID       string     `gorm:"column:change_id;type:char(36);primaryKey" json:"changeId"`
	ChangeBehavior int        `gorm:"column:change_behavior;type:tinyint;default:0" json:"changeBehavior"`
	ChangeRichText *string    `gorm:"column:change_rich_text;type:longtext" json:"changeRichText"`
	CreatorID      *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	BusinessID     *string    `gorm:"column:business_id;type:char(12)" json:"businessId"`
	BusinessType   int        `gorm:"column:business_type;type:tinyint;default:0" json:"businessType"`
	ExtendJson     *string    `gorm:"column:extend_json;type:varchar(512)" json:"extendJson"`
	CreateDate     *time.Time `gorm:"column:create_date" json:"createDate"`
	UpdateDate     *time.Time `gorm:"column:update_date" json:"updateDate"`
}

func (DevChangeHistory) TableName() string {
	return "dev_change_history"
}

type SysFile struct {
	FileID        string     `gorm:"column:file_id;type:char(36);primaryKey" json:"fileId"`
	URL           *string    `gorm:"column:url;type:varchar(512)" json:"url"`
	Name          *string    `gorm:"column:name;type:varchar(256)" json:"name"`
	Type          *string    `gorm:"column:type;type:varchar(64)" json:"type"`
	Size          int64      `gorm:"column:size;type:bigint;default:0" json:"size"`
	FileExt       *string    `gorm:"column:file_ext;type:varchar(16)" json:"fileExt"`
	OriginalName  *string    `gorm:"column:original_name;type:varchar(256)" json:"originalName"`
	Path          *string    `gorm:"column:path;type:varchar(512)" json:"path"`
	FullPath      *string    `gorm:"column:full_path;type:varchar(512)" json:"fullPath"`
	ThumbnailPath *string    `gorm:"column:thumbnail_path;type:varchar(512)" json:"thumbnailPath"`
	ThumbnailURL  *string    `gorm:"column:thumbnail_url;type:varchar(512)" json:"thumbnailUrl"`
	CreatorID     *string    `gorm:"column:creator_id;type:char(36)" json:"creatorId"`
	CreateDate    *time.Time `gorm:"column:create_date" json:"createDate"`
}

func (SysFile) TableName() string {
	return "sys_file"
}
