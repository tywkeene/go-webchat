package main

import (
	"encoding/json"
	"flag"
	"github.com/BurntSushi/toml"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type Client struct {
	Name    string
	Address string
}

type Message struct {
	Name      string
	Message   string
	Timestamp string
}

type Server struct {
	Clients []*Client
	Lines   []*Message
}

type Config struct {
	DocsDir         string `toml:"doc_dir"`
	StaticDir       string `toml:"static_dir"`
	Port            string `toml:"port"`
	Persist         bool   `toml:"persist"`
	PersistFile     string `toml:"persist_file_path"`
	PersistInterval string `toml:"persist_interval"`
	GetClients      bool   `toml:"get_clients_endpoint"`
	Ssl             bool   `toml:"ssl"`
	SslCert         string `toml:"ssl_cert_path"`
	SslKey          string `toml:"ssl_key_path"`
}

var config Config

func NewServer() *Server {
	return &Server{make([]*Client, 0), make([]*Message, 0)}
}

func NewMessage(who string, said string) *Message {
	return &Message{who, said, time.Now().String()}
}

func NewClient(name string, address string) *Client {
	return &Client{name, address}
}

func (s *Server) AddClient(c *Client) {
	s.Clients = append(s.Clients, c)
	log.Println("Added client:", c)
}

func (s *Server) FindClient(address string, username string) bool {
	if s.Clients == nil {
		return false
	}
	for _, client := range s.Clients {
		if client.Address == address && client.Name == username {
			return true
		}
	}
	return false
}

func getTemplate(filename string) *template.Template {
	html, err := ioutil.ReadFile(config.DocsDir + filename)
	if err != nil {
		log.Println(err)
		return nil
	}

	tmp, err := template.New(filename).Parse(string(html))
	if err != nil {
		log.Println(err)
		return nil
	}
	return tmp
}

func logHttp(r *http.Request) {
	log.Printf("%s %s %s %s", r.Method, r.URL, r.RemoteAddr, r.UserAgent())
}

func (s *Server) validateUser(resp http.ResponseWriter, req *http.Request) bool {
	cookie, err := req.Cookie("username")
	if cookie == nil {
		return false
	}
	if err != nil {
		log.Println("Error validating user:", err)
		return false
	}
	return s.FindClient(req.RemoteAddr, cookie.Value)
}

func (s *Server) Index(resp http.ResponseWriter, req *http.Request) {
	logHttp(req)
	if s.validateUser(resp, req) == false {
		tmp := getTemplate("index.html")
		tmp.Execute(resp, nil)
	} else {
		http.Redirect(resp, req, "/chat", 301)
	}
}

func (s *Server) Register(resp http.ResponseWriter, req *http.Request) {
	logHttp(req)
	if s.validateUser(resp, req) == true {
		http.Redirect(resp, req, "/chat", 301)
	}
	req.ParseForm()
	username := req.Form.Get("username")
	if username == "" {
		username = "Anonymous"
	}
	expiration := time.Now().Add(365 * 24 * time.Hour)
	cookie := http.Cookie{Name: "username", Value: username, Expires: expiration}
	http.SetCookie(resp, &cookie)

	client := NewClient(username, req.RemoteAddr)
	s.AddClient(client)
	http.Redirect(resp, req, "/chat", 301)
}

func (s *Server) Chat(resp http.ResponseWriter, req *http.Request) {
	logHttp(req)
	if s.validateUser(resp, req) == false {
		http.Redirect(resp, req, "/", 301)
	}
	tmp := getTemplate("chat.html")
	tmp.Execute(resp, nil)
}

func (s *Server) GetMessages(resp http.ResponseWriter, req *http.Request) {
	serial, err := json.MarshalIndent(&s.Lines, " ", " ")
	if err != nil {
		log.Println(err)
		return
	}
	io.WriteString(resp, string(serial))
}

func (s *Server) GetClients(resp http.ResponseWriter, req *http.Request) {
	logHttp(req)
	if config.GetClients == false {
		io.WriteString(resp, "Endpoint disabled")
		return
	}
	serial, err := json.MarshalIndent(&s.Clients, " ", " ")
	if err != nil {
		log.Println(err)
		return
	}
	io.WriteString(resp, string(serial))
}

func (s *Server) PostMessage(resp http.ResponseWriter, req *http.Request) {
	logHttp(req)
	req.ParseForm()
	cookie, err := req.Cookie("username")
	if cookie == nil || err != nil {
		return
	}
	user := cookie.Value
	message := req.Form.Get("message")
	s.Lines = append(s.Lines, NewMessage(user, message))
}

func (s *Server) Static(resp http.ResponseWriter, req *http.Request) {
	logHttp(req)
	http.ServeFile(resp, req, req.URL.Path[1:])
}

func (s *Server) RestoreMessages() error {
	if _, err := os.Stat(config.PersistFile); err != nil {
		return err
	}

	serial, err := ioutil.ReadFile(config.PersistFile)
	if err != nil {
		return err
	}
	json.Unmarshal(serial, &s.Lines)
	return nil
}

var persistFileSize int

func (s *Server) WriteToDisk(filename string) error {
	if len(s.Lines) == persistFileSize {
		return nil
	}
	log.Printf("Writing %d records to %s\n", len(s.Lines), filename)
	outfile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer outfile.Close()
	output, _ := json.MarshalIndent(s.Lines[:persistFileSize], " ", " ")
	_, err = outfile.WriteString(string(output))
	if err != nil {
		return err
	}
	persistFileSize = len(s.Lines)
	return nil
}

func (s *Server) PersistenceThread(filename string) {
	interval, err := time.ParseDuration(config.PersistInterval)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Persisting to %s every %s\n", config.PersistFile, config.PersistInterval)
	for {
		time.Sleep(interval)
		s.WriteToDisk(filename)
	}
}
func GetOptions() {

	var configPath string
	flag.StringVar(&configPath, "config", "", "Config file path (using the config will override command line arguments")

	flag.StringVar(&config.Port, "port", "80", "Bind port")
	flag.StringVar(&config.DocsDir, "docs-dir", "./docs/", "Documents directory")
	flag.StringVar(&config.StaticDir, "static-dir", "./static/", "Static directory")
	flag.BoolVar(&config.GetClients, "get-clients", false, "Enable or disable the /get_clients endpoint")
	flag.StringVar(&config.PersistInterval, "persist-interval", "5m", "How often to write data to disk")
	flag.StringVar(&config.PersistFile, "persist-file", "data.json", "Persistence data file")
	flag.BoolVar(&config.Persist, "persist", false, "Enable or disable persisting the store to disk")
	flag.BoolVar(&config.Ssl, "ssl", false, "Enable or disable serving over SSL")
	flag.StringVar(&config.SslCert, "cert", "", "Path to ssl certificate file")
	flag.StringVar(&config.SslKey, "key", "", "Path to ssl certificate key file")
	flag.Parse()

	flag.Parse()

	if configPath != "" {
		if _, err := toml.DecodeFile(configPath, &config); err != nil {
			log.Printf("Error reading config %s: %s\n", configPath, err.Error())
			os.Exit(-1)
		}
		log.Printf("Configration file options in %s overriding command line options\n", configPath)
	}
}
func main() {
	GetOptions()
	s := NewServer()
	if config.Persist == true {
		if err := s.RestoreMessages(); err != nil {
			log.Println(err)
		}
		go s.PersistenceThread(config.PersistFile)
	}

	http.HandleFunc("/", s.Index)
	http.HandleFunc("/register", s.Register)
	http.HandleFunc("/chat", s.Chat)
	http.HandleFunc("/get_messages", s.GetMessages)
	http.HandleFunc("/get_clients", s.GetClients)
	http.HandleFunc("/post_message", s.PostMessage)
	http.HandleFunc("/static/", s.Static)

	if config.Ssl == true {
		log.Printf("Using certificate (%s) and key (%s) for SSL\n", config.SslCert, config.SslKey)
		log.Print("Listening on port ", config.Port)
		log.Fatal(http.ListenAndServeTLS(":443", config.SslCert, config.SslKey, nil))
	} else {
		log.Print("Listening on port ", config.Port)
		log.Fatal(http.ListenAndServe(":"+config.Port, nil))
	}
}
