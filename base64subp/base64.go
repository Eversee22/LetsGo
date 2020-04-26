package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var (
	filename   = flag.String("filename", "inputb64", "input filename")
	outputfile = flag.String("ofilename", "", "output filename")
	suburl     = flag.String("url", "", "url of subscription")
)

type (
	// Vmess json struct
	Vmess struct {
		V    string      `json:"v"`
		Ps   string      `json:"ps"`
		Addr string      `json:"add"`
		Port string      `json:"port"`
		ID   string      `json:"id"`
		Aid  interface{} `json:"aid"` //int or string ?
		Net  string      `json:"net"`
		Typ  string      `json:"type"`
		Host string      `json:"host"`
		Path string      `json:"path"`
		TLS  string      `json:"tls"`
	}
)

func b64decode(strdata string, pad bool) []byte {
	if pad {
		if len(strdata)%4 != 0 {
			padn := (len(strdata)/4+1)*4 - len(strdata)
			padstr := "===="
			strdata += padstr[:padn]
		}
	}
	decodeBytes, err := base64.StdEncoding.DecodeString(strdata)
	if err != nil {
		decodeBytes, err = base64.URLEncoding.DecodeString(strdata)
		if err != nil {
			log.Fatalln(err)
		}
	}
	return decodeBytes
}

func parsevmes(vmes []string, writer io.Writer) {
	for i, v := range vmes {
		content := b64decode(v, false)
		var vs Vmess
		err := json.Unmarshal(content, &vs)
		if err != nil {
			log.Fatalln("json unmarshal error", err)
		}
		fmt.Fprintf(writer, "[%d]\n", i+1)
		fmt.Fprintln(writer, string(content))
		fmt.Fprintf(writer, "server: %s\nport: %s\nid: %s\nalterId: %s\nnet: %s\ntls: %s\npath: %s\nhost: %s\nps: %s\n\n",
			vs.Addr, vs.Port, vs.ID, vs.Aid, vs.Net, vs.TLS, vs.Path, vs.Host, vs.Ps)
	}
}

func parsetroj(troj []string, writer io.Writer) {
	for i, tr := range troj {
		// i := strings.Index(tr, "?")
		// part1 := tr[:i]
		// part2 := tr[i+1:]
		// //part1
		// i = strings.Index(part1, "@")
		// pass := part1[:i]
		// servport := part1[i+1:]
		// i = strings.Index(servport,":")
		// serv := servport[:i]
		// port := int(servport[i+1:])
		// // part2
		// i = strings.Index("#")
		// peer := part2[5:i]
		// name := part2[i+1:]

		/*scheme://[userinfo@]host/path[?query][#fragment]*/
		urlpart, _ := url.Parse(tr)
		//username := urlpart.User.Username()
		//password, _ := urlpart.User.Password()
		fmt.Fprintf(writer, "-%d-\nscheme:%s\npass:%s\nhost:%s\nquery:%s\nname:%s\n\n",
			i+1, urlpart.Scheme, urlpart.User.Username(), urlpart.Host,
			urlpart.RawQuery, urlpart.Fragment)
	}
}

func main() {
	flag.Parse()
	var data []byte
	var err error

	if *suburl != "" {
		r, err := http.Get(*suburl)
		if err != nil {
			log.Fatalln(err)
		}
		data, err = ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatalln(err)
		}
		if err = r.Body.Close(); err != nil {
			log.Println(err)
		}
	} else {
		data, err = ioutil.ReadFile(*filename)
		if err != nil {
			fmt.Println(err)
			flag.Usage()
			return
		}
	}
	fmt.Println("read bytes:", len(data))
	decodeBytes := b64decode(string(data), true)
	decodeStr := string(decodeBytes)
	items := strings.Split(decodeStr, "\n")

	vmes := []string{}
	troj := []string{}
	for _, item := range items {
		p := strings.Index(item, ":")
		scheme := item[:p]
		p = strings.LastIndex(item, "/")
		itemcontent := item[p+1:]
		switch scheme {
		case "vmess":
			vmes = append(vmes, itemcontent)
		case "trojan":
			troj = append(troj, item)
		}
	}

	var writer io.Writer
	if *outputfile != "" {
		writer, err = os.Create(*outputfile)
		if err != nil {
			log.Fatalln(err)
		}
		defer writer.(io.Closer).Close()
	} else {
		writer = os.Stdout
	}
	parsevmes(vmes, writer)
	parsetroj(troj, writer)
}
