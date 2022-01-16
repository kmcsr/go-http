
package http

import (
	io "io"
	strings "strings"
	strconv "strconv"
	net "net"
	url "net/url"
)

type Proto struct{
	Prefix string
	Major int
	Minor int
}

func ParseProto(s string)(p Proto){
	var (
		i int
		err error
	)
	if i = strings.IndexByte(s, '/'); i < 0 {
		return Proto{
			Prefix: s,
			Major: -1,
			Minor: -1,
		}
	}
	p.Prefix, s = s[:i], s[i + 1:]
	if i = strings.IndexByte(s, '.'); i >= 0 {
		if p.Major, err = strconv.Atoi(s[:i]); err == nil {
			if p.Minor, err = strconv.Atoi(s[i + 1:]); err == nil {
				return
			}
		}
	}
	p.Major = -1
	p.Minor = -1
	return
}

func (p Proto)String()(string){
	return p.Prefix + "/" + strconv.Itoa(p.Major) + "." + strconv.Itoa(p.Minor)
}

var DefaultProto Proto = Proto{ Prefix: "HTTP", Major: 1, Minor: 1 }

type Request struct{
	Method Method
	URL *url.URL
	Proto Proto
	Header Header
	Host string
	ContentLength int64
	
	Body io.ReadCloser
	PostForm url.Values
	
	LocalAddr net.Addr
	RemoteAddr net.Addr

	KeepAlive bool
}

func NewRequest(method Method, URL *url.URL, header Header)(req *Request){
	return &Request{
		Method: method,
		URL: URL,
		Proto: DefaultProto,
		Header: header,
		Host: header.Get("Host"),
		ContentLength: -1,
	}
}

func ParseRequest(r io.Reader)(req *Request, err error){
	req = new(Request)
	var ok bool
	req.Body, ok = r.(io.ReadCloser)
	if !ok {
		req.Body = io.NopCloser(r)
	}
	br := &wrapReader{r}
	req.Method, req.URL, req.Proto, err = ParseHttpMethod(br)
	if err != nil { return }
	req.Header, err = ParseHeader(br)
	if err != nil { return }
	return
}

func (req *Request)WriteTo(w io.Writer)(err error){
	if req.Method == "" {
		req.Method = MethodGet
	}

	_, err = io.WriteString(w, req.Method + " " + req.URL.EscapedPath() + " " + req.Proto.String())
	if err != nil { return }
	_, err = w.Write(crlf)
	if err != nil { return }
	if req.ContentLength >= 0 {
		req.Header.Set("Content-Length", strconv.FormatInt(req.ContentLength, 64))
	}else if len(req.Header.Get("Content-Length")) > 0 {
		req.ContentLength, err = strconv.ParseInt(req.Header.Get("Content-Length"), 10, 64)
		if err != nil { return }
	}
	err = req.Header.WriteTo(w)
	if err != nil { return }
	if req.ContentLength >= 0 {
		defer req.Body.Close()
		_, err = io.Copy(w, req.Body)
		if err != nil { return }
	}
	return nil
}

func (req *Request)WriteBody(w io.Writer)(err error){
	if req.ContentLength < 0 {
		defer req.Body.Close()
		_, err = io.Copy(w, req.Body)
		if err != nil { return }
	}
	return nil
}
