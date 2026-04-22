package service

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmsm/x509"
)

type ZSPayService struct {
	client        *ZSHttpClient
	MerID         string
	AppID         string
	AppSecret     string
	PrivateKeyPEM string
	PublicKeyPEM  string
	BaseURL       string
	privateKey    *sm2.PrivateKey
	publicKey     *sm2.PublicKey
}

func NewZSPayService() *ZSPayService {
	s := &ZSPayService{
		client:        NewZSHttpClient(30 * time.Second),
		MerID:         operation_setting.GetZSPayMerID(),
		AppID:         operation_setting.GetZSPayAppID(),
		AppSecret:     operation_setting.GetZSPayAppSecret(),
		PrivateKeyPEM: operation_setting.GetZSPayPrivateKey(),
		PublicKeyPEM:  operation_setting.GetZSPayPublicKey(),
		BaseURL:       operation_setting.GetZSPayBaseURL(),
	}

	if s.PrivateKeyPEM != "" {
		privateKey, err := s.parseSM2PrivateKey(s.PrivateKeyPEM)
		if err != nil {
			common.SysLog(fmt.Sprintf("[ZS Pay] 解析SM2私钥失败: %v", err))
		} else {
			s.privateKey = privateKey
			common.SysLog("[ZS Pay] SM2私钥解析成功")
		}
	}

	if s.PublicKeyPEM != "" {
		publicKey, err := s.parseSM2PublicKey(s.PublicKeyPEM)
		if err != nil {
			common.SysLog(fmt.Sprintf("[ZS Pay] 解析SM2公钥失败: %v", err))
		} else {
			s.publicKey = publicKey
			common.SysLog("[ZS Pay] SM2公钥解析成功")
		}
	}

	return s
}

func (s *ZSPayService) parseSM2PrivateKey(keyStr string) (*sm2.PrivateKey, error) {
	keyStr = strings.TrimSpace(keyStr)

	if strings.Contains(keyStr, "-----BEGIN") {
		return x509.ReadPrivateKeyFromPem([]byte(keyStr), nil)
	}

	keyBytes, err := hex.DecodeString(keyStr)
	if err != nil {
		keyBytes, err = base64.StdEncoding.DecodeString(keyStr)
		if err != nil {
			return nil, fmt.Errorf("无法解析私钥格式: %w", err)
		}
	}

	if len(keyBytes) == 32 {
		return s.buildSM2PrivateKeyFromD(keyBytes)
	}

	privKey, err := x509.ParsePKCS8PrivateKey(keyBytes, nil)
	if err == nil {
		return privKey, nil
	}

	return x509.ParseSm2PrivateKey(keyBytes)
}

func (s *ZSPayService) buildSM2PrivateKeyFromD(dBytes []byte) (*sm2.PrivateKey, error) {
	curve := sm2.P256Sm2()
	d := new(big.Int).SetBytes(dBytes)
	x, y := curve.ScalarBaseMult(dBytes)

	return &sm2.PrivateKey{
		PublicKey: sm2.PublicKey{
			Curve: curve,
			X:     x,
			Y:     y,
		},
		D: d,
	}, nil
}

func (s *ZSPayService) parseSM2PublicKey(keyStr string) (*sm2.PublicKey, error) {
	keyStr = strings.TrimSpace(keyStr)

	if strings.Contains(keyStr, "-----BEGIN") {
		return x509.ReadPublicKeyFromPem([]byte(keyStr))
	}

	keyBytes, err := hex.DecodeString(keyStr)
	if err != nil {
		keyBytes, err = base64.StdEncoding.DecodeString(keyStr)
		if err != nil {
			return nil, fmt.Errorf("无法解析公钥格式: %w", err)
		}
	}

	return x509.ParseSm2PublicKey(keyBytes)
}

type ZSBaseRequest struct {
	Version    string      `json:"version"`
	Encoding   string      `json:"encoding"`
	SignMethod string      `json:"signMethod"`
	Sign       string      `json:"sign"`
	BizContent interface{} `json:"biz_content"`
}

