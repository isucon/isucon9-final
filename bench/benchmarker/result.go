package benchmarker

import (
	"bufio"
	"bytes"
	"sync"
	"sync/atomic"

	"github.com/chibiegg/isucon9-final/bench/internal/scenario"
)

type BenchResult struct {
	TotalScore uint64

	errMu  sync.RWMutex
	Errors []error

	// FIXME: Logs をログ書き込み先に指定
	Logs bytes.Buffer

	scenarioMu sync.RWMutex
	Scenarios  []scenario.Scenario
}

func NewBenchResult() *BenchResult {
	return &BenchResult{
		Logs: bytes.Buffer{},
	}
}

func (r *BenchResult) AddScore(score uint64) {
	atomic.AddUint64(&r.TotalScore, score)
}

func (r *BenchResult) ErrorCount() int {
	r.errMu.RLock()
	defer r.errMu.RUnlock()

	return len(r.Errors)
}

func (r *BenchResult) AddError(err error) {
	if err == nil {
		return
	}

	r.errMu.Lock()
	defer r.errMu.Unlock()

	r.Errors = append(r.Errors, err)

	// TODO: エラー回数許容チェック
}

func (r *BenchResult) LogLines() (logs []string, err error) {
	scanner := bufio.NewScanner(&r.Logs)
	for scanner.Scan() {
		logs = append(logs, scanner.Text())
	}
	err = scanner.Err()

	return
}

func (r *BenchResult) AddScenario(scenario scenario.Scenario) {
	r.scenarioMu.Lock()
	defer r.scenarioMu.Unlock()

	r.Scenarios = append(r.Scenarios, scenario)
}

// TODO: ScenarioUsers は、シナリオで用いられたユーザ一覧を返します
func (r *BenchResult) ScenarioUsers() {
	r.scenarioMu.RLock()
	defer r.scenarioMu.RUnlock()
}

// TODO: ActiveUserCount は、IsRetiredが真にならない、有効なユーザの数を数えます
func (r *BenchResult) ActiveUserCount() int {
	r.scenarioMu.RLock()
	defer r.scenarioMu.RUnlock()

	return 0
}
