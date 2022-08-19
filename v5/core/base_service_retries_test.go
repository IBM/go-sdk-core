//go:build all || slow || basesvc || retries
// +build all slow basesvc retries

package core

// (C) Copyright IBM Corp. 2020, 2022.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var requestCountMutex sync.Mutex

// assertResponse is a convenience function for checking various parts of a response
func assertResponse(r *DetailedResponse, expectedStatusCode int, expectedContentType string) {
	Expect(r).ToNot(BeNil())
	Expect(r.StatusCode).To(Equal(expectedStatusCode))
	if expectedContentType != "" {
		Expect(r.Headers.Get("Content-Type")).To(Equal(expectedContentType))
	}
}

// clientInit is a convenience function for setting up a new service and request builder
func clientInit(method string, url string, maxRetries int, maxIntervalSecs int) (service *BaseService, builder *RequestBuilder) {
	var err error
	options := &ServiceOptions{
		URL:           url,
		Authenticator: &NoAuthAuthenticator{},
	}
	service, err = NewBaseService(options)
	Expect(err).To(BeNil())
	if maxRetries > 0 {
		service.EnableRetries(maxRetries, time.Duration(maxIntervalSecs)*time.Second)
	}
	builder = NewRequestBuilder(method)
	_, err = builder.ConstructHTTPURL(url, nil, nil)
	Expect(err).To(BeNil())
	return
}

