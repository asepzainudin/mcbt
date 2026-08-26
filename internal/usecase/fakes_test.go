package usecase

// Fakes untuk seluruh interface repository.
// Pola: embed interface (nil) sehingga fake otomatis memenuhi kontrak;
// method yang tidak dioverride akan PANIC bila terpanggil (deteksi dini
// pemanggilan tak terduga), dan perilaku dikontrol lewat field fungsi.

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/asepzainudin14/mcbt/internal/model"
	"github.com/asepzainudin14/mcbt/internal/repository"
)

// ---------- helpers ----------

func uuidNil() uuid.UUID        { return uuid.Nil }
func pstr(s string) *string     { return &s }
func fltPtr(f float64) *float64 { return &f }
func boolPtr(b bool) *bool      { return &b }
func notFound() error           { return gorm.ErrRecordNotFound }
func ctxBg() context.Context    { return context.Background() }
func pageOf[T any](items []T) *repository.PageResult[T] {
	return &repository.PageResult[T]{Items: items, TotalItems: int64(len(items))}
}

// mustErr: pastikan err tidak nil (untuk rantai assert ringkas).
func mustErr(err error) error {
	if err == nil {
		panic("diharapkan error, dapat nil")
	}
	return err
}

var _ time.Time // jaga import time saat beberapa fake belum memakai

// ---------- UserRepo ----------

type fakeUserRepo struct {
	UserRepo
	findByIdentifierFn func(context.Context, string) (*model.User, error)
	findByIDFn         func(context.Context, uuid.UUID) (*model.User, error)
	existsByIDFn       func(context.Context, uuid.UUID) (bool, error)
	updatePasswordFn   func(context.Context, uuid.UUID, string) error
	touchLastLoginFn   func(context.Context, uuid.UUID) error
	passwordUpdates    []uuid.UUID // rekam pemanggilan update password
}

// kontrak nyata repo: tidak ditemukan → (nil, nil), bukan error
func (f *fakeUserRepo) FindByIdentifier(ctx context.Context, s string) (*model.User, error) {
	if f.findByIdentifierFn != nil {
		return f.findByIdentifierFn(ctx, s)
	}
	return nil, nil
}
func (f *fakeUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	if f.findByIDFn != nil {
		return f.findByIDFn(ctx, id)
	}
	return nil, nil // kontrak repo: (nil, nil) bila tak ada
}
func (f *fakeUserRepo) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	if f.existsByIDFn != nil {
		return f.existsByIDFn(ctx, id)
	}
	return true, nil
}
func (f *fakeUserRepo) UpdatePassword(ctx context.Context, id uuid.UUID, hash string) error {
	f.passwordUpdates = append(f.passwordUpdates, id)
	if f.updatePasswordFn != nil {
		return f.updatePasswordFn(ctx, id, hash)
	}
	return nil
}
func (f *fakeUserRepo) TouchLastLogin(ctx context.Context, id uuid.UUID) error {
	if f.touchLastLoginFn != nil {
		return f.touchLastLoginFn(ctx, id)
	}
	return nil
}

// ---------- RoleRepo ----------

type fakeRoleRepo struct {
	RoleRepo
	findByNameFn func(context.Context, string) (*model.Role, error)
	findByIDsFn  func(context.Context, []uuid.UUID) ([]model.Role, error)
	replaced     map[uuid.UUID][]uuid.UUID
}

func (f *fakeRoleRepo) FindByName(ctx context.Context, name string) (*model.Role, error) {
	if f.findByNameFn != nil {
		return f.findByNameFn(ctx, name)
	}
	return nil, notFound()
}
func (f *fakeRoleRepo) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]model.Role, error) {
	if f.findByIDsFn != nil {
		return f.findByIDsFn(ctx, ids)
	}
	out := make([]model.Role, 0, len(ids))
	for _, id := range ids {
		out = append(out, model.Role{BaseModel: model.BaseModel{ID: id}})
	}
	return out, nil
}
func (f *fakeRoleRepo) ListPaged(ctx context.Context, page, limit int) (*repository.PageResult[model.Role], error) {
	return pageOf([]model.Role{}), nil
}
func (f *fakeRoleRepo) ReplaceUserRoles(ctx context.Context, userID uuid.UUID, roleIDs []uuid.UUID) error {
	if f.replaced == nil {
		f.replaced = map[uuid.UUID][]uuid.UUID{}
	}
	f.replaced[userID] = roleIDs
	return nil
}

// ---------- AcademicYearRepo ----------

type fakeAYRepo struct {
	AcademicYearRepo
	findByIDFn    func(context.Context, uuid.UUID) (*model.AcademicYear, error)
	activateCalls []uuid.UUID
	listFn        func(context.Context, string, int, int) (*repository.PageResult[model.AcademicYear], error)
	createFn      func(context.Context, *model.AcademicYear) error
	updateFn      func(context.Context, *model.AcademicYear) error
	deleteFn      func(context.Context, uuid.UUID) error
	activateFn    func(context.Context, uuid.UUID) error
	dupFn         func(context.Context, string, string, *uuid.UUID) (bool, error)
}

