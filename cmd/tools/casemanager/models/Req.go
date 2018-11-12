package models

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// Req a atmosphere api http request
type Req struct {
	APIName string        `json:"api_name"`
	FullURL string        `json:"url"`
	Method  string        `json:"method"`
	Payload string        `json:"payload"`
	Timeout time.Duration `json:"timeout"`
}

// GetReq : get *http.Request
func (r *Req) GetReq() *http.Request {
	var reqBody io.Reader
	if r.Payload == "" {
		reqBody = nil
	} else {
		reqBody = strings.NewReader(r.Payload)
	}
	req, err := http.NewRequest(r.Method, r.FullURL, reqBody)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "name=anny")
	return req
}

// Invoke : send a http request
func (r *Req) Invoke() (int, []byte, error) {
	client := http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				if r.Timeout == 0 {
					r.Timeout = time.Second * 180 // default timeout 3 min
				}
				c, err := net.DialTimeout(netw, addr, 5*time.Second) // default dial timeout 5 seconds
				if err != nil {
					return nil, err
				}
				err = c.SetDeadline(time.Now().Add(r.Timeout))
				if err != nil {
					return nil, err
				}
				return c, nil
			},
		},
	}
	req := r.GetReq()
	resp, err := client.Do(req)
	defer func() {
		if req.Body != nil {
			err = req.Body.Close()
		}
		if resp != nil && resp.Body != nil {
			err = resp.Body.Close()
		}
	}()
	if err != nil {
		return 0, nil, err
	}
	statusCode := resp.StatusCode
	var buf [4096 * 1024]byte
	n := 0
	n, err = resp.Body.Read(buf[:])
	if err.Error() == "EOF" {
		err = nil
	}
	return statusCode, buf[:n], err
}

// ToString : get json of this
func (r *Req) ToString() string {
	buf, err := json.MarshalIndent(r, "", "\t")
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("\n%s\n", string(buf))
}
