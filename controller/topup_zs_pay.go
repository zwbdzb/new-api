package controller

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type ZSPayRequest struct {
	Amount        int64  `json:"amount"`
	PaymentMethod string `json:"payment_method"`
}

func GetZSPayInfo(c *gin.Context) {
	enableZS := service.IsZSPayEnabled()

	data := gin.H{
		"enable_zs_pay_topup": enableZS,
	}
	common.ApiSuccess(c, data)
}

func RequestZSPay(c *gin.Context) {
	var req ZSPayRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "参数错误"})
		return
	}

	minTopup := getMinTopup()
	if req.Amount < minTopup {
		c.JSON(200, gin.H{"message": "error", "data": fmt.Sprintf("充值数量不能小于 %d", minTopup)})
		return
	}

	id := c.GetInt("id")
	group, err := model.GetUserGroup(id, true)
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "获取用户分组失败"})
		return
	}

	payMoney := getPayMoney(req.Amount, group)
	if payMoney < 0.01 {
		c.JSON(200, gin.H{"message": "error", "data": "充值金额过低"})
		return
	}

	if !operation_setting.ContainsPayMethod(req.PaymentMethod) {
		c.JSON(200, gin.H{"message": "error", "data": "支付方式不存在"})
		return
	}

	zsService := service.GetZSPayService()
	if zsService == nil {
		c.JSON(200, gin.H{"message": "error", "data": "招商银行聚合支付未启用"})
		return
	}

	notifyURL := zsService.GetNotifyURL()
	tradeNo := fmt.Sprintf("%s%d", common.GetRandomString(6), time.Now().Unix())
	tradeNo = fmt.Sprintf("USR%dZS%s", id, tradeNo)

	qrResult, err := zsService.QRCodeApply(tradeNo, payMoney, notifyURL)
	if err != nil {
		log.Printf("招商银行聚合支付申请二维码失败: %v", err)
		c.JSON(200, gin.H{"message": "error", "data": "申请支付二维码失败"})
		return
	}

	amount := req.Amount
	if operation_setting.GetQuotaDisplayType() == operation_setting.QuotaDisplayTypeTokens {
		dAmount := decimal.NewFromInt(int64(amount))
		dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
		amount = dAmount.Div(dQuotaPerUnit).IntPart()
	}

	topUp := &model.TopUp{
		UserId:        id,
		Amount:        amount,
		Money:         payMoney,
		TradeNo:       tradeNo,
		PaymentMethod: req.PaymentMethod,
		CreateTime:    time.Now().Unix(),
		Status:        "pending",
	}
	err = topUp.Insert()
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "创建订单失败"})
		return
	}

	c.JSON(200, gin.H{
		"message":      "success",
		"data":         qrResult.QRCodeURL,
		"cmb_order_id": qrResult.CmbOrderID,
		"trade_no":     tradeNo,
	})
}