func (f *fakeAYRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.AcademicYear, error) {
	if f.findByIDFn != nil {
		return f.findByIDFn(ctx, id)
	}
	return nil, notFound()
}
func (f *fakeAYRepo) List(ctx context.Context, s string, p, l int) (*repository.PageResult[model.AcademicYear], error) {
	if f.listFn != nil {
		return f.listFn(ctx, s, p, l)
	}
	return pageOf([]model.AcademicYear{}), nil
}
func (f *fakeAYRepo) Create(ctx context.Context, ay *model.AcademicYear) error {
	if f.createFn != nil {
		return f.createFn(ctx, ay)
	}
	return nil
}
func (f *fakeAYRepo) Update(ctx context.Context, ay *model.AcademicYear) error {
	if f.updateFn != nil {
		return f.updateFn(ctx, ay)
	}
	return nil
}
func (f *fakeAYRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if f.deleteFn != nil {
		return f.deleteFn(ctx, id)
	}
	return nil
}
func (f *fakeAYRepo) Activate(ctx context.Context, id uuid.UUID) error {
	f.activateCalls = append(f.activateCalls, id)
	if f.activateFn != nil {
		return f.activateFn(ctx, id)
	}
	return nil
}
func (f *fakeAYRepo) ExistsDuplicate(ctx context.Context, y, sem string, ex *uuid.UUID) (bool, error) {
	if f.dupFn != nil {
		return f.dupFn(ctx, y, sem, ex)
	}
	return false, nil
}

// ---------- ClassRepo ----------

type fakeClassRepo struct {
	ClassRepo
	findByIDFn func(context.Context, uuid.UUID) (*model.Class, error)
	listAllFn  func(context.Context) ([]model.Class, error)
	createFn   func(context.Context, *model.Class) error
	deleteFn   func(context.Context, uuid.UUID) error
	existsFn   func(context.Context, uuid.UUID) (bool, error)
	dupFn      func(context.Context, uuid.UUID, string, *uuid.UUID) (bool, error)
}

func (f *fakeClassRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Class, error) {
	if f.findByIDFn != nil {
		return f.findByIDFn(ctx, id)
	}
	return nil, notFound()
}
func (f *fakeClassRepo) List(ctx context.Context, p repository.ClassListParams) (*repository.PageResult[model.Class], error) {
	return pageOf([]model.Class{}), nil
}
func (f *fakeClassRepo) ListAll(ctx context.Context) ([]model.Class, error) {
	if f.listAllFn != nil {
		return f.listAllFn(ctx)
	}
	return nil, nil
}
func (f *fakeClassRepo) Create(ctx context.Context, c *model.Class) error {
	if f.createFn != nil {
		return f.createFn(ctx, c)
	}
	return nil
}
func (f *fakeClassRepo) Update(ctx context.Context, c *model.Class) error { return nil }
func (f *fakeClassRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if f.deleteFn != nil {
		return f.deleteFn(ctx, id)
	}
	return nil
}
func (f *fakeClassRepo) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	if f.existsFn != nil {
		return f.existsFn(ctx, id)
	}
	return true, nil
}
func (f *fakeClassRepo) ExistsDuplicate(ctx context.Context, ay uuid.UUID, name string, ex *uuid.UUID) (bool, error) {
	if f.dupFn != nil {
		return f.dupFn(ctx, ay, name, ex)
	}
	return false, nil
}

// ---------- SubjectRepo ----------

type fakeSubjectRepo struct {
	SubjectRepo
	findByIDFn func(context.Context, uuid.UUID) (*model.Subject, error)
	dupFn      func(context.Context, string, *uuid.UUID) (bool, error)
	createFn   func(context.Context, *model.Subject) error
	deleteErr  error
}

func (f *fakeSubjectRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Subject, error) {
	if f.findByIDFn != nil {
		return f.findByIDFn(ctx, id)
	}
	return nil, notFound()
}
func (f *fakeSubjectRepo) ExistsDuplicate(ctx context.Context, code string, ex *uuid.UUID) (bool, error) {
	if f.dupFn != nil {
		return f.dupFn(ctx, code, ex)
	}
	return false, nil
}
func (f *fakeSubjectRepo) Create(ctx context.Context, s *model.Subject) error {
	if f.createFn != nil {
		return f.createFn(ctx, s)
	}
	return nil
}
func (f *fakeSubjectRepo) Update(ctx context.Context, s *model.Subject) error { return nil }
func (f *fakeSubjectRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return f.deleteErr
}
func (f *fakeSubjectRepo) List(ctx context.Context, s string, p, l int) (*repository.PageResult[model.Subject], error) {
	return pageOf([]model.Subject{}), nil
}

// ---------- StudentRepo ----------

type fakeStudentRepo struct {
	StudentRepo
	findByIDFn    func(context.Context, uuid.UUID) (*model.Student, error)
	findByUID     func(context.Context, uuid.UUID) (*model.Student, error)
	listFn        func(context.Context, string, *uuid.UUID, int, int) (*repository.PageResult[model.Student], error)
	existsDupFn   func(context.Context, string, string, string, *uuid.UUID) (string, bool, error)
	changeClassFn func(context.Context, uuid.UUID, uuid.UUID) (*model.Student, error)
	deleteFn      func(context.Context, *model.Student) error
	createManyFn  func(context.Context, []repository.StudentUpsert) error
	resetted      []uuid.UUID
}

