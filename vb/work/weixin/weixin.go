package weixin

import (
	"fmt"
	"github.com/lgdzz/vingo-utils/vingo"
	"github.com/lgdzz/vingo-utils/vingo/cache"
	"github.com/lgdzz/vingo-utils/vingo/request"
	"reflect"
	"time"
)

type Config struct {
	Corpid    string // 企业ID
	AgentId   string // 企业应用ID
	AppSecret string // 企业应用密钥
}

type Client struct {
	Config
}

func NewClient(config Config) *Client {
	return &Client{Config: config}
}

type ErrorResponse struct {
	Errcode int64  `json:"errcode"`
	Errmsg  string `json:"errmsg"`
}

type AccessTokenResponse struct {
	ErrorResponse
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

func HttpGet(url string) string {
	return string(request.Get(url, request.Option{}))
}

func HttpResponse[T any](resp string, response *T) {
	vingo.StringToJson(resp, response)
	value := reflect.ValueOf(response).Elem()
	if value.FieldByName("Errcode").Int() == 0 {
		return
	}
	panic(value.FieldByName("Errmsg").Interface())
}

func (s *Client) GetAccessToken() string {
	return *cache.Fast(fmt.Sprintf("work.weixin.access_token:%v", s.AgentId), 6000*time.Second, func() *string {
		resp := HttpGet(fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%v&corpsecret=%v", s.Corpid, s.AppSecret))
		var response AccessTokenResponse
		HttpResponse[AccessTokenResponse](resp, &response)
		return &response.AccessToken
	})
}

type DepartmentListResponse struct {
	ErrorResponse
	Department []Department `json:"department"`
}

type Department struct {
	Id               int64    `json:"id"`
	Name             string   `json:"name"`
	NameEn           string   `json:"name_en"`
	DepartmentLeader []string `json:"department_leader"`
	ParentId         int64    `json:"parentid"`
	Order            int64    `json:"order"`
}

// 获取部门列表
func (s *Client) GetDepartmentList() []Department {
	resp := HttpGet(fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/department/list?access_token=%v", s.GetAccessToken()))
	var response DepartmentListResponse
	HttpResponse[DepartmentListResponse](resp, &response)
	return response.Department
}