func ZSPayNotify(c *gin.Context) {
	var notifyData service.ZSPaymentNotifyData

	if c.Request.Method == "POST" {
		if err := c.Request.ParseForm(); err != nil {
			log.Println("招商银行聚合支付回调POST解析失败:", err)
			c.Writer.Write([]byte("fail"))
			return
		}

		for key, values := range c.Request.PostForm {
			if len(values) > 0 {
				switch key {
				case "version":
					notifyData.Version = values[0]
				case "encoding":
					notifyData.Encoding = values[0]
				case "signMethod":
					notifyData.SignMethod = values[0]
				case "sign":
					notifyData.Sign = values[0]
				case "merId":
					notifyData.MerID = values[0]
				case "orderId":
					notifyData.OrderID = values[0]
				case "cmbOrderId":
					notifyData.CmbOrderID = values[0]
				case "userId":
					notifyData.UserID = values[0]
				case "txnAmt":
					notifyData.TxnAmt = values[0]
				case "dscAmt":
					notifyData.DscAmt = values[0]
				case "payType":
					notifyData.PayType = values[0]
				case "openId":
					notifyData.OpenID = values[0]
				case "payBank":
					notifyData.PayBank = values[0]
				case "thirdOrderId":
					notifyData.ThirdOrderID = values[0]
				case "txnTime":
					notifyData.TxnTime = values[0]
				case "endDate":
					notifyData.EndDate = values[0]
				case "endTime":
					notifyData.EndTime = values[0]
				case "mchReserved":
					notifyData.MchReserved = values[0]
				}
			}
		}
	} else {
		for key, values := range c.Request.URL.Query() {
			if len(values) > 0 {
				switch key {
				case "version":
					notifyData.Version = values[0]
				case "encoding":
					notifyData.Encoding = values[0]
				case "signMethod":
					notifyData.SignMethod = values[0]
				case "sign":
					notifyData.Sign = values[0]
				case "merId":
					notifyData.MerID = values[0]
				case "orderId":
					notifyData.OrderID = values[0]
				case "cmbOrderId":
					notifyData.CmbOrderID = values[0]
				case "userId":
					notifyData.UserID = values[0]
				case "txnAmt":
					notifyData.TxnAmt = values[0]
				case "dscAmt":
					notifyData.DscAmt = values[0]
				case "payType":
					notifyData.PayType = values[0]
				case "openId":
					notifyData.OpenID = values[0]
				case "payBank":
					notifyData.PayBank = values[0]
				case "thirdOrderId":
					notifyData.ThirdOrderID = values[0]
				case "txnTime":
					notifyData.TxnTime = values[0]
				case "endDate":
					notifyData.EndDate = values[0]
				case "endTime":
					notifyData.EndTime = values[0]
				case "mchReserved":
					notifyData.MchReserved = values[0]
				}
			}
		}
	}

	log.Printf("招商银行聚合支付回调: %+v", notifyData)

	if notifyData.OrderID == "" {
		log.Println("招商银行聚合支付回调订单号为空")
		c.Writer.Write([]byte("fail"))
		return
	}

	zsService := service.GetZSPayService()
	if zsService == nil {
		log.Println("招商银行聚合支付服务未初始化")
		c.Writer.Write([]byte("fail"))
		return
	}

	orderNo := notifyData.OrderID

	LockOrder(orderNo)
	defer UnlockOrder(orderNo)

	topUp := model.GetTopUpByTradeNo(orderNo)
	if topUp == nil {
		log.Printf("招商银行聚合支付回调未找到订单: %s", orderNo)
		c.Writer.Write([]byte("fail"))
		return
	}

	if topUp.Status == "pending" {
		topUp.Status = "success"
		if err := topUp.Update(); err != nil {
			log.Printf("招商银行聚合支付回调更新订单失败: %v", topUp)
			c.Writer.Write([]byte("fail"))
			return
		}

		dAmount := decimal.NewFromInt(int64(topUp.Amount))
		dQuotaPerUnit := decimal.NewFromFloat(common.QuotaPerUnit)
		quotaToAdd := int(dAmount.Mul(dQuotaPerUnit).IntPart())

		if err := model.IncreaseUserQuota(topUp.UserId, quotaToAdd, true); err != nil {
			log.Printf("招商银行聚合支付回调更新用户失败: %v", topUp)
			c.Writer.Write([]byte("fail"))
			return
		}

		log.Printf("招商银行聚合支付回调成功: %s, 用户: %d, 充值: %d", orderNo, topUp.UserId, quotaToAdd)
		model.RecordLog(topUp.UserId, model.LogTypeTopup, fmt.Sprintf("使用招商银行聚合支付成功，充值金额: %v", quotaToAdd))
	}

	c.Writer.Write([]byte("success"))
}

func QueryZSPayStatus(c *gin.Context) {
	tradeNo := c.Query("trade_no")
	if tradeNo == "" {
		c.JSON(200, gin.H{"message": "error", "data": "订单号不能为空"})
		return
	}

	zsService := service.GetZSPayService()
	if zsService == nil {
		c.JSON(200, gin.H{"message": "error", "data": "招商银行聚合支付未启用"})
		return
	}

	resp, err := zsService.OrderQuery(tradeNo)
	if err != nil {
		c.JSON(200, gin.H{"message": "error", "data": "查询失败"})
		return
	}

	status := convertTradeState(resp.TradeState)
	c.JSON(200, gin.H{
		"message":        "success",
		"status":         status,
		"trade_state":    resp.TradeState,
		"pay_type":       resp.PayType,
		"third_order_id": resp.ThirdOrderID,
		"txn_amt":        resp.TxnAmt,
	})
}

func convertTradeState(state string) string {
	switch state {
	case "S":
		return "PAID"
	case "F":
		return "FAILED"
	case "C", "D":
		return "CLOSED"
	case "R":
		return "REFUNDED"
	default:
		return "PENDING"
	}
}

func FormatZSMoney(fen string) string {
	f, _ := strconv.ParseFloat(fen, 64)
	return fmt.Sprintf("%.2f", f/100)
}
