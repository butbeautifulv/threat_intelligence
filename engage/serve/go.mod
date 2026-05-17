module github.com/butbeautifulv/veil/engage/serve

go 1.25.0

require (
	github.com/alicebob/miniredis/v2 v2.38.0
	github.com/butbeautifulv/veil/pkg v0.0.0
	github.com/butbeautifulv/veil/pkg/auth v0.0.0
	github.com/butbeautifulv/veil/pkg/engage v0.0.0
	github.com/jackc/pgx/v5 v5.7.2
	github.com/jung-kurt/gofpdf v1.16.2
	github.com/nats-io/nats-server/v2 v2.10.29
	github.com/nats-io/nats.go v1.48.0
	github.com/prometheus/client_golang v1.20.5
	github.com/redis/go-redis/v9 v9.19.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/MicahParks/jwkset v0.11.0 // indirect
	github.com/MicahParks/keyfunc/v3 v3.8.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/minio/highwayhash v1.0.3 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/nats-io/jwt/v2 v2.7.4 // indirect
	github.com/nats-io/nkeys v0.4.11 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.55.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/yuin/gopher-lua v1.1.1 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/crypto v0.48.0 // indirect
	golang.org/x/sync v0.19.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	golang.org/x/time v0.10.0 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
)

replace (
	github.com/butbeautifulv/veil/pkg => ../../pkg
	github.com/butbeautifulv/veil/pkg/auth => ../../pkg/auth
	github.com/butbeautifulv/veil/pkg/engage => ../../pkg/engage
)
