module github.com/butbeautifulv/veil/engage/serve

go 1.25.0

require (
	github.com/butbeautifulv/veil/pkg/auth v0.0.0
	github.com/butbeautifulv/veil/pkg/engage v0.0.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/MicahParks/jwkset v0.11.0 // indirect
	github.com/MicahParks/keyfunc/v3 v3.8.0 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	golang.org/x/time v0.9.0 // indirect
)

replace (
	github.com/butbeautifulv/veil/pkg/auth => ../../pkg/auth
	github.com/butbeautifulv/veil/pkg/engage => ../../pkg/engage
)
