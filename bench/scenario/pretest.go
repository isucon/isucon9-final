package scenario

import "errors"

// Guest訪問
var (
	ErrInitialTrainDatasetCount = errors.New("列車初期データセットの件数が一致しません")
)

// PreTestGuestVisitation はゲスト訪問を検証します
func PreTestGuestVisitation() error {
	// トップページに訪問可能

	// ~に訪問可能

	return nil
}

// PreTestAccountVisitation はログイン済みアカウントの訪問を検証します
func PreTestAccountVisitation() error {
	return nil
}

// PreTestInitialDataset は初期データセット件数を検証します
func PreTestInitialDataset() error {
	return nil
}
