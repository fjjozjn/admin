package rbac

import (
	m "fjjozjn/admin/src/models"
	"github.com/astaxie/beego/orm"
	"crypto/md5"
	"strings"
	"time"
	"fmt"
)

type UserController struct {
	CommonController
}

func (this *UserController) Index() {
	page, _ := this.GetInt64("page")
	page_size, _ := this.GetInt64("rows")
	sort := this.GetString("sort")
	order := this.GetString("order")
	if len(order) > 0 {
		if order == "desc" {
			sort = "-" + sort
		}
	} else {
		sort = "Id"
	}
	users, count := m.Getuserlist(page, page_size, sort)
	if this.IsAjax() {
		this.Data["json"] = &map[string]interface{}{"total": count, "rows": &users}
		this.ServeJSON()
		return
	} else {
		tree := this.GetTree()
		this.Data["tree"] = &tree
		this.Data["users"] = &users
		if this.GetTemplatetype() != "easyui" {
			this.Layout = this.GetTemplatetype() + "/public/layout.tpl"
		}
		this.TplName = this.GetTemplatetype() + "/rbac/user.tpl"
	}

}

func (this *UserController) AddUser() {
	u := m.User{}
	if err := this.ParseForm(&u); err != nil {
		//handle error
		this.Rsp(false, err.Error())
		return
	}
	id, err := m.AddUser(&u)
	if err == nil && id > 0 {

		//新增用户时同时插入用户数据到krnt_db
		myDb := orm.NewOrm()
		sNow := time.Now().Format("2006-01-02 15:04:05")
		bPassword := []byte("jdf932n" + u.Password)
		bEmailData := strings.Split(u.Email, "@")
		_, err := myDb.Raw("insert into krnt_db.tw_admin set AdminID = ?, AdminLogin = ?, AdminPassword = ?, AdminName = ?, AdminNameChi = ?, " +
			"AdminEmail = ?, AdminEmailRealName = ?, AdminEnabled = ?, AdminGrpID = ?, AdminCreateDate = ?, AdminPerm = ?, " +
			"AdminLuxGroup = ?, AdminJoinDate = ?, AdminPlatform = ?, mode = ?", id, u.Username, fmt.Sprintf("%x", md5.Sum(bPassword)),
			u.Nickname, u.Nickname, u.Email, bEmailData[0],
			1, 1, sNow, -1, "admin", sNow, "sys|fty", 1).Exec()
		if err != nil {
			this.Rsp(false, "Create tw_admin user false - " + err.Error())
			return
		}

		this.Rsp(true, "Success")
		return
	} else {
		this.Rsp(false, err.Error())
		return
	}

}

func (this *UserController) UpdateUser() {
	u := m.User{}
	if err := this.ParseForm(&u); err != nil {
		//handle error
		this.Rsp(false, err.Error())
		return
	}
	id, err := m.UpdateUser(&u)
	if err == nil && id > 0 {

		//修改用户时同时修改krnt_db用户数据
		myDb := orm.NewOrm()
		if len(u.Nickname) > 0 {
			_, err := myDb.Raw("update krnt_db.tw_admin set AdminName = ?, AdminNameChi = ? " +
				"where AdminID = ?", u.Nickname, u.Nickname, u.Id).Exec()
			if err != nil {
				this.Rsp(false, "Update tw_admin user (Nickname) false - " + err.Error())
				return
			}
		}
		if len(u.Email) > 0 {
			bEmailData := strings.Split(u.Email, "@")
			_, err := myDb.Raw("update krnt_db.tw_admin set AdminEmail = ?, " +
				"AdminEmailRealName = ? where AdminID = ?", u.Email, bEmailData[0], u.Id).Exec()
			if err != nil {
				this.Rsp(false, "Update tw_admin user (Email) false - " + err.Error())
				return
			}
		}
		if u.Status != 0 {
			iAdminEnabled := 0
			if u.Status == 1 {
				iAdminEnabled = 2
			} else if u.Status == 2 {
				iAdminEnabled = 1
			}
			_, err := myDb.Raw("update krnt_db.tw_admin set AdminEnabled = ? " +
				"where AdminID = ?", iAdminEnabled, u.Id).Exec()
			if err != nil {
				this.Rsp(false, "Update tw_admin user (Status) false - " + err.Error())
				return
			}
		}

		this.Rsp(true, "Success")
		return
	} else {
		this.Rsp(false, err.Error())
		return
	}

}

func (this *UserController) DelUser() {
	Id, _ := this.GetInt64("Id")
	status, err := m.DelUserById(Id)
	if err == nil && status > 0 {

		//删除用户时同时删除krnt_db用户数据
		myDb := orm.NewOrm()
		_, err := myDb.Raw("delete from krnt_db.tw_admin where AdminID = ?", Id).Exec()
		if err != nil {
			this.Rsp(false, "Delete tw_admin user false - " + err.Error())
			return
		}

		this.Rsp(true, "Success")
		return
	} else {
		this.Rsp(false, err.Error())
		return
	}
}
