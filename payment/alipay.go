package paymentSrv

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	NotifyUrl   string = "http://120.26.80.105:8080/alipay/notify"
	Partner     string = "2088811198755564"
	Key         string = "vfdr3c0kg2kzdac2kc3x9xmdrv4kzngm"
	SellerEmail string = "souvenir@chinarun.com"
)

type Alipay struct {
	Order          OrderInfo
	ReturnUrl      string
	NotifyUrl      string
	Partner        string
	Key            string
	WebSellerEmail string
	Result         map[string]interface{}
}

func NewAlipay(order OrderInfo, returnUrl string) *Alipay {
	alipay := &Alipay{
		Order:          order,
		ReturnUrl:      returnUrl,
		NotifyUrl:      NotifyUrl,
		Partner:        Partner,
		Key:            Key,
		WebSellerEmail: SellerEmail,
		Result:         make(map[string]interface{}),
	}
	return alipay
}
func (a *Alipay) IsOrderValid() bool {
	if a.Order.Valid() != nil {
		return false
	}
	return true
}

func (a *Alipay) Pay() map[string]interface{} {

	return a.CreateAlipaySign(a.Order.GetNo(), (float32(a.Order.GetPrice()) / 100), a.Order.GetDesc(), a.Order.GetDesc())
}

func (a *Alipay) HandleNotify(buyerEmail string) string {
	var resp string
	resp = "success"
	if a.Order.Valid() != nil {
		return resp + "OrderValid"
	}
	params := make(map[string]interface{})
	params["pay_time"] = time.Now().Local()
	params["pay_method"] = 1
	params["pay_account"] = buyerEmail
	err := a.Order.Update(params)
	if err != nil {
		resp += "UpdateOrderErr"
	}
	return resp
}

type AlipayParameters struct {
	InputCharset string  `json:"_input_charset"` //网站编码
	Body         string  `json:"body"`           //订单描述
	NotifyUrl    string  `json:"notify_url"`     //异步通知页面
	OutTradeNo   string  `json:"out_trade_no"`   //订单唯一id
	Partner      string  `json:"partner"`        //合作者身份ID
	PaymentType  uint8   `json:"payment_type"`   //支付类型 1：商品购买
	ReturnUrl    string  `json:"return_url"`     //回调url
	SellerEmail  string  `json:"seller_email"`   //卖家支付宝邮箱
	Service      string  `json:"service"`        //接口名称
	Subject      string  `json:"subject"`        //商品名称
	TotalFee     float32 `json:"total_fee"`      //总价
	Sign         string  `json:"sign"`           //签名，生成签名时忽略
	SignType     string  `json:"sign_type"`      //签名类型，生成签名时忽略
}

func Sign(param interface{}) string {
	//解析为字节数组
	paramBytes, err := json.Marshal(param)
	if err != nil {
		return ""
	}

	//重组字符串
	var sign string
	oldString := string(paramBytes)

	//为保证签名前特殊字符串没有被转码，这里解码一次
	oldString = strings.Replace(oldString, `\u003c`, "<", -1)
	oldString = strings.Replace(oldString, `\u003e`, ">", -1)

	//去除特殊标点
	oldString = strings.Replace(oldString, "\"", "", -1)
	oldString = strings.Replace(oldString, "{", "", -1)
	oldString = strings.Replace(oldString, "}", "", -1)
	paramArray := strings.Split(oldString, ",")

	for _, v := range paramArray {
		detail := strings.SplitN(v, ":", 2)
		//排除sign和sign_type
		if detail[0] != "sign" && detail[0] != "sign_type" {
			//total_fee转化为2位小数
			if detail[0] == "total_fee" {
				number, _ := strconv.ParseFloat(detail[1], 32)
				detail[1] = strconv.FormatFloat(number, 'f', 2, 64)
			}
			if sign == "" {
				sign = detail[0] + "=" + detail[1]
			} else {
				sign += "&" + detail[0] + "=" + detail[1]
			}
		}
	}

	//追加密钥
	sign += Key

	//md5加密
	m := md5.New()
	m.Write([]byte(sign))
	sign = hex.EncodeToString(m.Sum(nil))
	return sign
}
func (a *Alipay) CreateAlipaySign(orderId string, fee float32, nickname string, subject string) map[string]interface{} {
	//实例化参数
	ret := make(map[string]interface{})
	param := &AlipayParameters{}
	param.InputCharset = "utf-8"
	param.Body = "为" + nickname + "充值" + strconv.FormatFloat(float64(fee), 'f', 2, 32) + "元"
	param.NotifyUrl = a.NotifyUrl
	param.OutTradeNo = orderId
	param.Partner = a.Partner
	param.PaymentType = 1
	param.ReturnUrl = a.ReturnUrl
	param.SellerEmail = a.WebSellerEmail
	param.Service = "create_direct_pay_by_user"
	param.Subject = subject
	param.TotalFee = fee

	//生成签名
	sign := Sign(param)

	//追加参数
	param.Sign = sign
	param.SignType = "MD5"

	ret["alipay_url"] = fmt.Sprintf("https://mapi.alipay.com/gateway.do?_input_charset=utf-8&body=%s&notify_url=%s&out_trade_no=%s&partner=%s&payment_type=%s&return_url=%s&seller_email=%s&service=%s&subject=%s&total_fee=%s&sign=%s&sign_type=%s", param.Body, param.NotifyUrl, param.OutTradeNo, param.Partner, strconv.Itoa(int(param.PaymentType)), param.ReturnUrl, param.SellerEmail, param.Service, param.Subject, strconv.FormatFloat(float64(param.TotalFee), 'f', 2, 32), param.Sign, param.SignType)

	return ret

}
