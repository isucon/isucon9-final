package benchmarker

import (
	"context"
	"log"
	"time"

	"github.com/chibiegg/isucon9-final/bench/internal/consts"
	"github.com/chibiegg/isucon9-final/bench/internal/logger"
	"github.com/chibiegg/isucon9-final/bench/internal/scenario"
	"github.com/chibiegg/isucon9-final/bench/isutrain"
	"golang.org/x/sync/errgroup"
)

type Benchmarker struct {
	BaseURL string

	eg *errgroup.Group

	benchResult *BenchResult
	resultCh    chan chan *scenario.RequestResult

	client *isutrain.Isutrain

	// level はベンチ負荷レベルです
	level int

	// tick はシナリオ生成の際、どのシナリオを用いるか決定する際の判断材料として用います
	tick uint64
}

func NewBenchmarker(baseURL string) *Benchmarker {
	benchmarker := &Benchmarker{
		BaseURL:     baseURL,
		benchResult: NewBenchResult(),
		isutrain:    isutrain.NewIsutrain(baseURL),
		resultCh:    make(chan chan *scenario.RequestResult),
	}

	logger.InitBufferedZapLogger(&benchmarker.benchResult.Logs)

	return benchmarker
}

func (b *Benchmarker) Initialize(ctx context.Context) error {
	return nil
}

// PreTest は、ベンチマーク実行前の整合性チェックを実施します
func (b *Benchmarker) PreTest(ctx context.Context) error {
	return nil
}

// PostTest は、ベンチマーク実行後の整合性チェックを実施します
func (b *Benchmarker) PostTest(ctx context.Context) error {
	return nil
}

// // Bench はベンチマークを実施します
// func (b *Benchmarker) Bench(ctx context.Context) error {
// 	atomic.AddUint64(&b.tick, 1)
// 	s := scenario.NewScenario(b.tick, b.BaseURL)

// 	// FIXME: 初期並列度だけgoroutine生成
// 	err := s.Run(ctx)
// 	if err != nil {
// 		return err
// 	}

// 	var sum int
// 	for requestResult := range s.RequestResultStream() {
// 		sum += requestResult.Type.Score()
// 		if requestResult.Err != nil {
// 			log.Printf("error = %+v\n", requestResult.Err)
// 		}
// 	}

// 	log.Printf("score = %d\n", sum)

// 	return nil
// }

// increaseWorkload はクライアントを非同期実行で生成し、ウェブアプリへの負荷を高めます
func (b *Benchmarker) increaseWorkload() {
	for i := 0; i < consts.DefaultConcurrency; i++ {

	}
}

func (b *Benchmarker) Run(ctx context.Context) error {
	defer close(b.resultCh)

	ctx, cancel := context.WithTimeout(ctx, consts.BenchmarkTimeout)
	defer cancel()

	// initialize
	resp, err := b.isutrain.Initialize(ctx)
	if err != nil {
		return err
	}

	// pretest
	if err := b.Pretest(ctx); err != nil {
		return err
	}

	levelTicker := time.NewTicker(consts.BenchmarkerLevelInterval)
	defer levelTicker.Stop()

	for {
		select {
		case <-ctx.Done():
		case requestResultStream := <-b.resultCh:
			go func() {
				for {
					select {
					case <-ctx.Done():
					case requestResult := <-requestResultStream:
						// NOTE: リクエストごとrequestResultが送られてくるので、非同期化しないと詰まると思う
						log.Println(requestResult)
					}
				}
			}()
		case <-levelTicker.C:
			// TODO: レベル上げ処理
			// case <-checkBusySeason:
			// TODO: 繁忙期チェックによる負荷増大をイベントドリブンに処理する
		}
	}

	// posttest
	if err := b.Posttest(ctx); err != nil {
		return err
	}

	return nil
}
