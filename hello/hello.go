package hello

import (
  "html/template"
  "net/http"
  "time"
  "appengine"
  "appengine/datastore"
  "appengine/user"
)

type Greeting struct {
  Author string
  Content string
  Date time.Time
}

type GData struct {
  Greetings []Greeting
  Logout string
}

func init() {
  http.HandleFunc("/", root);
  http.HandleFunc("/sign", sign);
}


func guestbookKey(c appengine.Context) *datastore.Key {
  return datastore.NewKey(c, "Guestbook", "default_guestbook", 0, nil)
}

func root(w http.ResponseWriter, r *http.Request) {
  c := appengine.NewContext(r)
  u := user.Current(c)
  if u == nil {
    url, err := user.LoginURL(c, r.URL.String())
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Location", url)
    w.WriteHeader(http.StatusFound)
    return
  }
  q := datastore.NewQuery("Greeting").Ancestor(guestbookKey(c)).Order("-Date").Limit(10)
  greetings := make([]Greeting, 0, 10)
  if _, err := q.GetAll(c, &greetings); err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  logout, _ := user.LogoutURL(c, "/");
  gdata := GData {
    Greetings: greetings,
    Logout: logout,
  }
  if err := guestbookTemplate.Execute(w, gdata); err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
}

var guestbookTemplate = template.Must(template.New("book").Parse(guestbookTemplateHTML))

const guestbookTemplateHTML = `
<html>
  <body>
    {{range .Greetings}}
    {{with .Author}}
      <p><b>{{.}}</b> wrote:</p>
    {{else}}
      <p>An anonymous person wrote:</p>
    {{end}}
    <pre>{{.Content}}</pre>
    {{end}}
    <form action="/sign" method="post">
      <div><textarea name="content" rows="3" cols="60"></textarea></div>
      <div><input type="submit" value="Sign Guestbook"></div>
      <div><a href="{{.Logout}}">logout</a></div>
    </form>
  </body>
</html>
`

func sign(w http.ResponseWriter, r *http.Request) {
  c := appengine.NewContext(r)
  g := Greeting {
    Content: r.FormValue("content"),
    Date: time.Now(),
  }
  if u := user.Current(c); u != nil {
    g.Author = u.String()
  }

  key := datastore.NewIncompleteKey(c, "Greeting", guestbookKey(c))
  _, err := datastore.Put(c, key, &g)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  http.Redirect(w, r, "/", http.StatusFound)

}
