package explorer

import (
	"fmt"
	"log"
	"net/http"
	"text/template"

	"github.com/fantasticake/fantasticoin/blockchain"
)

type homeData struct {
	Blocks []*blockchain.Block
}

var templates *template.Template

func home(w http.ResponseWriter, r *http.Request) {
	data := homeData{blockchain.Blocks(blockchain.BC())}
	err := templates.ExecuteTemplate(w, "home", data)
	if err != nil {
		fmt.Println(err)
	}
}

func add(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		err := templates.ExecuteTemplate(w, "add", nil)
		if err != nil {
			fmt.Println(err)
		}
	case http.MethodPost:
		r.ParseForm()
		data := r.Form.Get("data")
		blockchain.BC().AddBlock(data)
		http.Redirect(w, r, "/", http.StatusPermanentRedirect)
	}
}

func Start(port int) {
	templates = template.Must(template.ParseGlob("explorer/templates/pages/*.html"))
	templates = template.Must(templates.ParseGlob("explorer/templates/partials/*.html"))

	router := http.NewServeMux()
	router.HandleFunc("/", home)
	router.HandleFunc("/add", add)

	fmt.Printf("Server listening on http://localhost:%d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), router))
}
