package models

import "time"

// AuthLogins 表示授权登录服务
type AuthLogins struct {
	Google   bool `json:"google"`
	Wechat   bool `json:"wechat"`
	Weibo    bool `json:"weibo"`
	Baidu    bool `json:"baidu"`
	Facebook bool `json:"facebook"`
	Github   bool `json:"github"`
	QQ       bool `json:"qq"`
	Alipay   bool `json:"alipay"`
	Taobao   bool `json:"taobao"`
	Dingtalk bool `json:"dingtalk"`
	Douyin   bool `json:"douyin"`
	Feishu   bool `json:"feishu"`
	Twitter  bool `json:"twitter"`
}

// Password 表示密码实体
type Password struct {
	ID         int        `json:"id"`
	Name       string     `json:"name"`
	Username   string     `json:"username"`
	Phone      string     `json:"phone"`
	Password   string     `json:"password"`
	Website    string     `json:"website"`
	AuthLogins AuthLogins `json:"authLogins"`
	Notes      string     `json:"notes"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}
