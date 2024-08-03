package main

import (
	"fmt"
	"net/mail"
	"time"
	"unicode"
	"unicode/utf8"
	"unsafe"

	"github.com/anton2920/gofa/database"
	"github.com/anton2920/gofa/net/http"
	"github.com/anton2920/gofa/prof"
	"github.com/anton2920/gofa/strings"
	"github.com/anton2920/gofa/syscall"
	"github.com/anton2920/gofa/util"
)

type User struct {
	ID    database.ID
	Flags int32

	FirstName string
	LastName  string
	Email     string
	Password  string
	Courses   []database.ID
	CreatedOn int64

	Data [1024]byte
}

const (
	UserActive int32 = iota
	UserDeleted
)

const (
	MinUserNameLen = 1
	MaxUserNameLen = 45

	MinPasswordLen = 5
	MaxPasswordLen = 45
)

func UserNameValid(l Language, name string) error {
	defer prof.End(prof.Begin(""))

	if !strings.LengthInRange(name, MinUserNameLen, MaxUserNameLen) {
		return http.BadRequest(Ls(l, "length of the name must be between %d and %d characters"), MinUserNameLen, MaxUserNameLen)
	}

	/* Fist character must be a letter. */
	r, nbytes := utf8.DecodeRuneInString(name)
	if !unicode.IsLetter(r) {
		return http.BadRequest(Ls(l, "first character of the name must be a letter"))
	}

	/* Latter characters may include: letters, spaces, dots, hyphens and apostrophes. */
	for _, r := range name[nbytes:] {
		if (!unicode.IsLetter(r)) && (r != ' ') && (r != '.') && (r != '-') && (r != '\'') {
			return http.BadRequest(Ls(l, "second and latter characters of the name must be letters, spaces, dots, hyphens or apostrophes"))
		}
	}

	return nil
}

func GetUserByEmail(email string, user *User) error {
	defer prof.End(prof.Begin(""))

	users := make([]User, 32)
	var pos int64

	for {
		n, err := GetUsers(&pos, users)
		if err != nil {
			return err
		}
		if n == 0 {
			break
		}
		for i := 0; i < n; i++ {
			if (users[i].Flags != UserDeleted) && (users[i].Email == email) {
				*user = users[i]
				return nil
			}
		}
	}

	return database.NotFound
}

func CreateUser(user *User) error {
	defer prof.End(prof.Begin(""))

	var err error

	user.ID, err = database.IncrementNextID(UsersDB)
	if err != nil {
		return fmt.Errorf("failed to increment user ID: %w", err)
	}

	return SaveUser(user)
}

func DBUser2User(user *User) {
	defer prof.End(prof.Begin(""))

	data := &user.Data[0]

	user.FirstName = database.Offset2String(user.FirstName, data)
	user.LastName = database.Offset2String(user.LastName, data)
	user.Email = database.Offset2String(user.Email, data)
	user.Password = database.Offset2String(user.Password, data)
	user.Courses = database.Offset2Slice(user.Courses, data)
}

func GetUserByID(id database.ID, user *User) error {
	defer prof.End(prof.Begin(""))

	if err := database.Read(UsersDB, id, user); err != nil {
		return err
	}

	DBUser2User(user)
	return nil
}

func GetUsers(pos *int64, users []User) (int, error) {
	defer prof.End(prof.Begin(""))

	n, err := database.ReadMany(UsersDB, pos, users)
	if err != nil {
		return 0, err
	}

	for i := 0; i < n; i++ {
		DBUser2User(&users[i])
	}
	return n, nil
}

func DeleteUserByID(id database.ID) error {
	defer prof.End(prof.Begin(""))

	flags := UserDeleted
	var user User

	offset := int64(int(id)*int(unsafe.Sizeof(user))) + database.DataOffset + int64(unsafe.Offsetof(user.Flags))
	_, err := syscall.Pwrite(UsersDB.FD, unsafe.Slice((*byte)(unsafe.Pointer(&flags)), unsafe.Sizeof(flags)), offset)
	if err != nil {
		return fmt.Errorf("failed to delete user from DB: %w", err)
	}

	return nil
}