var _ = Describe(`Retry scenarios`, func() {
	var server *httptest.Server

	BeforeEach(func() {
		goLogger := log.New(GinkgoWriter, "", log.LstdFlags)
		SetLogger(NewLogger(LevelDebug, goLogger, goLogger))
	})

	Describe(`Error scenarios`, func() {
		Describe(`Timeout errors`, func() {
			var requestCount int
			BeforeEach(func() {
				requestCount = 0
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					defer GinkgoRecover()
					time.Sleep(1 * time.Second)

					requestCountMutex.Lock()
					requestCount++
					requestCountMutex.Unlock()

					w.Header().Set("Retry-After", "1")
					w.WriteHeader(http.StatusTooManyRequests)
				}))
			})
			AfterEach(func() {
				server.Close()
			})
			It(`Timeout with retries disabled`, func() {
				service, builder := clientInit("GET", server.URL, 0, 0)
				ctx, cancelFunc := context.WithTimeout(context.Background(), 500*time.Millisecond)
				defer cancelFunc()
				builder.WithContext(ctx)
				req, _ := builder.Build()

				var foo *Foo
				resp, err := service.Request(req, &foo)
				Expect(err).ToNot(BeNil())
				fmt.Fprintf(GinkgoWriter, "Expected error: %s\n", err.Error())
				Expect(err.Error()).To(ContainSubstring("context deadline exceeded"))
				Expect(resp).To(BeNil())
				requestCountMutex.Lock()
				Expect(requestCount).To(Equal(0))
				requestCountMutex.Unlock()
			})
			It(`Timeout on initial request`, func() {
				service, builder := clientInit("GET", server.URL, 2, 0)
				ctx, cancelFunc := context.WithTimeout(context.Background(), 500*time.Millisecond)
				defer cancelFunc()
				builder.WithContext(ctx)
				req, _ := builder.Build()

				var foo *Foo
				resp, err := service.Request(req, &foo)
				Expect(err).ToNot(BeNil())
				fmt.Fprintf(GinkgoWriter, "Expected error: %s\n", err.Error())
				Expect(err.Error()).To(ContainSubstring("context deadline exceeded"))
				Expect(resp).To(BeNil())
				requestCountMutex.Lock()
				Expect(requestCount).To(Equal(0))
				requestCountMutex.Unlock()
			})
			It(`Timeout while doing retries`, func() {
				service, builder := clientInit("GET", server.URL, 2, 0)
				ctx, cancelFunc := context.WithTimeout(context.Background(), 3500*time.Millisecond)
				defer cancelFunc()
				builder.WithContext(ctx)
				req, _ := builder.Build()

				var foo *Foo
				resp, err := service.Request(req, &foo)
				Expect(err).ToNot(BeNil())
				fmt.Fprintf(GinkgoWriter, "Expected error: %s\n", err.Error())
				Expect(err.Error()).To(ContainSubstring("context deadline exceeded"))
				Expect(resp).To(BeNil())
				Expect(requestCount).To(Equal(2))
			})
		})
		Describe(`Connection error`, func() {
			It(`Cannot connect to server`, func() {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					defer GinkgoRecover()

					// Nothing to do here as we'll just shut down the server before invoking the request.
				}))

				service, builder := clientInit("GET", server.URL, 2, 30)
				req, _ := builder.Build()

				// Shut down the server to simulate a connection error.
				server.Close()

				resp, err := service.Request(req, nil)
				Expect(err).ToNot(BeNil())
				fmt.Fprintf(GinkgoWriter, "Expected error: %s\n", err.Error())
				Expect(err.Error()).To(ContainSubstring("connect: connection refused"))
				Expect(resp).To(BeNil())
			})
		})
		Describe(`Misc. errors`, func() {
			var requestCount int
			BeforeEach(func() {
				requestCount = 0
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					defer GinkgoRecover()

					requestCount++
					w.Header().Set("Retry-After", "1")
					w.WriteHeader(http.StatusTooManyRequests)
				}))
			})
			AfterEach(func() {
				server.Close()
			})
			It(`Retries not enabled`, func() {
				service, builder := clientInit("GET", server.URL, 0, 0)
				req, _ := builder.Build()

				var foo *Foo
				resp, err := service.Request(req, &foo)
				Expect(err).ToNot(BeNil())
				fmt.Fprintf(GinkgoWriter, "Expected error: %s\n", err.Error())
				Expect(err.Error()).To(ContainSubstring("Too Many Requests"))
				assertResponse(resp, http.StatusTooManyRequests, "")
				Expect(requestCount).To(Equal(1))
			})
			It(`Max retries exhausted`, func() {
				service, builder := clientInit("GET", server.URL, 3, 5)
				req, _ := builder.Build()

				var foo *Foo
				resp, err := service.Request(req, &foo)
				Expect(err).ToNot(BeNil())
				fmt.Fprintf(GinkgoWriter, "Expected error: %s\n", err.Error())
				Expect(err.Error()).To(ContainSubstring("Too Many Requests"))
				assertResponse(resp, http.StatusTooManyRequests, "")
				Expect(requestCount).To(Equal(4))
			})
			It(`Invalid URL scheme`, func() {
				service, builder := clientInit("GET", strings.Replace(server.URL, "http", "badscheme", 1), 1, 5)
				req, _ := builder.Build()

				var foo *Foo
				resp, err := service.Request(req, &foo)
				Expect(err).ToNot(BeNil())
				Expect(resp).To(BeNil())
				fmt.Fprintf(GinkgoWriter, "Expected error: %s\n", err.Error())
				Expect(err.Error()).To(ContainSubstring("unsupported protocol scheme"))
				Expect(requestCount).To(Equal(0))
			})
			It(`Invalid port`, func() {
				service, builder := clientInit("GET", server.URL+"99", 1, 5)
				req, _ := builder.Build()

				var foo *Foo
				resp, err := service.Request(req, &foo)
				Expect(err).ToNot(BeNil())
				Expect(resp).To(BeNil())
				fmt.Fprintf(GinkgoWriter, "Expected error: %s\n", err.Error())
				Expect(err.Error()).To(ContainSubstring("invalid port"))
				Expect(requestCount).To(Equal(0))
			})
			It(`Invalid host`, func() {
				service, builder := clientInit("GET", "http://notahost:12345", 1, 5)
				req, _ := builder.Build()

				var foo *Foo
				resp, err := service.Request(req, &foo)
				Expect(err).ToNot(BeNil())
				Expect(resp).To(BeNil())
				fmt.Fprintf(GinkgoWriter, "Expected error: %s\n", err.Error())
				Expect(requestCount).To(Equal(0))
			})
		})
		Describe(`Gateway errors`, func() {
			var requestCount int
			BeforeEach(func() {
				requestCount = 0
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					defer GinkgoRecover()

					requestCount++
					w.Header().Set("Retry-After", "1")
					w.Header().Set("Content-type", "application/json")
					w.WriteHeader(http.StatusBadGateway)
					fmt.Fprintf(w,
						`{"status_code": %d, "message": "Bad gateway error", "details": {"error":"BadRequest","description":"invalid request"}}`,
						http.StatusBadGateway)
				}))
			})
			AfterEach(func() {
				server.Close()
			})
			It(`Retries not enabled`, func() {
				service, builder := clientInit("GET", server.URL, 0, 0)
				req, _ := builder.Build()

				var foo *Foo
				resp, err := service.Request(req, &foo)
				Expect(err).ToNot(BeNil())
				fmt.Fprintf(GinkgoWriter, "Expected error: %s\n", err.Error())
				Expect(err.Error()).To(ContainSubstring("Bad gateway error"))
				assertResponse(resp, http.StatusBadGateway, "application/json")
				Expect(requestCount).To(Equal(1))
			})
			It(`Max retries exhausted`, func() {
				service, builder := clientInit("GET", server.URL, 3, 5)
				req, _ := builder.Build()

				var foo *Foo
				resp, err := service.Request(req, &foo)
				Expect(err).ToNot(BeNil())
				fmt.Fprintf(GinkgoWriter, "Expected error: %s\n", err.Error())
				Expect(err.Error()).To(ContainSubstring("Bad gateway error"))
				assertResponse(resp, http.StatusBadGateway, "application/json")
				Expect(requestCount).To(Equal(4))
			})
			It(`Invalid URL scheme`, func() {
				service, builder := clientInit("GET", strings.Replace(server.URL, "http", "badscheme", 1), 1, 5)
				req, _ := builder.Build()

				var foo *Foo
				resp, err := service.Request(req, &foo)
				Expect(err).ToNot(BeNil())
				Expect(resp).To(BeNil())
				fmt.Fprintf(GinkgoWriter, "Expected error: %s\n", err.Error())
				Expect(err.Error()).To(ContainSubstring("unsupported protocol scheme"))
				Expect(requestCount).To(Equal(0))
			})
		})
	})
	Describe(`Successful scenarios`, func() {

		Describe(`GET`, func() {
			var requestCount int
			var retryCode int
			BeforeEach(func() {
				requestCount = 0
				maxRequests := 3
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					defer GinkgoRecover()

					requestCount++
					if requestCount < maxRequests {
						w.Header().Set("Retry-After", "1")
						w.WriteHeader(retryCode)
					} else {
						w.Header().Set("Content-type", "application/json")
						w.WriteHeader(http.StatusOK)
						fmt.Fprint(w, `{"name": "Mookie Betts"}`)
					}
				}))
			})
			AfterEach(func() {
				server.Close()
			})
			It(`GET 429`, func() {
				retryCode = http.StatusTooManyRequests

				service, builder := clientInit("GET", server.URL, 5, 10)
				req, _ := builder.Build()

				var foo *Foo
				resp, err := service.Request(req, &foo)
				Expect(err).To(BeNil())
				assertResponse(resp, http.StatusOK, "application/json")

				result, ok := resp.Result.(*Foo)
				Expect(ok).To(BeTrue())
				Expect(result).ToNot(BeNil())
				Expect(foo).ToNot(BeNil())
				Expect(*result.Name).To(Equal("Mookie Betts"))
			})
			It(`GET 503`, func() {
				retryCode = http.StatusServiceUnavailable

				service, builder := clientInit("GET", server.URL, 5, 10)
				req, _ := builder.Build()

				var foo *Foo
				resp, err := service.Request(req, &foo)
				Expect(err).To(BeNil())
				assertResponse(resp, http.StatusOK, "application/json")

				result, ok := resp.Result.(*Foo)
				Expect(ok).To(BeTrue())
				Expect(result).ToNot(BeNil())
				Expect(foo).ToNot(BeNil())
				Expect(*result.Name).To(Equal("Mookie Betts"))
			})
		})

		Describe(`DELETE`, func() {
			var requestCount int
			var retryCode int
			BeforeEach(func() {
				requestCount = 0
				maxRequests := 3
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					defer GinkgoRecover()

					requestCount++
					if requestCount < maxRequests {
						w.Header().Set("Retry-After", "1")
						w.WriteHeader(retryCode)
					} else {
						w.Header().Set("Content-type", "application/json")
						w.WriteHeader(http.StatusNoContent)
					}
				}))
			})
			AfterEach(func() {
				server.Close()
			})
			It(`DELETE - 429`, func() {
				retryCode = http.StatusTooManyRequests

				service, builder := clientInit("DELETE", server.URL, 5, 10)
				req, _ := builder.Build()

				resp, err := service.Request(req, nil)
				Expect(err).To(BeNil())
				assertResponse(resp, http.StatusNoContent, "")
			})
			It(`DELETE - 503`, func() {
				retryCode = http.StatusServiceUnavailable

				service, builder := clientInit("DELETE", server.URL, 5, 10)
				req, _ := builder.Build()

				resp, err := service.Request(req, nil)
				Expect(err).To(BeNil())
				assertResponse(resp, http.StatusNoContent, "")
			})
		})

		Describe(`HEAD/OPTIONS`, func() {
			var requestCount int
			var retryCode int
			BeforeEach(func() {
				requestCount = 0
				maxRequests := 3
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					defer GinkgoRecover()

					requestCount++
					if requestCount < maxRequests {
						w.Header().Set("Retry-After", "1")
						w.WriteHeader(retryCode)
					} else {
						w.Header().Set("Server-Name", "My Server")
						w.WriteHeader(http.StatusOK)
					}
				}))
			})
			AfterEach(func() {
				server.Close()
			})
			It(`HEAD - 429`, func() {
				retryCode = http.StatusTooManyRequests

				service, builder := clientInit("HEAD", server.URL, 5, 10)
				req, _ := builder.Build()

				resp, err := service.Request(req, nil)
				Expect(err).To(BeNil())
				assertResponse(resp, http.StatusOK, "")
				Expect(resp.GetHeaders().Get("Server-Name")).To(Equal("My Server"))
			})
			It(`OPTIONS - 503`, func() {
				retryCode = http.StatusServiceUnavailable

				service, builder := clientInit("OPTIONS", server.URL, 5, 10)
				req, _ := builder.Build()

				resp, err := service.Request(req, nil)
				Expect(err).To(BeNil())
				assertResponse(resp, http.StatusOK, "")
				Expect(resp.GetHeaders().Get("Server-Name")).To(Equal("My Server"))
			})
		})
		Describe(`POST/PUT/PATCH`, func() {
			var requestCount int
			var retryCode int
			var successCode int
			var expectedRequestBody = "Good request body"

			BeforeEach(func() {
				requestCount = 0
				maxRequests := 3
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					defer GinkgoRecover()

					requestCount++

					// Validate the request. Make sure we got the correct request body for each retry.
					Expect(r.Header.Get("Content-Type")).To(Equal("text/plain"))
					bodyBuf := new(bytes.Buffer)
					_, _ = bodyBuf.ReadFrom(r.Body)
					Expect(bodyBuf.String()).To(Equal(expectedRequestBody))

					if requestCount < maxRequests {
						w.Header().Set("Retry-After", "1")
						w.WriteHeader(retryCode)
					} else {
						w.Header().Set("Content-type", "application/json")
						w.WriteHeader(successCode)
						fmt.Fprint(w, `{"name": "Mookie Betts"}`)
					}
				}))
			})
			AfterEach(func() {
				server.Close()
			})
			It(`POST 429`, func() {
				retryCode = http.StatusTooManyRequests
				successCode = http.StatusCreated

				service, builder := clientInit("POST", server.URL, 5, 10)
				_, _ = builder.SetBodyContentString(expectedRequestBody)
				builder.AddHeader("Content-Type", "text/plain")
				req, _ := builder.Build()

				service.DisableSSLVerification()
				Expect(service.IsSSLDisabled()).To(BeTrue())

				var foo *Foo
				resp, err := service.Request(req, &foo)
				Expect(err).To(BeNil())
				assertResponse(resp, http.StatusCreated, "application/json")

				result, ok := resp.Result.(*Foo)
				Expect(ok).To(BeTrue())
				Expect(result).ToNot(BeNil())
				Expect(foo).ToNot(BeNil())
				Expect(*result.Name).To(Equal("Mookie Betts"))
			})
			It(`PUT 503`, func() {
				retryCode = http.StatusServiceUnavailable
				successCode = http.StatusOK

				service, builder := clientInit("PUT", server.URL, 5, 10)
				_, _ = builder.SetBodyContentString(expectedRequestBody)
				builder.AddHeader("Content-Type", "text/plain")
				req, _ := builder.Build()

				var foo *Foo
				resp, err := service.Request(req, &foo)
				Expect(err).To(BeNil())
				assertResponse(resp, http.StatusOK, "application/json")

				result, ok := resp.Result.(*Foo)
				Expect(ok).To(BeTrue())
				Expect(result).ToNot(BeNil())
				Expect(foo).ToNot(BeNil())
				Expect(*result.Name).To(Equal("Mookie Betts"))
			})
			It(`PATCH 429`, func() {
				retryCode = http.StatusTooManyRequests
				successCode = http.StatusOK

				service, builder := clientInit("PATCH", server.URL, 5, 10)
				_, _ = builder.SetBodyContentString(expectedRequestBody)
				builder.AddHeader("Content-Type", "text/plain")
				req, _ := builder.Build()

				var foo *Foo
				resp, err := service.Request(req, &foo)
				Expect(err).To(BeNil())
				assertResponse(resp, http.StatusOK, "application/json")

				result, ok := resp.Result.(*Foo)
				Expect(ok).To(BeTrue())
				Expect(result).ToNot(BeNil())
				Expect(foo).ToNot(BeNil())
				Expect(*result.Name).To(Equal("Mookie Betts"))
			})
		})
	})
})
