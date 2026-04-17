package operation_setting

import (
	"os"
	"strconv"

	"github.com/QuantumNous/new-api/setting/config"
)

// ZSPaymentSetting 招商银行聚合支付配置
type ZSPaymentSetting struct {
	Enabled      bool   `json:"enabled"`
	MerID        string `json:"mer_id"`
	AppID        string `json:"app_id"`
	AppSecret    string `json:"app_secret"`
	PrivateKey   string `json:"private_key"`
	PublicKey    string `json:"public_key"`
	BaseURL      string `json:"base_url"`
	NotifyPath   string `json:"notify_path"`
	PayValidTime string `json:"pay_valid_time"`
}

// 默认配置
var zsPaymentSetting = ZSPaymentSetting{
	Enabled:      false, // 默认不启用，需要管理员手动配置
	MerID:        "",
	AppID:        "",
	AppSecret:    "",
	PrivateKey:   "",
	PublicKey:    "",
	BaseURL:      "",
	NotifyPath:   "/api/user/zs_pay/notify",
	PayValidTime: "1800",
}

func init() {
	// 从环境变量加载配置（此时 .env 可能还未加载）
	loadZSPayFromEnv()

	// 注册到全局配置管理器（仅用于 Enabled 开关，其他配置从环境变量读取）
	config.GlobalConfig.Register("zs_payment", &zsPaymentSetting)
}

// loadZSPayFromEnv 从环境变量加载招行支付配置
func loadZSPayFromEnv() {
	// 商户号
	if merID := os.Getenv("ZS_PAYMENT_MER_ID"); merID != "" {
		zsPaymentSetting.MerID = merID
	}

	// AppID
	if appID := os.Getenv("ZS_PAYMENT_APP_ID"); appID != "" {
		zsPaymentSetting.AppID = appID
	}

	// AppSecret
	if appSecret := os.Getenv("ZS_PAYMENT_APP_SECRET"); appSecret != "" {
		zsPaymentSetting.AppSecret = appSecret
	}

	// 私钥
	if privateKey := os.Getenv("ZS_PAYMENT_PRIVATE_KEY"); privateKey != "" {
		zsPaymentSetting.PrivateKey = privateKey
	}

	// 公钥
	if publicKey := os.Getenv("ZS_PAYMENT_PUBLIC_KEY"); publicKey != "" {
		zsPaymentSetting.PublicKey = publicKey
	}

	// 基础 URL
	if baseURL := os.Getenv("ZS_PAYMENT_BASE_URL"); baseURL != "" {
		zsPaymentSetting.BaseURL = baseURL
	}

	// 回调路径
	if notifyPath := os.Getenv("ZS_PAYMENT_NOTIFY_PATH"); notifyPath != "" {
		zsPaymentSetting.NotifyPath = notifyPath
	}

	// 支付有效期
	if payValidTime := os.Getenv("ZS_PAYMENT_PAY_VALID_TIME"); payValidTime != "" {
		zsPaymentSetting.PayValidTime = payValidTime
	}

	// 如果所有必要配置都已设置，则自动启用
	if zsPaymentSetting.MerID != "" &&
		zsPaymentSetting.AppID != "" &&
		zsPaymentSetting.AppSecret != "" &&
		zsPaymentSetting.PrivateKey != "" &&
		zsPaymentSetting.PublicKey != "" &&
		zsPaymentSetting.BaseURL != "" {
		zsPaymentSetting.Enabled = true
	}
}

// ReloadZSPayFromEnv 重新从环境变量加载招行支付配置
// 用于在 .env 文件加载后重新读取配置
func ReloadZSPayFromEnv() {
	// 先重置为默认值
	zsPaymentSetting = ZSPaymentSetting{
		Enabled:      false,
		MerID:        "",
		AppID:        "",
		AppSecret:    "",
		PrivateKey:   "",
		PublicKey:    "",
		BaseURL:      "",
		NotifyPath:   "/api/user/zs_pay/notify",
		PayValidTime: "1800",
	}
	// 重新加载
	loadZSPayFromEnv()
}

// GetZSPaymentSetting 获取招商银行聚合支付配置
func GetZSPaymentSetting() *ZSPaymentSetting {
	return &zsPaymentSetting
}

// IsZSPayEnabled 检查招商银行聚合支付是否启用
func IsZSPayEnabled() bool {
	return zsPaymentSetting.Enabled
}

// GetZSPayMerID 获取商户号
func GetZSPayMerID() string {
	return zsPaymentSetting.MerID
}

// GetZSPayAppID 获取AppID
func GetZSPayAppID() string {
	return zsPaymentSetting.AppID
}

// GetZSPayAppSecret 获取AppSecret
func GetZSPayAppSecret() string {
	return zsPaymentSetting.AppSecret
}

// GetZSPayPrivateKey 获取私钥
func GetZSPayPrivateKey() string {
	return zsPaymentSetting.PrivateKey
}

// GetZSPayPublicKey 获取公钥
func GetZSPayPublicKey() string {
	return zsPaymentSetting.PublicKey
}

// GetZSPayBaseURL 获取基础URL
func GetZSPayBaseURL() string {
	return zsPaymentSetting.BaseURL
}

// GetZSPayNotifyPath 获取回调路径
func GetZSPayNotifyPath() string {
	return zsPaymentSetting.NotifyPath
}

// GetZSPayPayValidTime 获取支付有效期
func GetZSPayPayValidTime() string {
	return zsPaymentSetting.PayValidTime
}

// UpdateZSPayEnabled 更新启用状态（供前端开关使用）
func UpdateZSPayEnabled(enabled bool) {
	zsPaymentSetting.Enabled = enabled
}

// UpdateZSPayNotifyPath 更新回调路径（供前端配置使用）
func UpdateZSPayNotifyPath(path string) {
	zsPaymentSetting.NotifyPath = path
}

// UpdateZSPayPayValidTime 更新支付有效期（供前端配置使用）
func UpdateZSPayPayValidTime(time string) {
	zsPaymentSetting.PayValidTime = time
}

// ParseBool 安全地解析布尔值
func ParseBool(s string) bool {
	b, _ := strconv.ParseBool(s)
	return b
}

// ZSEnvOption 环境变量配置项（避免循环导入）
type ZSEnvOption struct {
	Key   string
	Value string
}

// GetZSPayEnvOptions 获取环境变量配置，供前端显示用
func GetZSPayEnvOptions() []ZSEnvOption {
	return []ZSEnvOption{
		{Key: "zs_payment.MerID", Value: zsPaymentSetting.MerID},
		{Key: "zs_payment.AppID", Value: zsPaymentSetting.AppID},
		{Key: "zs_payment.AppSecret", Value: zsPaymentSetting.AppSecret},
		{Key: "zs_payment.PrivateKey", Value: zsPaymentSetting.PrivateKey},
		{Key: "zs_payment.PublicKey", Value: zsPaymentSetting.PublicKey},
		{Key: "zs_payment.BaseURL", Value: zsPaymentSetting.BaseURL},
	}
}