type ZSBaseResponse struct {
	Version    string `json:"version"`
	Encoding   string `json:"encoding"`
	SignMethod string `json:"signMethod"`
	Sign       string `json:"sign"`
	ReturnCode string `json:"returnCode"`
	RespCode   string `json:"respCode"`
	RespMsg    string `json:"respMsg"`
	ErrCode    string `json:"errCode"`
	BizContent string `json:"biz_content"`
}

type ZSQRCodeApplyReq struct {
	MerID        string `json:"merId"`
	OrderID      string `json:"orderId"`
	NotifyURL    string `json:"notifyUrl"`
	TxnAmt       string `json:"txnAmt"`
	Body         string `json:"body,omitempty"`
	PayValidTime string `json:"payValidTime,omitempty"`
	UserID       string `json:"userId,omitempty"`
	MchReserved  string `json:"mchReserved,omitempty"`
}

type ZSQRCodeApplyResp struct {
	MerID      string `json:"merId"`
	OrderID    string `json:"orderId"`
	CmbOrderID string `json:"cmbOrderId"`
	QRCodeURL  string `json:"qrCodeUrl"`
	QRCode     string `json:"qrCode"`
	QRCodeData string `json:"qrCodeData"`
}

type ZSOrderQueryReq struct {
	MerID      string `json:"merId"`
	OrderID    string `json:"orderId,omitempty"`
	CmbOrderID string `json:"cmbOrderId,omitempty"`
	OutOrderID string `json:"outOrderId,omitempty"`
}

type ZSOrderQueryResp struct {
	MerID               string `json:"merId"`
	OrderID             string `json:"orderId"`
	CmbOrderID          string `json:"cmbOrderId"`
	TxnAmt              string `json:"txnAmt"`
	DscAmt              string `json:"dscAmt"`
	PayType             string `json:"payType"`
	OpenID              string `json:"openId,omitempty"`
	PayBank             string `json:"payBank,omitempty"`
	ThirdOrderID        string `json:"thirdOrderId,omitempty"`
	TradeState          string `json:"tradeState"`
	TxnTime             string `json:"txnTime"`
	EndDate             string `json:"endDate,omitempty"`
	EndTime             string `json:"endTime,omitempty"`
	MchReserved         string `json:"mchReserved,omitempty"`
	PromotionDetail     string `json:"promotionDetail,omitempty"`
	EcnyPromotionDetail string `json:"ecnyPromotionDetail,omitempty"`
}

type ZSRefundReq struct {
	MerID          string `json:"merId"`
	OrderID        string `json:"orderId"`
	UserID         string `json:"userId"`
	OrigOrderID    string `json:"origOrderId,omitempty"`
	OrigCmbOrderID string `json:"origCmbOrderId,omitempty"`
	NotifyURL      string `json:"notifyUrl,omitempty"`
	TxnAmt         string `json:"txnAmt"`
	RefundAmt      string `json:"refundAmt"`
	RefundReason   string `json:"refundReason,omitempty"`
	MchReserved    string `json:"mchReserved,omitempty"`
}

type ZSRefundResp struct {
	MerID                string `json:"merId"`
	OrderID              string `json:"orderId"`
	CmbOrderID           string `json:"cmbOrderId"`
	RefundAmt            string `json:"refundAmt"`
	RefundDscAmt         string `json:"refundDscAmt"`
	RefundState          string `json:"refundState"`
	TxnTime              string `json:"txnTime"`
	EndDate              string `json:"endDate"`
	EndTime              string `json:"endTime"`
	RefundDetailItemList string `json:"refundDetailItemList,omitempty"`
}

type ZSCloseOrderReq struct {
	MerID          string `json:"merId"`
	OrigOrderID    string `json:"origOrderId,omitempty"`
	OrigCmbOrderID string `json:"origCmbOrderId,omitempty"`
	UserID         string `json:"userId"`
}

