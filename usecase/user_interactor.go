package usecase

import "github.com/lottotto/stdgrpc/model"

type UserInteract struct {
	repository UserRepository
}

type UserRepository interface {
	update(string) error
	findByName(string) ([]model.User, error)
}

func (i *UserInteract) addUser(name string) error {
	err := i.repository.update(name)
	return err
}

func (i *UserInteract) listUserByName(name string) ([]model.User, error) {
	users, err := i.repository.findByName(name)
	return users, err
}
