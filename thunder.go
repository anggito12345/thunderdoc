package thunderdoc

import (
	"bytes"
	"fmt"
	"net/http"
	"reflect"
	"text/template"
)

type ThunderDoc struct {
	html          string
	basicTemplate string
	rootTemplate  string
	Setup         struct {
		Version int
		Configs []ThunderDocConfig
	}
}

type ThunderDocConfig struct {
	ID         int
	Path       string
	Methods    []string
	Responses  []ThunderDocResponse
	Parameters ThunderDocRequest
}

type ThunderDocResponse struct {
	StatusCode   int
	DataAsString string
	Data         []prop
}

type ThunderDocRequest struct {
	DataAsString string
	Data         []prop
}

type config struct {
	Path          string
	Methods       []string
	RequestModel  interface{}
	ResponseModel []struct {
		StatusCode int
		Model      interface{}
	}
}

type prop struct {
	Name      string
	Type      string
	Required  bool
	Reference interface{}
}

func (t *ThunderDoc) New() *ThunderDoc {

	return &ThunderDoc{
		basicTemplate: `
		<!DOCTYPE html>
		<html>
		<body>
		
		{{ .Version }}
		<br>
		
		</body>
		</html>
		`,
	}
}

func (t *ThunderDoc) AddConfig(confs ...config) error {
	propRequest := []prop{}
	thunderResponses := []ThunderDocResponse{}

	for _, conf := range confs {
		reqFields := reflect.TypeOf(conf.RequestModel)
		for i := 0; i < reqFields.NumField(); i++ {
			propRequest = append(propRequest, prop{
				Name:     reqFields.Field(i).Name,
				Type:     reqFields.Field(i).Type.String(),
				Required: false,
			})
		}

		for _, responseModel := range conf.ResponseModel {

			propResponse := []prop{}

			resFields := reflect.TypeOf(responseModel.Model)
			for i := 0; i < resFields.NumField(); i++ {
				propResponse = append(propResponse, prop{
					Name:     reqFields.Field(i).Name,
					Type:     reqFields.Field(i).Type.String(),
					Required: false,
				})
			}

			thunderResponses = append(thunderResponses, ThunderDocResponse{
				StatusCode: responseModel.StatusCode,
				Data:       propResponse,
			})

		}

		t.Setup.Version = 0
		t.Setup.Configs = append(t.Setup.Configs, ThunderDocConfig{
			Methods: conf.Methods,
			Parameters: ThunderDocRequest{
				Data: propRequest,
			},
			Responses: thunderResponses,
			Path:      conf.Path,
		})
	}

	return nil
}

func (t *ThunderDoc) GenerateAsServe() (http.HandlerFunc, error) {

	tmpl, err := template.New("email template").Parse(t.basicTemplate)
	if err != nil {
		return nil, fmt.Errorf("error when parse body %v", err.Error())
	}

	var tpl bytes.Buffer

	err = tmpl.Execute(&tpl, t.Setup)
	if err != nil {
		return nil, fmt.Errorf("Error occured %v", err.Error())
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, tpl)
		return
	}, nil
}
