package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID       uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	Username string    `gorm:"uniqueIndex;not null" json:"username" binding:"required"`
	Email    string    `gorm:"uniqueIndex;not null" json:"email" binding:"required"`
	Password string    `gorm:"not null" json:"password" binding:"required"`
}

func (user *User) BeforeCreate(tx *gorm.DB) (err error) {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	return
}

type Match struct {
	ID       uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	HomeTeam string    `gorm:"not null" json:"home_team" binding:"required"`
	AwayTeam string    `gorm:"not null" json:"away_team" binding:"required"`
	Date     string    `gorm:"not null" json:"date" binding:"required"`
	League   string    `gorm:"not null" json:"league" binding:"required"`
	Season   string    `gorm:"not null" json:"season" binding:"required"`
	MatchDay int       `gorm:"not null" json:"match_day" binding:"required"`
	Result   string    `json:"result"` // Stores the full-time score in "home:away" format
}

func (match *Match) BeforeCreate(tx *gorm.DB) (err error) {
	if match.ID == uuid.Nil {
		match.ID = uuid.New()
	}
	return
}

type Prediction struct {
	ID                 uuid.UUID `gorm:"type:char(36);primaryKey" json:"id"`
	UserID             uuid.UUID `gorm:"type:char(36);not null;index:idx_user_match,unique" json:"user_id"`
	MatchID            uuid.UUID `gorm:"type:char(36);not null;index:idx_user_match,unique" json:"match_id"`
	PredictedScoreHome int       `gorm:"not null" json:"predicted_score_home" binding:"required"`
	PredictedScoreAway int       `gorm:"not null" json:"predicted_score_away" binding:"required"`
	Points             int       `gorm:"default:0" json:"points"`
	User               User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"-"`
	Match              Match     `gorm:"foreignKey:MatchID;constraint:OnDelete:CASCADE" json:"-"`
}

func (prediction *Prediction) BeforeCreate(tx *gorm.DB) (err error) {
	if prediction.ID == uuid.Nil {
		prediction.ID = uuid.New()
	}
	return
}