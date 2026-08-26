package usecase

// Consumer-defined interfaces (Opsi A): usecase mendefinisikan kebutuhannya sendiri,
// concrete *repository.X memenuhinya secara implisit sehingga router tidak berubah.
// Method-set adalah union persis dari yang dipakai usecase (tidak lebih).

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/asepzainudin14/mcbt/internal/model"
	"github.com/asepzainudin14/mcbt/internal/repository"
)

type AcademicYearRepo interface {
	Activate(ctx context.Context, id uuid.UUID) error
	Create(ctx context.Context, ay *model.AcademicYear) error
	Delete(ctx context.Context, id uuid.UUID) error
	ExistsDuplicate(ctx context.Context, year, semester string, excludeID *uuid.UUID) (bool, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.AcademicYear, error)
	List(ctx context.Context, search string, page, limit int) (*repository.PageResult[model.AcademicYear], error)
	Update(ctx context.Context, ay *model.AcademicYear) error
}

type ClassRepo interface {
	Create(ctx context.Context, c *model.Class) error
	Delete(ctx context.Context, id uuid.UUID) error
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Class, error)
	List(ctx context.Context, p repository.ClassListParams) (*repository.PageResult[model.Class], error)
	ListAll(ctx context.Context) ([]model.Class, error)
	Update(ctx context.Context, c *model.Class) error
}

type DashboardRepo interface {
	AdminStats(ctx context.Context) (*repository.AdminStats, error)
	StudentStats(ctx context.Context, studentID uuid.UUID) (*repository.StudentStats, error)
	TeacherStats(ctx context.Context, userID uuid.UUID) (*repository.TeacherStats, error)
}

type ExamAnswerRepo interface {
	AnsweredQuestionIDs(ctx context.Context, questionIDs []uuid.UUID) (map[uuid.UUID]bool, error)
	ListByAttempt(ctx context.Context, attemptID uuid.UUID) ([]model.ExamAnswer, error)
	SetFlag(ctx context.Context, attemptID, questionID uuid.UUID, flagged bool) (*model.ExamAnswer, error)
	UpsertAnswer(ctx context.Context, attemptID, questionID uuid.UUID, answerValue string, clientTimestamp int64) (*model.ExamAnswer, error)
}

type ExamAttemptRepo interface {
	CountByExam(ctx context.Context, examID uuid.UUID) (int64, error)
	CountByStudentExam(ctx context.Context, examID, studentID uuid.UUID) (int64, error)
	Create(ctx context.Context, a *model.ExamAttempt) error
	FinalizeSubmit(ctx context.Context, id uuid.UUID, submittedAt time.Time) error
	FindActive(ctx context.Context, examID, studentID uuid.UUID) (*model.ExamAttempt, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.ExamAttempt, error)
	ListCandidateExams(ctx context.Context, studentID uuid.UUID) ([]repository.CandidateExamRow, error)
	MarkExpired(ctx context.Context, id uuid.UUID) error
}

type ExamParticipantRepo interface {
	Assign(ctx context.Context, examID uuid.UUID, studentIDs []uuid.UUID, via string) (int, error)
	AssignedStudentIDsByExam(ctx context.Context, examID uuid.UUID) (map[uuid.UUID]bool, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.ExamParticipant, error)
	ListByExam(ctx context.Context, examID uuid.UUID) ([]model.ExamParticipant, error)
	RemoveWithCleanup(ctx context.Context, examID, participantID, studentID uuid.UUID) error
	StudentIDsByClasses(ctx context.Context, classIDs []uuid.UUID) ([]uuid.UUID, error)
	StudentsExist(ctx context.Context, ids []uuid.UUID) (bool, error)
}

type ExamRepo interface {
	CountBankQuestion(ctx context.Context, bankID, questionID uuid.UUID, dst *int64) error
	Create(ctx context.Context, e *model.Exam) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Exam, error)
	List(ctx context.Context, p repository.ExamListParams) (*repository.PageResult[model.Exam], error)
	SetStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateCore(ctx context.Context, e *model.Exam) error
	UpdateSettings(ctx context.Context, e *model.Exam) error
}

