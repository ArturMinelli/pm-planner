package api

import (
	"errors"
	"testing"
)

func TestUserFacingErrorParsesAPIBody(t *testing.T) {
	err := &HTTPStatusError{
		StatusCode: 422,
		Body:       `{"errors":["Já existe um registro neste minuto."]}`,
	}
	if got := UserFacingError(err); got != "Já existe um registro neste minuto." {
		t.Fatalf("message: %q", got)
	}
}

func TestUserFacingErrorWrapsRegisterFailure(t *testing.T) {
	wrapped := errors.New(UserFacingError(&HTTPStatusError{
		StatusCode: 403,
		Body:       `{}`,
	}))
	if wrapped.Error() != "sessão expirada ou sem permissão. Verifique login em Configurações." {
		t.Fatalf("message: %q", wrapped.Error())
	}
}
