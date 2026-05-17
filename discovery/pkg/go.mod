module github.com/butbeautifulv/veil/discovery/pkg

go 1.25.0

require (
	github.com/butbeautifulv/veil/pkg v0.0.0
	github.com/butbeautifulv/veil/pkg/exec v0.0.0
)

replace github.com/butbeautifulv/veil/pkg => ../../pkg

replace github.com/butbeautifulv/veil/pkg/exec => ../../pkg/exec
