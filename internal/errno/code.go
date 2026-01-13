package errno

type Code struct {
	Code    int
	Message string
}

var (
	SuccessCode           = 0
	InternalServerErrCode = 500
	ParamErrCode          = 400

	AuthErrCode         = 1000
	AuthErrCodeNotFound = 1001
	AuthErrCodeWrongPwd = 1002
	AuthErrCodeToken    = 1003
	AuthErrCodeExpired  = 1004
	AuthErrCodeNoLogin  = 1005

	BookErrCode         = 2000
	BookErrCodeNotFound = 2001
	BookErrCodeExist    = 2002
	BookErrCodeInvalid  = 2003

	ChapterErrCode         = 3000
	ChapterErrCodeNotFound = 3001

	TrimErrCode           = 4000
	TrimErrCodeNotFound   = 4001
	TrimErrCodeInvalid    = 4002
	TrimErrCodeGenerating = 4003

	TaskErrCode         = 5000
	TaskErrCodeNotFound = 5001
	TaskErrCodeRunning  = 5002
	TaskErrCodeFailed   = 5003
)

var (
	ErrSuccess        = &Code{Code: SuccessCode, Message: "success"}
	ErrInternalServer = &Code{Code: InternalServerErrCode, Message: "服务器内部错误"}
	ErrParam          = &Code{Code: ParamErrCode, Message: "参数错误"}

	ErrAuthNotFound = &Code{Code: AuthErrCodeNotFound, Message: "用户不存在"}
	ErrAuthWrongPwd = &Code{Code: AuthErrCodeWrongPwd, Message: "密码错误"}
	ErrAuthToken    = &Code{Code: AuthErrCodeToken, Message: "无效的 Token"}
	ErrAuthExpired  = &Code{Code: AuthErrCodeExpired, Message: "Token 已过期"}
	ErrAuthNoLogin  = &Code{Code: AuthErrCodeNoLogin, Message: "未登录"}

	ErrBookNotFound = &Code{Code: BookErrCodeNotFound, Message: "书籍不存在"}
	ErrBookExist    = &Code{Code: BookErrCodeExist, Message: "书籍已存在"}
	ErrBookInvalid  = &Code{Code: BookErrCodeInvalid, Message: "无效的书籍"}

	ErrChapterNotFound = &Code{Code: ChapterErrCodeNotFound, Message: "章节不存在"}

	ErrTrimNotFound   = &Code{Code: TrimErrCodeNotFound, Message: "精简结果不存在"}
	ErrTrimInvalid    = &Code{Code: TrimErrCodeInvalid, Message: "无效的精简参数"}
	ErrTrimGenerating = &Code{Code: TrimErrCodeGenerating, Message: "精简进行中"}

	ErrTaskNotFound = &Code{Code: TaskErrCodeNotFound, Message: "任务不存在"}
	ErrTaskRunning  = &Code{Code: TaskErrCodeRunning, Message: "任务进行中"}
	ErrTaskFailed   = &Code{Code: TaskErrCodeFailed, Message: "任务失败"}
)

var codeMsgMap = map[int]string{
	SuccessCode:           "success",
	InternalServerErrCode: "服务器内部错误",
	ParamErrCode:          "参数错误",
}

func init() {
	// Register all defined errors to the map
	register := func(c *Code) {
		codeMsgMap[c.Code] = c.Message
	}
	register(ErrAuthNotFound)
	register(ErrAuthWrongPwd)
	register(ErrAuthToken)
	register(ErrAuthExpired)
	register(ErrAuthNoLogin)
	register(ErrBookNotFound)
	register(ErrBookExist)
	register(ErrBookInvalid)
	register(ErrChapterNotFound)
	register(ErrTrimNotFound)
	register(ErrTrimInvalid)
	register(ErrTrimGenerating)
	register(ErrTaskNotFound)
	register(ErrTaskRunning)
	register(ErrTaskFailed)
}

func GetMsg(code int) string {
	if msg, ok := codeMsgMap[code]; ok {
		return msg
	}
	return "Unknown Error"
}

func (c *Code) Error() string {
	return c.Message
}

func New(code int, message string) *Code {
	return &Code{Code: code, Message: message}
}

func IsErrCode(err error, code int) bool {
	if e, ok := err.(*Code); ok {
		return e.Code == code
	}
	return false
}
