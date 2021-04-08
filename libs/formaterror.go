package libs

import (
	"errors"
	"strings"
)

func Formaterror(err string) error {

	if strings.Contains(err, "email") {
		return errors.New("Email Already Taken.")
	}

	if strings.Contains(err, "hasedPassword") {
		return errors.New("Incorrect Password.")
	}

	return errors.New("Incorrect Details.")

}