func User2DBUser(userDB *User, user *User, data []byte, n int) {
	defer prof.End(prof.Begin(""))

	userDB.ID = user.ID
	userDB.Flags = user.Flags

	/* TODO(anton2920): save up to a sizeof(user.Data). */
	n += database.String2DBString(&userDB.FirstName, user.FirstName, data, n)
	n += database.String2DBString(&userDB.LastName, user.LastName, data, n)
	n += database.String2DBString(&userDB.Email, user.Email, data, n)
	n += database.String2DBString(&userDB.Password, user.Password, data, n)
	n += database.Slice2DBSlice(&userDB.Courses, user.Courses, data, n)

	userDB.CreatedOn = user.CreatedOn
}

func SaveUser(user *User) error {
	defer prof.End(prof.Begin(""))

	var userDB User

	data := unsafe.Slice(&userDB.Data[0], len(userDB.Data))
	User2DBUser(&userDB, user, data, 0)

	return database.Write(UsersDB, userDB.ID, &userDB)
}

func UserOwnsCourse(user *User, courseID database.ID) bool {
	defer prof.End(prof.Begin(""))

	/* TODO(anton2920): move this out to caller? */
	if user.ID == AdminID {
		return true
	}
	for i := 0; i < len(user.Courses); i++ {
		cID := user.Courses[i]
		if cID == courseID {
			return true
		}
	}
	return false
}

func DisplayUserGroups(w *http.Response, l Language, userID database.ID) {
	groups := make([]Group, 32)
	var displayed bool
	var pos int64

	for {
		n, err := GetGroups(&pos, groups)
		if err != nil {
			/* TODO(anton2920): report error. */
		}
		if n == 0 {
			break
		}
		for i := 0; i < n; i++ {
			group := &groups[i]
			if (group.Flags == GroupDeleted) || (!UserInGroup(userID, group)) {
				continue
			}

			if !displayed {
				w.WriteString(`<h3>`)
				w.WriteString(Ls(l, "Groups"))
				w.WriteString(`</h3>`)
				w.WriteString(`<ul>`)
				displayed = true
			}

			w.WriteString(`<li>`)
			DisplayGroupLink(w, l, group)
			w.WriteString(`</li>`)
		}
	}
	if displayed {
		w.WriteString(`</ul>`)
		w.WriteString(`<br>`)
	}
}

func DisplayUserCourses(w *http.Response, l Language, user *User) {
	var course Course

	w.WriteString(`<h3>`)
	w.WriteString(Ls(l, "Courses"))
	w.WriteString(`</h3>`)
	w.WriteString(`<ul>`)

	const coursesOnPage = 10
	for i := 0; i < min(len(user.Courses), coursesOnPage); i++ {
		if err := GetCourseByID(user.Courses[i], &course); err != nil {
			/* TODO(anton2920): report error. */
		}
		if course.Flags == CourseDeleted {
			continue
		}

		w.WriteString(`<li>`)
		DisplayCourseLink(w, l, &course)
		w.WriteString(`</li>`)
	}
	if len(user.Courses) > coursesOnPage {
		w.WriteString(`<li>`)
		w.WriteString(`<a href="/courses">`)
		w.WriteString(Ls(l, "All"))
		w.WriteString(` `)
		w.WriteString(Ls(l, "Courses"))
		w.WriteString(`</a>`)
		w.WriteString(`</li>`)
	}
	w.WriteString(`</ul>`)
	w.WriteString(`<form method="POST" action="/course/create">`)
	DisplayButton(w, l, "", "Create course")
	w.WriteString(`</form>`)
	w.WriteString(`<br>`)
}

