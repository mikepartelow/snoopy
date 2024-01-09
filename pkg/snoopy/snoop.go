package snoopy

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type snoop struct {
	Upstream string `yaml:"upstream"`
	Local    string `yaml:"local"`
	Logfile  string `yaml:"logfile"`
	Cookies  []struct {
		Name  string `yaml:"name"`
		Value string `yaml:"value"`
	} `yaml:"cookies"`
	Headers []struct {
		Name  string `yaml:"name"`
		Value string `yaml:"value"`
	} `yaml:"headers"`
	RespnoseRewrites []struct {
		Old         string `yaml:"old"`
		New         string `yaml:"new"`
		MustRewrite bool   `yaml:"must-rewrite"`
	} `yaml:"response-rewrites"`

	logger *slog.Logger
}

func (s *snoop) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.logger.Debug("ServeHTTP", "req.Method", req.Method, "req.URL", req.URL, "req.Header", req.Header)

	upstreamURL, _ := url.JoinPath(s.Upstream, req.URL.Path)
	s.logger.Debug("ServeHTTP", "upstreamURL", upstreamURL)

	checkErr(s.logURL(upstreamURL), w, s.logger)

	client := &http.Client{}
	upstreamReq, err := s.newUpstreamRequest(upstreamURL, req)
	checkErr(err, w, s.logger)

	resp, err := client.Do(upstreamReq)
	checkErr(err, w, s.logger)
	defer resp.Body.Close()

	slog.Info("response", "upstreamURL", upstreamURL, "status", resp.Status, "header", resp.Header)

	copyHeader(w.Header(), resp.Header)

	if len(s.RespnoseRewrites) > 0 {
		body, err := io.ReadAll(resp.Body)
		checkErr(err, w, s.logger)

		newBody := string(body)

		for _, r := range s.RespnoseRewrites {
			s.logger.Debug("rewriting response", "old", r.Old, "new", r.New)

			oldNewBody, newBody := newBody, strings.ReplaceAll(newBody, r.Old, r.New)
			if r.MustRewrite && oldNewBody == newBody {
				checkErr(errors.New("must rewrite failed"), w, s.logger)
			}
		}

		w.Header().Set("content-length", strconv.Itoa(len(newBody)))
		w.WriteHeader(resp.StatusCode)
		_, _ = w.Write([]byte(newBody))
	} else {
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
	}
}

func (s *snoop) logURL(url string) error {
	file, err := os.OpenFile(s.Logfile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("couldn't optn %q: %w", s.Logfile, err)
	}
	defer file.Close()

	_, err = file.WriteString(string(url) + "\n")

	if err != nil {
		return fmt.Errorf("couldn't write to %q: %w", s.Logfile, err)
	}

	return nil
}

func (s *snoop) newUpstreamRequest(url string, req *http.Request) (*http.Request, error) {
	upstreamReq, err := http.NewRequest(req.Method, url, req.Body)
	if err != nil {
		return nil, fmt.Errorf("error creating http.Request: %w", err)
	}

	copyHeader(upstreamReq.Header, req.Header)

	for _, h := range s.Headers {
		upstreamReq.Header.Add(h.Name, h.Value)
	}

	for _, c := range s.Cookies {
		upstreamReq.AddCookie(&http.Cookie{
			Name:  c.Name,
			Value: c.Value,
		})
	}

	if len(s.RespnoseRewrites) > 0 {
		upstreamReq.Header.Set("accept-encoding", "text/plain")
	}

	return upstreamReq, nil
}

func checkErr(err error, w http.ResponseWriter, logger *slog.Logger) {
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		logger.Error(err.Error())
		panic(err)
	}
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
