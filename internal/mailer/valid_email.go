package mailer

import (
	"errors"
)

func (em *EmailInput) Validate() error {
	if em.To == "" {
		return errors.New("empty to")
	}

	if em.Subject == "" || em.Body == "" {
		return errors.New("empty subject/body")
	}

	return nil
}