func (f *fakeStudentRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Student, error) {
	if f.findByIDFn != nil {
		return f.findByIDFn(ctx, id)
	}
	return nil, notFound()
}
func (f *fakeStudentRepo) FindByUserID(ctx context.Context, id uuid.UUID) (*model.Student, error) {
	if f.findByUID != nil {
		return f.findByUID(ctx, id)
	}
	return nil, notFound()
}
func (f *fakeStudentRepo) List(ctx context.Context, s string, c *uuid.UUID, p, l int) (*repository.PageResult[model.Student], error) {
	if f.listFn != nil {
		return f.listFn(ctx, s, c, p, l)
	}
	return pageOf([]model.Student{}), nil
}
func (f *fakeStudentRepo) ExistsDuplicate(ctx context.Context, u, e, n string, ex *uuid.UUID) (string, bool, error) {
	if f.existsDupFn != nil {
		return f.existsDupFn(ctx, u, e, n, ex)
	}
	return "", false, nil
}
func (f *fakeStudentRepo) CreateWithUser(ctx context.Context, up repository.StudentUpsert, roleID uuid.UUID) (*model.Student, error) {
	return &model.Student{}, nil
}
func (f *fakeStudentRepo) CreateManyWithUsers(ctx context.Context, ups []repository.StudentUpsert, roleID uuid.UUID) error {
	if f.createManyFn != nil {
		return f.createManyFn(ctx, ups)
	}
	return nil
}
func (f *fakeStudentRepo) Update(ctx context.Context, st *model.Student, up repository.StudentUpdate) error {
	return nil
}
func (f *fakeStudentRepo) ChangeClass(ctx context.Context, sid, cid uuid.UUID) (*model.Student, error) {
	if f.changeClassFn != nil {
		return f.changeClassFn(ctx, sid, cid)
	}
	return &model.Student{}, nil
}
func (f *fakeStudentRepo) UpdatePasswordByUser(ctx context.Context, uid uuid.UUID, h string) error {
	f.resetted = append(f.resetted, uid)
	return nil
}
func (f *fakeStudentRepo) DeleteWithUser(ctx context.Context, st *model.Student) error {
	if f.deleteFn != nil {
		return f.deleteFn(ctx, st)
	}
	return nil
}

// ---------- TeacherRepo ----------

type fakeTeacherRepo struct {
	TeacherRepo
	findByIDFn   func(context.Context, uuid.UUID) (*model.Teacher, error)
	findByUID    func(context.Context, uuid.UUID) (*model.Teacher, error)
	listFn       func(context.Context, string, int, int) (*repository.PageResult[model.Teacher], error)
	createManyFn func(context.Context, []repository.TeacherUpsert) error
	resetted     []uuid.UUID
}

func (f *fakeTeacherRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Teacher, error) {
	if f.findByIDFn != nil {
		return f.findByIDFn(ctx, id)
	}
	return nil, notFound()
}
func (f *fakeTeacherRepo) FindByUserID(ctx context.Context, id uuid.UUID) (*model.Teacher, error) {
	if f.findByUID != nil {
		return f.findByUID(ctx, id)
	}
	return nil, notFound()
}
func (f *fakeTeacherRepo) List(ctx context.Context, s string, p, l int) (*repository.PageResult[model.Teacher], error) {
	if f.listFn != nil {
		return f.listFn(ctx, s, p, l)
	}
	return pageOf([]model.Teacher{}), nil
}
func (f *fakeTeacherRepo) ExistsDuplicate(ctx context.Context, u, e, n string, ex *uuid.UUID) (string, bool, error) {
	return "", false, nil
}
func (f *fakeTeacherRepo) CreateWithUser(ctx context.Context, up repository.TeacherUpsert, roleID uuid.UUID) (*model.Teacher, error) {
	return &model.Teacher{}, nil
}
func (f *fakeTeacherRepo) CreateManyWithUsers(ctx context.Context, ups []repository.TeacherUpsert, roleID uuid.UUID) error {
	if f.createManyFn != nil {
		return f.createManyFn(ctx, ups)
	}
	return nil
}
func (f *fakeTeacherRepo) Update(ctx context.Context, t *model.Teacher, up repository.TeacherUpdate) error {
	return nil
}
func (f *fakeTeacherRepo) UpdatePasswordByUser(ctx context.Context, uid uuid.UUID, h string) error {
	f.resetted = append(f.resetted, uid)
	return nil
}
func (f *fakeTeacherRepo) DeleteWithUser(ctx context.Context, t *model.Teacher) error { return nil }

// ---------- QuestionBankRepo ----------

