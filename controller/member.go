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

	"example.com/session"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Member struct {
	ID          uint64 `bson:"ID"`
	Name        string `bson:"Name"`
	DateofBirth string `bson:"DateofBirth"`
	DateofDeath string `bson:"DateofDeath"`
	Gender      string `bson:"Gender"`
	Passport    string `bson:"Passport"`
	Contact     string `bson:"Contact"`
}
type Members []Member

var (
	members Members
)

const (
	db   = "offering"
	coll = "members"
)

type MemberController struct {
	MemberPattern  *regexp.Regexp
	client         *mongo.Client
	sessionManager *session.SessionManager
}

func (member MemberController) FindMember(id uint64) (Member, error) {
	for _, m := range members {
		if m.ID == id {
			return m, nil
		}
	}
	return Member{}, fmt.Errorf("member ID %v not found", id)
}

func (member MemberController) fetchData() {
	_col := *member.client.Database(db).Collection(coll)
	result, err := _col.Find(context.TODO(), bson.M{})
	if err != nil {
		fmt.Println("fetchData error1 " + err.Error())
	} else {
		if err = result.All(context.TODO(), &members); err != nil {
			fmt.Println("fetchData error2 " + err.Error())
		}
	}
}

func (member MemberController) GetMembers() (Members, error) {
	return members, nil
}

func (member MemberController) FindNextID() uint64 {
	if len(members) == 0 {
		return 1
	}
	num := []uint64{}
	for _, x := range members {
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

func (member MemberController) AddMember(m Member) (Member, error) {
	if m.ID != 0 {
		return Member{}, fmt.Errorf("new Member cannot have a defined ID %v", m.ID)
	}
	m.ID = member.FindNextID()
	_col := member.client.Database(db).Collection(coll)
	_, err := _col.InsertOne(context.TODO(), m)
	if err != nil {
		return Member{}, fmt.Errorf(err.Error())
	}
	members = append(members, m)
	return m, nil
}

func (member MemberController) UpdateMember(m Member) (Member, error) {
	for i, me := range members {
		if m.ID == me.ID {
			_col := member.client.Database(db).Collection(coll)
			_, err := _col.UpdateOne(context.TODO(), bson.M{"ID": m.ID}, bson.M{"$set": bson.M{
				"Name":        m.Name,
				"DateofBirth": m.DateofBirth,
				"DateofDeath": m.DateofDeath,
				"Gender":      m.Gender,
				"Passport":    m.Passport,
				"Contact":     m.Contact}})
			if err != nil {
				return Member{}, fmt.Errorf(err.Error())
			}
			members[i] = m
			return members[i], nil
		}
	}
	return Member{}, fmt.Errorf("member ID %v not found", m.ID)
}

func (member MemberController) RemoveMember(id uint64) (Member, error) {
	for i, cow := range members {
		if cow.ID == id {
			_col := member.client.Database(db).Collection(coll)
			_, err := _col.DeleteOne(context.TODO(), bson.M{"ID": id})
			if err != nil {
				return Member{}, fmt.Errorf(err.Error())
			}
			members = append(members[:i], members[i+1:]...)
			return cow, nil
		}
	}
	return Member{}, fmt.Errorf("member ID %v not found", id)
}

func (member MemberController) EncodeResponseAsJson(data interface{}, w io.Writer) {
	enc := json.NewEncoder(w)
	enc.Encode(data)
}

func (member MemberController) ServeHttp(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/member" {
		switch r.Method {
		case http.MethodGet:
			{ //Get record of all the members
				data, err := member.GetMembers()
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				}
				member.EncodeResponseAsJson(data, w)
			}
		case http.MethodPost:
			{ //Add record of new member
				var m Member
				err := json.NewDecoder(r.Body).Decode(&m)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				}
				data, err := member.AddMember(m)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				}
				member.EncodeResponseAsJson(data, w)
			}
		case http.MethodPut:
			{ //Update records of members
				var m Member
				err := json.NewDecoder(r.Body).Decode(&m)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				}
				data, err := member.UpdateMember(m)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				}
				member.EncodeResponseAsJson(data, w)
			}
		default:
			{
				w.WriteHeader(http.StatusNotImplemented)
			}
		}
	} else {
		matches := member.MemberPattern.FindStringSubmatch(r.URL.Path)
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
				data, err := member.FindMember(uint64(id))
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				}
				member.EncodeResponseAsJson(data, w)
			}

		case http.MethodDelete:
			{ // Delete member by ID
				data, err := member.RemoveMember(uint64(id))
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				}
				member.EncodeResponseAsJson(data, w)
			}
		default:
			{
				w.WriteHeader(http.StatusNotImplemented)
			}
		}
	}
}

func NewMemberController(mclient *mongo.Client, s interface{}) *MemberController {

	return &MemberController{
		MemberPattern:  regexp.MustCompile(`^/member/(\d+)/?`),
		client:         mclient,
		sessionManager: s.(*session.SessionManager),
	}
}

func RegisterMemberController(client *mongo.Client, s interface{}) {
	membercontroller := NewMemberController(client, s)
	membercontroller.fetchData()
	http.HandleFunc("/member", membercontroller.ServeHttp)
	http.HandleFunc("/member/", membercontroller.ServeHttp)
}
