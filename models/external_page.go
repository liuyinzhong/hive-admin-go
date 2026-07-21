package models

type ExternalPageListRequest struct {
	Page     int    `form:"page" example:"1"`
	PageSize int    `form:"pageSize" example:"20"`
	Title    string `form:"title" example:"外部演示页面"`
	Name     string `form:"name" example:"externalDemo"`
	Path     string `form:"path" example:"/external/demo"`
	Status   *int   `form:"status" example:"1"`
}

type CreateExternalPageRequest struct {
	Title  string `json:"title" binding:"required" example:"外部演示页面"`
	Name   string `json:"name" binding:"required" example:"externalDemo"`
	Path   string `json:"path" binding:"required" example:"/external/demo"`
	Status *int   `json:"status" example:"1"`
}

type UpdateExternalPageRequest struct {
	Title string `json:"title" binding:"required" example:"外部演示页面"`
	Path  string `json:"path" binding:"required" example:"/external/demo"`
}

type UpdateExternalPageStatusRequest struct {
	Status int `json:"status" example:"1"`
}

type DeleteExternalPagesRequest struct {
	IDs []string `json:"ids" binding:"required,min=1" example:"[\"UUID\"]"`
}

type ExternalPageResponse struct {
	ID          string  `json:"id" example:"UUID"`
	Title       string  `json:"title" example:"外部演示页面"`
	Name        string  `json:"name" example:"externalDemo"`
	Path        string  `json:"path" example:"/external/demo"`
	Status      int     `json:"status" example:"1"`
	CreatorID   *string `json:"creatorId" example:"UUID"`
	CreatorName *string `json:"creatorName" example:"管理员"`
	CreateDate  *string `json:"createDate" example:"2026-07-18 12:00:00"`
	UpdateDate  *string `json:"updateDate" example:"2026-07-18 12:00:00"`
}

type PublicExternalPageResponse struct {
	Name string `json:"name" example:"externalDemo"`
	Path string `json:"path" example:"/external/demo"`
}
