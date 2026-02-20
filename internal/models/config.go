package models

import (
	"time"

	"gorm.io/gorm"
)

type ConfiguracaoSistema struct {
	ID                   uint           `gorm:"primaryKey" json:"id"`
	Chave                string         `gorm:"uniqueIndex;not null" json:"chave"`
	Valor                string         `gorm:"type:text" json:"valor"`
	Descricao            string         `gorm:"type:text" json:"descricao"`
	UsuarioAtualizacaoID *uint          `json:"usuarioAtualizacaoId"`
	UsuarioAtualizacao   *User          `gorm:"foreignKey:UsuarioAtualizacaoID" json:"usuarioAtualizacao,omitempty"`
	CreatedAt            time.Time      `json:"createdAt"`
	UpdatedAt            time.Time      `json:"updatedAt"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"-"`
}
