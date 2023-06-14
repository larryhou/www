package util

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path"
	"reflect"
	"strings"
	"time"
)

var httpClient *http.Client

func init() {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.TLSClientConfig = &tls.Config{InsecureSkipVerify: os.Getenv("TLS_ALLOW_INSECURE")==`true`}
	t.MaxIdleConnsPerHost = 100
	t.MaxConnsPerHost = 100
	t.MaxIdleConns = 100
	httpClient = &http.Client{Timeout: 60 * time.Minute, Transport: t}
}

type Attachment struct {
	Method string
	Header map[string]string
	Data   interface{}
}

/**
s := `WWL16.1+10.1 版本[trunk]`
log.Printf("url.QueryEscape %s", url.QueryEscape(s)) // encode '+' into %2B but encode ' ' into '+'
log.Printf("url.PathEscape  %s", url.PathEscape(s))  // encode ' ' into %20 but keep '+' as it is
# 2022/04/27 17:13:16 url.QueryEscape WWL16.1%2B10.1+%E7%89%88%E6%9C%AC%5Btrunk%5D
# 2022/04/27 17:13:16 url.PathEscape  WWL16.1+10.1%20%E7%89%88%E6%9C%AC%5Btrunk%5D
*/

func CookUrl(u string, params url.Values) string {
	if len(params) > 0 {
		buf := &bytes.Buffer{}
		buf.WriteString(u)
		p := strings.IndexByte(u, '?')
		n := 0
		for k, list := range params {
			k := url.QueryEscape(k)
			for _, v := range list {
				if n > 0 {buf.WriteByte('&')} else {
					if p == -1 {buf.WriteByte('?')} else {if p + 1 < len(u) { buf.WriteByte('&') }}
				}
				buf.WriteString(k)
				buf.WriteByte('=')
				buf.WriteString(strings.ReplaceAll(url.QueryEscape(v), "+", "%20"))
				n++
			}
		}
		return buf.String()
	}
	return u
}

func Request(model interface{}, u string, params url.Values, data interface{}, header map[string]string) (err error) {
	u = CookUrl(u, params)
	if header == nil { header = make(map[string]string) }

	var method string
	if attachment, ok := data.(*Attachment); ok {
		for k,v := range attachment.Header {header[k]=v}
		method = attachment.Method
		data = attachment.Data
	}

	const ContentType = "Content-Type"

	var request *http.Request
	if data != nil {
		var body io.Reader
		buf := &bytes.Buffer{}
		switch data := data.(type) {
		case string: buf.WriteString(data)
		case []byte: buf.Write(data)
		case url.Values:
			buf.WriteString(data.Encode())
			if _, ok := header[ContentType]; !ok {header[ContentType] = "application/x-www-form-urlencoded"}
		case io.Reader: body = data
		case nil:
		default:
			if _, ok := header[ContentType]; !ok {header[ContentType] = "application/json"}
			j := json.NewEncoder(buf)
			j.SetEscapeHTML(false)
			if err := j.Encode(data); err != nil {return err}
		}

		if len(method) == 0 { method = http.MethodPost }
		if buf.Len() > 0 { body = buf
			log.Printf("# %s %s \n", method, buf.String())
		} else {
			log.Printf("# %s %p \n", method, body)
		}
		request, err = http.NewRequest(method, u, body)
		if err != nil {return}
	} else {
		if len(method) == 0 { method = http.MethodGet }
		request, err = http.NewRequest(method, u, nil)
		if err != nil {return}
	}

	log.Printf("%p >> %s\n", request, request.URL)

	for k, v := range header {request.Header.Set(k, v)}

	rsp, err := httpClient.Do(request)

	if err != nil { return }
	switch m := model.(type) {
	case **http.Response: *m = rsp
	default:
		defer rsp.Body.Close()
		log.Printf("%p << %v #%d %s\n", request, rsp.StatusCode, rsp.ContentLength, rsp.Header.Get("Content-Type"))
		return unmarshal(model, rsp)
	}
	return
}