type ZSCloseOrderResp struct {
	MerID          string `json:"merId"`
	OrigOrderID    string `json:"origOrderId,omitempty"`
	OrigCmbOrderID string `json:"origCmbOrderId,omitempty"`
	CloseState     string `json:"closeState,omitempty"`
	TxnTime        string `json:"txnTime,omitempty"`
	ErrCode        string `json:"errCode,omitempty"`
	RespMsg        string `json:"respMsg,omitempty"`
}

type ZSPaymentNotifyData struct {
	Version      string `json:"version"`
	Encoding     string `json:"encoding"`
	SignMethod   string `json:"signMethod"`
	Sign         string `json:"sign"`
	MerID        string `json:"merId"`
	OrderID      string `json:"orderId"`
	CmbOrderID   string `json:"cmbOrderId"`
	UserID       string `json:"userId,omitempty"`
	TxnAmt       string `json:"txnAmt"`
	DscAmt       string `json:"dscAmt"`
	PayType      string `json:"payType"`
	OpenID       string `json:"openId,omitempty"`
	PayBank      string `json:"payBank,omitempty"`
	ThirdOrderID string `json:"thirdOrderId,omitempty"`
	TxnTime      string `json:"txnTime"`
	EndDate      string `json:"endDate,omitempty"`
	EndTime      string `json:"endTime,omitempty"`
	MchReserved  string `json:"mchReserved,omitempty"`
}

type ZSQRCodeResult struct {
	QRCodeURL  string
	CmbOrderID string
}

func (s *ZSPayService) QRCodeApply(orderNo string, amount float64, notifyURL string) (*ZSQRCodeResult, error) {
	url := s.BaseURL + "/polypay/v1.0/mchorders/qrcodeapply"

	common.SysLog(fmt.Sprintf("[ZS Pay] 申请二维码: orderNo=%s, amount=%.2f, notifyURL=%s, baseURL=%s", orderNo, amount, notifyURL, s.BaseURL))

	req := &ZSQRCodeApplyReq{
		MerID:        s.MerID,
		OrderID:      orderNo,
		NotifyURL:    notifyURL,
		TxnAmt:       s.amountToFen(amount),
		Body:         "账户充值",
		PayValidTime: operation_setting.GetZSPayPayValidTime(),
	}

	common.SysLog(fmt.Sprintf("[ZS Pay] 请求参数: %+v", req))

	resp := &ZSQRCodeApplyResp{}
	err := s.doRequest(url, req, resp)
	if err != nil {
		common.SysLog(fmt.Sprintf("[ZS Pay] 申请二维码失败: %v", err))
		return nil, fmt.Errorf("申请收款二维码失败: %w", err)
	}

	common.SysLog(fmt.Sprintf("[ZS Pay] 招行返回: cmbOrderId=%s, qrCodeURL=%s", resp.CmbOrderID, resp.QRCodeURL))

	qrCodeURL := s.processQRCodeData(resp.QRCodeURL)

	return &ZSQRCodeResult{
		QRCodeURL:  qrCodeURL,
		CmbOrderID: resp.CmbOrderID,
	}, nil
}

func (s *ZSPayService) OrderQuery(transactionNo string) (*ZSOrderQueryResp, error) {
	url := s.BaseURL + "/polypay/v1.0/mchorders/orderquery"

	req := &ZSOrderQueryReq{
		MerID:   s.MerID,
		OrderID: transactionNo,
	}

	resp := &ZSOrderQueryResp{}
	err := s.doRequest(url, req, resp)
	if err != nil {
		return nil, fmt.Errorf("查询支付状态失败: %w", err)
	}

	return resp, nil
}

