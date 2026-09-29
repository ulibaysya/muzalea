package storage

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/ulibaysya/muzalea/internal/pkg/compensate"
)

func assertFailureResults(t *testing.T, err, target error, id int64, compFunc compensate.Func) {
	t.Helper()

	if !errors.Is(err, target) {
		t.Fatalf("expected err %v, got %v", target, err)
	}

	if id != 0 {
		t.Fatalf("expected id 0, got %v", id)
	}

	if compFunc != nil {
		t.Fatalf("unexpected non-nil compensation func")
	}
}

func TestSave_FailBegin(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	ctx := t.Context()
	name := "name"
	author := "author"
	tags := []string{
		"tag1",
		"tag2",
	}

	someErr := errors.New("some error")

	mock.ExpectBegin().
		WillReturnError(someErr)

	returnedID, returnedCompFunc, returnedErr := New(slog.Default(), mock).Save(ctx, name, author, tags)

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("asserting mock expectations: %v", err)
	}
	assertFailureResults(t, returnedErr, someErr, returnedID, returnedCompFunc)
}

func TestSave_FailInsertName(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	ctx := t.Context()
	name := "name"
	author := "author"
	tags := []string{
		"tag1",
		"tag2",
	}

	someErr := errors.New("some error")

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO audio (name, author) VALUES ($1, $2) RETURNING id;`).
		WithArgs(name, author).
		WillReturnError(someErr)
	mock.ExpectRollback()

	returnedID, returnedCompFunc, returnedErr := New(slog.Default(), mock).Save(ctx, name, author, tags)

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("asserting mock expectations: %v", err)
	}
	assertFailureResults(t, returnedErr, someErr, returnedID, returnedCompFunc)
}

func TestSave_FailInsertTags(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	ctx := t.Context()
	name := "name"
	author := "author"
	tags := []string{
		"tag1",
		"tag2",
	}

	id := int64(123)
	someErr := errors.New("some error")

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO audio (name, author) VALUES ($1, $2) RETURNING id;`).
		WithArgs(name, author).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(id))
	mock.ExpectExec(`INSERT INTO audio_tags (audio_id, tag) VALUES ($1, $2), ($3, $4);`).
		WithArgs(id, tags[0], id, tags[1]).
		WillReturnError(someErr)
	mock.ExpectRollback()

	returnedID, returnedCompFunc, returnedErr := New(slog.Default(), mock).Save(ctx, name, author, tags)

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("asserting mock expectations: %v", err)
	}
	assertFailureResults(t, returnedErr, someErr, returnedID, returnedCompFunc)
}

func TestSave_FailCommit(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	ctx := t.Context()
	name := "name"
	author := "author"
	tags := []string{
		"tag1",
		"tag2",
	}

	id := int64(123)
	someErr := errors.New("some error")

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO audio (name, author) VALUES ($1, $2) RETURNING id;`).
		WithArgs(name, author).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(id))
	mock.ExpectExec(`INSERT INTO audio_tags (audio_id, tag) VALUES ($1, $2), ($3, $4);`).
		WithArgs(id, tags[0], id, tags[1]).
		WillReturnResult(pgxmock.NewResult("INSERT", 2))
	mock.ExpectCommit().
		WillReturnError(someErr)
	mock.ExpectRollback()

	returnedID, returnedCompFunc, returnedErr := New(slog.Default(), mock).Save(ctx, name, author, tags)

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("asserting mock expectations: %v", err)
	}
	assertFailureResults(t, returnedErr, someErr, returnedID, returnedCompFunc)
}

func TestSave_Success(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	ctx := t.Context()
	name := "name"
	author := "author"
	tags := []string{
		"tag1",
		"tag2",
	}

	id := int64(123)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO audio (name, author) VALUES ($1, $2) RETURNING id;`).
		WithArgs(name, author).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(id))
	mock.ExpectExec(`INSERT INTO audio_tags (audio_id, tag) VALUES ($1, $2), ($3, $4);`).
		WithArgs(id, tags[0], id, tags[1]).
		WillReturnResult(pgxmock.NewResult("INSERT", 2))
	mock.ExpectCommit()
	mock.ExpectRollback().WillReturnError(pgx.ErrTxClosed)

	storage := New(slog.Default(), mock)
	returnedID, returnedCompFunc, returnedErr := storage.Save(ctx, name, author, tags)

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("asserting mock expectations: %v", err)
	}

	if returnedErr != nil {
		t.Fatalf("expected err nil, got %v", returnedErr)
	}

	if returnedID != id {
		t.Fatalf("expected id %v, got %v", id, returnedID)
	}

	if returnedCompFunc == nil {
		t.Fatalf("unexpected nil compensation func")
	}

	mock.ExpectExec(`DELETE FROM audio WHERE id = $1;`).
		WithArgs(id).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	if returnCompErr := returnedCompFunc(ctx); returnCompErr != nil {
		t.Fatalf("expected compensation err nil, got %v", returnCompErr)
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("asserting mock expectations: %v", err)
	}
}
