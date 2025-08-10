package service

import (
	"crypto/rand"

	"golang.org/x/crypto/bcrypt"
)

type HashService interface {
	GenerateFromPassword(password []byte) ([]byte, error)
	CompareHashAndPassword(hashedPassword, password []byte) error
	GenerateRandomBytes() ([]byte, error)
}

type hashService struct {
}

func NewHashService() HashService {
	return &hashService{}
}

const defaultTotalRandomBytesSize = 128

func (s *hashService) GenerateFromPassword(password []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
}

func (s *hashService) CompareHashAndPassword(hashedPassword, password []byte) error {
	return bcrypt.CompareHashAndPassword(hashedPassword, password)
}

func (s *hashService) GenerateRandomBytes() ([]byte, error) {
	buffer := make([]byte, defaultTotalRandomBytesSize)
	_, err := rand.Read(buffer)
	if err != nil {
		return nil, err
	}

	return buffer, nil
}