type fakeBankRepo struct {
	QuestionBankRepo
	findByIDFn func(context.Context, uuid.UUID) (*model.QuestionBank, error)
	createFn   func(context.Context, *model.QuestionBank) error
	listFn     func(context.Context, repository.BankListParams) (*repository.PageResult[model.QuestionBank], error)
	cloneSrc   *model.QuestionBank
	cloneOwner *uuid.UUID
	setStatus  []struct {
		id     uuid.UUID
		status string
	}
	deleted   []uuid.UUID
	deleteErr error
}

func (f *fakeBankRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.QuestionBank, error) {
	if f.findByIDFn != nil {
		return f.findByIDFn(ctx, id)
	}
	return nil, notFound()
}
func (f *fakeBankRepo) List(ctx context.Context, p repository.BankListParams) (*repository.PageResult[model.QuestionBank], error) {
	if f.listFn != nil {
		return f.listFn(ctx, p)
	}
	return pageOf([]model.QuestionBank{}), nil
}
func (f *fakeBankRepo) Create(ctx context.Context, qb *model.QuestionBank) error {
	if f.createFn != nil {
		return f.createFn(ctx, qb)
	}
	return nil
}
func (f *fakeBankRepo) Update(ctx context.Context, qb *model.QuestionBank) error { return nil }
func (f *fakeBankRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	f.deleted = append(f.deleted, id)
	return nil
}
func (f *fakeBankRepo) SetStatus(ctx context.Context, id uuid.UUID, status string) error {
	f.setStatus = append(f.setStatus, struct {
		id     uuid.UUID
		status string
	}{id, status})
	return nil
}
func (f *fakeBankRepo) CloneWithQuestions(ctx context.Context, src *model.QuestionBank, owner *uuid.UUID) (*model.QuestionBank, error) {
	f.cloneSrc, f.cloneOwner = src, owner
	clone := *src
	clone.ID = uuid.New()
	return &clone, nil
}

// ---------- QuestionRepo ----------

type fakeQuestionRepo struct {
	QuestionRepo
	findByIDFn          func(context.Context, uuid.UUID) (*model.Question, error)
	createWithOptionsFn func(context.Context, *model.Question) error
	listFn              func(context.Context, repository.QuestionListParams) (*repository.PageResult[model.Question], error)
	countByBank         map[uuid.UUID]int64
	deleted             []uuid.UUID
	imported            []model.Question
}

func (f *fakeQuestionRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Question, error) {
	if f.findByIDFn != nil {
		return f.findByIDFn(ctx, id)
	}
	return nil, notFound()
}
func (f *fakeQuestionRepo) List(ctx context.Context, p repository.QuestionListParams) (*repository.PageResult[model.Question], error) {
	if f.listFn != nil {
		return f.listFn(ctx, p)
	}
	return pageOf([]model.Question{}), nil
}
func (f *fakeQuestionRepo) CountByBank(ctx context.Context, bankID uuid.UUID) (int64, error) {
	return f.countByBank[bankID], nil
}
func (f *fakeQuestionRepo) CreateWithOptions(ctx context.Context, q *model.Question) error {
	if f.createWithOptionsFn != nil {
		return f.createWithOptionsFn(ctx, q)
	}
	return nil
}
func (f *fakeQuestionRepo) UpdateWithOptions(ctx context.Context, q *model.Question, o []model.Option) error {
	return nil
}
func (f *fakeQuestionRepo) ReplaceOptions(ctx context.Context, id uuid.UUID, o []model.Option) error {
	return nil
}
func (f *fakeQuestionRepo) ReorderOptions(ctx context.Context, id uuid.UUID, ids []uuid.UUID) error {
	return nil
}
func (f *fakeQuestionRepo) SetCorrectOption(ctx context.Context, id, oid uuid.UUID) error {
	return nil
}
func (f *fakeQuestionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	f.deleted = append(f.deleted, id)
	return nil
}
func (f *fakeQuestionRepo) ImportQuestions(ctx context.Context, qs []model.Question) error {
	f.imported = append(f.imported, qs...)
	return nil
}

// ---------- ExamRepo ----------

type fakeExamRepo struct {
	ExamRepo
	findByIDFn  func(context.Context, uuid.UUID) (*model.Exam, error)
	listFn      func(ctx context.Context, p repository.ExamListParams) (*repository.PageResult[model.Exam], error)
	created     []*model.Exam
	setStatuses []struct {
		id     uuid.UUID
		status string
	}
	deleted      []uuid.UUID
	countByBankQ map[string]int64
	coreUpdated  *model.Exam
}

func (f *fakeExamRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Exam, error) {
	if f.findByIDFn != nil {
		return f.findByIDFn(ctx, id)
	}
	return nil, notFound()
}
func (f *fakeExamRepo) List(ctx context.Context, p repository.ExamListParams) (*repository.PageResult[model.Exam], error) {
	if f.listFn != nil {
		return f.listFn(ctx, p)
	}
	return pageOf([]model.Exam{}), nil
}
func (f *fakeExamRepo) Create(ctx context.Context, e *model.Exam) error {
	f.created = append(f.created, e)
	return nil
}
func (f *fakeExamRepo) UpdateCore(ctx context.Context, e *model.Exam) error {
	f.coreUpdated = e
	return nil
}
func (f *fakeExamRepo) UpdateSettings(ctx context.Context, e *model.Exam) error { return nil }
func (f *fakeExamRepo) SetStatus(ctx context.Context, id uuid.UUID, status string) error {
	f.setStatuses = append(f.setStatuses, struct {
		id     uuid.UUID
		status string
	}{id, status})
	return nil
}
func (f *fakeExamRepo) Delete(ctx context.Context, id uuid.UUID) error {
	f.deleted = append(f.deleted, id)
	return nil
}
func (f *fakeExamRepo) CountBankQuestion(ctx context.Context, bankID, questionID uuid.UUID, dst *int64) error {
	*dst = f.countByBankQ[bankID.String()+"|"+questionID.String()]
	return nil
}

