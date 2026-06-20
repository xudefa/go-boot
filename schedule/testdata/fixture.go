package schedule

import "context"

type FixtureService struct{}

func (s *FixtureService) DoJob(ctx context.Context) error {
	return nil
}

// @Scheduled(cron="0/5 * * * * ?")
func (s *FixtureService) Tick(ctx context.Context) error {
	return nil
}

// @Scheduled(cron="0 0 3 * * ?")
func (s *FixtureService) DailyRun(ctx context.Context) error {
	return nil
}

type NormalService struct{}

func (s *NormalService) Serve(ctx context.Context) error {
	return nil
}
