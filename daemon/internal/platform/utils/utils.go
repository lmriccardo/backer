package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

type WithDefault interface {
	ApplyDefault()
}

func createDir(path string) (string, error) {
	if err := os.MkdirAll(path, 0o700); err != nil {
		return "", fmt.Errorf("mkdir %s: %w", path, err)
	}
	return path, nil
}

// UserHome Returns the home folder of the current user
func UserHome() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return "", fmt.Errorf("unable to resolve home dir: %w", err)
	}
	return home, nil
}

// BackerHome Returns the default path for the
// backer folder where all configs, plans and registry
// are saved for the daemon to work.
func BackerHome() (string, error) {
	home, err := UserHome()
	if err != nil {
		return "", err
	}
	return createDir(filepath.Join(home, ".backer"))
}

// BackerLogHome Returns the default path where logs
// produced by runs are saved into
func BackerLogHome() (string, error) {
	home, err := BackerHome()
	if err != nil {
		return "", err
	}
	return createDir(filepath.Join(home, "log"))
}

// BackerExcludeHome returns the exclude home folder for
// any job configuration if the exclude output path is
// not given at creation time
func BackerExcludeHome() (string, error) {
	home, err := BackerHome()
	if err != nil {
		return "", err
	}
	return createDir(filepath.Join(home, "excludes"))
}

// RegistryFile Returns the path to the registry file
func RegistryFile() (string, error) {
	home, err := BackerHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "registry.db"), nil
}

// Exist checks if a file exists in the current machine
func Exist(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ToJson unmarshal the input string into the generic Type.
// If the type is not marshable, or json encodable then it
// returns an error.
func ToJson[T any](jsonStr string) (T, error) {
	var object T // Create the object of type T
	err := json.Unmarshal([]byte(jsonStr), &object)
	return object, err
}

func ToJsonWithObj[T any](dst *T, jsonStr string) error {
	err := json.Unmarshal([]byte(jsonStr), dst)
	return err
}

// UnmarshalJSONWithDefault is an helper function that performs classical
// json unmarshaling with the addition of DefaultField implemented
// interface types.
func UnmarshalJSONWithDefault[T any](dst *T, data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return err
	}

	return nil
}

// MustHaveBody reads all bytes from the request body and returns
// then with an error. If there are no bytes an error is returned.
func MustHaveBody(req *http.Request) ([]byte, error) {
	data, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("unabled to read body: %v", err)
	}

	if len(data) < 1 {
		return nil, fmt.Errorf("request must have body")
	}

	return data, nil
}

// applyDefaultR is a helper function for ApplyDefault
func applyDefaultR(vf reflect.Value) {
	if vf.Type().Kind() != reflect.Struct {
		return
	}

	if d, ok := vf.Addr().Interface().(WithDefault); ok {
		d.ApplyDefault()
	}

	for i := 0; i < vf.NumField(); i++ {
		field := vf.Field(i)
		if field.Kind() == reflect.Pointer {
			if field.IsNil() {
				continue
			}

			field = field.Elem()
		}

		applyDefaultR(field)
	}
}

// ApplyDefault recursively applies default values to obj. If the input object
// is nil, ApplyDefaults does nothing. This functions walks exported fields of
// obj recursively. For every struct encountered that implements WithDefault
// interface, its ApplyDefault method is invoked. The object is modified in-place.
func ApplyDefault[T any](obj *T) {
	if obj == nil {
		return
	}
	applyDefaultR(reflect.ValueOf(obj).Elem())
}

// MustBindWithJSON tries to unmarshal the JSON from a request
// into the input structure. It returns an error if there are
// no bytes in to the request body and if the decoding procedure
// has returned an error as well.
func MustBindWithJSON[T any](dst *T, req *http.Request) error {
	data, err := MustHaveBody(req)
	if err != nil {
		return err
	}

	// First perform simple unmarshaling of the data into the struct
	if err := UnmarshalJSONWithDefault(dst, data); err != nil {
		return err
	}

	// Finally, Apply default values
	ApplyDefault(dst)
	return nil
}

// GetDomainFromEmail splits the email string by the @ sign and
// returns the "right-hand side" of the split.
func GetDomainFromEmail(email string) string {
	return strings.Split(email, "@")[1]
}

// JSONToString marshals an input json structure into a string
func JSONToString[T any](in *T) string {
	data, _ := json.Marshal(in)
	return string(data)
}
