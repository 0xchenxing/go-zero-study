package xerr

var message map[uint32]string

func init() {
	message = make(map[uint32]string)
	message[OK] = "SUCCESS"
	message[SERVER_COMMON_ERROR] = "服务器开小差了，请稍后再试"
	message[REQUEST_PARAM_ERROR] = "参数错误"
	message[TOKEN_EXPIRE_ERROR] = "token失效，请重新登录"
	message[TOKEN_GENERATE_ERROR] = "生成token失败"
	message[DB_ERROR] = "数据库繁忙，请稍后再试"
	message[USER_NOT_FOUND] = "用户不存在"
	message[USER_ALREADY_EXISTS] = "用户已存在"
	message[PASSWORD_ERROR] = "密码错误"
}

// MapErrMsg 根据错误码返回对应的错误信息
func MapErrMsg(errcode uint32) string {
	if msg, ok := message[errcode]; ok {
		return msg
	}
	return "服务器开小差了，请稍后再试"
}

// IsCodeErr 判断错误码是否为已定义的自定义错误码
func IsCodeErr(errcode uint32) bool {
	_, ok := message[errcode]
	return ok
}
