package scenario

import (
	"context"

	"github.com/chibiegg/isucon9-final/bench/internal/xrandom"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
)

func registerUserAndLogin(ctx context.Context, client *isutrain.Client) error {

	user, err := xrandom.GetRandomUser()
	if err != nil {
		return err
	}

	err = client.Signup(ctx, user.Email, user.Password, nil)
	if err != nil {
		return err
	}

	err = client.Login(ctx, user.Email, user.Password, nil)
	if err != nil {
		return err
	}

	return nil
}