// ---------- ExamSectionRepo ----------

type fakeSectionRepo struct {
	ExamSectionRepo
	findByIDFn   func(context.Context, uuid.UUID) (*model.ExamSection, error)
	listByExamFn func(context.Context, uuid.UUID) ([]repository.SectionWithCount, error)
	listExamQFn  func(context.Context, *model.Exam) ([]repository.ExamQuestionGroup, error)
	inserted     []struct {
		sectionID uuid.UUID
		qids      []uuid.UUID
	}
	removed []struct {
		sectionID uuid.UUID
		qid       uuid.UUID
	}
	inExam      map[string]bool
	usedQ       map[uuid.UUID]bool
	qidsByBanks []uuid.UUID
	mappedSet   map[uuid.UUID]bool
	deleteErr   error
	removeErr   error
}

func (f *fakeSectionRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.ExamSection, error) {
	if f.findByIDFn != nil {
		return f.findByIDFn(ctx, id)
	}
	return nil, notFound()
}
func (f *fakeSectionRepo) ListByExam(ctx context.Context, examID uuid.UUID) ([]repository.SectionWithCount, error) {
	if f.listByExamFn != nil {
		return f.listByExamFn(ctx, examID)
	}
	return nil, nil
}
func (f *fakeSectionRepo) ListExamQuestions(ctx context.Context, exam *model.Exam) ([]repository.ExamQuestionGroup, error) {
	if f.listExamQFn != nil {
		return f.listExamQFn(ctx, exam)
	}
	return nil, nil
}
func (f *fakeSectionRepo) ListQuestions(ctx context.Context, sectionID uuid.UUID) ([]model.Question, error) {
	return nil, nil
}
func (f *fakeSectionRepo) Create(ctx context.Context, sec *model.ExamSection) error { return nil }
func (f *fakeSectionRepo) Update(ctx context.Context, sec *model.ExamSection) error { return nil }
func (f *fakeSectionRepo) Delete(ctx context.Context, id uuid.UUID) error           { return f.deleteErr }
func (f *fakeSectionRepo) InsertMappings(ctx context.Context, sectionID uuid.UUID, qids []uuid.UUID) (int, error) {
	f.inserted = append(f.inserted, struct {
		sectionID uuid.UUID
		qids      []uuid.UUID
	}{sectionID, qids})
	return len(qids), nil
}
func (f *fakeSectionRepo) RemoveQuestion(ctx context.Context, sectionID, questionID uuid.UUID) error {
	if f.removeErr != nil {
		return f.removeErr
	}
	f.removed = append(f.removed, struct {
		sectionID uuid.UUID
		qid       uuid.UUID
	}{sectionID, questionID})
	return nil
}
func (f *fakeSectionRepo) QuestionInExam(ctx context.Context, examID, questionID uuid.UUID) (bool, error) {
	return f.inExam[questionID.String()], nil
}
func (f *fakeSectionRepo) UsedQuestionIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]bool, error) {
	out := map[uuid.UUID]bool{}
	for _, id := range ids {
		if f.usedQ[id] {
			out[id] = true
		}
	}
	return out, nil
}
func (f *fakeSectionRepo) MappedQuestionIDsByExam(ctx context.Context, examID uuid.UUID) (map[uuid.UUID]bool, error) {
	return f.mappedSet, nil
}
func (f *fakeSectionRepo) QuestionIDsByBanks(ctx context.Context, bankIDs []uuid.UUID) ([]uuid.UUID, error) {
	return f.qidsByBanks, nil
}

// ---------- ExamAnswerRepo ----------

type fakeAnswerRepo struct {
	ExamAnswerRepo
	upserted []struct {
		attemptID uuid.UUID
		qid       uuid.UUID
		value     string
		ts        int64
		ret       *model.ExamAnswer
		err       error
	}
	flags []struct {
		attemptID uuid.UUID
		qid       uuid.UUID
		flagged   bool
	}
	byAttempt        map[uuid.UUID][]model.ExamAnswer
	answeredOverride map[uuid.UUID]bool
}