func DisplayUserSubjects(w *http.Response, l Language, userID database.ID) {
	subjects := make([]Subject, 32)
	var displayed bool
	var pos int64

	for {
		n, err := GetSubjects(&pos, subjects)
		if err != nil {
			/* TODO(anton2920): report error. */
		}
		if n == 0 {
			break
		}
		for i := 0; i < n; i++ {
			subject := &subjects[i]
			if subject.Flags == SubjectDeleted {
				continue
			}

			who, err := WhoIsUserInSubject(userID, subject)
			if err != nil {
				/* TODO(anton2920): report error. */
			}
			if who == SubjectUserNone {
				continue
			}

			if !displayed {
				w.WriteString(`<h3>`)
				w.WriteString(Ls(l, "Subjects"))
				w.WriteString(`</h3>`)
				w.WriteString(`<ul>`)
				displayed = true
			}

			w.WriteString(`<li>`)
			DisplaySubjectLink(w, l, subject)
			w.WriteString(`</li>`)
		}
	}
	if displayed {
		w.WriteString(`</ul>`)
	}
}

func DisplayUserTitle(w *http.Response, l Language, user *User) {
	w.WriteHTMLString(user.LastName)
	w.WriteString(` `)
	w.WriteHTMLString(user.FirstName)
	w.WriteString(` (ID: `)
	w.WriteID(user.ID)
	w.WriteString(`)`)
	DisplayDeleted(w, l, user.Flags == UserDeleted)
}

func DisplayUserLink(w *http.Response, l Language, user *User) {
	w.WriteString(`<a href="/user/`)
	w.WriteID(user.ID)
	w.WriteString(`">`)
	DisplayUserTitle(w, l, user)
	w.WriteString(`</a>`)
}

func UsersPageHandler(w *http.Response, r *http.Request) error {
	defer prof.End(prof.Begin(""))

	const width = WidthLarge

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}
	if session.ID != AdminID {
		return http.ForbiddenError
	}

	if err := r.URL.ParseQuery(); err != nil {
		return http.ClientError(err)
	}

	var page int
	if r.URL.Query.Has("Page") {
		page, err = r.URL.Query.GetInt("Page")
		if err != nil {
			return http.ClientError(err)
		}
	}

	nusers, err := database.GetNextID(UsersDB)
	if err != nil {
		return http.ServerError(err)
	}

	const usersPerPage = 10
	npages := int(nusers / usersPerPage)
	page = util.Clamp(page, 0, npages)

	DisplayHTMLStart(w)

	DisplayHeadStart(w)
	{
		w.WriteString(`<title>`)
		w.WriteString(Ls(GL, "Users"))
		w.WriteString(`</title>`)
	}
	DisplayHeadEnd(w)

	DisplayBodyStart(w)
	{
		DisplayHeader(w, GL)
		DisplaySidebar(w, GL, session)

		DisplayMainStart(w)

		DisplayCrumbsStart(w, width)
		{
			DisplayCrumbsItem(w, GL, "Users")
		}
		DisplayCrumbsEnd(w)

		DisplayPageStart(w, width)
		{
			w.WriteString(`<h2 class="text-center">`)
			w.WriteString(Ls(GL, "Users"))
			w.WriteString(`</h2>`)
			w.WriteString(`<br>`)

			DisplayTableStart(w, GL, []string{"ID", "First name", "Last name", "Email", "Created on", "Status"})
			{
				users := make([]User, usersPerPage)
				pos := database.GetOffsetForID[User](database.ID(page * usersPerPage))

				n, err := GetUsers(&pos, users)
				if err != nil {
					return http.ServerError(err)
				}

				for i := 0; i < n; i++ {
					user := &users[i]

					DisplayTableRowLinkIDStart(w, "/user", user.ID)

					const displayLen = 15
					DisplayTableItemShortenedString(w, user.FirstName, displayLen)
					DisplayTableItemShortenedString(w, user.LastName, displayLen)
					DisplayTableItemShortenedString(w, user.Email, 20)
					DisplayTableItemTime(w, user.CreatedOn)
					DisplayTableItemFlags(w, GL, user.Flags)

					DisplayTableRowEnd(w)
				}
			}
			DisplayTableEnd(w)

			DisplayPageSelector(w, "/users", page, npages)
			w.WriteString(`<br>`)

			w.WriteString(`<form method="POST" action="/user/create">`)
			DisplaySubmit(w, GL, "", "Create user", true)
			w.WriteString(`</form>`)
		}
		DisplayPageEnd(w)
		DisplayMainEnd(w)
	}
	DisplayBodyEnd(w)

	DisplayHTMLEnd(w)
	return nil
}

