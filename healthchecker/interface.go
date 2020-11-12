package healthchecker

type HealthChecker interface {
	Fail()
	Ok()
}
