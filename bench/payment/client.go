package payment

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/chibiegg/isucon9-final/bench/internal/bencherror"
	"github.com/chibiegg/isucon9-final/bench/internal/config"
)

var (
	ErrInitializePayment = errors.New("決済情報の初期化に失敗しました")
	ErrPaymentResult     = errors.New("課金APIの処理結果を取得できませんでした")
)

type Client struct {
	BaseURL *url.URL
}

func NewClient(baseURL string) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	return &Client{
		BaseURL: u,
	}, nil
}

func (c *Client) Initialize() error {
	u := c.BaseURL
	u.Path = filepath.Join(u.Path, config.PaymentInitializePath)

	req, err := http.NewRequest(http.MethodPost, u.String(), nil)
	if err != nil {
		return bencherror.NewCriticalError(err, "課金APIへのリクエストが失敗しました. 運営に確認をお願いいたします")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return bencherror.NewCriticalError(err, "課金APIへのリクエストが失敗しました. 運営に確認をお願いいたします")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return bencherror.NewCriticalError(ErrPaymentResult, "課金APIから決済結果取得時、不正なステータスコード(=%d)が返却されました. 運営に確認をお願いいたします", resp.StatusCode)
	}

	return nil
}

func (c *Client) Result(ctx context.Context) (*PaymentResult, error) {
	u := c.BaseURL
	u.Path = filepath.Join(u.Path, config.PaymentResultPath)

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		paymentErr := bencherror.NewCriticalError(err, "課金APIから決済結果を取得できませんでした. 運営に確認をお願いいたします")
		bencherror.InitializeErrs.AddError(paymentErr)
		return nil, paymentErr
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		paymentErr := bencherror.NewCriticalError(err, "課金APIへのリクエストに失敗しました. 運営に確認をお願いいたします")
		bencherror.InitializeErrs.AddError(paymentErr)
		return nil, paymentErr
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		paymentErr := bencherror.NewCriticalError(ErrPaymentResult, "課金APIから決済結果取得時、不正なステータスコード(=%d)が返却されました. 運営に確認をお願いいたします", resp.StatusCode)
		bencherror.InitializeErrs.AddError(paymentErr)
		return nil, paymentErr
	}

	var result *PaymentResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		paymentErr := bencherror.NewCriticalError(err, "課金APIのレスポンスが不正です.運営に確認をお願いいたします")
		bencherror.InitializeErrs.AddError(paymentErr)
		return nil, paymentErr
	}

	if !result.IsOK {
		paymentErr := bencherror.NewCriticalError(ErrPaymentResult, "課金APIで処理が失敗しました. 運営に確認をお願いいたします")
		bencherror.InitializeErrs.AddError(paymentErr)
		return nil, paymentErr
	}

	return result, nil
}