func UserPageHandler(w *http.Response, r *http.Request) error {
	defer prof.End(prof.Begin(""))

	const width = WidthLarge

	var user User

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}

	id, err := GetIDFromURL(GL, r.URL, "/user/")
	if err != nil {
		return http.ClientError(err)
	}
	if err := GetUserByID(id, &user); err != nil {
		if err == database.NotFound {
			return http.NotFound(Ls(GL, "user with this ID does not exist"))
		}
		return http.ServerError(err)
	}
	if (session.ID != AdminID) && (session.ID != user.ID) {
		return http.ForbiddenError
	}

	DisplayHTMLStart(w)

	DisplayHeadStart(w)
	{
		w.WriteString(`<title>`)
		DisplayUserTitle(w, GL, &user)
		w.WriteString(`</title>`)
	}
	DisplayHeadEnd(w)

	DisplayBodyStart(w)
	{
		DisplayHeader(w, GL)
		DisplaySidebar(w, GL, session)

		DisplayMainStart(w)

		DisplayCrumbsStart(w, width)
		{
			DisplayCrumbsItemStart(w)
			DisplayUserTitle(w, GL, &user)
			DisplayCrumbsItemEnd(w)
		}
		DisplayCrumbsEnd(w)

		DisplayPageStart(w, width)
		{
			w.WriteString(`<h2>`)
			DisplayUserTitle(w, GL, &user)
			w.WriteString(`</h2>`)
			w.WriteString(`<br>`)

			w.WriteString(`<h3>`)
			w.WriteString(Ls(GL, "Info"))
			w.WriteString(`</h3>`)

			w.WriteString(`<p>`)
			w.WriteString(Ls(GL, "Email"))
			w.WriteString(`: `)
			w.WriteHTMLString(user.Email)
			w.WriteString(`</p>`)

			w.WriteString(`<p>`)
			w.WriteString(Ls(GL, "Created on"))
			w.WriteString(`: `)
			DisplayFormattedTime(w, user.CreatedOn)
			w.WriteString(`</p>`)

			w.WriteString(`<div>`)
			w.WriteString(`<form style="display:inline" method="POST" action="/user/edit">`)
			DisplayHiddenID(w, "ID", user.ID)
			DisplayHiddenString(w, "FirstName", user.FirstName)
			DisplayHiddenString(w, "LastName", user.LastName)
			DisplayHiddenString(w, "Email", user.Email)
			DisplayButton(w, GL, "", "Edit")
			w.WriteString(`</form>`)

			if (session.ID == AdminID) && (id != AdminID) {
				w.WriteString(` <form style="display:inline" method="POST" action="/api/user/delete">`)
				DisplayHiddenID(w, "ID", user.ID)
				DisplayButton(w, GL, "", "Delete")
				w.WriteString(`</form>`)
			}
			w.WriteString(`</div>`)
			w.WriteString(`<br>`)

			DisplayUserGroups(w, GL, user.ID)
			DisplayUserCourses(w, GL, &user)
			DisplayUserSubjects(w, GL, user.ID)
		}
		DisplayPageEnd(w)

		DisplayMainEnd(w)
	}
	DisplayBodyEnd(w)

	DisplayHTMLEnd(w)
	return nil
}

