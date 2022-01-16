
package http

import (
	io "io"
	bytes "bytes"
	strings "strings"
	strconv "strconv"
	url "net/url"
)

var (
	crlf   = ([]byte)("\r\n")
	colonS = ([]byte)(": ")
	space  = ([]byte)(" ")
)

type Header map[string][]string

func NewHeader()(Header){
	return make(Header)
}

func (h Header)getRealK(key string)(string){
	_, ok := h[key]
	if ok {
		return key
	}
	for k, _ := range h {
		if strings.EqualFold(key, k) {
			return k
		}
	}
	return key
}

func (h Header)Values(key string)([]string){
	return h[h.getRealK(key)]
}

func (h Header)Get(key string)(string){
	vv := h.Values(key)
	if vv == nil || len(vv) == 0 {
		return ""
	}
	return vv[0]
}

func (h Header)Set(key string, values ...string)([]string){
	if len(values) == 0 {
		h.Del(key)
		return values
	}
	h[h.getRealK(key)] = values
	return values
}

func (h Header)Add(key string, values ...string)([]string){
	key = h.getRealK(key)
	values = append(h[key], values...)
	h[key] = values
	return values
}

func (h Header)Del(key string)([]string){
	for k, v := range h {
		if strings.EqualFold(key, k) {
			delete(h, k)
			return v
		}
	}
	return nil
}

func (h Header)Clone()(c Header){
	if h == nil {
		return nil
	}
	c = make(Header, len(h))
	for k, v := range h {
		c[k] = append(([]string)(nil), v...)
	}
	return
}

func (h Header)WriteTo(w io.Writer)(err error){
	for k, vv := range h {
		if len(vv) == 0 {
			continue
		}
		_, err = io.WriteString(w, k)
		if err != nil { return }
		_, err = w.Write(colonS)
		if err != nil { return }
		_, err = io.WriteString(w, strings.Join(vv, ","))
		if err != nil { return }
		_, err = w.Write(crlf)
		if err != nil { return }
	}
	_, err = w.Write(crlf)
	if err != nil { return }
	return nil
}

type LineReader interface{
	io.Reader
	ReadLine()(buf []byte, err error)
}

type wrapReader struct{
	r io.Reader
}

var _ LineReader = (*wrapReader)(nil)

func (r *wrapReader)Read(buf []byte)(n int, err error){
	return r.r.Read(buf)
}

func (r *wrapReader)ReadLine()(buf []byte, err error){
	var n int = 0
	buf = make([]byte, 1024)
	for {
		_, err = r.Read(buf[n:n + 1])
		if err != nil { return }
		if buf[n] == '\n' {
			if n > 0 && buf[n - 1] == '\r' {
				n--
			}
			buf = buf[:n]
			return
		}
		n++
	}
}

func ParseHeader(r LineReader)(h Header, err error){
	var (
		line []byte
		k string
		v []string
	)
	h = make(Header)
	for {
		line, err = r.ReadLine()
		if err != nil {
			return
		}
		if len(line) == 0 {
			break
		}
		k, v = parseHeaderKV(line)
		h[k] = v
	}
	return
}

func parseHeaderKV(line []byte)(k string, v []string){
	for i := 0; i < len(line); i++ {
		if line[i] == ':' {
			k = (string)(bytes.TrimSpace(line[:i]))
			vv := bytes.Split(line[i + 1:], ([]byte)(","))
			v = make([]string, len(vv))
			for i, b := range vv {
				v[i] = (string)(bytes.TrimSpace(b))
			}
			return
		}
	}
	return 
}

func ParseHttpMethod(r LineReader)(method Method, URL *url.URL, proto Proto, err error){
	var (
		line []byte
		i int
	)
	line, err = r.ReadLine()
	if err != nil { return }
	if i = bytes.IndexByte(line, ' '); i < 0 {
		err = ErrWrongFormat
		return
	}
	method, line = (Method)(bytes.ToUpper(line[:i])), line[i + 1:]
	if i = bytes.IndexByte(line, ' '); i < 0 {
		err = ErrWrongFormat
		return
	}
	if method == MethodConnect {
		URL = &url.URL{
			Host: (string)(line[:i]),
		}
	}else{
		URL, err = url.ParseRequestURI((string)(line[:i]))
		if err != nil { return }
	}
	line = line[i + 1:]
	if i < 0 {
		i = len(line)
	}
	proto = ParseProto((string)(line[:i]))
	return
}

func ParseHttpCode(r LineReader)(proto Proto, code StatusCode, err error){
	var (
		line []byte
		i int
		n int
	)
	line, err = r.ReadLine()
	if err != nil { return }
	i = bytes.IndexByte(line, ' ')
	if i < 0 {
		err = ErrWrongFormat
		return
	}
	proto, line = ParseProto((string)(line[:i])), line[i + 1:]
	i = bytes.IndexByte(line, ' ')
	if i < 0 {
		i = len(line)
	}
	n, err = strconv.Atoi((string)(line[:i]))
	if err != nil { return }
	code = (StatusCode)(n)
	return
}
