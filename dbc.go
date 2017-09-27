package dbc

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"github.com/pkg/errors"
	"math/rand"
	"net"
	"strconv"
	"time"
	"os"
	"log"
)

const (
	API_SERVER_HOST       = "api.dbcapi.me"
	API_SERVER_FIRST_PORT = 8123
	API_SERVER_LAST_PORT  = 8130
)

var Dev = false

func init() {
	Dev = os.Getenv("DBC_MODE") == "dev"
}

var API_CMD_TERMINATOR = []byte{'\r', '\n'}

var (
	ErrTimeout        = errors.New("captcha timeout")
	ErrAccountInvalid = errors.New("invalid username or password")
	ErrLoginFail      = errors.New("fail to login")
)

type CaptchaRes struct {
	IsCorrect bool
	Status    int
	Captcha   int64
	Text      string
}

type UserInfo struct {
	IsBanned bool
	Status   int
	User     int64
	Rate     float64
	Balance  float64
	Error 	 string
}

type LoginCmd struct {
	Cmd      string `json:"cmd"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type CaptchaCmd struct {
	Cmd     string `json:"cmd"`
	Captcha string `json:"captcha"`
}

type Client struct {
	login    bool
	Username string
	Password string
	Conn     net.Conn
	UserInfo UserInfo
}

func NewClient(username, password string) (*Client, error) {
	if username == "" || password == "" {
		return nil, ErrAccountInvalid
	}

	return &Client{
		Username: username, Password: password,
	}, nil
}

func (c *Client) Login() error {
	if c.login {
		return nil
	}

	rand.Seed(time.Now().Unix())
	var port = API_SERVER_FIRST_PORT + rand.Intn(API_SERVER_LAST_PORT-API_SERVER_FIRST_PORT)
	conn, err := net.Dial("tcp", API_SERVER_HOST+":"+strconv.Itoa(port))
	if err != nil {
		return  err
	}

	c.Conn = conn

	res, err := c.call(&LoginCmd{
		Cmd: "login", Username: c.Username, Password: c.Password,
	})
	if err != nil {
		return err
	}

	err = json.Unmarshal(res, &c.UserInfo)
	if err != nil {
		return err
	}

	if c.UserInfo.Status != 0 {
		return errors.New(c.UserInfo.Error)
	}

	c.login = true

	return err
}

func (c *Client) Decode(b []byte) (string, error) {
	res, err := c.call(&CaptchaCmd{
		Cmd:     "upload",
		Captcha: base64.StdEncoding.EncodeToString(b),
	})
	if err != nil {
		return "", err
	}

	var code CaptchaRes
	if err = json.Unmarshal(res, &code); err != nil {
		return "", nil
	}

	if code.Text != "" {
		return code.Text, nil
	}

	cid := strconv.FormatInt(code.Captcha, 10)

	t := time.NewTicker(1 * time.Second)
	timeout := time.After(60 * time.Second)
	for {
		code, err := c.GetCaptcha(cid)
		if err != nil {
			return "", err
		}

		if code != "" {
			return code, nil
		}

		select {
		case <-timeout:
			return "", ErrTimeout
		case <-t.C:
		}
	}
}

func (c *Client) GetCaptcha(cid string) (string, error) {
	res, err := c.call(&CaptchaCmd{
		Cmd:     "captcha",
		Captcha: cid,
	})

	if err != nil {
		return "", err
	}

	var code CaptchaRes
	if err = json.Unmarshal(res, &code); err != nil {
		return "", err
	}

	return code.Text, nil
}

func (c *Client) Close() {
	c.Conn.Close()
}

func (c *Client) call(command interface{}) ([]byte, error) {
	if _, ok := command.(*LoginCmd); !ok && c.Login() != nil {
		return nil, ErrLoginFail
	}

	cmd, err := json.Marshal(command)
	if err != nil {
		return nil, err
	}

	if Dev {
		log.Println("->", string(cmd[:]))
	}

	cmd = append(cmd, API_CMD_TERMINATOR...)
	if _, err := c.Conn.Write(cmd); err != nil {
		return nil, err
	}

	res, _, err := bufio.NewReader(c.Conn).ReadLine()
	if err != nil {
		return nil, err
	}

	if Dev {
		log.Println("<-", string(res[:]))
	}

	return res, nil
}