func UserCreateEditPageHandler(w *http.Response, r *http.Request, session *Session, user *User, endpoint string, title string, action string, err error) error {
	defer prof.End(prof.Begin(""))

	const width = WidthSmall

	DisplayHTMLStart(w)

	DisplayHeadStart(w)
	{
		w.WriteString(`<title>`)
		w.WriteString(Ls(GL, title))
		w.WriteString(`</title>`)
	}
	DisplayHeadEnd(w)

	DisplayBodyStart(w)
	{
		DisplayHeader(w, GL)
		DisplaySidebar(w, GL, session)

		DisplayMainStart(w)

		DisplayCrumbsStart(w, width)
		{
			switch title {
			case "Create user":
				DisplayCrumbsLink(w, GL, "/users", "Users")
			case "Edit user":
				DisplayCrumbsLinkIDStart(w, "/user", user.ID)
				DisplayUserTitle(w, GL, user)
				DisplayCrumbsLinkEnd(w)
			}
			DisplayCrumbsItem(w, GL, title)
		}
		DisplayCrumbsEnd(w)

		DisplayFormPageStart(w, r, GL, width, title, endpoint, err)
		{
			DisplayLabel(w, GL, "First name")
			DisplayConstraintInput(w, "text", MinNameLen, MaxNameLen, "FirstName", r.Form.Get("FirstName"), true)
			w.WriteString(`<br>`)

			DisplayLabel(w, GL, "Last name")
			DisplayConstraintInput(w, "text", MinNameLen, MaxNameLen, "LastName", r.Form.Get("LastName"), true)
			w.WriteString(`<br>`)

			DisplayLabel(w, GL, "Email")
			DisplayInput(w, "email", "Email", r.Form.Get("Email"), true)
			w.WriteString(`<br>`)

			DisplayLabel(w, GL, "Password")
			DisplayConstraintInput(w, "password", MinPasswordLen, MaxPasswordLen, "Password", "", true)
			w.WriteString(`<br>`)

			DisplayLabel(w, GL, "Repeat password")
			DisplayConstraintInput(w, "password", MinPasswordLen, MaxPasswordLen, "RepeatPassword", "", true)
			w.WriteString(`<br>`)

			DisplaySubmit(w, GL, "", action, true)
		}
		DisplayFormPageEnd(w)

		DisplayMainEnd(w)
	}
	DisplayBodyEnd(w)

	DisplayHTMLEnd(w)
	return nil
}

func UserCreatePageHandler(w *http.Response, r *http.Request, e error) error {
	defer prof.End(prof.Begin(""))

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}
	if session.ID != AdminID {
		return http.ForbiddenError
	}

	if err := r.ParseForm(); err != nil {
		return http.ClientError(err)
	}

	return UserCreateEditPageHandler(w, r, session, nil, APIPrefix+"/user/create", "Create user", "Create", e)
}

func UserEditPageHandler(w *http.Response, r *http.Request, e error) error {
	defer prof.End(prof.Begin(""))

	var user User

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}

	if err := r.ParseForm(); err != nil {
		return http.ClientError(err)
	}

	userID, err := r.Form.GetID("ID")
	if err != nil {
		return http.ClientError(err)
	}
	if (session.ID != userID) && (session.ID != AdminID) {
		return http.ForbiddenError
	}

	if err := GetUserByID(userID, &user); err != nil {
		if err == database.NotFound {
			return http.NotFound(Ls(GL, "user with this ID does not exist"))
		}
		return http.ServerError(err)
	}

	return UserCreateEditPageHandler(w, r, session, &user, APIPrefix+"/user/edit", "Edit user", "Save", e)
}

func UserSigninPageHandler(w *http.Response, r *http.Request, err error) error {
	defer prof.End(prof.Begin(""))

	DisplayHTMLStart(w)

	DisplayHeadStart(w)
	{
		w.WriteString(`<title>`)
		w.WriteString(Ls(GL, "Sign in"))
		w.WriteString(`</title>`)

		if CSSEnabled {
			w.WriteString(`<style> html, body { height: 100%; } body { display: flex; align-items: center; padding-top: 40px; padding-bottom: 40px; }  .form-signin { max-width: 330px; padding: 15px; } </style>`)
		}
	}
	DisplayHeadEnd(w)

	DisplayBodyStart(w)
	{
		w.WriteString(`<main class="form-signin w-100 m-auto rounded-4 shadow bg-body-tertiary">`)

		w.WriteString(`<h2 class="text-center fw-normal"><b>`)
		w.WriteString(Ls(GL, "Sign in"))
		w.WriteString(`</b></h2>`)

		DisplayError(w, GL, err)

		w.WriteString(`<form class="form-signin" method="POST" action="/api/user/signin">`)

		DisplayLabel(w, GL, "Email")
		DisplayInput(w, "email", "Email", r.Form.Get("Email"), true)
		w.WriteString(`<br>`)

		DisplayLabel(w, GL, "Password")
		DisplayInput(w, "password", "Password", r.Form.Get("Password"), true)
		w.WriteString(`<br>`)

		DisplaySubmit(w, GL, "", "Sign in", true)
		w.WriteString(`</form>`)

		w.WriteString(`</main>`)
	}
	DisplayBodyEnd(w)

	DisplayHTMLEnd(w)
	return nil
}

