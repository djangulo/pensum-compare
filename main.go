package main

import (
	_ "embed"
	"encoding/csv"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
)

//go:embed "index.html"
var indexStr string

var (
	t       *template.Template
	csvData [][]string
)

func main() {
	server := newServer()
	mux := http.NewServeMux()

	mux.Handle("/", server)

	fmt.Println("Listening on port :51777")
	http.ListenAndServe(":51777", mux)

}

type server struct {
	Records     []*record
	tmpl        *template.Template
	compareTmpl *template.Template
}

func newServer() *server {
	fh, err := os.Open("pensum.csv")
	if err != nil {
		panic(err)
	}

	r := csv.NewReader(fh)

	csvData, err = r.ReadAll()
	if err != nil {
		panic(err)
	}

	funcMap := template.FuncMap{
		"HasID": hasID,
		"Sub": func(i, j int) (k int) {
			k = i - j
			return
		},
		"Percentage": func(num, den int) string {
			p := float64(num) / float64(den) * 100.0
			return fmt.Sprintf("%.1f%%", p)
		},
	}

	t, err = template.New("index.html").Funcs(funcMap).Parse(indexStr)
	if err != nil {
		panic(err)
	}
	return &server{tmpl: t, Records: parseCSV(csvData)}
}

func (s *server) redirectToHome(w http.ResponseWriter, share uint64, name string) {
	u, _ := url.Parse("/")
	q := url.Values{}
	q.Set("share", strconv.FormatUint(share, 10))
	if len(name) > 0 {
		q.Set("name", name)
	}
	u.RawQuery = q.Encode()
	w.Header().Add("location", u.String())
	w.WriteHeader(http.StatusSeeOther)
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		q := r.URL.Query()
		data := struct {
			Records []*record
			IDs     string
			Name    string
			Storage uint64
			Total   int
			Done    int
		}{Records: s.Records, Total: len(s.Records)}

		if str := q.Get("share"); str != "" {
			k, _ := strconv.ParseUint(str, 10, 64)
			data.Storage = k
		}
		if name := q.Get("name"); name != "" {
			data.Name = name
		}
		data.IDs = strings.Join(decodeList(data.Storage), ",")
		var done int
		for _, rec := range s.Records {
			if hasID(data.Storage, rec.ID) {
				done++
			}
		}
		data.Done = done

		s.tmpl.Execute(w, data)
	case http.MethodPost:
		r.ParseForm()
		ids := r.Form.Get("ids")
		name := r.Form.Get("name")
		share := encodeList(strings.Split(ids, ",")...)
		s.redirectToHome(w, share, name)
	}
}

func hasID(k, id uint64) bool {
	return k&id == id
}

func encodeList(ids ...string) uint64 {
	var k uint64
	for _, id := range ids {
		n, _ := strconv.ParseUint(id, 10, 64)
		k += n
	}
	return k
}

func decodeList(k uint64) []string {
	var res = make([]string, 0)
	var i uint64
	var long uint64 = uint64(len(csvData) - 1)
	for i = 0; i < long; i++ {
		if (k & (1 << i)) == (1 << i) {
			res = append(res, strconv.FormatUint(1<<i, 10))
			k -= (1 << i)
		}
	}
	return res
}

type record struct {
	ID              uint64
	Codigo          string
	Titulo          string
	Creditos        int
	PrerrequisitoID string
	Prerrequisito   *record
	Cuatrimestre    int
}

func parseCSV(in [][]string) []*record {
	var set = make(map[string]*record)
	for i, r := range in[1:] {
		creditos, _ := strconv.Atoi(r[2])
		cuatrimestre, _ := strconv.Atoi(r[4])
		var prerequisitoID string
		if r[3] != "NULL" {
			prerequisitoID = r[3]
		}
		record := &record{
			ID:              1 << i,
			Codigo:          r[0],
			Titulo:          r[1],
			Creditos:        creditos,
			PrerrequisitoID: prerequisitoID,
			Cuatrimestre:    cuatrimestre,
		}
		set[record.Codigo] = record
	}
	for _, record := range set {
		if record.PrerrequisitoID != "" {
			record.Prerrequisito = set[record.PrerrequisitoID]
		}
	}
	var records = make([]*record, 0)
	for _, r := range set {
		records = append(records, r)
	}
	sort.SliceStable(records, func(i, j int) bool {
		return records[i].ID < records[j].ID
	})
	return records
}
