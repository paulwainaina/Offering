package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserType int

const (
	db1             = "offering"
	col1           = "users"
	Admin UserType = iota
	Normal
	Guest
)

type User struct {
	ID       uint64
	Name     string
	Type     UserType
	Password string
	Email    string
	Passport string
}

type Users []User

var (
	systemUsers   = Users{}
	loggedInUsers = Users{}
)

func GetLoggedInUsers() Users {
	return loggedInUsers
}

func AddLoggedInUsers(user User )(User,error){
	loggedInUsers=append(loggedInUsers,user)
	return user ,nil
}

func RemoveLoggedInUsers(id uint64)(error){
	for i,user :=range loggedInUsers{
		if user.ID==id {
			loggedInUsers=append(loggedInUsers[:i],loggedInUsers[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("user ID %v not found ",id)
}

type UserController struct {
	UserPattern *regexp.Regexp
	client      *mongo.Client
}

func (user UserController) GetUsers() (Users, error) {
	return systemUsers, nil
}

func (user UserController) FindUser(id uint64) (User, error) {
	for _, m := range systemUsers {
		if m.ID == id {
			return m, nil
		}
	}
	return User{}, fmt.Errorf("user ID %v not found", id)
}

func (user UserController) fetchData() {
	_col := *user.client.Database(db).Collection(coll)
	result, err := _col.Find(context.TODO(), bson.M{})
	if err != nil {
		fmt.Println("fetchData error1 " + err.Error())
	} else {
		if err = result.All(context.TODO(), &systemUsers); err != nil {
			fmt.Println("fetchData error2 " + err.Error())
		}
	}
}

func (user UserController) FindNextID() uint64 {
	if len(systemUsers) == 0 {
		return 1
	}
	num := []uint64{}
	for _, x := range systemUsers {
		num = append(num, x.ID)
	}
	sort.Slice(num, func(a, b int) bool { return num[a] > num[b] })
	for i := 0; i < len(num)-1; i++ {
		if num[i] > 0 && num[i+1] != num[i]+1 && num[i+1] != num[i] {
			return num[i] + 1
		}
	}
	return num[len(num)-1] + 1
}

func (user UserController) AddUser(m User) (User, error) {
	if m.ID != 0 {
		return User{}, fmt.Errorf("new User cannot have a defined ID %v", m.ID)
	}
	m.ID = user.FindNextID()
	_col := user.client.Database(db1).Collection(col1)
	_, err := _col.InsertOne(context.TODO(), m)
	if err != nil {
		return User{}, fmt.Errorf(err.Error())
	}
	systemUsers = append(systemUsers, m)
	return m, nil
}

func (user UserController) UpdateUser(m User) (User, error) {
	for i, me := range systemUsers {
		if m.ID == me.ID {
			_col := user.client.Database(db1).Collection(col1)
			_, err := _col.UpdateOne(context.TODO(), bson.M{"ID": m.ID}, bson.M{"$set": bson.M{
				"Name":     m.Name,
				"Type":     m.Type,
				"Email":    m.Email,
				"Password": m.Password,
				"Passport": m.Passport}})
			if err != nil {
				return User{}, fmt.Errorf(err.Error())
			}
			systemUsers[i] = m
			return systemUsers[i], nil
		}
	}
	return User{}, fmt.Errorf("user ID %v not found", m.ID)
}

func (user UserController) RemoveUser(id uint64) (User, error) {
	for i, cow := range systemUsers {
		if cow.ID == id {
			_col := user.client.Database(db1).Collection(col1)
			_, err := _col.DeleteOne(context.TODO(), bson.M{"ID": id})
			if err != nil {
				return User{}, fmt.Errorf(err.Error())
			}
			systemUsers = append(systemUsers[:i], systemUsers[i+1:]...)
			return cow, nil
		}
	}
	return User{}, fmt.Errorf("user ID %v not found", id)
}

func (user UserController) EncodeResponseAsJson(data interface{}, w io.Writer) {
	enc := json.NewEncoder(w)
	enc.Encode(data)
}

func (user UserController) ServeHttp(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/user" {
		switch r.Method {
		case http.MethodGet:
			{ //Get record of all the members
				data, err := user.GetUsers()
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				}
				user.EncodeResponseAsJson(data, w)
			}
		case http.MethodPost:
			{ //Add record of new member
				var m User
				err := json.NewDecoder(r.Body).Decode(&m)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				}
				data, err := user.AddUser(m)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				}
				user.EncodeResponseAsJson(data, w)
			}
		case http.MethodPut:
			{ //Update records of members
				var m User
				err := json.NewDecoder(r.Body).Decode(&m)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				}
				data, err := user.UpdateUser(m)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				}
				user.EncodeResponseAsJson(data, w)
			}
		default:
			{
				w.WriteHeader(http.StatusNotImplemented)
			}
		}
	} else {
		matches := user.UserPattern.FindStringSubmatch(r.URL.Path)
		if len(matches) == 0 {
			w.WriteHeader(http.StatusNotFound)
		}
		id, err := strconv.Atoi(matches[1])
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
		}
		switch r.Method {
		case http.MethodGet:
			{ //Get member by ID
				data, err := user.FindUser(uint64(id))
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				}
				user.EncodeResponseAsJson(data, w)
			}

		case http.MethodDelete:
			{ // Delete member by ID
				data, err := user.RemoveUser(uint64(id))
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				}
				user.EncodeResponseAsJson(data, w)
			}
		default:
			{
				w.WriteHeader(http.StatusNotImplemented)
			}
		}
	}
}

func NewUserController(mclient *mongo.Client) *UserController {

	return &UserController{
		UserPattern: regexp.MustCompile(`^/member/(\d+)/?`),
		client:      mclient,
	}
}

func RegisterUserController(client *mongo.Client) {
	membercontroller := NewUserController(client)
	membercontroller.fetchData()
	http.HandleFunc("/user", membercontroller.ServeHttp)
	http.HandleFunc("/user/", membercontroller.ServeHttp)
}
