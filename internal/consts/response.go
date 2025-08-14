package consts

const (
	ResponseCodeSuccess      = iota // 通用成功
	ResponseCodeParamError          // 参数错误
	ResponseCodeNoPermission        // 没有权限
	ResponseCodeError               // 通用错误
	ResponseCodeSystemError         // 系统错误

	ResponseCodeLoginFailed      // 登录失败
	ResponseCodeLoginExpired     // 登录过期
	ResponseCodeRegisterFailed   // 注册失败
	ResponseCodeUserDetailFailed // 用户详情失败
	ResponseCodeUserSaveFailed   // 用户保存失败

	ResponseCodeGetVideoSourceListFailed     // 获取视频源列表失败
	ResponseCodeGetVideoSourceDetailFailed   // 获取视频源详情失败
	ResponseCodeSaveVideoSourceFailed        // 保存视频源失败
	ResponseCodeDeleteVideoSourceFailed      // 删除视频源失败
	ResponseCodeCheckVideoSourceStatusFailed // 检查视频源状态失败
	ResponseCodeVideoSourceStatusUnavailable // 视频源状态不可用
)
