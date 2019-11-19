package handler

import (
	"context"
	"errors"
	"fmt"
	"log"

	json "github.com/json-iterator/go"

	"github.com/micro/go-micro/broker"
	"golang.org/x/crypto/bcrypt"

	"github.com/John-Tonny/mnhost/conf"

	logPB "github.com/John-Tonny/mnhost/interface/out/log"
	pb "github.com/John-Tonny/mnhost/interface/out/user"

	"github.com/John-Tonny/mnhost/user/model"
)

type Handler struct {
	tokenService model.Authable
	Broker       broker.Broker
}

var (
	topic       string
	serviceName string
)

func init() {
	topic = config.GetBrokerTopic("log")
	serviceName = config.GetServiceName("user")
}

func GetHandler(bk broker.Broker) *Handler {
	return &Handler{
		tokenService: model.GetTokenService(),
		Broker:       bk,
	}
}

// 从主会话中 Clone() 出新会话处理查询
func (h *Handler) GetRepo() model.Repository {
	return model.GetUserRepository()
}

func (h *Handler) Create(ctx context.Context, req *pb.User, resp *pb.Response) error {
	// 哈希处理用户输入的密码
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	req.Password = string(hashedPwd)

	repo := h.GetRepo()
	if err := repo.Create(req); err != nil {
		return err
	}
	resp.User = req

	go func() {
		msg := fmt.Sprintf("[user] id:%s mobile: %s Created", req.GetId(), req.GetName(), req.GetMobile())
		h.pubLog(req.GetId(), "Create", msg)
	}()

	return nil
}

func (h *Handler) Get(ctx context.Context, req *pb.User, resp *pb.Response) error {
	repo := h.GetRepo()
	u, err := repo.Get(req.GetId())
	if err != nil {
		return err
	}
	resp.User = u

	go func() {
		msg := fmt.Sprintf("[user] id:%s", req.GetId())
		h.pubLog(req.GetId(), "Get", msg)
	}()

	return nil
}

func (h *Handler) GetAll(ctx context.Context, req *pb.Request, resp *pb.Response) error {
	repo := h.GetRepo()
	users, err := repo.GetAll()
	if err != nil {
		return err
	}
	resp.Users = users

	go func() {
		msg := ""
		h.pubLog("", "GetAll", msg)
	}()

	return nil
}

func (h *Handler) Auth(ctx context.Context, req *pb.User, resp *pb.Token) error {
	// 在 part3 中直接传参 &pb.User 去查找用户
	// 会导致 req 的值完全是数据库中的记录值
	// 即 req.Password 与 u.Password 都是加密后的密码
	// 将无法通过验证
	repo := h.GetRepo()
	u, err := repo.GetByMobile(req.Mobile)
	if err != nil {
		return err
	}

	// 进行密码验证
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)); err != nil {
		return err
	}
	t, err := h.tokenService.Encode(u)
	if err != nil {
		return err
	}
	resp.Token = t

	go func() {
		msg := fmt.Sprintf("[user] mobile: %s", req.GetMobile())
		h.pubLog("", "Auth", msg)
	}()

	return nil
}

func (h *Handler) ValidateToken(ctx context.Context, req *pb.Token, resp *pb.Token) error {
	// Decode token
	claims, err := h.tokenService.Decode(req.Token)
	if err != nil {
		return err
	}
	if claims.User.Id == "" {
		return errors.New("invalid user")
	}
	resp.Valid = true
	resp.UserId = claims.User.Id

	go func() {
		msg := fmt.Sprintf("[user] id: %s", claims.User.Id)
		h.pubLog(claims.User.Id, "ValidateToken", msg)
	}()

	return nil
}

// 发送log
func (h *Handler) pubLog(userID, method, msg string) error {
	logPB := logPB.Log{
		Method: method,
		Origin: serviceName,
		Msg:    msg,
	}
	body, err := json.Marshal(logPB)
	if err != nil {
		return err
	}

	data := &broker.Message{
		Header: map[string]string{
			"user_id": userID,
		},
		Body: body,
	}

	fmt.Println("publish log:", method, serviceName, msg)
	if err := h.Broker.Publish(topic, data); err != nil {
		log.Fatalf("[pub] failed: %v\n", err)
	}
	return nil
}