func (s *ZSPayService) Refund(refundNo string, originalTransactionNo string, origCmbOrderID string, txnAmt float64, refundAmt float64, refundReason string) (*ZSRefundResp, error) {
	url := s.BaseURL + "/polypay/v1.0/mchorders/refund"

	req := &ZSRefundReq{
		MerID:          s.MerID,
		OrderID:        refundNo,
		UserID:         s.MerID,
		OrigOrderID:    originalTransactionNo,
		OrigCmbOrderID: origCmbOrderID,
		TxnAmt:         s.amountToFen(txnAmt),
		RefundAmt:      s.amountToFen(refundAmt),
		RefundReason:   refundReason,
	}

	resp := &ZSRefundResp{}
	err := s.doRequest(url, req, resp)
	if err != nil {
		return nil, fmt.Errorf("退款申请失败: %w", err)
	}

	return resp, nil
}

func (s *ZSPayService) CloseOrder(origOrderID string, origCmbOrderID string) (*ZSCloseOrderResp, error) {
	url := s.BaseURL + "/polypay/v1.0/mchorders/close"

	req := &ZSCloseOrderReq{
		MerID:          s.MerID,
		OrigOrderID:    origOrderID,
		OrigCmbOrderID: origCmbOrderID,
		UserID:         s.MerID,
	}

	resp := &ZSCloseOrderResp{}
	err := s.doRequest(url, req, resp)
	if err != nil {
		return nil, fmt.Errorf("关闭订单失败: %w", err)
	}

	return resp, nil
}

func (s *ZSPayService) doRequest(reqURL string, reqData interface{}, respData interface{}) error {
	bizContent, err := json.Marshal(reqData)
	if err != nil {
		return fmt.Errorf("序列化请求数据失败: %w", err)
	}

	common.SysLog(fmt.Sprintf("[ZS Pay] 业务请求数据: %s", string(bizContent)))

	sign, err := s.sign(string(bizContent))
	if err != nil {
		common.SysLog(fmt.Sprintf("[ZS Pay] 签名生成失败: %v", err))
		return fmt.Errorf("生成签名失败: %w", err)
	}

	baseReq := ZSBaseRequest{
		Version:    "0.0.1",
		Encoding:   "UTF-8",
		SignMethod: "02",
		Sign:       sign,
		BizContent: json.RawMessage(bizContent),
	}

	reqBody, err := json.Marshal(baseReq)
	if err != nil {
		return fmt.Errorf("序列化基础请求失败: %w", err)
	}

	common.SysLog(fmt.Sprintf("[ZS Pay] 完整请求体: %s", string(reqBody)))
	common.SysLog(fmt.Sprintf("[ZS Pay] 请求URL: %s", reqURL))

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	body, err := s.client.Post(reqURL, reqBody, headers)
	if err != nil {
		common.SysLog(fmt.Sprintf("[ZS Pay] HTTP请求失败: %v", err))
		return fmt.Errorf("发送请求失败: %w", err)
	}

	common.SysLog(fmt.Sprintf("[ZS Pay] 原始响应: %s", string(body)))

	var baseResp ZSBaseResponse
	if err := json.Unmarshal(body, &baseResp); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	common.SysLog(fmt.Sprintf("[ZS Pay] 基础响应: returnCode=%s, respCode=%s, errCode=%s, respMsg=%s", baseResp.ReturnCode, baseResp.RespCode, baseResp.ErrCode, baseResp.RespMsg))

	if baseResp.ReturnCode != "SUCCESS" || baseResp.RespCode != "SUCCESS" {
		err := fmt.Errorf("招行返回错误: %s - %s", baseResp.ErrCode, baseResp.RespMsg)
		common.SysLog(fmt.Sprintf("[ZS Pay] %v", err))
		return err
	}

	if baseResp.BizContent == "" {
		err := fmt.Errorf("招行返回业务数据为空")
		common.SysLog(fmt.Sprintf("[ZS Pay] %v", err))
		return err
	}

	common.SysLog(fmt.Sprintf("[ZS Pay] 业务响应数据: %s", baseResp.BizContent))

	if err := json.Unmarshal([]byte(baseResp.BizContent), respData); err != nil {
		return fmt.Errorf("解析业务响应失败: %w", err)
	}

	return nil
}

