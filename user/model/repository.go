package model

import (
	"errors"
	"log"
	"strconv"

	"github.com/John-Tonny/mnhost/model"
	"github.com/astaxie/beego/orm"

	pb "github.com/John-Tonny/mnhost/interface/out/user"
)

type Repository interface {
	Get(id string) (*pb.User, error)
	GetAll() ([]*pb.User, error)
	Create(*pb.User) error
	GetByMobile(mobile string) (*pb.User, error)
}

type UserRepository struct {
}

const (
	DB_NAME        = "MnHost"
	CON_COLLECTION = "users"
)

func GetUserRepository() *UserRepository {
	return &UserRepository{}
}

func (repo *UserRepository) Get(id string) (*pb.User, error) {
	log.Println("get user from id", id)
	iid, err := strconv.Atoi(id)
	if err != nil {
		log.Println("get user from id", id, "error:", err)
		return nil, errors.New("id format is wrong")
	}
	var user models.User
	o := orm.NewOrm()
	qs := o.QueryTable("user")
	err = qs.Filter("id", iid).One(&user)
	if err != nil {
		log.Println("get user from id", id, "error:", err)
		return nil, err
	}

	pbUser := User2PBUser(&user)
	log.Println("success get user from id", id)
	return &pbUser, nil
}

func (repo *UserRepository) GetAll() ([]*pb.User, error) {
	log.Println("get all user")
	var users []models.User
	o := orm.NewOrm()
	qs := o.QueryTable("user")
	nums, err := qs.All(&users)
	if err != nil {
		log.Println("get all user, error:", err)
		return nil, err
	}
	if nums == 0 {
		log.Println("get all user, error:", err)
		return nil, errors.New("no data")
	}

	pbUsers := make([]*pb.User, len(users))
	for i, user := range users {
		pbUser := User2PBUser(&user)
		pbUsers[i] = &pbUser
	}
	log.Println("success get all user")
	return pbUsers, nil
}

func (repo *UserRepository) Create(u *pb.User) error {
	log.Println("create user")
	var user models.User
	o := orm.NewOrm()
	qs := o.QueryTable("user")
	err := qs.Filter("mobile", u.Mobile).One(&user)
	if err == nil {
		log.Println("create user, error:", err)
		return errors.New("mobile is exist")
	}

	o = orm.NewOrm()

	user = PBUser2User(u)

	userId, err := o.Insert(&user)
	if err != nil {
		log.Println("create user, error:", err)
		return err
	}

	u.Id = strconv.FormatInt(userId, 10)
	log.Println("success create user id", u.Id)
	return nil
}

func (repo *UserRepository) GetByMobile(mobile string) (*pb.User, error) {
	log.Println("get user from mobile", mobile)
	var user models.User
	o := orm.NewOrm()
	qs := o.QueryTable("user")
	err := qs.Filter("mobile", mobile).One(&user)
	if err != nil {
		log.Println("get user from mobile", mobile, "error:", err)
		return nil, err
	}

	pbUser := User2PBUser(&user)
	log.Println("success get user from mobile", mobile)
	return &pbUser, nil
}

func PBUser2User(u *pb.User) models.User {
	data := models.User{
		Password: u.Password,
		Mobile:   u.Mobile,
	}
	log.Println("*******:", u.Name)
	if len(u.Name) > 0 {
		data.Name = u.Name
	}
	if len(u.RealName) > 0 {
		data.RealName = u.RealName
	}
	if len(u.IdCard) > 0 {
		data.IdCard = u.IdCard
	}
	if len(u.RewardAddress) > 0 {
		data.RewardAddress = u.RewardAddress
	}

	return data
}

func User2PBUser(u *models.User) pb.User {
	return pb.User{
		Id:            strconv.Itoa(u.Id),
		Name:          u.Name,
		Password:      u.Password,
		Mobile:        u.Mobile,
		RealName:      u.RealName,
		IdCard:        u.IdCard,
		RewardAddress: u.RewardAddress,
	}
}
