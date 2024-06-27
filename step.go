package main

import "github.com/anton2920/gofa/net/http"

func StepsPageHandler(w *http.Response, r *http.Request) error {
	const width = WidthLarge

	session, err := GetSessionFromRequest(r)
	if err != nil {
		return http.UnauthorizedError
	}

	DisplayHTMLStart(w)

	DisplayHeadStart(w)
	{
		w.AppendString(`<title>`)
		w.AppendString(Ls(GL, "Steps"))
		w.AppendString(`</title>`)
	}
	DisplayHeadEnd(w)

	DisplayBodyStart(w)
	{
		DisplayHeader(w, GL)
		DisplaySidebar(w, GL, session)

		DisplayMainStart(w)

		DisplayCrumbsStart(w, width)
		{
			DisplayCrumbsItem(w, GL, "Steps")
		}
		DisplayCrumbsEnd(w)

		DisplayPageStart(w, width)
		{
			w.AppendString(`<h2 class="text-center">`)
			w.AppendString(Ls(GL, "Steps"))
			w.AppendString(`</h2>`)
			w.AppendString(`<br>`)

			/* TODO(anton2920): this is very slow!!! */
			subjects := make([]Subject, 32)
			var displayed bool
			var pos int64

			for {
				n, err := GetSubjects(&pos, subjects)
				if err != nil {
					return http.ServerError(err)
				}
				if n == 0 {
					break
				}
				for i := 0; i < n; i++ {
					subject := &subjects[i]

					who, err := WhoIsUserInSubject(session.ID, subject)
					if err != nil {
						return http.ServerError(err)
					}
					if who != SubjectUserStudent {
						continue
					}

					for j := 0; j < len(subject.Lessons); j++ {
						var lesson Lesson
						if err := GetLessonByID(subject.Lessons[j], &lesson); err != nil {
							return http.ServerError(err)
						}

						if len(lesson.Steps) > 0 {
							var found bool
							for k := 0; k < len(lesson.Submissions); k++ {
								var submission Submission
								if err := GetSubmissionByID(lesson.Submissions[k], &submission); err != nil {
									return http.ServerError(err)
								}
								if submission.UserID == session.ID {
									found = true
									break
								}
							}
							if !found {
								if !displayed {
									DisplayTableStart(w, GL, []string{"ID", "Subject", "Lesson", "Steps"})
									displayed = true
								}

								DisplayTableRowLinkIDStart(w, "/lesson", lesson.ID)

								DisplayTableItemString(w, subject.Name)
								DisplayTableItemString(w, lesson.Name)
								DisplayTableItemInt(w, len(lesson.Steps))

								DisplayTableRowEnd(w)
							}
						}
					}
				}
			}

			if displayed {
				DisplayTableEnd(w)
			} else {
				w.AppendString(`<h4>`)
				w.AppendString(Ls(GL, "You don't have any unfinished steps"))
				w.AppendString(`</h4>`)
			}

			w.AppendString(`<br>`)
			w.AppendString(`<h2 class="text-center">`)
			w.AppendString(Ls(GL, "Submissions"))
			w.AppendString(`</h2>`)
			w.AppendString(`<br>`)

			DisplayTableStart(w, GL, []string{"ID", "Subject", "Lesson", "Score"})
			{
				submissions := make([]Submission, 32)
				var pos int64

				for {
					n, err := GetSubmissions(&pos, submissions)
					if err != nil {
						return http.ServerError(err)
					}
					if n == 0 {
						break
					}
					for i := 0; i < n; i++ {
						submission := &submissions[i]

						if (submission.UserID != session.ID) || (submission.Flags != SubmissionActive) {
							continue
						}

						var lesson Lesson
						if err := GetLessonByID(submission.LessonID, &lesson); err != nil {
							return http.ServerError(err)
						}

						var subject Subject
						if err := GetSubjectByID(lesson.ContainerID, &subject); err != nil {
							return http.ServerError(err)
						}

						DisplayTableRowLinkIDStart(w, "/submission", submission.ID)

						DisplayTableItemString(w, subject.Name)
						DisplayTableItemString(w, lesson.Name)

						DisplayTableItemStart(w)
						DisplaySubmissionTotalScore(w, submission)
						DisplayTableItemEnd(w)

						DisplayTableRowEnd(w)
					}
				}
			}
			DisplayTableEnd(w)
		}
		DisplayPageEnd(w)
		DisplayMainEnd(w)
	}
	DisplayBodyEnd(w)

	DisplayHTMLEnd(w)
	return nil
}
