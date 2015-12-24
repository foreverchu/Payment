package paymentSrv

import ()

type OrderInfo interface {
	Valid() error
	GetNo() string
	GetDesc() string
	GetDetail() string
	GetPrice() int
	GetProductId() string
	Update(map[string]interface{}) error
}

type Pay interface {
	Pay() map[string]interface{}
	IsOrderValid() bool
}
