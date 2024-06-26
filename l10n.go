package main

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

/* TODO(anton2920): remove '([A-Z]|[a-z])[a-z]+' duplicates. */
var Localizations = map[string]*[XX]string{
	"Administration": {
		RU: "Управление",
	},
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
		RU: "Создание пользователя",
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
	"Edit lessons": {
		RU: "Редактирование уроков",
		FR: "",
	},
	"Edit user": {
		RU: "Редактирование пользователя",
		FR: "",
	},
	"Email": {
		RU: "Электронная почта",
		FR: "",
	},
	"Error": {
		RU: "Ошибка",
	},
	"Evaluation": {
		RU: "Задания",
		FR: "",
	},
	"Evaluation pass": {
		RU: "Выполнение заданий",
	},
	"Examples": {
		RU: "Примеры",
		FR: "",
	},
	"Finish": {
		RU: "Отправить",
		FR: "",
	},
	"First name": {
		RU: "Имя",
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
	"Last name": {
		RU: "Фамилия",
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
	"Password": {
		RU: "Пароль",
		FR: "",
	},
	"Pending": {
		RU: "Ожидается",
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
	"Repeat password": {
		RU: "Повторите пароль",
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
	"Unnamed": {
		RU: "Безымянный",
	},
	"User": {
		RU: "Пользователь",
		FR: "",
	},
	"Users": {
		RU: "Пользователи",
		FR: "Utilisateurs",
	},
	"Verification": {
		RU: "Проверка",
		FR: "",
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
	"example %d: %s": {
		RU: "пример %d: %s",
		FR: "",
	},
	"expected %q, got %q": {
		RU: "ожидалось %q, получено %q",
		FR: "",
	},
	"failed to compile program: exceeded compilation timeout of %d seconds": {
		RU: "неудалось собрать программу: превышено время ожидания в %d секунд",
		FR: "",
	},
	"failed to compile program: %s %w": {
		RU: "неудалось собрать программу: %s %w",
		FR: "",
	},
	"failed to run program: exceeded timeout of %d seconds": {
		RU: "неудалось выполнить программу: превышено время ожидания в %d секунд",
		FR: "",
	},
	"failed to run program: %s %w": {
		RU: "неудалось выполнить программу: %s %w",
		FR: "",
	},
	"first character of the name must be a letter": {
		RU: "первый символ имени/фамилии должен быть буквой",
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
	"in progress": {
		RU: "в процессе",
		FR: "",
	},
	"index out of range": {
		RU: "индекс вне допустимого диапазона",
	},
	"invalid ID for %q": {
		RU: "некорректный ID для %q",
	},
	"length of the name must be between %d and %d characters": {
		RU: "имя и фамилия должны содержать от %d до %d символов",
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
	"passwords do not match each other": {
		RU: "пароли не совпадают",
	},
	"password length must be between %d and %d characters long": {
		RU: "пароль должен содержать от %d до %d символов",
	},
	"pending": {
		RU: "ожидается",
		FR: "",
	},
	"programming task %d is a draft": {
		RU: "задание по программированию %d всё ещё черновик",
	},
	"provided email is not valid": {
		RU: "недопустимый адрес электронной почты",
	},
	"provided password is incorrect": {
		RU: "неверный пароль",
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
	"requested file does not exist": {
		RU: "запрашиваемоего файла не существует",
		FR: "",
	},
	"requested page does not exist": {
		RU: "запрашиваемой страницы не существует",
		FR: "",
	},
	"score": {
		RU: "оценка",
	},
	"second and latter characters of the name must be letters, spaces, dots, hyphens or apostrophes": {
		RU: "второй и последующий символы имени/фамилии должны быть буквы, пробелы, точки, дефисы и апострофы",
	},
	"selected language is not available": {
		RU: "выбранный язык недоступен",
	},
	"solution length must be between %d and %d characters long": {
		RU: "решение должно сожержать от %d до %d символов",
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
	"test %d is a draft": {
		RU: "тест %d всё ещё черновик",
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
	"user with this ID does not exist": {
		RU: "пользователя с таким ID не существует",
	},
	"user with this email already exists": {
		RU: "пользователь с такой электронной почтой уже существует",
	},
	"user with this email does not exist": {
		RU: "пользователя с такой электронной почтой не существует",
	},
	"verification": {
		RU: "проверка",
		FR: "",
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
		/*
			switch s {
			default:
				log.Errorf("Not localized %q", s)
			case "↑", "↓", "^|", "|v", "-", "Command":
			}
		*/
		return s
	}

	return ls[l]
}