type ExamScheduleRepo interface {
	Create(ctx context.Context, s *model.ExamSchedule) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByExam(ctx context.Context, examID uuid.UUID) (*model.ExamSchedule, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.ExamSchedule, error)
	TokenExists(ctx context.Context, token string, excludeID *uuid.UUID) (bool, error)
	Update(ctx context.Context, s *model.ExamSchedule) error
}

type ExamSectionRepo interface {
	Create(ctx context.Context, s *model.ExamSection) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.ExamSection, error)
	InsertMappings(ctx context.Context, sectionID uuid.UUID, questionIDs []uuid.UUID) (int, error)
	ListByExam(ctx context.Context, examID uuid.UUID) ([]repository.SectionWithCount, error)
	ListExamQuestions(ctx context.Context, exam *model.Exam) ([]repository.ExamQuestionGroup, error)
	ListQuestions(ctx context.Context, sectionID uuid.UUID) ([]model.Question, error)
	MappedQuestionIDsByExam(ctx context.Context, examID uuid.UUID) (map[uuid.UUID]bool, error)
	QuestionIDsByBanks(ctx context.Context, bankIDs []uuid.UUID) ([]uuid.UUID, error)
	QuestionInExam(ctx context.Context, examID, questionID uuid.UUID) (bool, error)
	RemoveQuestion(ctx context.Context, sectionID, questionID uuid.UUID) error
	Update(ctx context.Context, s *model.ExamSection) error
	UsedQuestionIDs(ctx context.Context, questionIDs []uuid.UUID) (map[uuid.UUID]bool, error)
}

type GradingRepo interface {
	FindAnswerByID(ctx context.Context, id uuid.UUID) (*model.ExamAnswer, error)
	FindAnswerByIDByAttempt(ctx context.Context, attemptID, questionID uuid.UUID) (*model.ExamAnswer, error)
	ListAnswersByAttempt(ctx context.Context, attemptID uuid.UUID) ([]model.ExamAnswer, error)
	ListQuestionsByIDs(ctx context.Context, ids []uuid.UUID) ([]model.Question, error)
	ListSubmittedAttempts(ctx context.Context, examID uuid.UUID) ([]model.ExamAttempt, error)
	ListUngradedEssays(ctx context.Context, examID uuid.UUID) ([]repository.UngradedEssayRow, error)
	QuestionWithGradingInfo(ctx context.Context, questionID uuid.UUID) (*model.Question, error)
	SumScoresByAttempt(ctx context.Context, attemptID uuid.UUID) (float64, error)
	UpdateAttemptScore(ctx context.Context, attemptID uuid.UUID, score float64) error
	UpdateGrading(ctx context.Context, answerID uuid.UUID, score float64, isCorrect *bool, feedback *string, via string) error
}

type MediaRepo interface {
	Create(ctx context.Context, m *model.Media) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Media, error)
}

type ProfileRepo interface {
	FindByUserID(ctx context.Context, userID uuid.UUID) (*repository.ProfileRow, error)
	UpdateName(ctx context.Context, userID uuid.UUID, name string) error
	UpdateStudentPhone(ctx context.Context, userID uuid.UUID, phone *string) error
	UpdateTeacherPhone(ctx context.Context, userID uuid.UUID, phone *string) error
}

type QuestionBankRepo interface {
	CloneWithQuestions(ctx context.Context, source *model.QuestionBank, createdBy *uuid.UUID) (*model.QuestionBank, error)
	Create(ctx context.Context, qb *model.QuestionBank) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.QuestionBank, error)
	List(ctx context.Context, p repository.BankListParams) (*repository.PageResult[model.QuestionBank], error)
	SetStatus(ctx context.Context, id uuid.UUID, status string) error
	Update(ctx context.Context, qb *model.QuestionBank) error
}

type QuestionReportRepo interface {
	Create(ctx context.Context, report *model.QuestionReport) error
	FindByAttemptQuestion(ctx context.Context, attemptID, questionID uuid.UUID) (*model.QuestionReport, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.QuestionReport, error)
	ListAll(ctx context.Context, status string, ownerUserID *uuid.UUID) ([]repository.ReportRow, error)
	Update(ctx context.Context, report *model.QuestionReport) error
}