func UserCreateHandler(w *http.Response, r *http.Request) error {
	defer prof.End(prof.Begin(""))

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}
	if session.ID != AdminID {
		return http.ForbiddenError
	}

	if err := r.ParseForm(); err != nil {
		return http.ClientError(err)
	}

	firstName := r.Form.Get("FirstName")
	if err := UserNameValid(GL, firstName); err != nil {
		return UserCreatePageHandler(w, r, err)
	}

	lastName := r.Form.Get("LastName")
	if err := UserNameValid(GL, lastName); err != nil {
		return UserCreatePageHandler(w, r, err)
	}

	address, err := mail.ParseAddress(r.Form.Get("Email"))
	if err != nil {
		return UserCreatePageHandler(w, r, http.BadRequest(Ls(GL, "provided email is not valid")))
	}
	email := address.Address

	password := r.Form.Get("Password")
	repeatPassword := r.Form.Get("RepeatPassword")
	if !strings.LengthInRange(password, MinPasswordLen, MaxPasswordLen) {
		return UserCreatePageHandler(w, r, http.BadRequest(Ls(GL, "password length must be between %d and %d characters long"), MinPasswordLen, MaxPasswordLen))
	}
	if password != repeatPassword {
		return UserCreatePageHandler(w, r, http.BadRequest(Ls(GL, "passwords do not match each other")))
	}

	var user User
	if err := GetUserByEmail(email, &user); err == nil {
		return UserCreatePageHandler(w, r, http.Conflict(Ls(GL, "user with this email already exists")))
	}

	user.FirstName = firstName
	user.LastName = lastName
	user.Email = email
	user.Password = password
	user.CreatedOn = time.Now().Unix()

	if err := CreateUser(&user); err != nil {
		return http.ServerError(err)
	}

	w.Redirect("/users", http.StatusSeeOther)
	return nil

}

func UserDeleteHandler(w *http.Response, r *http.Request) error {
	defer prof.End(prof.Begin(""))

	var user User

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}

	if err := r.ParseForm(); err != nil {
		return http.ClientError(err)
	}

	userID, err := r.Form.GetID("ID")
	if err != nil {
		return http.ClientError(err)
	}
	if err := GetUserByID(userID, &user); err != nil {
		if err == database.NotFound {
			return http.NotFound(Ls(GL, "user with this ID does not exist"))
		}
		return http.ServerError(err)
	}
	if session.ID != AdminID {
		return http.ForbiddenError
	}
	if userID == AdminID {
		return http.Conflict(Ls(GL, "cannot delete Admin user"))
	}

	/* TODO(anton2920): maybe in race with 'UserSigninHandler'. */
	RemoveAllUserSessions(userID)
	if err := DeleteUserByID(userID); err != nil {
		return http.ServerError(err)
	}
	for i := 0; i < len(user.Courses); i++ {
		_ = DeleteCourseByID(user.Courses[i])
	}

	w.Redirect("/users", http.StatusSeeOther)
	return nil
}

