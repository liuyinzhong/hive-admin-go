package models

// externalFeedbackType 工单类型枚举
const (
	ExternalFeedbackTypeStory = "story"
	ExternalFeedbackTypeBug   = "bug"
)

// CreateStoryFeedbackRequest 外部反馈工单提交请求体。
// 不绑定登录态，由公开接口 /api/public/feedback 接收；
// type 决定写入 dev_story 还是 dev_bug，source 固定写 10（外部反馈）。
type CreateStoryFeedbackRequest struct {
	Type      string   `json:"type" binding:"required" example:"story"` // 工单类型: story=需求, bug=缺陷
	Title     string   `json:"title" binding:"required" example:"反馈标题"` // 工单标题，对应 storyTitle/bugTitle
	RichText  *string  `json:"richText" example:"<p>反馈描述</p>"`          // 工单描述(富文本)，对应 storyRichText/bugRichText
	FileIDs   []string `json:"fileIds" example:"[\"UUID\"]"`            // 附件id数组，由 /api/public/upload 公开上传产生
	UserAgent *string  `json:"userAgent" example:"Mozilla/5.0"`         // 提交方浏览器UA，仅 type=bug 时写入 bug_ua
}

// CreateStoryFeedbackResponse 外部反馈提交成功响应，返回工单编号供提交方作为凭据。
type CreateStoryFeedbackResponse struct {
	Num  int    `json:"num" example:"1"`      // 工单编号，对应 storyNum/bugNum
	Type string `json:"type" example:"story"` // 工单类型，回显请求 type
}
