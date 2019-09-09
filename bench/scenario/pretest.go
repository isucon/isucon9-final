package scenario

import "errors"

// Guest訪問
var (
	ErrInitialTrainDatasetCount = errors.New("列車初期データセットの件数が一致しません")
)

func PreTest() error {
	if err := preTestGuestVisitation(); err != nil {
		return err
	}

	if err := preTestAccountVisitation(); err != nil {
		return err
	}

	if err := preTestInitialDataset(); err != nil {
		return err
	}

	return nil
}

// PreTestGuestVisitation はゲスト訪問を検証します
func preTestGuestVisitation() error {
	// トップページに訪問可能

	// ~に訪問可能

	return nil
}

// PreTestAccountVisitation はログイン済みアカウントの訪問を検証します
func preTestAccountVisitation() error {
	return nil
}

// PreTestInitialDataset は初期データセット件数を検証します
func preTestInitialDataset() error {
	return nil
}