type Pipable interface {
	Pipe(r *http.Response)
}

func unmarshal(o interface{}, r *http.Response) error {
	body, err := io.ReadAll(r.Body)
	dump := func(ctx interface{}) {
		w := log.Writer()
		defer w.Write([]byte{'\n'})
		t := r.Header.Get(`Content-Type`)
		fmt.Fprintf(w, "%v %s [%d] %+v ", ctx, r.Request.URL, r.StatusCode, r.Request.Header)
		switch {
		case strings.Contains(t,`json`):
		case strings.Contains(t,`text`):
		case strings.Contains(t,`xml`):
		case strings.Contains(t,`url`):
		default: return
		}

		w.Write(body)
	}

	if r.StatusCode >= 400 { dump(`HTTP_ERROR`) }
	if err == nil {
		if len(body) == 0 {return nil}
		switch data := o.(type) {
		case io.Writer:
			if w, ok := data.(http.ResponseWriter); ok {
				for k := range r.Header { w.Header().Set(k, r.Header.Get(k)) }
			}
			_, err = data.Write(body)
		default:
			if data == nil { dump(`NIL_WRITER`) } else {
				if err = json.Unmarshal(body, data); err == nil {
					if p, ok := data.(Pipable); ok {p.Pipe(r)}
				}
			}
		}
	}

	if err != nil {dump(err)}
	return err
}

type FormFile struct {
	Type   string
	Name   string
	Size   int64
	Source interface{}
	Fields map[string]string
	Header map[string]string
	Md5sum string
}

func (x *FormFile) check() error {
	if len(x.Type) == 0 {x.Type = "file"}
	if len(x.Name) == 0 {return fmt.Errorf("filename is empty")}
	if x.Size <= 0 {return fmt.Errorf("size is not set")}
	return nil
}

type Counter int64
func (c *Counter) Write(p []byte) (int, error) {
	*c += Counter(len(p))
	return len(p), nil
}

func (c *Counter) WriteString(p string) (int, error) {
	*c += Counter(len(p))
	return len(p), nil
}

func NewFormFile(rsp *http.Response) *FormFile {
	name := path.Base(rsp.Request.URL.Path)
	if _, opts, err := mime.ParseMediaType(rsp.Header.Get("Content-Disposition")); err == nil {
		if s, ok := opts["filename"]; ok { name = s }
	}
	return &FormFile{Name: name, Size: rsp.ContentLength, Source: rsp.Body}
}

func FormRequest(u string, model interface{}, params url.Values, data map[string]string, header map[string]string) error {
	b := &bytes.Buffer{}
	m := multipart.NewWriter(b)

	u = CookUrl(u, params)
	for k := range data {
		m.WriteField(k, data[k])
		log.Printf("part %s: %s", k, data[k])
	}
	m.Close()

	req, err := http.NewRequest(http.MethodPost, u, b)
	if err != nil {return err}
	for k, v := range header { req.Header.Set(k, v) }
	req.Header.Set("Content-Type", m.FormDataContentType())

	rsp, err := httpClient.Do(req)
	if err != nil {return err}
	defer rsp.Body.Close()

	return unmarshal(model, rsp)
}

type mpWriter multipart.Writer

func (x *mpWriter) CreateFormFile(name, filename string, header map[string]string) (io.Writer, error) {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, name, filename))
	ctype := "application/octet-stream"
	if p := strings.LastIndexByte(filename, '.'); p > 0 {
		if t := mime.TypeByExtension(filename[p:]); len(t) > 0 {ctype = t}
	}
	h.Set("Content-Type", ctype)
	for k, v := range header { h.Set(k, v) }
	return (*multipart.Writer)(x).CreatePart(h)
}

