package server

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/cisco-eti/sre-go-helloworld/pkg/models"
	"github.com/cisco-eti/sre-go-helloworld/pkg/utils"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Get godoc
// @Summary Get Home
// @Description get a response from home endpoint
// @ID get-home
// @Tags Home
// @Error 401
// @Router / [get]
func (s *Server) RootHandler(w http.ResponseWriter, _ *http.Request) {
	s.log.Info("/ request received")

	_ = utils.OKResponse(w, "root")
	return
}

// Get godoc
// @Summary Get Ping
// @Description get helloworld status
// @Produce json
// @Success 200 {object} models.PingResponse
// @Router /ping [get]
func (s *Server) PingHandler(w http.ResponseWriter, r *http.Request) {
	s.log.Info("/ping request received")

	hostname := os.Getenv("HOSTNAME")
	if hostname == "" {
		hostname = "HOSTNAME env missing"
	}

	// Pure sample data, value need to be changed
	myService := models.Service{
		ServiceName:  "HelloWorld",
		ServiceType:  "OPTIONAL",
		ServiceState: "online",
		Message:      "Healthy",
		ServiceInstance: models.ServiceInstance{
			InstanceID: hostname,
			Host:       "172.18.231.5",
			Port:       21455,
		},
		LastUpdated:    "2020-10-20T08:42:07.290Z",
		BaseURL:        "http://helloworld.int.scratch-aws-1.prod.eticloud.io/",
		DurationPretty: "91ms",
		Duration:       91350005,
		UpstreamServices: []models.Service{
			{
				ServiceName:  "Postgres",
				ServiceType:  "REQUIRED",
				ServiceState: "online",
				ServiceInstance: models.ServiceInstance{
					InstanceID: "e3c16830-2c65-c6c8-68ab-30d728d6179e[9]",
					Host:       "172.18.244.25",
					Port:       5432,
				},
				Message:          "PostgresDataSource is online",
				LastUpdated:      "2021-01-06T13:43:42.984Z",
				DurationPretty:   "3ms",
				Duration:         3631684,
				UpstreamServices: []models.Service{},
				DefaultCharset:   "UTF-8",
			},
		},
		DefaultCharset: "UTF-8",
	}
	_ = utils.OKResponse(w, myService)
	return
}

// Get godoc
// @Summary Get API Docs
// @Description get Swagger API documentation
// @Produce Yaml
// @Success 200
// @Router /docs [get]
func (s *Server) DocsHandler(w http.ResponseWriter, r *http.Request) {
	s.log.Info("/docs request received")

	filePath, err := filepath.Abs("./docs/openapi.yaml")
	if err != nil {
		utils.ServerErrorResponse(w, err.Error())
		return
	}

	s.log.Info("filePath: %s", filePath)
	http.ServeFile(w, r, filePath)
}

// Get godoc
// @Summary Get Prometheus Metrics
// @Description get helloworld status
// @Produce Yaml
// @Success 200
// @Router /metrics [get]
func (s *Server) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	s.log.Info("/metrics request received")

	promhttp.Handler().ServeHTTP(w, r)
}

type PageData struct {
	Message string
}

// Get godoc
// @Summary Get S3 Test result
// @Description get S3 Test page
// @Produce json
// @Success 200
// @Router /s3 [get]
func (s *Server) S3Handler(w http.ResponseWriter, r *http.Request) {
	s.log.Info("/s3 request received")
	var (
		bucket      string = os.Getenv("S3_BUCKET")
		key         string = "sre-go-helloworld-s3-test"
		filename    string = "s3_object.txt"
		web_message string = ""
	)
	renderTemplate := func(msg string) {
		s.log.Info("web_message: %s", msg)
		data := PageData{
			Message: msg,
		}
		tmpl := template.Must(template.ParseFiles("./web/s3.html"))
		tmpl.Execute(w, data)
	}

	// All clients require a Session. The Session provides the client with
	// shared configuration such as region, endpoint, and credentials. A
	// Session should be shared where possible to take advantage of
	// configuration and credential caching. See the session package for
	// more information.
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION_OVERRIDE")),
	}))

	// Create an uploader with the session and default options
	uploader := s3manager.NewUploader(sess)

	// Create a context with a timeout that will abort the upload if it takes
	// more than the passed in timeout.
	timeout, _ := time.ParseDuration("1m")
	ctx := context.Background()
	var cancelFn func()
	if timeout > 0 {
		ctx, cancelFn = context.WithTimeout(ctx, timeout)
	}
	// Ensure the context is canceled to prevent leaking.
	// See context package for more information, https://golang.org/pkg/context/
	if cancelFn != nil {
		defer cancelFn()
	}

	// Uploads the object to S3. The Context will interrupt the request if the
	// timeout expires.
	f, err := os.Open("/" + filename)
	if err != nil {
		s.log.Error("failed to open file %q, %v", filename, err)
		web_message = "An error occurred while trying to open the file to upload. Check logs for details."
		renderTemplate(web_message)
		return
	}
	result, err := uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   f,
	})
	if err != nil {
		s.log.Error("failed to upload file to S3, %v", err)
		web_message = "An error occurred while trying to upload to S3. Check logs for details."
		renderTemplate(web_message)
		return
	}
	s.log.Info("file successfully uploaded to: %s", result.Location)

	if web_message == "" {
		web_message = fmt.Sprintf("Successfully uploaded file to S3 at %s", result.Location)
	}
	renderTemplate(web_message)
}

// Get godoc
// @Summary GetCallerIdentity result
// @Description get GetCallerIdentity page
// @Produce html
// @Success 200
// @Router /gci [get]
func (s *Server) GciHandler(w http.ResponseWriter, r *http.Request) {
	s.log.Info("/gci request received")
	var web_message string
	renderTemplate := func(msg string) {
		s.log.Debug("web_message: %s", msg)
		data := PageData{
			Message: msg,
		}
		tmpl := template.Must(template.ParseFiles("./web/gci.html"))
		tmpl.Execute(w, data)
	}

	mySession := session.Must(session.NewSession())
	svc := sts.New(mySession)
	input := &sts.GetCallerIdentityInput{}

	result, err := svc.GetCallerIdentity(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			s.log.Error(aerr.Error())
			web_message = fmt.Sprintf("Error: %s\n", aerr.Error())
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			s.log.Error(err.Error())
			web_message = fmt.Sprintf("Error: %s\n", err.Error())
		}
	} else {
		s.log.Info("GetCallerIdentityOutput: %+v", result)
		resultJSON, err := json.MarshalIndent(result, "", "    ")
		if err != nil {
			s.log.Error(err.Error())
			web_message = fmt.Sprintf("%+v", result)
		} else {
			web_message = fmt.Sprint(string(resultJSON))
		}
	}
	renderTemplate(web_message)
}