func (f *fakeAnswerRepo) UpsertAnswer(ctx context.Context, attemptID, qid uuid.UUID, value string, ts int64) (*model.ExamAnswer, error) {
	rec := struct {
		attemptID uuid.UUID
		qid       uuid.UUID
		value     string
		ts        int64
		ret       *model.ExamAnswer
		err       error
	}{attemptID, qid, value, ts, nil, nil}
	for i := range f.upserted {
		if f.upserted[i].attemptID == rec.attemptID && f.upserted[i].err != nil {
			rec.err = f.upserted[i].err
		}
	}
	f.upserted = append(f.upserted, rec)
	if rec.err != nil {
		return nil, rec.err
	}
	return &model.ExamAnswer{AttemptID: attemptID, QuestionID: qid}, nil
}
func (f *fakeAnswerRepo) ListByAttempt(ctx context.Context, attemptID uuid.UUID) ([]model.ExamAnswer, error) {
	return f.byAttempt[attemptID], nil
}
func (f *fakeAnswerRepo) AnsweredQuestionIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]bool, error) {
	if f.answeredOverride != nil {
		return f.answeredOverride, nil
	}
	return map[uuid.UUID]bool{}, nil
}
func (f *fakeAnswerRepo) SetFlag(ctx context.Context, attemptID, qid uuid.UUID, flagged bool) (*model.ExamAnswer, error) {
	f.flags = append(f.flags, struct {
		attemptID uuid.UUID
		qid       uuid.UUID
		flagged   bool
	}{attemptID, qid, flagged})
	return &model.ExamAnswer{}, nil
}

// ---------- ExamAttemptRepo ----------

type fakeAttemptRepo struct {
	ExamAttemptRepo
	findByIDFn func(context.Context, uuid.UUID) (*model.ExamAttempt, error)
	findActive map[string]*model.ExamAttempt // "exam|student"
	created    []*model.ExamAttempt
	finalized  []uuid.UUID
	marked     []uuid.UUID
	counts     map[string]int64 // "exam|student" atau "exam"
	candExams  []repository.CandidateExamRow
}

func key2(a, b uuid.UUID) string { return a.String() + "|" + b.String() }

func (f *fakeAttemptRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.ExamAttempt, error) {
	if f.findByIDFn != nil {
		return f.findByIDFn(ctx, id)
	}
	return nil, notFound()
}
func (f *fakeAttemptRepo) FindActive(ctx context.Context, examID, studentID uuid.UUID) (*model.ExamAttempt, error) {
	if a, ok := f.findActive[key2(examID, studentID)]; ok {
		return a, nil
	}
	return nil, notFound()
}
func (f *fakeAttemptRepo) CountByStudentExam(ctx context.Context, examID, studentID uuid.UUID) (int64, error) {
	return f.counts[key2(examID, studentID)], nil
}
func (f *fakeAttemptRepo) CountByExam(ctx context.Context, examID uuid.UUID) (int64, error) {
	return f.counts[examID.String()], nil
}
func (f *fakeAttemptRepo) Create(ctx context.Context, a *model.ExamAttempt) error {
	f.created = append(f.created, a)
	return nil
}
func (f *fakeAttemptRepo) FinalizeSubmit(ctx context.Context, id uuid.UUID, at time.Time) error {
	f.finalized = append(f.finalized, id)
	return nil
}
func (f *fakeAttemptRepo) MarkExpired(ctx context.Context, id uuid.UUID) error {
	f.marked = append(f.marked, id)
	return nil
}
func (f *fakeAttemptRepo) ListCandidateExams(ctx context.Context, studentID uuid.UUID) ([]repository.CandidateExamRow, error) {
	return f.candExams, nil
}

// ---------- ExamScheduleRepo ----------

type fakeScheduleRepo struct {
	ExamScheduleRepo
	findByExam map[uuid.UUID]*model.ExamSchedule
	findByIDFn func(context.Context, uuid.UUID) (*model.ExamSchedule, error)
	tokenTaken map[string]bool
}

func (f *fakeScheduleRepo) FindByExam(ctx context.Context, examID uuid.UUID) (*model.ExamSchedule, error) {
	if s, ok := f.findByExam[examID]; ok {
		return s, nil
	}
	return nil, notFound()
}
func (f *fakeScheduleRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.ExamSchedule, error) {
	if f.findByIDFn != nil {
		return f.findByIDFn(ctx, id)
	}
	return nil, notFound()
}
func (f *fakeScheduleRepo) TokenExists(ctx context.Context, token string, ex *uuid.UUID) (bool, error) {
	return f.tokenTaken[token], nil
}
func (f *fakeScheduleRepo) Create(ctx context.Context, sc *model.ExamSchedule) error { return nil }
func (f *fakeScheduleRepo) Update(ctx context.Context, sc *model.ExamSchedule) error { return nil }
func (f *fakeScheduleRepo) Delete(ctx context.Context, id uuid.UUID) error           { return nil }

// ---------- ExamParticipantRepo ----------

type fakeParticipantRepo struct {
	ExamParticipantRepo
	findByIDFn          func(context.Context, uuid.UUID) (*model.ExamParticipant, error)
	listByExam          map[uuid.UUID][]model.ExamParticipant
	assignedOverride    map[uuid.UUID]bool
	studentsByClassesFn func(context.Context, []uuid.UUID) ([]uuid.UUID, error)
	studentsExistFn     func(context.Context, []uuid.UUID) (bool, error)
	assigned            []struct {
		examID uuid.UUID
		stuIDs []uuid.UUID
		via    string
	}
	removed []uuid.UUID
}