func UserEditHandler(w *http.Response, r *http.Request) error {
	defer prof.End(prof.Begin(""))

	var user User

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}

	if err := r.ParseForm(); err != nil {
		return http.ClientError(err)
	}

	userID, err := r.Form.GetID("ID")
	if err != nil {
		return http.ClientError(err)
	}
	if err := GetUserByID(userID, &user); err != nil {
		if err == database.NotFound {
			return http.NotFound(Ls(GL, "user with this ID does not exist"))
		}
		return http.ServerError(err)
	}
	if (session.ID != userID) && (session.ID != AdminID) {
		return http.ForbiddenError
	}

	firstName := r.Form.Get("FirstName")
	if err := UserNameValid(GL, firstName); err != nil {
		return UserEditPageHandler(w, r, err)
	}

	lastName := r.Form.Get("LastName")
	if err := UserNameValid(GL, lastName); err != nil {
		return UserEditPageHandler(w, r, err)
	}

	address, err := mail.ParseAddress(r.Form.Get("Email"))
	if err != nil {
		return UserEditPageHandler(w, r, http.BadRequest(Ls(GL, "provided email is not valid")))
	}
	email := address.Address

	password := r.Form.Get("Password")
	repeatPassword := r.Form.Get("RepeatPassword")
	if !strings.LengthInRange(password, MinPasswordLen, MaxPasswordLen) {
		return UserEditPageHandler(w, r, http.BadRequest(Ls(GL, "password length must be between %d and %d characters long"), MinPasswordLen, MaxPasswordLen))
	}
	if password != repeatPassword {
		return UserEditPageHandler(w, r, http.BadRequest(Ls(GL, "passwords do not match each other")))
	}

	var user2 User
	if err := GetUserByEmail(email, &user2); (err == nil) && (user2.ID != userID) {
		return UserEditPageHandler(w, r, http.Conflict(Ls(GL, "user with this email already exists")))
	}

	user.FirstName = firstName
	user.LastName = lastName
	user.Email = email
	user.Password = password

	if err := SaveUser(&user); err != nil {
		return http.ServerError(err)
	}

	UpdateAllUserSessions(&user)

	w.RedirectID("/user/", userID, http.StatusSeeOther)
	return nil
}

func UserSigninHandler(w *http.Response, r *http.Request) error {
	defer prof.End(prof.Begin(""))

	if err := r.ParseForm(); err != nil {
		return http.ClientError(err)
	}

	address, err := mail.ParseAddress(r.Form.Get("Email"))
	if err != nil {
		return UserSigninPageHandler(w, r, http.BadRequest(Ls(GL, "provided email is not valid")))
	}
	email := address.Address

	var user User
	if err := GetUserByEmail(email, &user); err != nil {
		if err == database.NotFound {
			return UserSigninPageHandler(w, r, http.NotFound(Ls(GL, "user with this email does not exist")))
		}
		return http.ServerError(err)
	}

	password := r.Form.Get("Password")
	if user.Password != password {
		return UserSigninPageHandler(w, r, http.Conflict(Ls(GL, "provided password is incorrect")))
	}

	token, err := GenerateSessionToken()
	if err != nil {
		return http.ServerError(err)
	}
	expiry := time.Now().Add(OneWeek)

	session := &Session{
		ID:     user.ID,
		Expiry: expiry,
	}
	User2DBUser(&session.User, &user, unsafe.Slice(&session.User.Data[0], len(session.User.Data)), 0)
	DBUser2User(&session.User)

	SessionsLock.Lock()
	Sessions[token] = session
	SessionsLock.Unlock()

	if Debug {
		w.SetCookieUnsafe("Token", token, int(expiry.Unix()))
	} else {
		w.SetCookie("Token", token, int(expiry.Unix()))
	}
	w.Redirect("/", http.StatusSeeOther)
	return nil
}

func UserSignoutHandler(w *http.Response, r *http.Request) error {
	defer prof.End(prof.Begin(""))

	token := r.Cookie("Token")
	if token == "" {
		return http.UnauthorizedError
	}

	if _, err := GetSessionFromToken(token); err != nil {
		return http.UnauthorizedError
	}

	SessionsLock.Lock()
	delete(Sessions, token)
	SessionsLock.Unlock()

	w.DelCookie("Token")
	w.Redirect("/", http.StatusSeeOther)
	return nil
}
