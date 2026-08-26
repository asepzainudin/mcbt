package handler

import (
	"testing"

	"github.com/google/uuid"

	"github.com/asepzainudin14/mcbt/internal/model"
)

func TestOptionKey(t *testing.T) {
	cases := []struct{ in, want string }{
		{"a", "A"},
		{"A", "A"},
		{" b ", "B"},
		{"true", "TRUE"},
	}
	for _, c := range cases {
		if got := optionKey(c.in); got != c.want {
			t.Errorf("optionKey(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestAnswerKeysList(t *testing.T) {
	if got := answerKeysList(nil); len(got) != 0 {
		t.Errorf("nil → %v, want kosong", got)
	}
	if got := answerKeysList(stringPtr("")); len(got) != 0 {
		t.Errorf("kosong → %v, want kosong", got)
	}
	got := answerKeysList(stringPtr("Jawaban satu\n  Jawaban dua \n"))
	want := []string{"Jawaban satu", "Jawaban dua"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("got %v, want %v", got, want)
	}
}

func stringPtr(s string) *string { return &s }

func TestBankResponseActor_CanManage(t *testing.T) {
	owner := uuid.Must(uuid.Parse("11111111-1111-1111-1111-111111111111"))
	bank := &model.QuestionBank{CreatedBy: &owner}

	if !bankResponseActor(bank, owner, false)["can_manage"].(bool) {
		t.Error("pemilik bank harus can_manage=true")
	}
	if bankResponseActor(bank, uuid.New(), false)["can_manage"].(bool) {
		t.Error("bukan pemilik harus can_manage=false")
	}
	if !bankResponseActor(bank, uuid.Nil, true)["can_manage"].(bool) {
		t.Error("admin selalu can_manage=true")
	}

	orphan := &model.QuestionBank{} // created_by NULL (ujian lama / seed)
	if orphanResp, _ := bankResponseActor(orphan, uuid.New(), false)["can_manage"].(bool); orphanResp {
		t.Error("bank tanpa pemilik: non-admin tidak boleh can_manage")
	}
}