func (s *ZSPayService) sign(data string) (string, error) {
	if s.privateKey == nil {
		return "", fmt.Errorf("SM2私钥未初始化")
	}

	signature, err := s.privateKey.Sign(rand.Reader, []byte(data), nil)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

func (s *ZSPayService) verifySign(data string, signStr string) bool {
	if s.publicKey == nil {
		return false
	}

	signBytes, err := base64.StdEncoding.DecodeString(signStr)
	if err != nil {
		return false
	}

	return s.publicKey.Verify([]byte(data), signBytes)
}

func (s *ZSPayService) amountToFen(amount float64) string {
	return fmt.Sprintf("%.0f", amount*100)
}

func (s *ZSPayService) fenToAmount(fen string) float64 {
	f, _ := strconv.ParseFloat(fen, 64)
	return f / 100
}

func (s *ZSPayService) processQRCodeData(qrCode string) string {
	if qrCode == "" {
		return qrCode
	}

	baseURL := operation_setting.GetZSPayBaseURL()
	if baseURL == "" || !regexp.MustCompile(`api\.cmburl\.cn:8065`).MatchString(baseURL) {
		return qrCode
	}

	re := regexp.MustCompile(`https?://[^/]+`)
	return re.ReplaceAllString(qrCode, "http://payment-uat.cs.cmburl.cn")
}

func (s *ZSPayService) GetNotifyURL() string {
	notifyPath := operation_setting.GetZSPayNotifyPath()
	if notifyPath == "" {
		notifyPath = "/api/user/zs_pay/notify"
	}
	serverAddress := operation_setting.PayAddress
	if serverAddress == "" {
		serverAddress = "http://localhost:3000"
	}
	return strings.TrimSuffix(serverAddress, "/") + notifyPath
}

func IsZSPayEnabled() bool {
	return operation_setting.IsZSPayEnabled() &&
		operation_setting.GetZSPayMerID() != "" &&
		operation_setting.GetZSPayAppID() != "" &&
		operation_setting.GetZSPayAppSecret() != "" &&
		operation_setting.GetZSPayPrivateKey() != "" &&
		operation_setting.GetZSPayPublicKey() != "" &&
		operation_setting.GetZSPayBaseURL() != ""
}

func GetZSPayService() *ZSPayService {
	if !IsZSPayEnabled() {
		return nil
	}
	return NewZSPayService()
}

type ZSHttpClient struct {
	timeout time.Duration
}

func NewZSHttpClient(timeout time.Duration) *ZSHttpClient {
	return &ZSHttpClient{timeout: timeout}
}

func (c *ZSHttpClient) Post(reqURL string, reqBody []byte, headers map[string]string) ([]byte, error) {
	// 创建 HTTP 客户端，跳过 TLS 验证（因为招行测试环境使用自签名证书）
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Timeout:   c.timeout,
		Transport: transport,
	}

	req, err := http.NewRequest("POST", reqURL, strings.NewReader(string(reqBody)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	common.SysLog(fmt.Sprintf("[ZS Pay] 发送请求到: %s", reqURL))
	common.SysLog(fmt.Sprintf("[ZS Pay] 请求头: %+v", req.Header))

	resp, err := client.Do(req)
	if err != nil {
		common.SysLog(fmt.Sprintf("[ZS Pay] HTTP 请求失败: %v", err))
		return nil, err
	}
	defer resp.Body.Close()

	common.SysLog(fmt.Sprintf("[ZS Pay] HTTP 状态码: %d", resp.StatusCode))

	buf := new(strings.Builder)
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return nil, err
	}

	respBody := buf.String()
	common.SysLog(fmt.Sprintf("[ZS Pay] 响应长度: %d 字节", len(respBody)))
	common.SysLog(fmt.Sprintf("[ZS Pay] 响应内容: %s", respBody))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("HTTP 状态码异常: %d, 响应: %s", resp.StatusCode, respBody)
	}

	return []byte(respBody), nil
}

func GenerateRandomString(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)[:length]
}
