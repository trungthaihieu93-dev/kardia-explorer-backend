// Package private
package private

type Staking interface {
	ReloadValidators() error
}

func (s *service) ReloadValidators() error {
	return nil
}
