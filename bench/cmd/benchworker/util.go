package main

import "errors"

var (
	errTargetServerNotFound = errors.New("ベンチマーク対象サーバが見つかりませんでした")
)

func getTargetServer(job *Job) (*Server, error) {
	for _, server := range job.Team.Servers {
		if server.IsBenchTarget {
			return server, nil
		}
	}
	return nil, errTargetServerNotFound
}
