
package http

import (
	io "io"
	strconv "strconv"
)

type ResponseWriter interface{
	Header()(Header)
	WriteHeader(status StatusCode)(err error)
	Write(buf []byte)(n int, err error)
}

type responseWriter struct{
	writer io.Writer
	status StatusCode
	header Header
}

func NewResponseWriter(w io.Writer)(ResponseWriter){
	return &responseWriter{
		writer: w,
		status: -1,
		header: NewHeader(),
	}
}

func (rw *responseWriter)Header()(Header){
	return rw.header
}

func (rw *responseWriter)WriteHeader(status StatusCode)(err error){
	if rw.status > 0 {
		return ErrHeaderHasWritten
	}
	rw.status = status
	_, err = io.WriteString(rw.writer, DefaultProto.String() + " " + strconv.Itoa((int)(status)) + " " + status.String())
	if err != nil { return }
	_, err = rw.writer.Write(crlf)
	if err != nil { return }
	err = rw.header.WriteTo(rw.writer)
	if err != nil { return }
	return nil
}

func (rw *responseWriter)Write(buf []byte)(n int, err error){
	return rw.writer.Write(buf)
}

type Response struct{
	Proto Proto
	Status StatusCode
	Header Header
	Body io.ReadCloser
}

func ParseResponse(r io.Reader)(res *Response, err error){
	res = new(Response)
	var ok bool
	res.Body, ok = r.(io.ReadCloser)
	if !ok {
		res.Body = io.NopCloser(r)
	}
	br := &wrapReader{r}
	res.Proto, res.Status, err = ParseHttpCode(br)
	if err != nil { return }
	res.Header, err = ParseHeader(br)
	if err != nil { return }
	return
}
