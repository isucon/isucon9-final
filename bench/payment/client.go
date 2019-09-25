package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/internal/config"
	"github.com/chibiegg/isucon9-final/bench/internal/endpoint"
)

var (
	ErrInitializePayment = errors.New("決済情報の初期化に失敗しました")
	ErrPaymentResult     = errors.New("課金APIの処理結果を取得できませんでした")
	ErrRegistCard        = errors.New("クレジットカードの登録及びトークン発行に失敗しました")
)

type Client struct {
	BaseURL *url.URL
}

func NewClient() (*Client, error) {
	u, err := url.Parse(config.PaymentBaseURL)
	if err != nil {
		return nil, err
	}

	return &Client{
		BaseURL: u,
	}, nil
}

func (c *Client) Initialize() error {
	u := *c.BaseURL
	u.Path = filepath.Join(u.Path, endpoint.PaymentInitializePath)

	req, err := http.NewRequest(http.MethodPost, u.String(), nil)
	if err != nil {
		return bencherror.InitializeErrs.AddError(bencherror.NewCriticalError(err, "課金APIへのinitializeリクエストが失敗しました. 運営に確認をお願いいたします"))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return bencherror.InitializeErrs.AddError(bencherror.NewCriticalError(err, "課金APIへのinitializeリクエストが失敗しました. 運営に確認をお願いいたします"))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return bencherror.InitializeErrs.AddError(bencherror.NewCriticalError(
			ErrPaymentResult,
			"課金APIへのinitialize時、不正なステータスコードが返却されました(got=%d, want=%d). 運営に確認をお願いいたします",
			resp.StatusCode,
			http.StatusOK,
		))
	}

	return nil
}

func (c *Client) RegistCard(ctx context.Context, cardNumber, cvv, expiryDate string) (string, error) {
	u := *c.BaseURL
	u.Path = filepath.Join(u.Path, endpoint.PaymentRegistCardPath)

	b, err := json.Marshal(map[string]*CardInformation{
		"card_information": &CardInformation{
			CardNumber: cardNumber,
			Cvv:        cvv,
			ExpiryDate: expiryDate,
		},
	})
	if err != nil {
		return "", bencherror.NewCriticalError(ErrRegistCard, "課金APIへのRegistCard時、Marshal処理で失敗しました. 運営に確認をお願いいたします")
	}

	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewBuffer(b))
	if err != nil {
		return "", bencherror.NewCriticalError(ErrRegistCard, "課金APIにクレジットカードを登録できませんでした. 運営に確認をお願いいたします")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", bencherror.NewCriticalError(ErrRegistCard, "課金APIにクレジットカードを登録できませんでした. 運営に確認をお願いいたします")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", bencherror.NewCriticalError(
			ErrRegistCard,
			"クレジットカードの登録時、不正なステータスコードが返却されました(got=%d, want=%d). 運営に確認をお願いいたします",
			resp.StatusCode,
			http.StatusOK,
		)
	}

	registCardResp := &RegistCardResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&registCardResp); err != nil {
		return "", bencherror.NewCriticalError(ErrRegistCard, "課金APIにクレジットカードを登録できませんでした. 運営に確認をお願いいたします")
	}

	if !registCardResp.IsOK {
		return "", bencherror.NewCriticalError(ErrRegistCard, "課金APIがIsOK=falseでレスポンスを返しました. 運営に確認をお願いいたします")
	}

	return registCardResp.CardToken, nil
}

func (c *Client) Result(ctx context.Context) (*PaymentResult, error) {
	u := *c.BaseURL
	u.Path = filepath.Join(u.Path, endpoint.PaymentResultPath)

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, bencherror.NewCriticalError(err, "課金APIから決済結果を取得できませんでした. 運営に確認をお願いいたします")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, bencherror.NewCriticalError(err, "課金APIへのリクエストに失敗しました. 運営に確認をお願いいたします")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, bencherror.NewCriticalError(
			ErrPaymentResult,
			"課金APIから決済結果取得時、不正なステータスコード(got=%d, want=%d)が返却されました. 運営に確認をお願いいたします",
			resp.StatusCode,
			http.StatusOK,
		)
	}

	result := &PaymentResult{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, bencherror.NewCriticalError(err, "課金APIのレスポンスが不正です.運営に確認をお願いいたします")
	}

	if !result.IsOK {
		return nil, bencherror.NewCriticalError(ErrPaymentResult, "課金APIがIsOK=falseでレスポンスを返しました. 運営に確認をお願いいたします")
	}

	return result, nil
}
