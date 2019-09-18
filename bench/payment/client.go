package payment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

var (
	ErrInitializePayment = errors.New("決済情報の初期化に失敗しました")
	ErrPaymentResult     = errors.New("課金APIの処理結果を取得できませんでした")
)

type Client struct {
	BaseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
	}
}

func (c *Client) Initialize() error {
	uri := fmt.Sprintf("%s/initialize", c.BaseURL)

	req, err := http.NewRequest(http.MethodPost, uri, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ErrInitializePayment
	}

	return nil
}

func (c *Client) Result(ctx context.Context) (*PaymentResult, error) {
	lgr := zap.S()
	uri := fmt.Sprintf("%s/result", c.BaseURL)

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		lgr.Warnf("課金APIの結果取得でステータスコードが不正: %d", resp.StatusCode)
		return nil, ErrPaymentResult
	}

	var result *PaymentResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.IsOK {
		lgr.Warn("課金APIの結果でis_okがfalse")
		return nil, ErrPaymentResult
	}

	return result, nil
}