func (f *fakeParticipantRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.ExamParticipant, error) {
	if f.findByIDFn != nil {
		return f.findByIDFn(ctx, id)
	}
	return nil, notFound()
}
func (f *fakeParticipantRepo) ListByExam(ctx context.Context, examID uuid.UUID) ([]model.ExamParticipant, error) {
	return f.listByExam[examID], nil
}
func (f *fakeParticipantRepo) Assign(ctx context.Context, examID uuid.UUID, stuIDs []uuid.UUID, via string) (int, error) {
	f.assigned = append(f.assigned, struct {
		examID uuid.UUID
		stuIDs []uuid.UUID
		via    string
	}{examID, stuIDs, via})
	return len(stuIDs), nil
}
func (f *fakeParticipantRepo) RemoveWithCleanup(ctx context.Context, examID, pid, sid uuid.UUID) error {
	f.removed = append(f.removed, pid)
	return nil
}
func (f *fakeParticipantRepo) AssignedStudentIDsByExam(ctx context.Context, examID uuid.UUID) (map[uuid.UUID]bool, error) {
	if f.assignedOverride != nil {
		return f.assignedOverride, nil
	}
	return map[uuid.UUID]bool{}, nil
}
func (f *fakeParticipantRepo) StudentIDsByClasses(ctx context.Context, classIDs []uuid.UUID) ([]uuid.UUID, error) {
	if f.studentsByClassesFn != nil {
		return f.studentsByClassesFn(ctx, classIDs)
	}
	return nil, nil
}
func (f *fakeParticipantRepo) StudentsExist(ctx context.Context, ids []uuid.UUID) (bool, error) {
	if f.studentsExistFn != nil {
		return f.studentsExistFn(ctx, ids)
	}
	return true, nil
}

// ---------- GradingRepo ----------

type fakeGradingRepo struct {
	GradingRepo
	findAnswerByIDFn      func(context.Context, uuid.UUID) (*model.ExamAnswer, error)
	findAnswerByAttemptFn func(context.Context, uuid.UUID, uuid.UUID) (*model.ExamAnswer, error)
	answersByAttempt      map[uuid.UUID][]model.ExamAnswer
	ungraded              []repository.UngradedEssayRow
	sumScores             map[uuid.UUID]float64
	updatedScore          []struct {
		attemptID uuid.UUID
		score     float64
	}
	updatedGrading []uuid.UUID
	submitted      []model.ExamAttempt
	questionsByID  map[uuid.UUID]*model.Question
}

func (f *fakeGradingRepo) ListUngradedEssays(ctx context.Context, examID uuid.UUID) ([]repository.UngradedEssayRow, error) {
	return f.ungraded, nil
}
func (f *fakeGradingRepo) SumScoresByAttempt(ctx context.Context, attemptID uuid.UUID) (float64, error) {
	return f.sumScores[attemptID], nil
}
func (f *fakeGradingRepo) UpdateAttemptScore(ctx context.Context, attemptID uuid.UUID, score float64) error {
	f.updatedScore = append(f.updatedScore, struct {
		attemptID uuid.UUID
		score     float64
	}{attemptID, score})
	return nil
}
func (f *fakeGradingRepo) UpdateGrading(ctx context.Context, answerID uuid.UUID, score float64, isCorrect *bool, feedback *string, via string) error {
	f.updatedGrading = append(f.updatedGrading, answerID)
	return nil
}
func (f *fakeGradingRepo) ListSubmittedAttempts(ctx context.Context, examID uuid.UUID) ([]model.ExamAttempt, error) {
	return f.submitted, nil
}
func (f *fakeGradingRepo) QuestionWithGradingInfo(ctx context.Context, questionID uuid.UUID) (*model.Question, error) {
	if q, ok := f.questionsByID[questionID]; ok {
		return q, nil
	}
	return nil, notFound()
}
func (f *fakeGradingRepo) ListQuestionsByIDs(ctx context.Context, ids []uuid.UUID) ([]model.Question, error) {
	var out []model.Question
	for _, id := range ids {
		if q, ok := f.questionsByID[id]; ok {
			out = append(out, *q)
		}
	}
	return out, nil
}
func (f *fakeGradingRepo) FindAnswerByID(ctx context.Context, id uuid.UUID) (*model.ExamAnswer, error) {
	if f.findAnswerByIDFn != nil {
		return f.findAnswerByIDFn(ctx, id)
	}
	return nil, notFound()
}
func (f *fakeGradingRepo) FindAnswerByIDByAttempt(ctx context.Context, attemptID, questionID uuid.UUID) (*model.ExamAnswer, error) {
	if f.findAnswerByAttemptFn != nil {
		return f.findAnswerByAttemptFn(ctx, attemptID, questionID)
	}
	return nil, notFound()
}
func (f *fakeGradingRepo) ListAnswersByAttempt(ctx context.Context, attemptID uuid.UUID) ([]model.ExamAnswer, error) {
	return f.answersByAttempt[attemptID], nil
}

// ---------- ResultRepo ----------