func Upload(f *FormFile, u string, model interface{}, params url.Values, header map[string]string) error {
	if err := f.check(); err != nil {return err}
	u = CookUrl(u, params)

	size := Counter(0)
	{ // caculate multipart form size
		m := multipart.NewWriter(&size)
		for k := range f.Fields { m.WriteField(k, f.Fields[k]) }
		(*mpWriter)(m).CreateFormFile(f.Type, f.Name, f.Header)
		m.Close()
	}

	r, w := io.Pipe()
	m := multipart.NewWriter(w)
	go func() {
		defer w.Close()
		defer m.Close()
		if err := doMultipartUpload(f, m); err != nil {log.Printf("%p upload: %v", &u, err)}
	}()

	log.Printf("%p >> %s %s \n", &u, u, f.Name)
	req, err := http.NewRequest(http.MethodPost, u, r)
	if err != nil {return err}
	if f.Size > 0 { req.ContentLength = f.Size + int64(size) }
	for k, v := range header { req.Header.Set(k, v) }
	req.Header.Set("Content-Type", m.FormDataContentType())
	rsp, err := httpClient.Do(req)
	if err != nil {return err}
	defer rsp.Body.Close()

	log.Printf("%p << %v #%d %s\n", &u, rsp.StatusCode, rsp.ContentLength, rsp.Header.Get("Content-Type"))
	return unmarshal(model, rsp)
}

func doMultipartUpload(f *FormFile, w *multipart.Writer) error {
	var input io.Reader
	switch source := f.Source.(type) {
	case []byte: input = bytes.NewReader(source)
	case io.Reader: input = source
	default:return fmt.Errorf("unsupported source type: %v", reflect.ValueOf(f.Source).Type().Name())
	}

	if c, ok := input.(io.Closer); ok { defer c.Close() }

	for k := range f.Fields { w.WriteField(k, f.Fields[k]) }
	part, err := (*mpWriter)(w).CreateFormFile(f.Type, f.Name, f.Header)
	if err != nil { return err }

	h := md5.New()
	n, err := io.Copy(io.MultiWriter(part, h), input)
	if err == nil {
		md5sum := hex.EncodeToString(h.Sum(nil))
		if len(f.Md5sum) == 32 && f.Md5sum != md5sum {return fmt.Errorf("md5sum mismatch: (expect=%s, actual=%s) %s", f.Md5sum, md5sum, f.Name)}
		log.Printf("sent %s %s %d/%d", md5sum, f.Name, n, f.Size)
	} else {
		return fmt.Errorf("%s %d/%d %v", f.Name, n, f.Size, err)
	}
	return nil
}

func MultiUpload(files []*FormFile, u string, model interface{}, params url.Values, header map[string]string) error {
	u = CookUrl(u, params)

	size := Counter(0)
	{
		w := multipart.NewWriter(&size)
		for _, f := range files { // caculate multipart form size
			if err := f.check(); err != nil {return err}
			size += Counter(f.Size)
			for k := range f.Fields { w.WriteField(k, f.Fields[k]) }
			(*mpWriter)(w).CreateFormFile(f.Type, f.Name, f.Header)
		}
		w.Close()
	}

	r, w := io.Pipe()
	m := multipart.NewWriter(w)
	go func() {
		defer w.Close()
		defer m.Close()
		for _, f := range files {
			if err := doMultipartUpload(f, m); err != nil {log.Printf("%p upload: %v", &u, err)}
		}
	}()

	req, err := http.NewRequest(http.MethodPost, u, r)
	if err != nil {return err}
	for k, v := range header { req.Header.Set(k, v) }
	if size > 0 { req.ContentLength = int64(size) }
	req.Header.Set("Content-Type", m.FormDataContentType())
	rsp, err := httpClient.Do(req)
	if err != nil {return err}
	defer rsp.Body.Close()

	log.Printf("%p << %v #%d %s\n", &u, rsp.StatusCode, rsp.ContentLength, rsp.Header.Get("Content-Type"))
	return unmarshal(model, rsp)
}


