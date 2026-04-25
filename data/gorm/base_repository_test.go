package gorm

import (
	"github.com/xudefa/go-boot/data"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type TestModel struct {
	data.BaseModel
	Name string
}

func (TestModel) TableName() string {
	return "test_models"
}

func TestBaseRepository_ImplementsRepository(t *testing.T) {
	is := assert.New(t)
	is.Implements((*data.Repository[TestModel])(nil), &BaseRepository[TestModel]{})
}

func TestGetTableName(t *testing.T) {
	is := assert.New(t)
	is.Equal("test_models", data.GetTableName[TestModel]())
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty", "", ""},
		{"single word", "User", "user"},
		{"two words", "UserName", "user_name"},
		{"three words", "UserNameTest", "user_name_test"},
		{"all caps", "USER", "u_s_e_r"},
		{"already snake", "user_name", "user_name"},
		{"mixed", "myUserName", "my_user_name"},
		{"single char", "A", "a"},
		{"single lower", "a", "a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := assert.New(t)
			result := data.ToSnakeCase(tt.input)
			is.Equal(tt.expected, result)
		})
	}
}

func TestBaseModel_TableName(t *testing.T) {
	is := assert.New(t)
	m := data.BaseModel{}
	is.Equal("", m.TableName())
}

func TestGetTableName_AutoSnakeCase(t *testing.T) {
	type AutoModel struct {
		ID   uint
		Name string
	}
	is := assert.New(t)
	is.Equal("auto_model", data.GetTableName[AutoModel]())
}

func TestBaseRepository_NewBaseRepository(t *testing.T) {
	is := assert.New(t)

	db, _ := gorm.Open(nil)
	repo := NewBaseRepository[TestModel](db)

	is.NotNil(repo)
}

func TestRepository_Interface(t *testing.T) {
	type NoTableModel struct {
		ID   uint
		Name string
	}

	var _ data.Repository[NoTableModel] = &BaseRepository[NoTableModel]{}
}