type QuestionRepo interface {
	CountByBank(ctx context.Context, bankID uuid.UUID) (int64, error)
	CreateWithOptions(ctx context.Context, question *model.Question) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*model.Question, error)
	ImportQuestions(ctx context.Context, questions []model.Question) error
	List(ctx context.Context, p repository.QuestionListParams) (*repository.PageResult[model.Question], error)
	ReorderOptions(ctx context.Context, questionID uuid.UUID, orderedIDs []uuid.UUID) error
	SetCorrectOption(ctx context.Context, questionID, optionID uuid.UUID) error
	UpdateWithOptions(ctx context.Context, question *model.Question, options []model.Option) error
}

type ResultRepo interface {
	ExamResults(ctx context.Context, examID uuid.UUID, classID *uuid.UUID) ([]repository.ExamResultRow, error)
	SetResultsPublished(ctx context.Context, examID uuid.UUID, published bool) error
	StudentResults(ctx context.Context, studentID uuid.UUID) ([]repository.StudentResultRow, error)
}

type RoleRepo interface {
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]model.Role, error)
	FindByName(ctx context.Context, name string) (*model.Role, error)
	ListPaged(ctx context.Context, page, limit int) (*repository.PageResult[model.Role], error)
	ReplaceUserRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error
}

type StudentRepo interface {
	ChangeClass(ctx context.Context, studentID, targetClassID uuid.UUID) (*model.Student, error)
	CreateManyWithUsers(ctx context.Context, ups []repository.StudentUpsert, roleID uuid.UUID) error
	CreateWithUser(ctx context.Context, up repository.StudentUpsert, roleID uuid.UUID) (*model.Student, error)
	DeleteWithUser(ctx context.Context, student *model.Student) error
	ExistsDuplicate(ctx context.Context, username, email, nis string, excludeUserID *uuid.UUID) (string, bool, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Student, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) (*model.Student, error)
	List(ctx context.Context, search string, classID *uuid.UUID, page, limit int) (*repository.PageResult[model.Student], error)
	Update(ctx context.Context, student *model.Student, up repository.StudentUpdate) error
	UpdatePasswordByUser(ctx context.Context, userID uuid.UUID, passwordHash string) error
}

type SubjectRepo interface {
	Create(ctx context.Context, s *model.Subject) error
	Delete(ctx context.Context, id uuid.UUID) error
	ExistsDuplicate(ctx context.Context, code string, excludeID *uuid.UUID) (bool, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Subject, error)
	List(ctx context.Context, search string, page, limit int) (*repository.PageResult[model.Subject], error)
	Update(ctx context.Context, s *model.Subject) error
}

type TeacherRepo interface {
	CreateManyWithUsers(ctx context.Context, ups []repository.TeacherUpsert, roleID uuid.UUID) error
	CreateWithUser(ctx context.Context, up repository.TeacherUpsert, roleID uuid.UUID) (*model.Teacher, error)
	DeleteWithUser(ctx context.Context, teacher *model.Teacher) error
	ExistsDuplicate(ctx context.Context, username, email, nip string, excludeUserID *uuid.UUID) (string, bool, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Teacher, error)
	FindByUserID(ctx context.Context, userID uuid.UUID) (*model.Teacher, error)
	List(ctx context.Context, search string, page, limit int) (*repository.PageResult[model.Teacher], error)
	Update(ctx context.Context, teacher *model.Teacher, up repository.TeacherUpdate) error
	UpdatePasswordByUser(ctx context.Context, userID uuid.UUID, passwordHash string) error
}

type UserRepo interface {
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	FindByIdentifier(ctx context.Context, identifier string) (*model.User, error)
	TouchLastLogin(ctx context.Context, id uuid.UUID) error
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
}

// ---- dependensi lintas-usecase ----

// ExamRanker dipakai ExportUsecase (dipenuhi *ResultUsecase).
type ExamRanker interface {
	ExamResults(ctx context.Context, examID uuid.UUID, classID *uuid.UUID) ([]RankedResult, error)
}

// AttemptOwnerAssertor dipakai QuestionReportUsecase (dipenuhi *AccessUsecase).
type AttemptOwnerAssertor interface {
	AssertAttemptOwner(ctx context.Context, userID uuid.UUID, isAdmin bool, attemptID uuid.UUID) error
}
