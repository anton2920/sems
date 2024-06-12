package main

import "github.com/anton2920/gofa/log"

type Language int32

const (
	EN Language = iota
	RU
	FR
	XX
)

var Language2String = [...]string{
	EN: "English",
	RU: "Русский",
	FR: "Français",
}

var Localizations = map[string]*[XX]string{
	"Add another answer": {
		RU: "Добавить вариант ответа",
		FR: "",
	},
	"Add another question": {
		RU: "Добавить вопрос",
		FR: "",
	},
	"Add example": {
		RU: "Добавить пример",
		FR: "",
	},
	"Add lesson": {
		RU: "Добавить урок",
	},
	"Add programming task": {
		RU: "Добавить задание по программированию",
	},
	"Add test": {
		RU: "Добавить тест",
	},
	"Answers": {
		RU: "Ответы",
		FR: "",
	},
	"Answers (mark the correct ones)": {
		RU: "Ответы (пометьте галочкой правильные)",
		FR: "",
	},
	"Continue": {
		RU: "Продолжить",
	},
	"Course": {
		RU: "Курс",
	},
	"Courses": {
		RU: "Курсы",
	},
	"Create": {
		RU: "Создать",
		FR: "",
	},
	"Create course": {
		RU: "Создать курс",
	},
	"Create group": {
		RU: "Создать группу",
	},
	"Create lesson": {
		RU: "Создать урок",
		FR: "",
	},
	"Create subject": {
		RU: "Создать предмет",
	},
	"Create user": {
		RU: "Создать пользователя",
	},
	"Created on": {
		RU: "Дата создания",
		FR: "",
	},
	"Delete": {
		RU: "Удалить",
	},
	"Description": {
		RU: "Описание",
		FR: "",
	},
	"Discard": {
		RU: "Отменить",
		FR: "",
	},
	"Edit": {
		RU: "Редактировать",
	},
	"Edit group": {
		RU: "Редактирование группы",
		FR: "",
	},
	"Edit subject": {
		RU: "Редактирование предмета",
		FR: "",
	},
	"Edit subject lessons": {
		RU: "Редактирование уроков предмета",
		FR: "",
	},
	"Error": {
		RU: "Ошибка",
	},
	"Evaluation": {
		RU: "Задания",
		FR: "",
	},
	"Examples": {
		RU: "Примеры",
		FR: "",
	},
	"Finish": {
		RU: "Отправить",
		FR: "",
	},
	"Group": {
		RU: "Группа",
	},
	"Groups": {
		RU: "Группы",
	},
	"ID out of range": {
		RU: "ID вне допустимого диапазона",
	},
	"Info": {
		RU: "Информация",
		FR: "",
	},
	"Input": {
		RU: "Входные данные",
		FR: "",
	},
	"Lesson": {
		RU: "Урок",
	},
	"Lessons": {
		RU: "Уроки",
	},
	"Master's degree": {
		RU: "Магистерская диссертация",
		FR: "Une maîtrise",
	},
	"Name": {
		RU: "Название",
	},
	"Next": {
		RU: "Далее",
	},
	"Note: answers marked with [x] are correct": {
		RU: "Подсказка: правильные ответы помечены [x]",
		FR: "",
	},
	"Open": {
		RU: "Открыть",
	},
	"Pass": {
		RU: "Приступить к выполнению",
		FR: "",
	},
	"Programming language": {
		RU: "Язык программирования",
		FR: "",
	},
	"Programming task": {
		RU: "Задание по программированию",
		FR: "",
	},
	"Profile": {
		RU: "Профиль",
		FR: "Profil",
	},
	"Question": {
		RU: "Вопрос",
		FR: "",
	},
	"Re-check": {
		RU: "Перепроверить",
		FR: "",
	},
	"Save": {
		RU: "Сохранить",
	},
	"Score": {
		RU: "Оценка",
		FR: "",
	},
	"Sign in": {
		RU: "Войти",
		FR: "Se connecter",
	},
	"Sign out": {
		RU: "Выйти",
		FR: "Se déconnecter",
	},
	"Solution": {
		RU: "Решение",
		FR: "",
	},
	"Step": {
		RU: "Задание",
		FR: "",
	},
	"Students": {
		RU: "Студенты",
		FR: "",
	},
	"Subject": {
		RU: "Предмет",
	},
	"Subjects": {
		RU: "Предметы",
	},
	"Submission": {
		RU: "Решение",
		FR: "",
	},
	"Submissions": {
		RU: "Решения",
		FR: "",
	},
	"Submitted programming task": {
		RU: "Решённое задание по программированию",
		FR: "",
	},
	"Submitted test": {
		RU: "Решённый тест",
		FR: "",
	},
	"Teacher": {
		RU: "Преподаватель",
		FR: "",
	},
	"Test": {
		RU: "Тест",
		FR: "",
	},
	"Tests": {
		RU: "Тесты",
		FR: "",
	},
	"Title": {
		RU: "Название",
		FR: "",
	},
	"Theory": {
		RU: "Теория",
	},
	"This step has been skipped": {
		RU: "Задание было пропущено",
		FR: "",
	},
	"Total score": {
		RU: "Суммарная оценка",
		FR: "",
	},
	"Type": {
		RU: "Тип",
		FR: "",
	},
	"Users": {
		RU: "Пользователи",
		FR: "Utilisateurs",
	},

	"add at least one student": {
		RU: "добавьте хотя бы одного студента",
		FR: "",
	},
	"by": {
		RU: "от",
		FR: "",
	},
	"course name length must be between %d and %d characters long": {
		RU: "название курса должно содержать от %d до %d символов",
	},
	"course with this ID does not exist": {
		RU: "курса с таким ID не существует",
	},
	"create at least one lesson": {
		RU: "создайте хотя бы один урок",
	},
	"create new from scratch": {
		RU: "наполнить предмет с нуля",
		FR: "",
	},
	"create from": {
		RU: "взять за основу",
		FR: "",
	},
	"for": {
		RU: "для",
		FR: "",
	},
	"give as is": {
		RU: "выдать как есть",
		FR: "",
	},
	"group name length must be between %d and %d characters long": {
		RU: "название группы должно содержать от %d до %d символов",
	},
	"group with this ID does not exist": {
		RU: "группы с таким ID не существует",
	},
	"deleted": {
		RU: "удалён",
	},
	"draft": {
		RU: "черновик",
	},
	"index out of range": {
		RU: "индекс вне допустимого диапазона",
	},
	"invalid ID for %q": {
		RU: "некорректный ID для %q",
	},
	"lesson %d is a draft": {
		RU: "урок %d всё ещё черновик",
		FR: "",
	},
	"lesson name length must be between %d and %d characters long": {
		RU: "название урока должно содержать от %d до %d символов",
	},
	"lesson theory length must be between %d and %d characters long": {
		RU: "теория урока должна содержать от %d до %d символов",
	},
	"lesson with this ID does not exist": {
		RU: "урока с таким ID не существует",
	},
	"or": {
		RU: "или",
		FR: "",
	},
	"output": {
		RU: "выходные данные",
		FR: "",
	},
	"question %d: answer %d: length must be between %d and %d characters long": {
		RU: "вопрос %d: ответ %d: длина должна быть от %d до %d символов",
	},
	"question %d: select at least one correct answer": {
		RU: "вопрос %d: выберите хотя бы один правильный ответ",
		FR: "",
	},
	"question %d: select at least one answer": {
		RU: "вопрос %d: выберите хотя бы один ответ",
		FR: "",
	},
	"question %d: title length must be between %d and %d characters long": {
		RU: "вопрос %d: название должно содержать от %d до %d символов",
	},
	"requested API endpoint does not exist": {
		RU: "запрашиваемой команды не существует",
		FR: "",
	},
	"requested page does not exist": {
		RU: "запрашиваемой страницы не существует",
		FR: "",
	},
	"step %d is still a draft": {
		RU: "задание %d всё ещё черновик",
		FR: "",
	},
	"subject name length must be between %d and %d characters long": {
		RU: "название предмета должно содержать от %d до %d символов",
	},
	"subject with this ID does not exist": {
		RU: "предмета с таким ID не существует",
	},
	"test error": {
		RU: "тестовая ошибка",
		FR: "",
	},
	"test panic": {
		RU: "тестовая паника",
		FR: "",
	},
	"test name length must be between %d and %d characters long": {
		RU: "имя теста должно содержать от %d до %d символов",
	},
	"with": {
		RU: "с",
		FR: "",
	},
	"whoops... Something went wrong. Please reload this page or try again later": {
		RU: "упс... Что-то пошло не так. Перезагрузите страницу и повторите опрерацию ещё раз",
	},
	"whoops... Something went wrong. Please try again later": {
		RU: "упс... Что-то пошло не так. Пожалуйста, попробуйте ещё раз",
	},
	"whoops... You have to sign in to see this page": {
		RU: "упс... Войдите в систему для просмотра этой страницы",
	},
	"whoops... Your permissions are insufficient": {
		RU: "упс... Ваших прав недостаточно для просмотра этой страницы",
	},
	"you have to pass at least one step": {
		RU: "вы должны выполнить хотя бы одно задание",
		FR: "",
	},
}

var GL = RU

func (l Language) String() string {
	return Language2String[l]
}

func Ls(l Language, s string) string {
	if l == EN {
		return s
	}

	ls := Localizations[s]
	if (ls == nil) || (ls[l] == "") {
		switch s {
		default:
			log.Errorf("String %q is not localized in %q", s, l.String())
		case "↑", "^|", "↓", "|v", "-", "Command":
		}
		return s
	}

	return ls[l]
}
