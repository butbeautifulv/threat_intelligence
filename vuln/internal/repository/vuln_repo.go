package repository

type VulnRepository interface {
    Save(ctx context.Context, v *Vulnerability) error
    FindByCVE(ctx context.Context, id string) (*Vulnerability, error)
}
