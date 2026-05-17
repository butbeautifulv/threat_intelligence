package job

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	domain "github.com/butbeautifulv/veil/pkg/engage/domain/job"
	"github.com/nats-io/nats.go"
)

const (
	natsStreamName    = "ENGAGE_JOBS"
	natsSubjectPending = "engage.jobs.pending"
	natsKVBucket      = "engage_jobs"
	natsConsumer      = "engage-worker"
)

// NATSStore persists jobs in JetStream KV and dispatches work via engage.jobs.pending.
type NATSStore struct {
	nc  *nats.Conn
	js  nats.JetStreamContext
	kv  nats.KeyValue
	sub *nats.Subscription
}

func NewNATSStore(url string) (*NATSStore, error) {
	if url == "" {
		return nil, fmt.Errorf("nats url required")
	}
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}
	js, err := nc.JetStream()
	if err != nil {
		_ = nc.Drain()
		return nil, err
	}
	if err := ensureEngageStream(js); err != nil {
		_ = nc.Drain()
		return nil, err
	}
	kv, err := js.CreateKeyValue(&nats.KeyValueConfig{Bucket: natsKVBucket})
	if err != nil {
		kv, err = js.KeyValue(natsKVBucket)
		if err != nil {
			_ = nc.Drain()
			return nil, fmt.Errorf("kv bucket: %w", err)
		}
	}
	s := &NATSStore{nc: nc, js: js, kv: kv}
	if err := s.ensureConsumer(); err != nil {
		_ = nc.Drain()
		return nil, err
	}
	return s, nil
}

func ensureEngageStream(js nats.JetStreamContext) error {
	_, err := js.StreamInfo(natsStreamName)
	if err == nats.ErrStreamNotFound {
		_, err = js.AddStream(&nats.StreamConfig{
			Name:     natsStreamName,
			Subjects: []string{"engage.jobs.>"},
			Storage:  nats.FileStorage,
		})
	}
	return err
}

func (s *NATSStore) ensureConsumer() error {
	_, err := s.js.AddConsumer(natsStreamName, &nats.ConsumerConfig{
		Durable:       natsConsumer,
		FilterSubject: natsSubjectPending,
		AckPolicy:     nats.AckExplicitPolicy,
		MaxDeliver:    5,
	})
	if err != nil && err != nats.ErrConsumerNameAlreadyInUse {
		return err
	}
	sub, err := s.js.PullSubscribe(natsSubjectPending, natsConsumer, nats.Bind(natsStreamName, natsConsumer))
	if err != nil {
		return err
	}
	s.sub = sub
	return nil
}

func (s *NATSStore) Close() {
	if s.nc != nil {
		_ = s.nc.Drain()
	}
}

func (s *NATSStore) Put(j *domain.Job) error {
	if j == nil || j.ID == "" {
		return fmt.Errorf("invalid job")
	}
	data, err := json.Marshal(j)
	if err != nil {
		return err
	}
	if _, err := s.kv.Put(j.ID, data); err != nil {
		return err
	}
	if j.Status == domain.StatusPending {
		_, err = s.js.Publish(natsSubjectPending, []byte(j.ID), nats.MsgId(j.ID))
	}
	return err
}

func (s *NATSStore) Get(id string) (*domain.Job, bool) {
	entry, err := s.kv.Get(id)
	if err != nil {
		return nil, false
	}
	var j domain.Job
	if json.Unmarshal(entry.Value(), &j) != nil {
		return nil, false
	}
	return &j, true
}

func (s *NATSStore) ListByStatus(status domain.Status, limit int) ([]*domain.Job, error) {
	keys, err := s.kv.Keys()
	if err != nil {
		return nil, err
	}
	var out []*domain.Job
	for _, id := range keys {
		j, ok := s.Get(id)
		if !ok {
			continue
		}
		if status != "" && j.Status != status {
			continue
		}
		out = append(out, j)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	sortJobsByCreated(out)
	return trimLimit(out, limit), nil
}

func (s *NATSStore) ListPending() ([]*domain.Job, error) {
	return s.ListByStatus(domain.StatusPending, 0)
}

func (s *NATSStore) CountByStatus(status domain.Status) (int, error) {
	list, err := s.ListByStatus(status, 0)
	return len(list), err
}

func (s *NATSStore) TryClaim(id string) (*domain.Job, bool) {
	entry, err := s.kv.Get(id)
	if err != nil {
		return nil, false
	}
	var j domain.Job
	if json.Unmarshal(entry.Value(), &j) != nil || j.Status != domain.StatusPending {
		return nil, false
	}
	j.Status = domain.StatusRunning
	j.UpdatedAt = time.Now().UTC()
	data, err := json.Marshal(j)
	if err != nil {
		return nil, false
	}
	if _, err := s.kv.Update(id, data, entry.Revision()); err != nil {
		return nil, false
	}
	return &j, true
}

// FetchPending pulls the next pending job id from JetStream (caller should TryClaim).
func (s *NATSStore) FetchPending(ctx context.Context, timeout time.Duration) (string, *nats.Msg, error) {
	if s.sub == nil {
		return "", nil, fmt.Errorf("nats consumer not ready")
	}
	msgs, err := s.sub.Fetch(1, nats.MaxWait(timeout))
	if err != nil {
		return "", nil, err
	}
	if len(msgs) == 0 {
		return "", nil, nats.ErrTimeout
	}
	m := msgs[0]
	return string(m.Data), m, nil
}

func (s *NATSStore) Ack(msg *nats.Msg) error {
	return msg.Ack()
}

func (s *NATSStore) Nak(msg *nats.Msg) error {
	return msg.Nak()
}
