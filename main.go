package main

import (
	"bytes"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/larryhou/www/api/module/urlmap"
	"github.com/larryhou/www/api/util/db"
	"github.com/larryhou/www/api/util/svr"
)

func main() {
	opts := struct {
		key string
		pem string
	}{}
	flag.StringVar(&opts.key, `key`, `res/rapidsir.key`, `private key`)
	flag.StringVar(&opts.pem, `pem`, `res/rapidsir.pem`, `certificate`)
	flag.Parse()

	defer db.Close()
	if f, err := os.OpenFile(path.Base(os.Args[0])+`.log`, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
		log.SetOutput(f)
	}

	mux := http.NewServeMux()
	mux.Handle(`/cloud/`, svr.FileHandle(`./static`))
	mux.Handle(`/url`, svr.Handle(urlmap.Handle))
	mux.Handle(`/url/`, svr.Handle(urlmap.Handle))
	mux.Handle("/", svr.Handle(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case `/`, `/index.html`, `index.htm`:
			io.Copy(w, bytes.NewReader([]byte(index)))
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))

	go http.ListenAndServe(":80", mux)
	log.Fatal(http.ListenAndServeTLS(":443", opts.pem, opts.key, mux))
}

const index = `
<!doctype html>
<html data-n-head-ssr lang="zh-CN" lang="en-US">

<head>
    <title>快先生慢思考</title>
    <meta charset="UTF-8">
    <style>
        footer {
            position: fixed;
            padding: 10px 10px 0px 10px;
            bottom: 0;
            width: 100%;
            height: 35px;
        }

        div {
            margin: 2px 10px 0px 0px;
        }

        #title {
            font-size: 5em;
        }
    </style>
</head>

<body>
    <h1 id="title">快先生慢思考</h1>
    <hr style="width: 50%;" align="left">
    <div>
        <footer>
            <div align="right" style="margin-right: 30pt;">
                <a target="_blank"
                    href="http://www.beian.gov.cn/portal/registerSystemInfo?recordcode=44030902003969"
                    style="display:inline-block;text-decoration:none;height:20px;line-height:20px;margin-right:8px;"><img
                        src="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABQAAAAUCAMAAAC6V+0/AAAC3FBMVEUAAAD+/ODz6Kr//+PeqFfYrn3x167k0JXoxnyaaVzhs2ifaFXbrGLkvFnpyF7v2X/kwm3cp1nhsGfqw3rZqG3ntVzjrFPt3oDjvGnfr2fbnFGti3q0lH7ktoLryXn9v1T4znr/74bnvGz034v+2I/ktoDz6ZLkwY/Dfz7buoftzYbq2IPr0pjs3bLv6KPRrnbKhFv79ND488n/+dDZr4Lx38f/+cH/95f42oL7/97s2Y3++uzw1rvTk3DmuloAAHkBAm7uzWYAAGXktV3qvFr/0ljksE7fo0rWHxhrdocAAIAABHf143Pyy27w1GwGA2jtymHpwWDqxV/qyVyTeFrrwFflwFPislP+xVLpsErbmUfVkEbysETemUTpgj7ThT3XdTg5FDjdhTXWZTDaTCm7TCbTOCLXPiD9LA/QFg3UAwnOAQOEj5kcPpdyhZSptJEACJFpfo4AG44XMInFvYfTvIejmYSVkINyeoJzdoK9un6SjX7FrnwAEHp8enny2HjWwHjKtnhcX3jYzHeNhnfu2HWUjHWsonPNwnH70m9WTm8AAW//723pym3dtmn/0mbnxGa0o2ZeWWb8zGT/4mPtwmJuYmL/22D/vmB5ZGC9kF7/2l0MAF3uyFqnjVn4xFjYnli0mVi5i1jiqVfyyVbmtlbXkVNUOFPlvFLpt1LNrFKjfVLuvlBgHlDsuU/ouU9ONU/ov05ODk7/2E02Gk3jqkqEaUr/tUngjkf7n0bXikb6xERCJETdn0LckUG1gD/ooD3Ulj3jkz3TZT3WjjzOeDqBWDr3pDnglTlMADnbbTf2gjbkbzaTYDZpAjbplzTtcTTEazPXXzOeXzDscS3MPi38jizJWSrVSCrrXynzfCjVdCjZRyjTQCbFUiTlYCPXPSHLPSHWMR/wXh7iRh7GPh3PLBrSIRrWGhfMJxPGJxPRDBG/ABG2ABCxDg7BDAvEGArZAAbJAALPAADa4ry/AAAAPnRSTlMACEIaxqxpAvv7+ff19PDs7Ovn5uXk5OHg29LRy8fEw8G+vLqysaufnJiVk4yDfG9dXFpMSEFBNTApJyEcFO3QiBQAAAFzSURBVBjTYoACZjYZaTZmBmRgxsp9+di21ZysxggxxlmJZy/ev9LXnriIEa5VYUPIray0lOyd+ctVoKKWXFsmXXvu8exO5vsZnnuErcCC5m1e8x5nPXrxOu3TzSqHFguQmI18tff+Jx89HqR7fE5v7q5TtAYK6h8v81p4Ovv6wbAdmRc6HMpddYGCmudrCqbtTn2anHBq15SZ9iUx6kBBkSTfXIfUuBsPL909c9i/uP6EJFAQMJ6j2/Ps32Yk30uIy3jjXxgRLwEUVN07ubTo5LsPr16mXD1X29gZrgUUlN23uD/H28lp09o5TvYVs523ygEFORYsO+TbEOI5cVVTV+XUA1Fu/EBBoxXu0bfnT98cEePa45oUHR7MBHK9IV9Y/BFHFzc7R7/YqF4BsBiDqVBw0NLQoMAAF3c7vwmCEEFln1ZnZxe3wJWx7nZ2jj5qkNDU5l2/ZE3kusjQuRsDxPXYoQFqa6DBIiUmyqKkYwIWAgD35oZAL/mkFwAAAABJRU5ErkJggg=="
                        style="float:left;" />粤公网安备 44030902003969号</a>
                <a href="https://beian.miit.gov.cn/" target="_blank"><img
                        class="nn-national-emblem"
                        src="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABIAAAASCAYAAABWzo5XAAAD1UlEQVQ4jU2US2icZRSGn+/y///MZGaSmSG2TWKTemlsbSMVSy/gQtLS4qIWFReidNWFCILo0oV7ERQXdlGxgkp14UJQEIy1RbC21l5AaG3jJYkdJs1MkpnM/Lfv+z+ZWMR3dV5438PLOYcjjDGzQoghISSZ4C6WSHvfEp17j/TsNbx92/D2vIyuHED7o0jgP+m/WBbOuQ5Q7NPe7TPYzgeIP24RfXoV81lElvjIwKCedOijW/H2bMff/BJe/uD/m3U0kParjA62+TZrx76CyyAmNiDe3IncZJFLGeaTRdJjvxHfe4P8qTrZ/jF8dT/S8/v2tJ8SwwKdS0cJXz+Dvgzu8ATi4y3YqI5r1LGt23jvFHHHx5HzivCV83RPH6AXXrwbyIHN2q2lm8+5hRdwDYSrVwPXmNnpFp7f4OaGc272qYqbGy25+elBt3hu3C1OFt0SuPq0cH//tNel4a/OOdeU3aUZ6t98SVYrInBkOwNku4v6vIkoSPxfDMIXqHNdxNUO2f7+GEDV8jS/v8jKzQ/XM0mbRIi8T24qWBe4QgZJgjQWT4J6XKNKFmUMLtG4slo3euOKYGNAGoXrXIcr1yiOeORHKqwNLCMuxKSvWvQbw8gHBvCmc2SXDPzQxu4exH7UQCCQDw1S3mqJurNY20UGyXeELcNiI0JP+aimxb6VYEc94r8s8ZmE+GqM2VElOrGCd62DGNB0c5bm75YBfYG0t4BM0kkqxZSC38M9qugHdzMha+8aUhdAzWHzAe0TEfb0GpIMd59EbkyplmLSaBPoDWiNRuYMjAyTHYqxJ+fx04R0KcSc7GG+UNAyiFCSE2l/0YgDZfTWEbLrt3BeGZF20TlzHbsQ4Hlr2F1Vomkf93VMTmcI38e0HNIPEL5CxykE4D9TQy12SesalZuD6CxaP/IYgjuIzjKq0UA9nSOe12SRQpcdMrSI/vFqTZIvExyKEXoV3YsRtQHkwxOo6gRSqd3oHT3shCCKJWrfFuzxAfwxg0rBLzg0Fh3HyCdCsiMjuHyFns7QuxLkPRWE2I504jBZ7zBS1OjMWZJmi/jBAumRHHLS4eYN0lrUs5re3jIib2jfWIGohBOT0HsRIYbQyeIM5sqPaNempgXpbJfhqiabLNAqZvglRTqm8KZ8hj0Bc6sEPYduQ3p+Hjd4CrlnNzo/etCzlQam8z6it4qKJGl7FZHbhAoiMlISZSiO5cisJdBdRKGKKCTIwjZ0+TW0v9nr/6NZ5xjK7BUyd5ks/pnwzzrhHUHU0ZQ8Q5gInJfhFSWlcYm/cRtKToHci1KjAMv/AOnkyBlAs+QxAAAAAElFTkSuQmCC"
                        alt="快先生慢思考备案号" title="快先生慢思考备案号" />粤ICP备2023064967号-1</a>
            </div>
        </footer>
    </div>
</body>

</html>
`
