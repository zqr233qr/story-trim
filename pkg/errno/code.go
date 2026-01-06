package errno

const (
	SuccessCode = 0
	InternalServerErrCode = 5000
	ParamErrCode = 4000
	
	// 业务模块代码
	AuthErrCode = 1000
	UploadErrCode = 2000
	BookNotFoundCode = 2001
	ChapterNotFoundCode = 2002
	LLMErrCode = 3000
	TaskErrCode = 4001
)

var MsgFlags = map[int]string{
	SuccessCode:           "Success",
	InternalServerErrCode: "Internal Server Error",
	ParamErrCode:          "Parameter Error",
	AuthErrCode:           "Authentication Failed",
	UploadErrCode:         "Upload Failed",
	BookNotFoundCode:      "Book Not Found",
	ChapterNotFoundCode:   "Chapter Not Found",
	LLMErrCode:            "AI Engine Error",
	TaskErrCode:           "Task Execution Error",
}

func GetMsg(code int) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}
	return MsgFlags[InternalServerErrCode]
}