type fakeResultRepo struct {
	ResultRepo
	examResults []repository.ExamResultRow
	published   []struct {
		examID uuid.UUID
		val    bool
	}
	studentResults []repository.StudentResultRow
}

func (f *fakeResultRepo) ExamResults(ctx context.Context, examID uuid.UUID, classID *uuid.UUID) ([]repository.ExamResultRow, error) {
	return f.examResults, nil
}
func (f *fakeResultRepo) SetResultsPublished(ctx context.Context, examID uuid.UUID, val bool) error {
	f.published = append(f.published, struct {
		examID uuid.UUID
		val    bool
	}{examID, val})
	return nil
}
func (f *fakeResultRepo) StudentResults(ctx context.Context, studentID uuid.UUID) ([]repository.StudentResultRow, error) {
	return f.studentResults, nil
}

// ---------- QuestionReportRepo ----------

type fakeReportRepo struct {
	QuestionReportRepo
	findByIDFn  func(context.Context, uuid.UUID) (*model.QuestionReport, error)
	byAttemptQ  *model.QuestionReport
	created     []*model.QuestionReport
	updated     []*model.QuestionReport
	rowsFor     map[string][]repository.ReportRow // status -> rows
	rowsOwner   map[uuid.UUID][]repository.ReportRow
	lastOwnerFl *uuid.UUID
}

func (f *fakeReportRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.QuestionReport, error) {
	if f.findByIDFn != nil {
		return f.findByIDFn(ctx, id)
	}
	return nil, notFound()
}
func (f *fakeReportRepo) FindByAttemptQuestion(ctx context.Context, attemptID, questionID uuid.UUID) (*model.QuestionReport, error) {
	if f.byAttemptQ != nil {
		return f.byAttemptQ, nil
	}
	return nil, notFound()
}
func (f *fakeReportRepo) ListAll(ctx context.Context, status string, owner *uuid.UUID) ([]repository.ReportRow, error) {
	f.lastOwnerFl = owner
	if owner != nil {
		return f.rowsOwner[*owner], nil
	}
	return f.rowsFor[status], nil
}
func (f *fakeReportRepo) Create(ctx context.Context, r *model.QuestionReport) error {
	f.created = append(f.created, r)
	return nil
}
func (f *fakeReportRepo) Update(ctx context.Context, r *model.QuestionReport) error {
	f.updated = append(f.updated, r)
	return nil
}

// ---------- DashboardRepo ----------

type fakeDashRepo struct {
	DashboardRepo
	adminStats   *repository.AdminStats
	teacherStats *repository.TeacherStats
	studentStats *repository.StudentStats
}

func (f *fakeDashRepo) AdminStats(ctx context.Context) (*repository.AdminStats, error) {
	return f.adminStats, nil
}
func (f *fakeDashRepo) TeacherStats(ctx context.Context, uid uuid.UUID) (*repository.TeacherStats, error) {
	return f.teacherStats, nil
}
func (f *fakeDashRepo) StudentStats(ctx context.Context, sid uuid.UUID) (*repository.StudentStats, error) {
	return f.studentStats, nil
}

// ---------- ProfileRepo ----------

type fakeProfileRepo struct {
	ProfileRepo
	row        *repository.ProfileRow
	nameUpdate string
	phoneSet   **string
}

func (f *fakeProfileRepo) FindByUserID(ctx context.Context, uid uuid.UUID) (*repository.ProfileRow, error) {
	if f.row != nil {
		return f.row, nil
	}
	return &repository.ProfileRow{ID: uid, Username: "u", Name: "N", Email: "e@x.id"}, nil
}
func (f *fakeProfileRepo) UpdateName(ctx context.Context, uid uuid.UUID, name string) error {
	f.nameUpdate = name
	return nil
}
func (f *fakeProfileRepo) UpdateStudentPhone(ctx context.Context, uid uuid.UUID, phone *string) error {
	f.phoneSet = &phone
	return nil
}
func (f *fakeProfileRepo) UpdateTeacherPhone(ctx context.Context, uid uuid.UUID, phone *string) error {
	return nil
}

// ---------- MediaRepo ----------

type fakeMediaRepo struct {
	MediaRepo
	found *model.Media
}

func (f *fakeMediaRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Media, error) {
	if f.found != nil {
		return f.found, nil
	}
	return nil, notFound()
}
func (f *fakeMediaRepo) Create(ctx context.Context, m *model.Media) error { return nil }

// ---------- lintas-usecase ----------

type fakeExamRanker struct {
	ranked []RankedResult
}

func (f *fakeExamRanker) ExamResults(ctx context.Context, examID uuid.UUID, classID *uuid.UUID) ([]RankedResult, error) {
	return f.ranked, nil
}

type fakeOwnerAssertor struct {
	err        error
	lastUserID uuid.UUID
	lastAdmin  bool
}

func (f *fakeOwnerAssertor) AssertAttemptOwner(ctx context.Context, userID uuid.UUID, isAdmin bool, attemptID uuid.UUID) error {
	f.lastUserID, f.lastAdmin = userID, isAdmin
	return f.err
}
